package project

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type Service struct {
	repo domainProject.ProjectRepository
}

func New(repo domainProject.ProjectRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(name string, creatorID uuid.UUID) (*domainProject.Project, error) {
	proj := &domainProject.Project{
		Name:      name,
		CreatorID: creatorID,
		Status:    domainProject.StatusPlanned,
	}

	if err := s.repo.Create(proj); err != nil {
		return nil, err
	}
	return proj, nil
}

func (s *Service) List(page, limit int, search string) (*domain.PaginatedList[domainProject.Project], error) {
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}
