package repositories

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/snowplow-incubator/avalanche-api/pkg/domain"
)

type AttributesRepositoryInterface interface {
	Create(ctx context.Context, name, value string) error
	CreateBatch(ctx context.Context, attributes []domain.Attribute) error
	FindByNameValue(ctx context.Context, nameValuePairs map[string]string) ([]domain.Attribute, error)
	CreateAndGetIDs(ctx context.Context, attributes []domain.Attribute) ([]int, error)
	FindByNameValueBatch(ctx context.Context, eventDataMap map[string]map[string]string) (map[string][]domain.Attribute, error)
	CreateBatchOptimized(ctx context.Context, eventDataMap map[string]map[string]string) (map[string][]int, error)
}

type AttributesRepository struct {
	db *pgxpool.Pool
}

func NewAttributesRepository(db *pgxpool.Pool) *AttributesRepository {
	return &AttributesRepository{
		db: db,
	}
}

func (r *AttributesRepository) Create(ctx context.Context, name, value string) error {
	query := `INSERT INTO attributes (name, value) VALUES ($1, $2) ON CONFLICT (name, value) DO NOTHING`
	_, err := r.db.Exec(ctx, query, name, value)
	return err
}

func (r *AttributesRepository) CreateBatch(ctx context.Context, attributes []domain.Attribute) error {
	if len(attributes) == 0 {
		return nil
	}

	query := `INSERT INTO attributes (name, value) VALUES ($1, $2) ON CONFLICT (name, value) DO NOTHING`
	batch := &pgx.Batch{}

	for _, attr := range attributes {
		batch.Queue(query, attr.Name, attr.Value)
	}

	results := r.db.SendBatch(ctx, batch)
	//nolint:errcheck
	defer results.Close()

	for i := 0; i < len(attributes); i++ {
		_, err := results.Exec()
		if err != nil {
			return err
		}
	}

	return nil
}

func (r *AttributesRepository) FindByNameValue(ctx context.Context, nameValuePairs map[string]string) ([]domain.Attribute, error) {
	if len(nameValuePairs) == 0 {
		return []domain.Attribute{}, nil
	}

	query := `SELECT id, name, value, created_at FROM attributes WHERE (name, value) IN (`
	args := make([]any, 0, len(nameValuePairs)*2)
	placeholders := make([]string, 0, len(nameValuePairs))

	i := 1
	for name, value := range nameValuePairs {
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", i, i+1))
		args = append(args, name, value)
		i += 2
	}

	query += strings.Join(placeholders, ", ") + ")"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var attributes []domain.Attribute
	for rows.Next() {
		var attr domain.Attribute
		if err := rows.Scan(&attr.ID, &attr.Name, &attr.Value, &attr.CreatedAt); err != nil {
			return nil, err
		}
		attributes = append(attributes, attr)
	}

	return attributes, rows.Err()
}

func (r *AttributesRepository) CreateAndGetIDs(ctx context.Context, attributes []domain.Attribute) ([]int, error) {
	if len(attributes) == 0 {
		return []int{}, nil
	}

	query := `INSERT INTO attributes (name, value) VALUES ($1, $2) ON CONFLICT (name, value) DO UPDATE SET name = EXCLUDED.name RETURNING id`
	var ids []int

	for _, attr := range attributes {
		var id int
		err := r.db.QueryRow(ctx, query, attr.Name, attr.Value).Scan(&id)
		if err != nil {
			return nil, err
		}
		ids = append(ids, id)
	}

	return ids, nil
}

func (r *AttributesRepository) FindByNameValueBatch(ctx context.Context, eventDataMap map[string]map[string]string) (map[string][]domain.Attribute, error) {
	if len(eventDataMap) == 0 {
		return make(map[string][]domain.Attribute), nil
	}

	uniquePairs := make(map[string]string)
	for _, pairs := range eventDataMap {
		for name, value := range pairs {
			key := name + ":" + value
			uniquePairs[key] = value
		}
	}

	if len(uniquePairs) == 0 {
		return make(map[string][]domain.Attribute), nil
	}

	query := `SELECT id, name, value, created_at FROM attributes WHERE (name, value) IN (`
	args := make([]any, 0, len(uniquePairs)*2)
	placeholders := make([]string, 0, len(uniquePairs))

	i := 1
	for key := range uniquePairs {
		parts := strings.SplitN(key, ":", 2)
		if len(parts) != 2 {
			continue
		}
		name, value := parts[0], parts[1]
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d)", i, i+1))
		args = append(args, name, value)
		i += 2
	}

	query += strings.Join(placeholders, ", ") + ")"

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	allAttributes := make([]domain.Attribute, 0)
	for rows.Next() {
		var attr domain.Attribute
		if err := rows.Scan(&attr.ID, &attr.Name, &attr.Value, &attr.CreatedAt); err != nil {
			return nil, err
		}
		allAttributes = append(allAttributes, attr)
	}

	result := make(map[string][]domain.Attribute)
	for eventId, pairs := range eventDataMap {
		eventAttrs := make([]domain.Attribute, 0)

		for name, value := range pairs {
			for _, attr := range allAttributes {
				if attr.Name == name && attr.Value == value {
					eventAttrs = append(eventAttrs, attr)
					break
				}
			}
		}
		result[eventId] = eventAttrs
	}

	return result, rows.Err()
}

func (r *AttributesRepository) CreateBatchOptimized(ctx context.Context, eventDataMap map[string]map[string]string) (map[string][]int, error) {
	if len(eventDataMap) == 0 {
		return make(map[string][]int), nil
	}

	uniqueAttrs := make(map[string]domain.Attribute)
	for _, pairs := range eventDataMap {
		for name, value := range pairs {
			key := name + ":" + value
			uniqueAttrs[key] = domain.Attribute{Name: name, Value: value}
		}
	}

	if len(uniqueAttrs) == 0 {
		return make(map[string][]int), nil
	}

	query := `INSERT INTO attributes (name, value) VALUES ($1, $2) ON CONFLICT (name, value) DO UPDATE SET name = EXCLUDED.name RETURNING id, name, value`

	createdAttrs := make(map[string]int)

	for _, attr := range uniqueAttrs {
		var id int
		var name, value string
		err := r.db.QueryRow(ctx, query, attr.Name, attr.Value).Scan(&id, &name, &value)
		if err != nil {
			return nil, err
		}
		key := name + ":" + value
		createdAttrs[key] = id
	}

	result := make(map[string][]int)
	for eventId, pairs := range eventDataMap {
		eventIDs := make([]int, 0)

		for name, value := range pairs {
			key := name + ":" + value
			if id, exists := createdAttrs[key]; exists {
				eventIDs = append(eventIDs, id)
			}
		}
		result[eventId] = eventIDs
	}

	return result, nil
}
