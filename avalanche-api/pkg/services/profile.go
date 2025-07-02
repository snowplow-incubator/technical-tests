package services

import (
	"context"
	"database/sql"
	"sync"

	"github.com/google/uuid"
	"github.com/snowplow-incubator/avalanche-api/pkg/config"
	"github.com/snowplow-incubator/avalanche-api/pkg/domain"
	"github.com/snowplow-incubator/avalanche-api/pkg/errors"
	"github.com/snowplow-incubator/avalanche-api/pkg/repositories"
)

const (
	OpFindAttributes   = "find_attributes"
	OpCreateAttributes = "create_attributes"
	OpFindProfiles     = "find_profiles"
	OpCreateProfiles   = "create_profiles"
	OpLinkAttributes   = "link_attributes"
)

type ProfileServiceInterface interface {
	ProcessEvent(ctx context.Context, event map[string]string) (*domain.Profile, error)
	ProcessEvents(ctx context.Context, events []map[string]string) ([]EventResult, error)
}

type EventResult struct {
	Profile *domain.Profile `json:"profile"`
	Error   string          `json:"error,omitempty"`
}

type ProfileService struct {
	attributeRepo repositories.AttributesRepositoryInterface
	profileRepo   repositories.ProfilesRepositoryInterface
	extractKeys   []string
	cfg           ServiceConfig
}

type ServiceConfig struct {
	BatchSize  int
	MaxWorkers int
}

func NewProfileService(
	attributeRepo repositories.AttributesRepositoryInterface,
	profileRepo repositories.ProfilesRepositoryInterface,
	extractKeys []string,
	cfg ServiceConfig,
) *ProfileService {
	if cfg.BatchSize <= 0 {
		cfg.BatchSize = config.DefaultBatchSize
	}
	if cfg.MaxWorkers <= 0 {
		cfg.MaxWorkers = config.DefaultMaxWorkers
	}

	return &ProfileService{
		attributeRepo: attributeRepo,
		profileRepo:   profileRepo,
		extractKeys:   extractKeys,
		cfg:           cfg,
	}
}

func (s *ProfileService) ProcessEvent(ctx context.Context, event map[string]string) (*domain.Profile, error) {
	nameValuePairs := domain.ExtractAttributes(event, s.extractKeys)
	if len(nameValuePairs) == 0 {
		return nil, nil // No attributes to process
	}

	existingAttributes, err := s.attributeRepo.FindByNameValue(ctx, nameValuePairs)
	if err != nil {
		return nil, errors.NewDatabaseError(OpFindAttributes, err)
	}

	targetProfile, err := s.findOrCreateProfile(ctx, existingAttributes)
	if err != nil {
		return nil, err
	}

	if err := s.createAndLinkAttributes(ctx, targetProfile.ID, nameValuePairs); err != nil {
		return nil, err
	}

	return targetProfile, nil
}

func (s *ProfileService) findOrCreateProfile(ctx context.Context, existingAttributes []domain.Attribute) (*domain.Profile, error) {
	if len(existingAttributes) > 0 {
		attributeIDs := make([]int, len(existingAttributes))
		for i, attr := range existingAttributes {
			attributeIDs[i] = attr.ID
		}

		profile, err := s.profileRepo.FindOldestByAttributeIDs(ctx, attributeIDs)
		if err != nil && err != sql.ErrNoRows {
			return nil, errors.NewDatabaseError(OpFindProfiles, err)
		}
		if profile != nil {
			return profile, nil
		}
	}

	profile, err := s.profileRepo.Create(ctx)
	if err != nil {
		return nil, errors.NewDatabaseError(OpCreateProfiles, err)
	}

	return profile, nil
}

func (s *ProfileService) createAndLinkAttributes(ctx context.Context, profileID uuid.UUID, nameValuePairs map[string]string) error {
	attributesToCreate := make([]domain.Attribute, 0, len(nameValuePairs))
	for name, value := range nameValuePairs {
		attributesToCreate = append(attributesToCreate, domain.Attribute{
			Name:  name,
			Value: value,
		})
	}

	attributeIDs, err := s.attributeRepo.CreateAndGetIDs(ctx, attributesToCreate)
	if err != nil {
		return errors.NewDatabaseError(OpCreateAttributes, err)
	}

	if err := s.profileRepo.LinkAttributeBatch(ctx, profileID, attributeIDs); err != nil {
		return errors.NewDatabaseError(OpLinkAttributes, err)
	}

	return nil
}

func (s *ProfileService) ProcessEvents(ctx context.Context, events []map[string]string) ([]EventResult, error) {
	if len(events) == 0 {
		return []EventResult{}, nil
	}

	batches := s.createBatches(events)
	results := make([]EventResult, len(events))

	semaphore := make(chan struct{}, s.cfg.MaxWorkers)
	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, batch := range batches {
		wg.Add(1)
		go func(b eventBatch) {
			defer wg.Done()

			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			batchResults, err := s.processBatch(ctx, b.events)

			mu.Lock()
			defer mu.Unlock()

			if err != nil {
				for j := 0; j < len(b.events); j++ {
					results[b.startIndex+j] = EventResult{Error: err.Error()}
				}
			} else {
				for j, result := range batchResults {
					results[b.startIndex+j] = result
				}
			}
		}(batch)
	}

	wg.Wait()
	return results, nil

}

func (s *ProfileService) processBatch(ctx context.Context, events []map[string]string) ([]EventResult, error) {
	if len(events) == 0 {
		return []EventResult{}, nil
	}

	eventDataMap, eventOrder, err := s.extractEventData(events)
	if err != nil {
		return nil, err
	}

	existingAttrGroups, err := s.attributeRepo.FindByNameValueBatch(ctx, eventDataMap)
	if err != nil {
		return nil, err
	}

	eventAttributeIDs := make(map[string][]int)
	for eventId, attrs := range existingAttrGroups {
		if len(attrs) > 0 {
			ids := make([]int, len(attrs))
			for j, attr := range attrs {
				ids[j] = attr.ID
			}
			eventAttributeIDs[eventId] = ids
		}
	}

	profileGroups, err := s.profileRepo.FindOldestByAttributeIDsBatch(ctx, eventAttributeIDs)
	if err != nil {
		return nil, err
	}

	var newProfilesNeeded int
	profileAssignments := make(map[string]*domain.Profile)

	for eventId := range eventDataMap {
		if profile, exists := profileGroups[eventId]; exists && profile != nil {
			profileAssignments[eventId] = profile
		} else {
			newProfilesNeeded++
		}
	}

	var newProfiles []*domain.Profile
	if newProfilesNeeded > 0 {
		newProfiles, err = s.profileRepo.CreateBatch(ctx, newProfilesNeeded)
		if err != nil {
			return nil, err
		}
	}

	newProfileIndex := 0
	for eventId := range eventDataMap {
		if profileAssignments[eventId] == nil {
			profileAssignments[eventId] = newProfiles[newProfileIndex]
			newProfileIndex++
		}
	}

	attributeIDGroups, err := s.attributeRepo.CreateBatchOptimized(ctx, eventDataMap)
	if err != nil {
		return nil, err
	}

	profileAttributeMappings := make(map[uuid.UUID][]int)
	for eventId, profile := range profileAssignments {
		if profile != nil {
			if attrIDs, exists := attributeIDGroups[eventId]; exists {
				if existing, hasExisting := profileAttributeMappings[profile.ID]; hasExisting {
					profileAttributeMappings[profile.ID] = append(existing, attrIDs...)
				} else {
					profileAttributeMappings[profile.ID] = attrIDs
				}
			}
		}
	}

	err = s.profileRepo.LinkAttributesBatch(ctx, profileAttributeMappings)
	if err != nil {
		return nil, err
	}

	results := make([]EventResult, len(events))
	for i, eventId := range eventOrder {
		if profile := profileAssignments[eventId]; profile != nil {
			results[i] = EventResult{Profile: profile}
		} else {
			results[i] = EventResult{Error: "Failed to assign profile"}
		}
	}

	return results, nil
}

func (s *ProfileService) extractEventData(events []map[string]string) (map[string]map[string]string, []string, error) {
	eventDataMap := make(map[string]map[string]string)
	eventOrder := make([]string, len(events))

	for i, event := range events {
		eventID, err := domain.ExtractEventID(event, i)
		if err != nil {
			return nil, nil, err
		}

		nameValuePairs := domain.ExtractAttributes(event, s.extractKeys)
		eventDataMap[eventID] = nameValuePairs
		eventOrder[i] = eventID
	}

	return eventDataMap, eventOrder, nil
}

func (s *ProfileService) createBatches(events []map[string]string) []eventBatch {
	var batches []eventBatch
	batchSize := s.cfg.BatchSize

	for i := 0; i < len(events); i += batchSize {
		end := i + batchSize
		if end > len(events) {
			end = len(events)
		}

		batches = append(batches, eventBatch{
			events:     events[i:end],
			startIndex: i,
		})
	}

	return batches
}

type eventBatch struct {
	events     []map[string]string
	startIndex int
}
