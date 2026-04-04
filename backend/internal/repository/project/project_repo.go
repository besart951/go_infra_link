package project

import (
	"context"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/besart951/go_infra_link/backend/internal/repository/gormbase"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectRepo struct {
	*gormbase.BaseRepository[*domainProject.Project]
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) domainProject.ProjectRepository {
	searchCallback := func(query *gorm.DB, search string) *gorm.DB {
		pattern := "%" + strings.ToLower(strings.TrimSpace(search)) + "%"
		return query.Where("LOWER(name) LIKE ? OR LOWER(description) LIKE ?", pattern, pattern)
	}

	baseRepo := gormbase.NewBaseRepository[*domainProject.Project](db, searchCallback)
	return &projectRepo{
		BaseRepository: baseRepo,
		db:             db,
	}
}

func (r *projectRepo) Update(ctx context.Context, entity *domainProject.Project) error {
	entity.Base.TouchForUpdate(time.Now().UTC())
	return r.db.WithContext(ctx).Model(&domainProject.Project{}).
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

func (r *projectRepo) GetPaginatedList(ctx context.Context, params domain.PaginationParams) (*domain.PaginatedList[domainProject.Project], error) {
	result, err := r.BaseRepository.GetPaginatedList(ctx, params, 10)
	if err != nil {
		return nil, err
	}
	return gormbase.DerefPaginatedList(result), nil
}

func (r *projectRepo) GetPaginatedListForUser(ctx context.Context, params domain.PaginationParams, userID uuid.UUID) (*domain.PaginatedList[domainProject.Project], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.WithContext(ctx).Model(&domainProject.Project{}).
		Joins("LEFT JOIN project_users pu ON pu.project_id = projects.id").
		Where("pu.user_id = ? OR projects.creator_id = ?", userID, userID)

	if params.Search != "" {
		pattern := "%" + strings.ToLower(strings.TrimSpace(params.Search)) + "%"
		query = query.Where("LOWER(projects.name) LIKE ? OR LOWER(projects.description) LIKE ?", pattern, pattern)
	}

	var total int64
	if err := query.Distinct("projects.id").Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainProject.Project
	if err := query.
		Distinct("projects.*").
		Order("projects.created_at DESC").
		Limit(limit).
		Offset(offset).
		Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainProject.Project]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}

func (r *projectRepo) AddUser(ctx context.Context, projectID, userID uuid.UUID) error {
	project := &domainProject.Project{Base: domain.Base{ID: projectID}}
	user := &domainUser.User{Base: domain.Base{ID: userID}}
	return r.db.WithContext(ctx).Model(project).Association("Users").Append(user)
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
	project := &domainProject.Project{Base: domain.Base{ID: projectID}}
	user := &domainUser.User{Base: domain.Base{ID: userID}}
	return r.db.WithContext(ctx).Model(project).Association("Users").Delete(user)
}

func (r *projectRepo) ListUsers(ctx context.Context, projectID uuid.UUID) ([]domainUser.User, error) {
	project := &domainProject.Project{Base: domain.Base{ID: projectID}}
	var users []domainUser.User
	if err := r.db.WithContext(ctx).Model(project).Association("Users").Find(&users); err != nil {
		return nil, err
	}
	return users, nil
}
