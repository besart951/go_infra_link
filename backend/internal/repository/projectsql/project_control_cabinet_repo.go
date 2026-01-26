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

func (r *projectControlCabinetRepo) GetByIds(ids []uuid.UUID) ([]*project.ProjectControlCabinet, error) {
	return r.BaseRepository.GetByIds(ids)
}

func (r *projectControlCabinetRepo) Create(entity *project.ProjectControlCabinet) error {
	return r.BaseRepository.Create(entity)
}

func (r *projectControlCabinetRepo) Update(entity *project.ProjectControlCabinet) error {
	return r.BaseRepository.Update(entity)
}

func (r *projectControlCabinetRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.BaseRepository.DeleteByIds(ids)
}

func (r *projectControlCabinetRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[project.ProjectControlCabinet], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}

	// Convert []*ProjectControlCabinet to []ProjectControlCabinet for the interface
	items := make([]project.ProjectControlCabinet, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}

	return &domain.PaginatedList[project.ProjectControlCabinet]{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
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
