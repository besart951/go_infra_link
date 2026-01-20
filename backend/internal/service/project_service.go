package service

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type ProjectService struct {
	repo project.ProjectRepository
}

func NewProjectService(repo project.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) CreateProject(name string, creatorID uuid.UUID) (*project.Project, error) {
	proj := &project.Project{
		Name:      name,
		CreatorID: creatorID,
		Status:    project.StatusPlanned,
	}

	if err := s.repo.Create(proj); err != nil {
		return nil, err
	}
	return proj, nil
}

func (s *ProjectService) ListProjects(page, limit int, search string) (*domain.PaginatedList[project.Project], error) {
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}
