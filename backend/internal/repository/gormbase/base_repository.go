package gormbase

import (
	"context"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

type VersionLock struct {
	Model   any
	ID      uuid.UUID
	Version uint64
}

func LockVersion(ctx context.Context, db *gorm.DB, request VersionLock) error {
	if request.Version == 0 {
		return domain.ErrInvalidArgument
	}
	result := db.WithContext(ctx).
		Clauses(clause.Locking{Strength: "UPDATE"}).
		Where("id = ? AND version = ?", request.ID, request.Version).
		First(request.Model)
	if result.Error == gorm.ErrRecordNotFound {
		return domain.ErrConflict
	}
	return result.Error
}

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
func (r *BaseRepository[T]) GetByIds(ctx context.Context, ids []uuid.UUID) ([]T, error) {
	if len(ids) == 0 {
		return []T{}, nil
	}
	var items []T
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&items).Error
	return items, err
}

// Create creates a new entity
func (r *BaseRepository[T]) Create(ctx context.Context, entity T) error {
	now := time.Now().UTC()
	if err := entity.GetBase().InitForCreate(now); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(entity).Error
}

// Update updates an existing entity
func (r *BaseRepository[T]) Update(ctx context.Context, entity T) error {
	base := entity.GetBase()
	expectedVersion := base.Version
	base.TouchForUpdate(time.Now().UTC())

	if expectedVersion == 0 {
		return domain.ErrInvalidArgument
	}

	result := r.db.WithContext(ctx).
		Model(entity).
		Where("id = ? AND version = ?", base.ID, expectedVersion).
		Select("*").
		Omit("created_at", clause.Associations).
		Updates(entity)
	if result.Error != nil {
		base.Version = expectedVersion
		return result.Error
	}
	if result.RowsAffected == 0 {
		base.Version = expectedVersion
		return domain.ErrConflict
	}
	return nil
}

// DeleteAtVersion deletes exactly the revision observed by the caller.
func (r *BaseRepository[T]) DeleteAtVersion(ctx context.Context, id uuid.UUID, version uint64) error {
	if version == 0 {
		return domain.ErrInvalidArgument
	}
	var model T
	result := r.db.WithContext(ctx).Where("id = ? AND version = ?", id, version).Delete(&model)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return domain.ErrConflict
	}
	return nil
}

// LockAtVersion serializes a destructive aggregate command and validates its
// optimistic-concurrency token inside the caller's transaction.
func (r *BaseRepository[T]) LockAtVersion(ctx context.Context, id uuid.UUID, version uint64) error {
	var model T
	return LockVersion(ctx, r.db, VersionLock{Model: &model, ID: id, Version: version})
}

// DeleteByIds hard deletes entities by their IDs
func (r *BaseRepository[T]) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	var model T
	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&model).Error
}

// GetPaginatedList retrieves a paginated list of entities with search support
func (r *BaseRepository[T]) GetPaginatedList(ctx context.Context, params domain.PaginationParams, defaultLimit int) (*domain.PaginatedList[T], error) {
	var model T
	query := r.db.WithContext(ctx).Model(&model)

	if r.searchCallback != nil && params.Search != "" {
		query = r.searchCallback(query, params.Search)
	}

	return ExactOffsetPage[T](
		query,
		NormalizeOffsetPage(params, defaultLimit),
		defaultCreatedAtOrder(query),
	)
}

// BulkCreate creates multiple entities in batches
func (r *BaseRepository[T]) BulkCreate(ctx context.Context, entities []T, batchSize int) error {
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

	return r.db.WithContext(ctx).CreateInBatches(entities, batchSize).Error
}

// BulkUpdate updates multiple entities with optional upsert support
func (r *BaseRepository[T]) BulkUpdate(ctx context.Context, entities []T) error {
	if len(entities) == 0 {
		return nil
	}

	now := time.Now().UTC()
	for i := range entities {
		entities[i].GetBase().TouchForUpdate(now)
	}

	// Use transaction for bulk updates
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
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

func defaultCreatedAtOrder(query *gorm.DB) clause.OrderByColumn {
	return clause.OrderByColumn{
		Column: clause.Column{Table: clause.CurrentTable, Name: "created_at"},
		Desc:   true,
	}
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
