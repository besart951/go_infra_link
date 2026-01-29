package facilitysql

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type objectDataRepo struct {
	*gormbase.BaseRepository[*domainFacility.ObjectData]
	db *gorm.DB
}

func NewObjectDataRepository(db *gorm.DB) domainFacility.ObjectDataStore {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where("LOWER(description) LIKE ?", pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainFacility.ObjectData](db, searchCallback)
	return &objectDataRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *objectDataRepo) GetByIds(ids []uuid.UUID) ([]*domainFacility.ObjectData, error) {
	var items []*domainFacility.ObjectData
	if err := r.db.Where("id IN ?", ids).Preload("BacnetObjects").Preload("Apparats").Find(&items).Error; err != nil {
		return nil, err
	}
	return items, nil
}

func (r *objectDataRepo) Create(entity *domainFacility.ObjectData) error {
	return r.BaseRepository.Create(entity)
}

func (r *objectDataRepo) Update(entity *domainFacility.ObjectData) error {
	return r.BaseRepository.Update(entity)
}

func (r *objectDataRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.BaseRepository.DeleteByIds(ids)
}

func (r *objectDataRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.Model(&domainFacility.ObjectData{}).
		Where("project_id IS NULL")

	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.ToLower(strings.TrimSpace(params.Search)) + "%"
		query = query.Where("LOWER(description) LIKE ?", pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainFacility.ObjectData
	if err := query.
		Order("created_at DESC").
		Preload("BacnetObjects").
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.ObjectData]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
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
	err := r.db.Where("deleted_at IS NULL").Where("is_active = ? AND project_id IS NULL", true).Preload("BacnetObjects").Find(&items).Error
	return items, err
}

func (r *objectDataRepo) GetPaginatedListForProject(projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ObjectData], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.Model(&domainFacility.ObjectData{}).
		Where("deleted_at IS NULL").
		Where("project_id = ?", projectID)

	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.ToLower(strings.TrimSpace(params.Search)) + "%"
		query = query.Where("LOWER(description) LIKE ?", pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainFacility.ObjectData
	if err := query.Order("created_at DESC").Preload("BacnetObjects").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainFacility.ObjectData]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}
