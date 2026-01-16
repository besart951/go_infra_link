package repo

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"gorm.io/gorm"
)

type ProjectRepository struct {
	db *gorm.DB
}

func NewProjectRepository(db *gorm.DB) *ProjectRepository {
	return &ProjectRepository{db: db}
}

func (r *ProjectRepository) Create(ctx context.Context, p *domain.Project) error {
	return r.db.WithContext(ctx).Create(p).Error
}

func (r *ProjectRepository) FindByID(ctx context.Context, id string) (*domain.Project, error) {
	var p domain.Project
	res := r.db.WithContext(ctx).First(&p, "id = ?", id)
	if res.Error != nil {
		if res.Error == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, res.Error
	}
	return &p, nil
}

func (r *ProjectRepository) AddMember(ctx context.Context, m *domain.ProjectMember) error {
	return r.db.WithContext(ctx).Create(m).Error
}
