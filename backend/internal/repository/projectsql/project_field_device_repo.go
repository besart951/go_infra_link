package projectsql

import (
	"context"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectFieldDeviceRepo struct {
	db *gorm.DB
}

func NewProjectFieldDeviceRepository(db *gorm.DB) project.ProjectFieldDeviceRepository {
	return &projectFieldDeviceRepo{
		db: db,
	}
}

func (r *projectFieldDeviceRepo) GetByIds(ctx context.Context, ids []uuid.UUID) ([]*project.ProjectFieldDevice, error) {
	if len(ids) == 0 {
		return []*project.ProjectFieldDevice{}, nil
	}

	var records []*ProjectFieldDeviceRecord
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&records).Error
	return toProjectFieldDeviceDomains(records), err
}

func (r *projectFieldDeviceRepo) Create(ctx context.Context, entity *project.ProjectFieldDevice) error {
	if err := entity.Base.InitForCreate(time.Now().UTC()); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(toProjectFieldDeviceRecord(entity)).Error
}

func (r *projectFieldDeviceRepo) Update(ctx context.Context, entity *project.ProjectFieldDevice) error {
	entity.Base.TouchForUpdate(time.Now().UTC())
	return r.db.WithContext(ctx).Model(&ProjectFieldDeviceRecord{}).
		Where("id = ?", entity.ID).
		Updates(map[string]any{
			"updated_at":      entity.UpdatedAt,
			"project_id":      entity.ProjectID,
			"field_device_id": entity.FieldDeviceID,
		}).Error
}

func (r *projectFieldDeviceRepo) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&ProjectFieldDeviceRecord{}).Error
}

func (r *projectFieldDeviceRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[project.ProjectFieldDevice], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&ProjectFieldDeviceRecord{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []ProjectFieldDeviceRecord
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[project.ProjectFieldDevice]{
		Items:      projectFieldDeviceDomainValues(records),
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *projectFieldDeviceRepo) GetPaginatedListByProjectID(ctx context.Context, projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[project.ProjectFieldDevice], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&ProjectFieldDeviceRecord{}).
		Where("project_id = ?", projectID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []ProjectFieldDeviceRecord
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[project.ProjectFieldDevice]{
		Items:      projectFieldDeviceDomainValues(records),
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *projectFieldDeviceRepo) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*project.ProjectFieldDevice, error) {
	var records []*ProjectFieldDeviceRecord
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&records).Error
	return toProjectFieldDeviceDomains(records), err
}

func (r *projectFieldDeviceRepo) GetByFieldDeviceID(ctx context.Context, fieldDeviceID uuid.UUID) ([]*project.ProjectFieldDevice, error) {
	return r.GetByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID})
}

func (r *projectFieldDeviceRepo) GetByFieldDeviceIDs(ctx context.Context, fieldDeviceIDs []uuid.UUID) ([]*project.ProjectFieldDevice, error) {
	if len(fieldDeviceIDs) == 0 {
		return []*project.ProjectFieldDevice{}, nil
	}

	var records []*ProjectFieldDeviceRecord
	err := r.db.WithContext(ctx).Where("field_device_id IN ?", fieldDeviceIDs).Find(&records).Error
	return toProjectFieldDeviceDomains(records), err
}

func (r *projectFieldDeviceRepo) DeleteByFieldDeviceIDs(ctx context.Context, fieldDeviceIDs []uuid.UUID) error {
	if len(fieldDeviceIDs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).
		Where("field_device_id IN ?", fieldDeviceIDs).
		Delete(&ProjectFieldDeviceRecord{}).Error
}

func (r *projectFieldDeviceRepo) DeleteByProjectAndFieldDevice(ctx context.Context, projectID, fieldDeviceID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("project_id = ? AND field_device_id = ?", projectID, fieldDeviceID).
		Delete(&ProjectFieldDeviceRecord{}).Error
}
