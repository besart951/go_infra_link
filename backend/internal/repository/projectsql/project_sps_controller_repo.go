package projectsql

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectSPSControllerRepo struct {
	*gormbase.BaseRepository[*project.ProjectSPSController]
	db *gorm.DB
}

func NewProjectSPSControllerRepository(db *gorm.DB) project.ProjectSPSControllerRepository {
	baseRepo := gormbase.NewBaseRepository[*project.ProjectSPSController](db, nil)
	return &projectSPSControllerRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *projectSPSControllerRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[project.ProjectSPSController], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *projectSPSControllerRepo) GetPaginatedListByProjectID(ctx context.Context, projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[project.ProjectSPSController], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&project.ProjectSPSController{}).
		Where("project_id = ?", projectID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []project.ProjectSPSController
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[project.ProjectSPSController]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *projectSPSControllerRepo) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*project.ProjectSPSController, error) {
	var items []*project.ProjectSPSController
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&items).Error
	return items, err
}

func (r *projectSPSControllerRepo) GetBySPSControllerID(ctx context.Context, spsControllerID uuid.UUID) ([]*project.ProjectSPSController, error) {
	var items []*project.ProjectSPSController
	err := r.db.WithContext(ctx).Where("sps_controller_id = ?", spsControllerID).Find(&items).Error
	return items, err
}

func (r *projectSPSControllerRepo) DeleteByProjectAndSPSController(ctx context.Context, projectID, spsControllerID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("project_id = ? AND sps_controller_id = ?", projectID, spsControllerID).
		Delete(&project.ProjectSPSController{}).Error
}
