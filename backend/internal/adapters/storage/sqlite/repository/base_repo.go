package storage

import (
	"context"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type BaseRepository[T any] struct {
	DB *gorm.DB
}

func (r *BaseRepository[T]) Create(ctx context.Context, entity *T) error {
	return r.DB.WithContext(ctx).Create(entity).Error
}

func (r *BaseRepository[T]) GetByID(ctx context.Context, id uuid.UUID) (*T, error) {
	var entity T
	err := r.DB.WithContext(ctx).First(&entity, "id = ?", id).Error
	return &entity, err
}

func (r *BaseRepository[T]) GetAll(ctx context.Context) ([]T, error) {
	var entities []T
	err := r.DB.WithContext(ctx).Find(&entities).Error
	return entities, err
}

func (r *BaseRepository[T]) Update(ctx context.Context, entity *T) error {
	// .Save updates all fields, including associations
	return r.DB.WithContext(ctx).Save(entity).Error
}

func (r *BaseRepository[T]) Delete(ctx context.Context, id uuid.UUID) error {
	var entity T
	return r.DB.WithContext(ctx).Delete(&entity, "id = ?", id).Error
}