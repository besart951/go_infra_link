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

func (r *projectFieldDeviceRepo) GetByIds(ids []uuid.UUID) ([]*project.ProjectFieldDevice, error) {
	return r.BaseRepository.GetByIds(ids)
}

func (r *projectFieldDeviceRepo) Create(entity *project.ProjectFieldDevice) error {
	return r.BaseRepository.Create(entity)
}

func (r *projectFieldDeviceRepo) Update(entity *project.ProjectFieldDevice) error {
	return r.BaseRepository.Update(entity)
}

func (r *projectFieldDeviceRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.BaseRepository.DeleteByIds(ids)
}

func (r *projectFieldDeviceRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[project.ProjectFieldDevice], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}

	// Convert []*ProjectFieldDevice to []ProjectFieldDevice for the interface
	items := make([]project.ProjectFieldDevice, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}

	return &domain.PaginatedList[project.ProjectFieldDevice]{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}, nil
}

// GetByProjectID retrieves all field devices associated with a project
func (r *projectFieldDeviceRepo) GetByProjectID(projectID uuid.UUID) ([]*project.ProjectFieldDevice, error) {
	var items []*project.ProjectFieldDevice
	err := r.db.Where("deleted_at IS NULL").Where("project_id = ?", projectID).Find(&items).Error
	return items, err
}

// GetByFieldDeviceID retrieves all projects associated with a field device
func (r *projectFieldDeviceRepo) GetByFieldDeviceID(fieldDeviceID uuid.UUID) ([]*project.ProjectFieldDevice, error) {
	var items []*project.ProjectFieldDevice
	err := r.db.Where("deleted_at IS NULL").Where("field_device_id = ?", fieldDeviceID).Find(&items).Error
	return items, err
}

// DeleteByProjectAndFieldDevice deletes a specific association
func (r *projectFieldDeviceRepo) DeleteByProjectAndFieldDevice(projectID, fieldDeviceID uuid.UUID) error {
	return r.db.Model(&project.ProjectFieldDevice{}).
		Where("project_id = ? AND field_device_id = ?", projectID, fieldDeviceID).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}
