package facilitysql

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/besart951/go_infra_link/backend/internal/repository/searchspec"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type bacnetObjectRepo struct {
	*gormbase.BaseRepository[*domainFacility.BacnetObject]
	db *gorm.DB
}

func NewBacnetObjectRepository(db *gorm.DB) domainFacility.BacnetObjectStore {
	baseRepo := gormbase.NewBaseRepository(db,
		gormbase.TrigramSearchCallback[*domainFacility.BacnetObject](searchspec.BacnetObjects.SearchColumns("")...),
	)
	return &bacnetObjectRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *bacnetObjectRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.BacnetObject], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *bacnetObjectRepo) BulkCreate(ctx context.Context, entities []*domainFacility.BacnetObject, batchSize int) error {
	return r.BaseRepository.BulkCreate(ctx, entities, batchSize)
}

func (r *bacnetObjectRepo) GetByFieldDeviceIDs(ctx context.Context, ids []uuid.UUID) ([]*domainFacility.BacnetObject, error) {
	if len(ids) == 0 {
		return []*domainFacility.BacnetObject{}, nil
	}
	var items []*domainFacility.BacnetObject
	err := r.db.WithContext(ctx).
		Where("field_device_id IN ?", ids).
		Preload("StateText").
		Preload("NotificationClass").
		Preload("AlarmType").
		Find(&items).Error
	return items, err
}

func (r *bacnetObjectRepo) DeleteByFieldDeviceIDs(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Where("field_device_id IN ?", ids).Delete(&domainFacility.BacnetObject{}).Error
}

func (r *bacnetObjectRepo) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&domainFacility.BacnetObject{}).Error
}
