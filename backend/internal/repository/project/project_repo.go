package project

import (
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

func (r *projectRepo) GetByIds(ids []uuid.UUID) ([]*domainProject.Project, error) {
	return r.BaseRepository.GetByIds(ids)
}

func (r *projectRepo) Create(entity *domainProject.Project) error {
	return r.BaseRepository.Create(entity)
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
	return r.BaseRepository.DeleteByIds(ids)
}

func (r *projectRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainProject.Project], error) {
	result, err := r.BaseRepository.GetPaginatedList(params, 10)
	if err != nil {
		return nil, err
	}

	// Convert []*Project to []Project for the interface
	items := make([]domainProject.Project, len(result.Items))
	for i, item := range result.Items {
		items[i] = *item
	}

	return &domain.PaginatedList[domainProject.Project]{
		Items:      items,
		Total:      result.Total,
		Page:       result.Page,
		TotalPages: result.TotalPages,
	}, nil
}

func (r *projectRepo) AddUser(projectID, userID uuid.UUID) error {
	project := &domainProject.Project{Base: domain.Base{ID: projectID}}
	user := &domainUser.User{Base: domain.Base{ID: userID}}
	return r.db.Model(project).Association("Users").Append(user)
}

func (r *projectRepo) RemoveUser(projectID, userID uuid.UUID) error {
	project := &domainProject.Project{Base: domain.Base{ID: projectID}}
	user := &domainUser.User{Base: domain.Base{ID: userID}}
	return r.db.Model(project).Association("Users").Delete(user)
}

func (r *projectRepo) ListUsers(projectID uuid.UUID) ([]domainUser.User, error) {
	project := &domainProject.Project{Base: domain.Base{ID: projectID}}
	var users []domainUser.User
	if err := r.db.Model(project).Association("Users").Find(&users); err != nil {
		return nil, err
	}
	return users, nil
}
