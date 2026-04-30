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

type specificationRepo struct {
	*gormbase.BaseRepository[*domainFacility.Specification]
	db *gorm.DB
}

func NewSpecificationRepository(db *gorm.DB) domainFacility.SpecificationStore {
	baseRepo := gormbase.NewBaseRepository(db,
		gormbase.TrigramSearchCallback[*domainFacility.Specification](searchspec.Specifications.SearchColumns("")...),
	)
	return &specificationRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *specificationRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Specification], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *specificationRepo) GetByFieldDeviceIDs(ctx context.Context, fieldDeviceIDs []uuid.UUID) ([]*domainFacility.Specification, error) {
	if len(fieldDeviceIDs) == 0 {
		return []*domainFacility.Specification{}, nil
	}
	var items []*domainFacility.Specification
	err := r.db.WithContext(ctx).Where("field_device_id IN ?", fieldDeviceIDs).Find(&items).Error
	return items, err
}

func (r *specificationRepo) DeleteByFieldDeviceIDs(ctx context.Context, fieldDeviceIDs []uuid.UUID) error {
	if len(fieldDeviceIDs) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).
		Where("field_device_id IN ?", fieldDeviceIDs).
		Delete(&domainFacility.Specification{}).Error
}
