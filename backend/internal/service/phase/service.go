package phase

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type Service struct {
	repo project.PhaseRepository
}

func NewPhaseService(repo project.PhaseRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(phase *project.Phase) error {
	if err := phase.InitForCreate(time.Now().UTC()); err != nil {
		return err
	}
	return s.repo.Create(phase)
}

func (s *Service) GetByID(id uuid.UUID) (*project.Phase, error) {
	phases, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(phases) == 0 {
		return nil, domain.ErrNotFound
	}
	return phases[0], nil
}

func (s *Service) List(page, limit int, search string) (*domain.PaginatedList[project.Phase], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	params := domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	}
	return s.repo.GetPaginatedList(params)
}

func (s *Service) Update(phase *project.Phase) error {
	return s.repo.Update(phase)
}

func (s *Service) DeleteByID(id uuid.UUID) error {
	return s.repo.DeleteByIds([]uuid.UUID{id})
}
