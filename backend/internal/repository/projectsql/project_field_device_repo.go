package projectsql

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectFieldDeviceRepo struct {
	*gormbase.BaseRepository[*project.ProjectFieldDevice]
	db *gorm.DB
}

func NewProjectFieldDeviceRepository(db *gorm.DB) project.ProjectFieldDeviceRepository {
	baseRepo := gormbase.NewBaseRepository[*project.ProjectFieldDevice](db, nil)
	return &projectFieldDeviceRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *projectFieldDeviceRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[project.ProjectFieldDevice], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

// GetPaginatedListByProjectID retrieves field devices for a project with pagination
func (r *projectFieldDeviceRepo) GetPaginatedListByProjectID(projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[project.ProjectFieldDevice], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.Model(&project.ProjectFieldDevice{}).
		Where("project_id = ?", projectID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []project.ProjectFieldDevice
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[project.ProjectFieldDevice]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

// GetByProjectID retrieves all field devices associated with a project
func (r *projectFieldDeviceRepo) GetByProjectID(projectID uuid.UUID) ([]*project.ProjectFieldDevice, error) {
	var items []*project.ProjectFieldDevice
	err := r.db.Where("project_id = ?", projectID).Find(&items).Error
	return items, err
}

// GetByFieldDeviceID retrieves all projects associated with a field device
func (r *projectFieldDeviceRepo) GetByFieldDeviceID(fieldDeviceID uuid.UUID) ([]*project.ProjectFieldDevice, error) {
	var items []*project.ProjectFieldDevice
	err := r.db.Where("field_device_id = ?", fieldDeviceID).Find(&items).Error
	return items, err
}

// DeleteByProjectAndFieldDevice deletes a specific association
func (r *projectFieldDeviceRepo) DeleteByProjectAndFieldDevice(projectID, fieldDeviceID uuid.UUID) error {
	return r.db.
		Where("project_id = ? AND field_device_id = ?", projectID, fieldDeviceID).
		Delete(&project.ProjectFieldDevice{}).Error
}
