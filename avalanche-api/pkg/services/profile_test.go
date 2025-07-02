package services

import (
	"context"
	"database/sql"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/snowplow-incubator/avalanche-api/pkg/domain"
)

type mockAttributesRepository struct {
	createFunc               func(ctx context.Context, name, value string) error
	createBatchFunc          func(ctx context.Context, attributes []domain.Attribute) error
	findByNameValueFunc      func(ctx context.Context, nameValuePairs map[string]string) ([]domain.Attribute, error)
	findByNameValueBatchFunc func(ctx context.Context, eventDataMap map[string]map[string]string) (map[string][]domain.Attribute, error)
	createAndGetIDsFunc      func(ctx context.Context, attributes []domain.Attribute) ([]int, error)
	createBatchOptimizedFunc func(ctx context.Context, eventDataMap map[string]map[string]string) (map[string][]int, error)
}

func (m *mockAttributesRepository) Create(ctx context.Context, name, value string) error {
	if m.createFunc != nil {
		return m.createFunc(ctx, name, value)
	}
	return nil
}

func (m *mockAttributesRepository) CreateBatch(ctx context.Context, attributes []domain.Attribute) error {
	if m.createBatchFunc != nil {
		return m.createBatchFunc(ctx, attributes)
	}
	return nil
}

func (m *mockAttributesRepository) FindByNameValue(ctx context.Context, nameValuePairs map[string]string) ([]domain.Attribute, error) {
	if m.findByNameValueFunc != nil {
		return m.findByNameValueFunc(ctx, nameValuePairs)
	}
	return nil, nil
}

func (m *mockAttributesRepository) FindByNameValueBatch(ctx context.Context, eventDataMap map[string]map[string]string) (map[string][]domain.Attribute, error) {
	if m.findByNameValueBatchFunc != nil {
		return m.findByNameValueBatchFunc(ctx, eventDataMap)
	}
	return make(map[string][]domain.Attribute), nil
}

func (m *mockAttributesRepository) CreateAndGetIDs(ctx context.Context, attributes []domain.Attribute) ([]int, error) {
	if m.createAndGetIDsFunc != nil {
		return m.createAndGetIDsFunc(ctx, attributes)
	}
	return []int{1, 2, 3}, nil
}

func (m *mockAttributesRepository) CreateBatchOptimized(ctx context.Context, eventDataMap map[string]map[string]string) (map[string][]int, error) {
	if m.createBatchOptimizedFunc != nil {
		return m.createBatchOptimizedFunc(ctx, eventDataMap)
	}
	result := make(map[string][]int)
	for eventID := range eventDataMap {
		result[eventID] = []int{1, 2}
	}
	return result, nil
}

type mockProfilesRepository struct {
	createFunc                        func(ctx context.Context) (*domain.Profile, error)
	getByIDFunc                       func(ctx context.Context, id uuid.UUID) (*domain.Profile, error)
	linkAttributeFunc                 func(ctx context.Context, profileID uuid.UUID, attributeID int) error
	linkProfilesFunc                  func(ctx context.Context, fromProfileID, toProfileID uuid.UUID) error
	findOldestByAttributeIDsFunc      func(ctx context.Context, attributeIDs []int) (*domain.Profile, error)
	findOldestByAttributeIDsBatchFunc func(ctx context.Context, eventAttributeIDs map[string][]int) (map[string]*domain.Profile, error)
	createBatchFunc                   func(ctx context.Context, count int) ([]*domain.Profile, error)
	linkAttributeBatchFunc            func(ctx context.Context, profileID uuid.UUID, attributeIDs []int) error
	linkAttributesBatchFunc           func(ctx context.Context, profileAttributeMappings map[uuid.UUID][]int) error
}

func (m *mockProfilesRepository) Create(ctx context.Context) (*domain.Profile, error) {
	if m.createFunc != nil {
		return m.createFunc(ctx)
	}
	return &domain.Profile{
		ID:        uuid.New(),
		CreatedAt: time.Now(),
	}, nil
}

func (m *mockProfilesRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error) {
	if m.getByIDFunc != nil {
		return m.getByIDFunc(ctx, id)
	}
	return &domain.Profile{
		ID:        id,
		CreatedAt: time.Now(),
	}, nil
}

func (m *mockProfilesRepository) LinkAttribute(ctx context.Context, profileID uuid.UUID, attributeID int) error {
	if m.linkAttributeFunc != nil {
		return m.linkAttributeFunc(ctx, profileID, attributeID)
	}
	return nil
}

func (m *mockProfilesRepository) LinkProfiles(ctx context.Context, fromProfileID, toProfileID uuid.UUID) error {
	if m.linkProfilesFunc != nil {
		return m.linkProfilesFunc(ctx, fromProfileID, toProfileID)
	}
	return nil
}

func (m *mockProfilesRepository) FindOldestByAttributeIDs(ctx context.Context, attributeIDs []int) (*domain.Profile, error) {
	if m.findOldestByAttributeIDsFunc != nil {
		return m.findOldestByAttributeIDsFunc(ctx, attributeIDs)
	}
	return nil, sql.ErrNoRows
}

func (m *mockProfilesRepository) FindOldestByAttributeIDsBatch(ctx context.Context, eventAttributeIDs map[string][]int) (map[string]*domain.Profile, error) {
	if m.findOldestByAttributeIDsBatchFunc != nil {
		return m.findOldestByAttributeIDsBatchFunc(ctx, eventAttributeIDs)
	}
	return make(map[string]*domain.Profile), nil
}

func (m *mockProfilesRepository) CreateBatch(ctx context.Context, count int) ([]*domain.Profile, error) {
	if m.createBatchFunc != nil {
		return m.createBatchFunc(ctx, count)
	}
	profiles := make([]*domain.Profile, count)
	for i := 0; i < count; i++ {
		profiles[i] = &domain.Profile{
			ID:        uuid.New(),
			CreatedAt: time.Now(),
		}
	}
	return profiles, nil
}

func (m *mockProfilesRepository) LinkAttributeBatch(ctx context.Context, profileID uuid.UUID, attributeIDs []int) error {
	if m.linkAttributeBatchFunc != nil {
		return m.linkAttributeBatchFunc(ctx, profileID, attributeIDs)
	}
	return nil
}

func (m *mockProfilesRepository) LinkAttributesBatch(ctx context.Context, profileAttributeMappings map[uuid.UUID][]int) error {
	if m.linkAttributesBatchFunc != nil {
		return m.linkAttributesBatchFunc(ctx, profileAttributeMappings)
	}
	return nil
}

func TestProfileService_ProcessEvent(t *testing.T) {
	tests := []struct {
		name        string
		event       map[string]string
		extractKeys []string
		setupMocks  func(*mockAttributesRepository, *mockProfilesRepository)
		expectError bool
	}{
		{
			name: "process event with no attributes",
			event: map[string]string{
				"eventId": "test-1",
				"other":   "data",
			},
			extractKeys: []string{"userId"},
			setupMocks: func(attrRepo *mockAttributesRepository, profileRepo *mockProfilesRepository) {
				// No attributes to extract, should return nil
			},
			expectError: false,
		},
		{
			name: "process event creates new profile",
			event: map[string]string{
				"eventId": "test-1",
				"userId":  "123",
			},
			extractKeys: []string{"userId"},
			setupMocks: func(attrRepo *mockAttributesRepository, profileRepo *mockProfilesRepository) {
				attrRepo.findByNameValueFunc = func(ctx context.Context, nameValuePairs map[string]string) ([]domain.Attribute, error) {
					return []domain.Attribute{}, nil // No existing attributes
				}
				attrRepo.createAndGetIDsFunc = func(ctx context.Context, attributes []domain.Attribute) ([]int, error) {
					return []int{1}, nil
				}
				profileRepo.createFunc = func(ctx context.Context) (*domain.Profile, error) {
					return &domain.Profile{
						ID:        uuid.New(),
						CreatedAt: time.Now(),
					}, nil
				}
			},
			expectError: false,
		},
		{
			name: "process event finds existing profile",
			event: map[string]string{
				"eventId": "test-1",
				"userId":  "123",
			},
			extractKeys: []string{"userId"},
			setupMocks: func(attrRepo *mockAttributesRepository, profileRepo *mockProfilesRepository) {
				existingProfile := &domain.Profile{
					ID:        uuid.New(),
					CreatedAt: time.Now(),
				}
				attrRepo.findByNameValueFunc = func(ctx context.Context, nameValuePairs map[string]string) ([]domain.Attribute, error) {
					return []domain.Attribute{{ID: 1, Name: "userId", Value: "123"}}, nil
				}
				profileRepo.findOldestByAttributeIDsFunc = func(ctx context.Context, attributeIDs []int) (*domain.Profile, error) {
					return existingProfile, nil
				}
				attrRepo.createAndGetIDsFunc = func(ctx context.Context, attributes []domain.Attribute) ([]int, error) {
					return []int{1}, nil
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrRepo := &mockAttributesRepository{}
			profileRepo := &mockProfilesRepository{}
			tt.setupMocks(attrRepo, profileRepo)

			service := NewProfileService(attrRepo, profileRepo, tt.extractKeys, ServiceConfig{
				BatchSize:  10,
				MaxWorkers: 2,
			})

			result, err := service.ProcessEvent(context.Background(), tt.event)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}

				if len(domain.ExtractAttributes(tt.event, tt.extractKeys)) == 0 {
					if result != nil {
						t.Error("expected nil result for event with no attributes")
					}
				} else if result == nil {
					t.Error("expected profile result but got nil")
				}
			}
		})
	}
}

func TestProfileService_ProcessEvents(t *testing.T) {
	tests := []struct {
		name        string
		events      []map[string]string
		extractKeys []string
		setupMocks  func(*mockAttributesRepository, *mockProfilesRepository)
		expectError bool
	}{
		{
			name:        "empty events",
			events:      []map[string]string{},
			extractKeys: []string{"userId"},
			setupMocks:  func(attrRepo *mockAttributesRepository, profileRepo *mockProfilesRepository) {},
			expectError: false,
		},
		{
			name: "single event",
			events: []map[string]string{
				{"eventId": "test-1", "userId": "123"},
			},
			extractKeys: []string{"userId"},
			setupMocks: func(attrRepo *mockAttributesRepository, profileRepo *mockProfilesRepository) {
				attrRepo.findByNameValueBatchFunc = func(ctx context.Context, eventDataMap map[string]map[string]string) (map[string][]domain.Attribute, error) {
					return make(map[string][]domain.Attribute), nil
				}
				profileRepo.findOldestByAttributeIDsBatchFunc = func(ctx context.Context, eventAttributeIDs map[string][]int) (map[string]*domain.Profile, error) {
					return make(map[string]*domain.Profile), nil
				}
				profileRepo.createBatchFunc = func(ctx context.Context, count int) ([]*domain.Profile, error) {
					profiles := make([]*domain.Profile, count)
					for i := 0; i < count; i++ {
						profiles[i] = &domain.Profile{
							ID:        uuid.New(),
							CreatedAt: time.Now(),
						}
					}
					return profiles, nil
				}
				attrRepo.createBatchOptimizedFunc = func(ctx context.Context, eventDataMap map[string]map[string]string) (map[string][]int, error) {
					result := make(map[string][]int)
					for eventID := range eventDataMap {
						result[eventID] = []int{1}
					}
					return result, nil
				}
			},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			attrRepo := &mockAttributesRepository{}
			profileRepo := &mockProfilesRepository{}
			tt.setupMocks(attrRepo, profileRepo)

			service := NewProfileService(attrRepo, profileRepo, tt.extractKeys, ServiceConfig{
				BatchSize:  10,
				MaxWorkers: 2,
			})

			results, err := service.ProcessEvents(context.Background(), tt.events)

			if tt.expectError {
				if err == nil {
					t.Error("expected error but got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(results) != len(tt.events) {
					t.Errorf("expected %d results, got %d", len(tt.events), len(results))
				}
			}
		})
	}
}

func TestProfileService_extractEventData(t *testing.T) {
	service := NewProfileService(nil, nil, []string{"userId", "sessionId"}, ServiceConfig{})

	events := []map[string]string{
		{"eventId": "test-1", "userId": "123", "sessionId": "abc"},
		{"eventId": "test-2", "userId": "456"},
	}

	eventDataMap, eventOrder, err := service.extractEventData(events)

	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	if len(eventDataMap) != 2 {
		t.Errorf("expected 2 events in map, got %d", len(eventDataMap))
	}

	if len(eventOrder) != 2 {
		t.Errorf("expected 2 events in order, got %d", len(eventOrder))
	}

	if data, exists := eventDataMap["test-1"]; !exists {
		t.Error("expected test-1 in event data map")
	} else {
		if data["userId"] != "123" || data["sessionId"] != "abc" {
			t.Errorf("unexpected data for test-1: %v", data)
		}
	}

	if data, exists := eventDataMap["test-2"]; !exists {
		t.Error("expected test-2 in event data map")
	} else {
		if data["userId"] != "456" {
			t.Errorf("unexpected data for test-2: %v", data)
		}
		if _, hasSession := data["sessionId"]; hasSession {
			t.Error("did not expect sessionId in test-2 data")
		}
	}
}

func TestProfileService_createBatches(t *testing.T) {
	service := NewProfileService(nil, nil, []string{}, ServiceConfig{
		BatchSize: 3,
	})

	events := []map[string]string{
		{"eventId": "1"}, {"eventId": "2"}, {"eventId": "3"},
		{"eventId": "4"}, {"eventId": "5"},
	}

	batches := service.createBatches(events)

	if len(batches) != 2 {
		t.Errorf("expected 2 batches, got %d", len(batches))
	}

	if len(batches[0].events) != 3 {
		t.Errorf("expected first batch to have 3 events, got %d", len(batches[0].events))
	}
	if batches[0].startIndex != 0 {
		t.Errorf("expected first batch start index 0, got %d", batches[0].startIndex)
	}

	if len(batches[1].events) != 2 {
		t.Errorf("expected second batch to have 2 events, got %d", len(batches[1].events))
	}
	if batches[1].startIndex != 3 {
		t.Errorf("expected second batch start index 3, got %d", batches[1].startIndex)
	}
}
