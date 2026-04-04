package projectsql

import (
	"context"

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

func (r *projectControlCabinetRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[project.ProjectControlCabinet], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *projectControlCabinetRepo) GetPaginatedListByProjectID(ctx context.Context, projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[project.ProjectControlCabinet], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&project.ProjectControlCabinet{}).
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

func (r *projectControlCabinetRepo) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*project.ProjectControlCabinet, error) {
	var items []*project.ProjectControlCabinet
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&items).Error
	return items, err
}

func (r *projectControlCabinetRepo) GetByControlCabinetID(ctx context.Context, controlCabinetID uuid.UUID) ([]*project.ProjectControlCabinet, error) {
	var items []*project.ProjectControlCabinet
	err := r.db.WithContext(ctx).Where("control_cabinet_id = ?", controlCabinetID).Find(&items).Error
	return items, err
}

func (r *projectControlCabinetRepo) DeleteByProjectAndControlCabinet(ctx context.Context, projectID, controlCabinetID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("project_id = ? AND control_cabinet_id = ?", projectID, controlCabinetID).
		Delete(&project.ProjectControlCabinet{}).Error
}
