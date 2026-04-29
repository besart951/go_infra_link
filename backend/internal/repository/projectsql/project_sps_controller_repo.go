package projectsql

import (
	"context"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type projectSPSControllerRepo struct {
	db *gorm.DB
}

func NewProjectSPSControllerRepository(db *gorm.DB) project.ProjectSPSControllerRepository {
	return &projectSPSControllerRepo{
		db: db,
	}
}

func (r *projectSPSControllerRepo) GetByIds(ctx context.Context, ids []uuid.UUID) ([]*project.ProjectSPSController, error) {
	if len(ids) == 0 {
		return []*project.ProjectSPSController{}, nil
	}

	var records []*ProjectSPSControllerRecord
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&records).Error
	return toProjectSPSControllerDomains(records), err
}

func (r *projectSPSControllerRepo) Create(ctx context.Context, entity *project.ProjectSPSController) error {
	if err := entity.Base.InitForCreate(time.Now().UTC()); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(toProjectSPSControllerRecord(entity)).Error
}

func (r *projectSPSControllerRepo) BulkCreate(ctx context.Context, entities []*project.ProjectSPSController, batchSize int) error {
	if len(entities) == 0 {
		return nil
	}
	if batchSize <= 0 {
		batchSize = 200
	}

	now := time.Now().UTC()
	records := make([]projectSPSControllerRecord, 0, len(entities))
	for _, entity := range entities {
		if entity == nil {
			continue
		}
		if err := entity.Base.InitForCreate(now); err != nil {
			return err
		}
		records = append(records, *toProjectSPSControllerRecord(entity))
	}
	if len(records) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		CreateInBatches(records, batchSize).Error
}

func (r *projectSPSControllerRepo) Update(ctx context.Context, entity *project.ProjectSPSController) error {
	entity.Base.TouchForUpdate(time.Now().UTC())
	return r.db.WithContext(ctx).Model(&ProjectSPSControllerRecord{}).
		Where("id = ?", entity.ID).
		Updates(map[string]any{
			"updated_at":        entity.UpdatedAt,
			"project_id":        entity.ProjectID,
			"sps_controller_id": entity.SPSControllerID,
		}).Error
}

func (r *projectSPSControllerRepo) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&ProjectSPSControllerRecord{}).Error
}

func (r *projectSPSControllerRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[project.ProjectSPSController], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&ProjectSPSControllerRecord{})

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []ProjectSPSControllerRecord
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[project.ProjectSPSController]{
		Items:      projectSPSControllerDomainValues(records),
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *projectSPSControllerRepo) GetPaginatedListByProjectID(ctx context.Context, projectID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[project.ProjectSPSController], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&ProjectSPSControllerRecord{}).
		Where("project_id = ?", projectID)

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []ProjectSPSControllerRecord
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&records).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[project.ProjectSPSController]{
		Items:      projectSPSControllerDomainValues(records),
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *projectSPSControllerRepo) GetByProjectID(ctx context.Context, projectID uuid.UUID) ([]*project.ProjectSPSController, error) {
	var records []*ProjectSPSControllerRecord
	err := r.db.WithContext(ctx).Where("project_id = ?", projectID).Find(&records).Error
	return toProjectSPSControllerDomains(records), err
}

func (r *projectSPSControllerRepo) GetBySPSControllerID(ctx context.Context, spsControllerID uuid.UUID) ([]*project.ProjectSPSController, error) {
	return r.GetBySPSControllerIDs(ctx, []uuid.UUID{spsControllerID})
}

func (r *projectSPSControllerRepo) GetBySPSControllerIDs(ctx context.Context, spsControllerIDs []uuid.UUID) ([]*project.ProjectSPSController, error) {
	if len(spsControllerIDs) == 0 {
		return []*project.ProjectSPSController{}, nil
	}

	var records []*ProjectSPSControllerRecord
	err := r.db.WithContext(ctx).Where("sps_controller_id IN ?", spsControllerIDs).Find(&records).Error
	return toProjectSPSControllerDomains(records), err
}

func (r *projectSPSControllerRepo) DeleteBySPSControllerIDs(ctx context.Context, spsControllerIDs []uuid.UUID) error {
	if len(spsControllerIDs) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).
		Where("sps_controller_id IN ?", spsControllerIDs).
		Delete(&ProjectSPSControllerRecord{}).Error
}

func (r *projectSPSControllerRepo) DeleteByProjectAndSPSController(ctx context.Context, projectID, spsControllerID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("project_id = ? AND sps_controller_id = ?", projectID, spsControllerID).
		Delete(&ProjectSPSControllerRecord{}).Error
}
