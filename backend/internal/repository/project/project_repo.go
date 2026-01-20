package project

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/repository"

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
	var projects []*domainProject.Project
	err := r.db.Preload("Creator").Preload("Phase").Where("id IN ?", ids).Find(&projects).Error
	return projects, err
}

func (r *projectRepo) Create(entity *domainProject.Project) error {
	return r.db.Create(entity).Error
}

func (r *projectRepo) Update(entity *domainProject.Project) error {
	return r.db.Save(entity).Error
}

func (r *projectRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&domainProject.Project{}).Error
}

func (r *projectRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domainProject.Project], error) {
	searchFields := []string{"name", "description"}
	return repository.Paginate[domainProject.Project](r.db, params, searchFields)
}
