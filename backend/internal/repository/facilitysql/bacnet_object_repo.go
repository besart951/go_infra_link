package facilitysql

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type bacnetObjectRepo struct {
	db *gorm.DB
}

func NewBacnetObjectRepository(db *gorm.DB) domainFacility.BacnetObjectStore {
	return &bacnetObjectRepo{db: db}
}

func (r *bacnetObjectRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.BacnetObject, error) {
	var items []*domainFacility.BacnetObject
	err := r.db.Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *bacnetObjectRepo) Create(entity *domainFacility.BacnetObject) error {
	return r.db.Create(entity).Error
}

func (r *bacnetObjectRepo) Update(entity *domainFacility.BacnetObject) error {
	return r.db.Save(entity).Error
}

func (r *bacnetObjectRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.db.Delete(&domainFacility.BacnetObject{}, ids).Error
}

func (r *bacnetObjectRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.BacnetObject], error) {
	var items []domainFacility.BacnetObject
	var total int64

	db := r.db.Model(&domainFacility.BacnetObject{})

	if params.Search != "" {
		db = db.Where("text_fix ILIKE ?", "%"+params.Search+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (params.Page - 1) * params.Limit
	if err := db.Limit(params.Limit).Offset(offset).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.BacnetObject]{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		TotalPages: domain.CalculateTotalPages(total, params.Limit),
	}, nil
}

func (r *bacnetObjectRepo) GetByFieldDeviceIDs(ids []uuid.UUID) ([]*domainFacility.BacnetObject, error) {
	var items []*domainFacility.BacnetObject
	err := r.db.Where("field_device_id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *bacnetObjectRepo) SoftDeleteByFieldDeviceIDs(ids []uuid.UUID) error {
	return r.db.Where("field_device_id IN ?", ids).Delete(&domainFacility.BacnetObject{}).Error
}
