package facilitysql

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type objectDataRepo struct {
	db *gorm.DB
}

func NewObjectDataRepository(db *gorm.DB) domainFacility.ObjectDataStore {
	return &objectDataRepo{db: db}
}

func (r *objectDataRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.ObjectData, error) {
	var items []*domainFacility.ObjectData
	err := r.db.Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *objectDataRepo) Create(entity *domainFacility.ObjectData) error {
	return r.db.Create(entity).Error
}

func (r *objectDataRepo) Update(entity *domainFacility.ObjectData) error {
	return r.db.Save(entity).Error
}

func (r *objectDataRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.db.Delete(&domainFacility.ObjectData{}, ids).Error
}

func (r *objectDataRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	var items []domainFacility.ObjectData
	var total int64

	db := r.db.Model(&domainFacility.ObjectData{})

	if params.Search != "" {
		db = db.Where("description ILIKE ?", "%"+params.Search+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (params.Page - 1) * params.Limit
	if err := db.Limit(params.Limit).Offset(offset).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.ObjectData]{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		TotalPages: domain.CalculateTotalPages(total, params.Limit),
	}, nil
}

func (r *objectDataRepo) GetBacnetObjectIDs(objectDataID uuid.UUID) ([]uuid.UUID, error) {
	var ids []uuid.UUID
	err := r.db.Table("object_data_bacnet_objects").
		Select("bacnet_object_id").
		Where("object_data_id = ?", objectDataID).
		Scan(&ids).Error
	return ids, err
}

func (r *objectDataRepo) GetTemplates() ([]*domainFacility.ObjectData, error) {
	var items []*domainFacility.ObjectData
	err := r.db.Where("is_active = ? AND project_id IS NULL", true).Preload("BacnetObjects").Find(&items).Error
	return items, err
}

func (r *objectDataRepo) GetPaginatedListForProject(projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	var items []domainFacility.ObjectData
	var total int64

	db := r.db.Model(&domainFacility.ObjectData{}).Where("project_id = ?", projectID)

	if params.Search != "" {
		db = db.Where("description ILIKE ?", "%"+params.Search+"%")
	}

	if err := db.Count(&total).Error; err != nil {
		return nil, err
	}

	offset := (params.Page - 1) * params.Limit
	if err := db.Limit(params.Limit).Offset(offset).Order("created_at DESC").Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.ObjectData]{
		Items:      items,
		Total:      total,
		Page:       params.Page,
		TotalPages: domain.CalculateTotalPages(total, params.Limit),
	}, nil
}
