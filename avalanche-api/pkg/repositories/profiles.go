package repositories

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/snowplow-incubator/avalanche-api/pkg/domain"
)

type ProfilesRepositoryInterface interface {
	Create(ctx context.Context) (*domain.Profile, error)
	GetByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error)
	LinkAttribute(ctx context.Context, profileID uuid.UUID, attributeID int) error
	LinkProfiles(ctx context.Context, fromProfileID, toProfileID uuid.UUID) error
	FindOldestByAttributeIDs(ctx context.Context, attributeIDs []int) (*domain.Profile, error)
	LinkAttributeBatch(ctx context.Context, profileID uuid.UUID, attributeIDs []int) error
	FindOldestByAttributeIDsBatch(ctx context.Context, eventAttributeIDs map[string][]int) (map[string]*domain.Profile, error)
	CreateBatch(ctx context.Context, count int) ([]*domain.Profile, error)
	LinkAttributesBatch(ctx context.Context, profileAttributeMappings map[uuid.UUID][]int) error
}

type ProfilesRepository struct {
	db *pgxpool.Pool
}

func NewProfilesRepository(db *pgxpool.Pool) *ProfilesRepository {
	return &ProfilesRepository{
		db: db,
	}
}

func (r *ProfilesRepository) Create(ctx context.Context) (*domain.Profile, error) {
	profile := &domain.Profile{}
	query := `INSERT INTO profiles DEFAULT VALUES RETURNING id, created_at`
	err := r.db.QueryRow(ctx, query).Scan(&profile.ID, &profile.CreatedAt)
	return profile, err
}

func (r *ProfilesRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Profile, error) {
	profile := &domain.Profile{}
	query := `SELECT id, created_at FROM profiles WHERE id = $1`
	err := r.db.QueryRow(ctx, query, id).Scan(&profile.ID, &profile.CreatedAt)
	return profile, err
}

func (r *ProfilesRepository) LinkAttribute(ctx context.Context, profileID uuid.UUID, attributeID int) error {
	query := `INSERT INTO profile_attributes (profile_id, attribute_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.Exec(ctx, query, profileID, attributeID)
	return err
}

func (r *ProfilesRepository) LinkProfiles(ctx context.Context, fromProfileID, toProfileID uuid.UUID) error {
	query := `INSERT INTO profile_relationships (from_profile_id, to_profile_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	_, err := r.db.Exec(ctx, query, fromProfileID, toProfileID)
	return err
}

func (r *ProfilesRepository) FindOldestByAttributeIDs(ctx context.Context, attributeIDs []int) (*domain.Profile, error) {
	if len(attributeIDs) == 0 {
		return nil, nil
	}

	query := `
		SELECT DISTINCT p.id, p.created_at 
		FROM profiles p 
		JOIN profile_attributes pa ON p.id = pa.profile_id 
		WHERE pa.attribute_id = ANY($1) 
		ORDER BY p.created_at ASC 
		LIMIT 1`

	profile := &domain.Profile{}
	err := r.db.QueryRow(ctx, query, attributeIDs).Scan(&profile.ID, &profile.CreatedAt)
	if err != nil {
		return nil, err
	}

	return profile, nil
}

func (r *ProfilesRepository) LinkAttributeBatch(ctx context.Context, profileID uuid.UUID, attributeIDs []int) error {
	if len(attributeIDs) == 0 {
		return nil
	}

	query := `INSERT INTO profile_attributes (profile_id, attribute_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	batch := &pgx.Batch{}
	for _, attributeID := range attributeIDs {
		batch.Queue(query, profileID, attributeID)
	}

	results := r.db.SendBatch(ctx, batch)
	//nolint:errcheck
	defer results.Close()

	for i := 0; i < len(attributeIDs); i++ {
		_, err := results.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *ProfilesRepository) FindOldestByAttributeIDsBatch(ctx context.Context, eventAttributeIDs map[string][]int) (map[string]*domain.Profile, error) {
	result := make(map[string]*domain.Profile)

	for eventId, attributeIDs := range eventAttributeIDs {
		if len(attributeIDs) == 0 {
			result[eventId] = nil
			continue
		}

		profile, err := r.FindOldestByAttributeIDs(ctx, attributeIDs)
		if err != nil && err != sql.ErrNoRows {
			return nil, err
		}
		result[eventId] = profile
	}

	return result, nil
}

func (r *ProfilesRepository) CreateBatch(ctx context.Context, count int) ([]*domain.Profile, error) {
	if count == 0 {
		return []*domain.Profile{}, nil
	}

	profiles := make([]*domain.Profile, 0, count)
	query := `INSERT INTO profiles DEFAULT VALUES RETURNING id, created_at`

	for i := 0; i < count; i++ {
		profile := &domain.Profile{}
		err := r.db.QueryRow(ctx, query).Scan(&profile.ID, &profile.CreatedAt)
		if err != nil {
			return nil, err
		}
		profiles = append(profiles, profile)
	}

	return profiles, nil
}

func (r *ProfilesRepository) LinkAttributesBatch(ctx context.Context, profileAttributeMappings map[uuid.UUID][]int) error {
	if len(profileAttributeMappings) == 0 {
		return nil
	}

	query := `INSERT INTO profile_attributes (profile_id, attribute_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`
	batch := &pgx.Batch{}

	for profileID, attributeIDs := range profileAttributeMappings {
		for _, attributeID := range attributeIDs {
			batch.Queue(query, profileID, attributeID)
		}
	}

	results := r.db.SendBatch(ctx, batch)
	//nolint:errcheck
	defer results.Close()

	for i := 0; i < batch.Len(); i++ {
		_, err := results.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}
