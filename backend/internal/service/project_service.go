package service

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	projectservice "github.com/besart951/go_infra_link/backend/internal/service/project"
	"github.com/google/uuid"
)

// Backwards-compatible wrapper so existing imports of internal/service keep working.
// New code should use internal/service/project.
type ProjectService struct {
	svc *projectservice.Service
}

func NewProjectService(repo project.ProjectRepository) *ProjectService {
	return &ProjectService{svc: projectservice.New(repo)}
}

func (s *ProjectService) CreateProject(name string, creatorID uuid.UUID) (*project.Project, error) {
	return s.svc.Create(name, creatorID)
}

func (s *ProjectService) ListProjects(page, limit int, search string) (*domain.PaginatedList[project.Project], error) {
	return s.svc.List(page, limit, search)
}
