package projectsql

import (
	"context"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectControlCabinetRepo struct {
	db *gorm.DB
}

func NewProjectControlCabinetRepository(db *gorm.DB) project.ProjectControlCabinetRepository {
	return &projectControlCabinetRepo{
		db: db,
	}
}

func (r *projectControlCabinetRepo) GetByIds(ctx context.Context, ids []uuid.UUID) ([]*project.ProjectControlCabinet, error) {
	if len(ids) == 0 {
		return []*project.ProjectControlCabinet{}, nil
	}

	var records []*ProjectControlCabinetRecord
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&records).Error
	return toProjectControlCabinetDomains(records), err
}

func (r *projectControlCabinetRepo) Create(ctx context.Context, entity *project.ProjectControlCabinet) error {
	if err := entity.Base.InitForCreate(time.Now().UTC()); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(toProjectControlCabinetRecord(entity)).Error
}

func (r *projectControlCabinetRepo) Update(ctx context.Context, entity *project.ProjectControlCabinet) error {
	entity.Base.TouchForUpdate(time.Now().UTC())
	return r.db.WithContext(ctx).Model(&ProjectControlCabinetRecord{}).
		Where("id = ?", entity.ID).
		Updates(map[string]any{
			"updated_at":         entity.UpdatedAt,
			"project_id":         entity.ProjectID,
			"control_cabinet_id": entity.ControlCabinetID,
		}).Error
}

func (r *projectControlCabinetRepo) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&ProjectControlCabinetRecord{}).Error
}

func (r *projectControlCabinetRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[project.ProjectControlCabinet], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&ProjectControlCabinetRecord{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []ProjectControlCabinetRecord
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[project.ProjectControlCabinet]{
		Items:      projectControlCabinetDomainValues(records),
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *projectControlCabinetRepo) GetPaginatedListByProjectID(ctx context.Context, projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[project.ProjectControlCabinet], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&ProjectControlCabinetRecord{}).
		Where("project_id = ?", projectID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []ProjectControlCabinetRecord
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[project.ProjectControlCabinet]{
		Items:      projectControlCabinetDomainValues(records),
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *projectControlCabinetRepo) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*project.ProjectControlCabinet, error) {
	var records []*ProjectControlCabinetRecord
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&records).Error
	return toProjectControlCabinetDomains(records), err
}

func (r *projectControlCabinetRepo) GetByControlCabinetID(ctx context.Context, controlCabinetID uuid.UUID) ([]*project.ProjectControlCabinet, error) {
	return r.GetByControlCabinetIDs(ctx, []uuid.UUID{controlCabinetID})
}

func (r *projectControlCabinetRepo) GetByControlCabinetIDs(ctx context.Context, controlCabinetIDs []uuid.UUID) ([]*project.ProjectControlCabinet, error) {
	if len(controlCabinetIDs) == 0 {
		return []*project.ProjectControlCabinet{}, nil
	}

	var records []*ProjectControlCabinetRecord
	err := r.db.WithContext(ctx).Where("control_cabinet_id IN ?", controlCabinetIDs).Find(&records).Error
	return toProjectControlCabinetDomains(records), err
}

func (r *projectControlCabinetRepo) DeleteByControlCabinetIDs(ctx context.Context, controlCabinetIDs []uuid.UUID) error {
	if len(controlCabinetIDs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).
		Where("control_cabinet_id IN ?", controlCabinetIDs).
		Delete(&ProjectControlCabinetRecord{}).Error
}

func (r *projectControlCabinetRepo) DeleteByProjectAndControlCabinet(ctx context.Context, projectID, controlCabinetID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("project_id = ? AND control_cabinet_id = ?", projectID, controlCabinetID).
		Delete(&ProjectControlCabinetRecord{}).Error
}
