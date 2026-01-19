package repository

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

type projectRepo struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) domain.ProjectRepository {
	return &projectRepo{db: db}
}

func (r *projectRepo) GetByIds(ids []uuid.UUID) ([]*domain.Project, error) {
	var projects []*domain.Project
	err := r.db.Preload("Creator").Preload("Phase").Where("id IN ?", ids).Find(&projects).Error
	return projects, err
}

func (r *projectRepo) Create(entity *domain.Project) error {
	return r.db.Create(entity).Error
}

func (r *projectRepo) Update(entity *domain.Project) error {
	return r.db.Save(entity).Error
}

func (r *projectRepo) DeleteByIds(ids []uuid.UUID) error {
	return r.db.Where("id IN ?", ids).Delete(&domain.Project{}).Error
}

func (r *projectRepo) GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[domain.Project], error) {
	searchFields := []string{"name", "description"}
	return paginate[domain.Project](r.db, params, searchFields)
}
