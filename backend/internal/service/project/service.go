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

func (s *Service) Create(project *domainProject.Project) error {
	if project.Status == "" {
		project.Status = domainProject.StatusPlanned
	}

	return s.repo.Create(project)
}

func (s *Service) GetByIds(ids []uuid.UUID) ([]*domainProject.Project, error) {
	return s.repo.GetByIds(ids)
}

func (s *Service) GetByID(id uuid.UUID) (*domainProject.Project, error) {
	projects, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(projects) == 0 {
		return nil, domain.ErrNotFound
	}
	return projects[0], nil
}

func (s *Service) Update(project *domainProject.Project) error {
	return s.repo.Update(project)
}

func (s *Service) DeleteByIds(ids []uuid.UUID) error {
	return s.repo.DeleteByIds(ids)
}

func (s *Service) List(page, limit int, search string) (*domain.PaginatedList[domainProject.Project], error) {
	page, limit = normalizePagination(page, limit)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func normalizePagination(page, limit int) (int, int) {
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 10
	}

	return page, limit
}
