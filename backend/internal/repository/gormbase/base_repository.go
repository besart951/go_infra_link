package gormbase

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	// DefaultBatchSize is the default batch size for bulk operations
	DefaultBatchSize = 100
)

// Entity is the interface that all entities must implement to work with BaseRepository
type Entity interface {
	GetBase() *domain.Base
}

// SearchCallback is a function type for custom search logic
type SearchCallback[T any] func(query *gorm.DB, search string) *gorm.DB

// BaseRepository provides common CRUD operations for entities with hard delete support
type BaseRepository[T Entity] struct {
	db             *gorm.DB
	searchCallback SearchCallback[T]
}

// NewBaseRepository creates a new base repository with optional search callback
func NewBaseRepository[T Entity](db *gorm.DB, searchCallback SearchCallback[T]) *BaseRepository[T] {
	return &BaseRepository[T]{
		db:             db,
		searchCallback: searchCallback,
	}
}

// GetByIds retrieves entities by their IDs
func (r *BaseRepository[T]) GetByIds(ids []uuid.UUID) ([]T, error) {
	if len(ids) == 0 {
		return []T{}, nil
	}
	var items []T
	err := r.db.Where("id IN ?", ids).Find(&items).Error
	return items, err
}

// Create creates a new entity
func (r *BaseRepository[T]) Create(entity T) error {
	now := time.Now().UTC()
	if err := entity.GetBase().InitForCreate(now); err != nil {
		return err
	}
	return r.db.Create(entity).Error
}

// Update updates an existing entity
func (r *BaseRepository[T]) Update(entity T) error {
	entity.GetBase().TouchForUpdate(time.Now().UTC())
	return r.db.Save(entity).Error
}

// DeleteByIds hard deletes entities by their IDs
func (r *BaseRepository[T]) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	var model T
	return r.db.Where("id IN ?", ids).Delete(&model).Error
}

// GetPaginatedList retrieves a paginated list of entities with search support
func (r *BaseRepository[T]) GetPaginatedList(params domain.PaginationParams, defaultLimit int) (*domain.PaginatedList[T], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, defaultLimit)
	offset := (page - 1) * limit

	var model T
	query := r.db.Model(&model)

	// Apply custom search if callback is provided and search term is not empty
	if r.searchCallback != nil && params.Search != "" {
		query = r.searchCallback(query, params.Search)
	}

	// Count total items
	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	// Retrieve paginated items
	var items []T
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[T]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

// BulkCreate creates multiple entities in batches
func (r *BaseRepository[T]) BulkCreate(entities []T, batchSize int) error {
	if len(entities) == 0 {
		return nil
	}

	now := time.Now().UTC()
	for i := range entities {
		if err := entities[i].GetBase().InitForCreate(now); err != nil {
			return err
		}
	}

	if batchSize <= 0 {
		batchSize = DefaultBatchSize
	}

	return r.db.CreateInBatches(entities, batchSize).Error
}

// BulkUpdate updates multiple entities with optional upsert support
func (r *BaseRepository[T]) BulkUpdate(entities []T) error {
	if len(entities) == 0 {
		return nil
	}

	now := time.Now().UTC()
	for i := range entities {
		entities[i].GetBase().TouchForUpdate(now)
	}

	// Use transaction for bulk updates
	return r.db.Transaction(func(tx *gorm.DB) error {
		for _, entity := range entities {
			if err := tx.Save(entity).Error; err != nil {
				return err
			}
		}
		return nil
	})
}

// DB returns the underlying GORM database instance for custom queries
func (r *BaseRepository[T]) DB() *gorm.DB {
	return r.db
}

// DerefPaginatedList converts a PaginatedList[*T] to PaginatedList[T]
// by dereferencing each item pointer. This bridges BaseRepository (which
// operates on pointer types) with domain interfaces (which use value types
// in PaginatedList).
func DerefPaginatedList[T any](src *domain.PaginatedList[*T]) *domain.PaginatedList[T] {
	items := make([]T, len(src.Items))
	for i, item := range src.Items {
		items[i] = *item
	}
	return &domain.PaginatedList[T]{
		Items:      items,
		Total:      src.Total,
		Page:       src.Page,
		TotalPages: src.TotalPages,
	}
}
