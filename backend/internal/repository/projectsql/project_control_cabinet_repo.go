package projectsql

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectControlCabinetRepo struct {
	*gormbase.BaseRepository[*project.ProjectControlCabinet]
	db *gorm.DB
}

func NewProjectControlCabinetRepository(db *gorm.DB) project.ProjectControlCabinetRepository {
	baseRepo := gormbase.NewBaseRepository[*project.ProjectControlCabinet](db, nil)
	return &projectControlCabinetRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *projectControlCabinetRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[project.ProjectControlCabinet], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

// GetPaginatedListByProjectID retrieves control cabinets for a project with pagination
func (r *projectControlCabinetRepo) GetPaginatedListByProjectID(projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[project.ProjectControlCabinet], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.Model(&project.ProjectControlCabinet{}).
		Where("deleted_at IS NULL").
		Where("project_id = ?", projectID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []project.ProjectControlCabinet
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[project.ProjectControlCabinet]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

// GetByProjectID retrieves all control cabinets associated with a project
func (r *projectControlCabinetRepo) GetByProjectID(projectID uuid.UUID) ([]*project.ProjectControlCabinet, error) {
	var items []*project.ProjectControlCabinet
	err := r.db.Where("deleted_at IS NULL").Where("project_id = ?", projectID).Find(&items).Error
	return items, err
}

// GetByControlCabinetID retrieves all projects associated with a control cabinet
func (r *projectControlCabinetRepo) GetByControlCabinetID(controlCabinetID uuid.UUID) ([]*project.ProjectControlCabinet, error) {
	var items []*project.ProjectControlCabinet
	err := r.db.Where("deleted_at IS NULL").Where("control_cabinet_id = ?", controlCabinetID).Find(&items).Error
	return items, err
}

// DeleteByProjectAndControlCabinet deletes a specific association
func (r *projectControlCabinetRepo) DeleteByProjectAndControlCabinet(projectID, controlCabinetID uuid.UUID) error {
	return r.db.Model(&project.ProjectControlCabinet{}).
		Where("project_id = ? AND control_cabinet_id = ?", projectID, controlCabinetID).
		Update("deleted_at", gorm.Expr("CURRENT_TIMESTAMP")).Error
}
