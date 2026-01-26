package project

import (
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectRepo struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) domainProject.ProjectRepository {
	return &projectRepo{db: db}
}

func (r *projectRepo) GetByIds(ids []uuid.UUID) ([]*domainProject.Project, error) {
	if len(ids) == 0 {
		return []*domainProject.Project{}, nil
	}
	var items []*domainProject.Project
	err := r.db.Where("deleted_at IS NULL").Where("id IN ?", ids).Find(&items).Error
	return items, err
}

func (r *projectRepo) Create(entity *domainProject.Project) error {
	now := time.Now().UTC()
	if err := entity.Base.InitForCreate(now); err != nil {
		return err
	}

	return r.db.Create(entity).Error
}

func (r *projectRepo) Update(entity *domainProject.Project) error {
	entity.Base.TouchForUpdate(time.Now().UTC())
	return r.db.Model(&domainProject.Project{}).
		Where("deleted_at IS NULL AND id = ?", entity.ID).
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

func (r *projectRepo) DeleteByIds(ids []uuid.UUID) error {
	if len(ids) == 0 {
		return nil
	}
	now := time.Now().UTC()
	return r.db.Model(&domainProject.Project{}).
		Where("id IN ?", ids).
		Updates(map[string]any{"deleted_at": now, "updated_at": now}).Error
}

func (r *projectRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainProject.Project], error) {
	page, limit := domain.NormalizePagination(params.Page, params.Limit, 10)
	offset := (page - 1) * limit

	query := r.db.Model(&domainProject.Project{}).Where("deleted_at IS NULL")
	if strings.TrimSpace(params.Search) != "" {
		pattern := "%" + strings.TrimSpace(params.Search) + "%"
		query = query.Where("name ILIKE ? OR description ILIKE ?", pattern, pattern)
	}

	var total int64
	if err := query.Count(&total).Error; err != nil {
		return nil, err
	}

	var items []domainProject.Project
	if err := query.Order("created_at DESC").Limit(limit).Offset(offset).Find(&items).Error; err != nil {
		return nil, err
	}

	return &domain.PaginatedList[domainProject.Project]{
		Items:      items,
		Total:      total,
		Page:       page,
		TotalPages: domain.CalculateTotalPages(total, limit),
	}, nil
}
