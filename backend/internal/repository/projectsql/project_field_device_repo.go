package projectsql

import (
	"context"

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

func (r *projectFieldDeviceRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[project.ProjectFieldDevice], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *projectFieldDeviceRepo) GetPaginatedListByProjectID(ctx context.Context, projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[project.ProjectFieldDevice], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&project.ProjectFieldDevice{}).
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

func (r *projectFieldDeviceRepo) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*project.ProjectFieldDevice, error) {
	var items []*project.ProjectFieldDevice
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&items).Error
	return items, err
}

func (r *projectFieldDeviceRepo) GetByFieldDeviceID(ctx context.Context, fieldDeviceID uuid.UUID) ([]*project.ProjectFieldDevice, error) {
	var items []*project.ProjectFieldDevice
	err := r.db.WithContext(ctx).Where("field_device_id = ?", fieldDeviceID).Find(&items).Error
	return items, err
}

func (r *projectFieldDeviceRepo) DeleteByProjectAndFieldDevice(ctx context.Context, projectID, fieldDeviceID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("project_id = ? AND field_device_id = ?", projectID, fieldDeviceID).
		Delete(&project.ProjectFieldDevice{}).Error
}
