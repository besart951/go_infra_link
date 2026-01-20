package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type fieldDeviceRepo struct {
	db *gorm.DB
}

func NewFieldDeviceRepository(db *gorm.DB) facility.FieldDeviceRepository {
	return &fieldDeviceRepo{db: db}
}

func (r *fieldDeviceRepo) GetByIds(ids []uuid.UUID) ([]*facility.FieldDevice, error) {
	var items []*facility.FieldDevice
	err := r.db.
		Preload("Apparat").
		Preload("Specification").
		Preload("SystemPart").
		Preload("BacnetObjects").
		Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *fieldDeviceRepo) Create(entity *facility.FieldDevice) error {
	return r.db.Create(entity).Error
}

func (r *fieldDeviceRepo) Update(entity *facility.FieldDevice) error {
	return r.db.Save(entity).Error
}

func (r *fieldDeviceRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&facility.FieldDevice{}).Error
}

func (r *fieldDeviceRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[facility.FieldDevice], error) {
	return repository.Paginate[facility.FieldDevice](r.db, params, []string{"bmk", "description"})
}
