package facilitysql

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type fieldDeviceRepo struct {
	db *gorm.DB
}

func NewFieldDeviceRepository(db *gorm.DB) domainFacility.FieldDeviceStore {
	return &fieldDeviceRepo{db: db}
}

func (r *fieldDeviceRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.FieldDevice, error) {
	var items []*domainFacility.FieldDevice
	err := r.db.Preload("Specification").Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *fieldDeviceRepo) Create(entity *domainFacility.FieldDevice) error {
	return r.db.Create(entity).Error
}

func (r *fieldDeviceRepo) Update(entity *domainFacility.FieldDevice) error {
	return r.db.Save(entity).Error
}

func (r *fieldDeviceRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.db.Delete(&domainFacility.FieldDevice{}, ids).Error
}

func (r *fieldDeviceRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	var items []domainFacility.FieldDevice
	var total int64

	db := r.db.Model(&domainFacility.FieldDevice{})

	if params.Search != "" {
		db = db.Where("bmk ILIKE ? OR description ILIKE ?", "%"+params.Search+"%", "%"+params.Search+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (params.Page - 1) * params.Limit
	if err := db.Limit(params.Limit).Offset(offset).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.FieldDevice]{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		TotalPages: domain.CalculateTotalPages(total, params.Limit),
	}, nil
}

func (r *fieldDeviceRepo) ExistsApparatNrConflict(spsControllerSystemTypeID uuid.UUID, systemPartID *uuid.UUID, apparatID uuid.UUID, apparatNr int, excludeID *uuid.UUID) (bool, error) {
	db := r.db.Model(&domainFacility.FieldDevice{}).
		Where("sps_controller_system_type_id = ?", spsControllerSystemTypeID).
		Where("apparat_id = ?", apparatID).
		Where("apparat_nr = ?", apparatNr)

	if systemPartID != nil {
		db = db.Where("system_part_id = ?", *systemPartID)
	} else {
		db = db.Where("system_part_id IS NULL")
	}

	if excludeID != nil {
		db = db.Where("id != ?", *excludeID)
	}

	var count int64
	err := db.Count(&count).Error
	return count > 0, err
}
