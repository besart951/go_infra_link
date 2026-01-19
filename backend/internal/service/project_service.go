package service

import (
	"besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type ProjectService struct {
	repo domain.ProjectRepository
}

func NewProjectService(repo domain.ProjectRepository) *ProjectService {
	return &ProjectService{repo: repo}
}

func (s *ProjectService) CreateProject(name string, creatorID uuid.UUID) (*domain.Project, error) {
	proj := &domain.Project{
		Name:      name,
		CreatorID: creatorID,
		Status:    domain.StatusPlanned,
	}

	if err := s.repo.Create(proj); err != nil {
		return nil, err
	}
	return proj, nil
}

func (s *ProjectService) ListProjects(page, limit int, search string) (*domain.PaginatedList[domain.Project], error) {
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}