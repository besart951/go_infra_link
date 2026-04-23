package project

import (
	"context"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectRepo struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) domainProject.ProjectRepository {
	return &projectRepo{
		db: db,
	}
}

func (r *projectRepo) GetByIds(ctx context.Context, ids []uuid.UUID) ([]*domainProject.Project, error) {
	if len(ids) == 0 {
		return []*domainProject.Project{}, nil
	}

	var records []*projectRecord
	err := r.db.WithContext(ctx).Where("id IN ?", ids).Find(&records).Error
	return toProjectDomains(records), err
}

func (r *projectRepo) Create(ctx context.Context, entity *domainProject.Project) error {
	if err := entity.Base.InitForCreate(time.Now().UTC()); err != nil {
		return err
	}

	return r.db.WithContext(ctx).Create(toProjectRecord(entity)).Error
}

func (r *projectRepo) Update(ctx context.Context, entity *domainProject.Project) error {
	entity.Base.TouchForUpdate(time.Now().UTC())
	return r.db.WithContext(ctx).Model(&projectRecord{}).
		Where("id = ?", entity.ID).
		Updates(map[string]any{
			"updated_at":  entity.UpdatedAt,
			"name":        entity.Name,
			"description": entity.Description,
			"status":      entity.Status,
			"start_date":  entity.StartDate,
			"phase_id":    entity.PhaseID,
			"creator_id":  entity.CreatorID,
		}).Error
}

func (r *projectRepo) DeleteByIds(ctx context.Context, ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}

	return r.db.WithContext(ctx).Where("id IN ?", ids).Delete(&projectRecord{}).Error
}

func (r *projectRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainProject.Project], error) {
	return r.GetPaginatedListWithStatus(ctx, params, nil)
}

func (r *projectRepo) GetPaginatedListForUser(ctx context.Context, params domain.PaginationParams, userID uuid.UUID) (*domain.PaginatedList[domainProject.Project], error) {
	return r.GetPaginatedListForUserWithStatus(ctx, params, userID, nil)
}

func (r *projectRepo) GetPaginatedListWithStatus(ctx context.Context, params domain.PaginationParams, status *domainProject.ProjectStatus) (*domain.PaginatedList[domainProject.Project], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&projectRecord{})

	if params.Search != "" {
		pattern := "%" + strings.ToLower(strings.TrimSpace(params.Search)) + "%"
		query = query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", pattern, pattern)
	}

	if status != nil && *status != "" {
		query = query.Where("status = ?", *status)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var records []projectRecord
	if err := query.
		Order("created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainProject.Project]{
		Items:      projectDomainValues(records),
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *projectRepo) GetPaginatedListForUserWithStatus(ctx context.Context, params domain.PaginationParams, userID uuid.UUID, status *domainProject.ProjectStatus) (*domain.PaginatedList[domainProject.Project], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&projectRecord{}).
		Joins("LEFT JOIN project_users pu ON pu.project_id = projects.id").
		Where("pu.user_id = ?", userID)

	if params.Search != "" {
		pattern := "%" + strings.ToLower(strings.TrimSpace(params.Search)) + "%"
		query = query.Where("LOWER(projects.name) LIKE ? OR LOWER(projects.description) LIKE ?", pattern, pattern)
	}

	if status != nil && *status != "" {
		query = query.Where("projects.status = ?", *status)
	}

	var total int64
	if err := query.Distinct("projects.id").Count(&total).Error; err != nil {
		return nil, err
	}

	var records []projectRecord
	if err := query.
		Distinct("projects.*").
		Order("projects.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&records).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainProject.Project]{
		Items:      projectDomainValues(records),
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *projectRepo) AddUser(ctx context.Context, projectID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).Create(&projectUserRecord{ProjectID: projectID, UserID: userID}).Error
}

func (r *projectRepo) HasUser(ctx context.Context, projectID, userID uuid.UUID) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).Table("project_users").
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Count(&count).Error
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *projectRepo) RemoveUser(ctx context.Context, projectID, userID uuid.UUID) error {
	return r.db.WithContext(ctx).
		Where("project_id = ? AND user_id = ?", projectID, userID).
		Delete(&projectUserRecord{}).Error
}

func (r *projectRepo) ListUsers(ctx context.Context, projectID uuid.UUID) ([]domainUser.User, error) {
	var users []domainUser.User
	if err := r.db.WithContext(ctx).
		Model(&domainUser.User{}).
		Joins("JOIN project_users pu ON pu.user_id = users.id").
		Where("pu.project_id = ?", projectID).
		Find(&users).Error; err != nil {
		return nil, err
	}
	return users, nil
}
