package projectservice

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/core/domain/project"
	"github.com/google/uuid"
)

type ProjectService struct {
	repo project.ProjectRepository
}

func NewProjectService(repo project.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) Create(ctx context.Context, p *project.Project) error {
	return s.repo.Create(ctx, p)
}

func (s *ProjectService) GetByID(ctx context.Context, id uuid.UUID) (*project.Project, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *ProjectService) GetAll(ctx context.Context) ([]project.Project, error) {
	return s.repo.GetAll(ctx)
}

func (s *ProjectService) Update(ctx context.Context, p *project.Project) error {
	return s.repo.Update(ctx, p)
}

func (s *ProjectService) Delete(ctx context.Context, id uuid.UUID) error {
	return s.repo.Delete(ctx, id)
}
