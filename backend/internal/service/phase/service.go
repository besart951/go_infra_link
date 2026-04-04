package phase

import (
	"context"
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

func (s *Service) Create(ctx context.Context, phase *project.Phase) error {
	if err := phase.InitForCreate(time.Now().UTC()); err != nil {
		return err
	}
	return s.repo.Create(ctx, phase)
}

func (s *Service) GetByID(ctx context.Context, id uuid.UUID) (*project.Phase, error) {
	return domain.GetByID(ctx, s.repo, id)
}

func (s *Service) List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[project.Phase], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	params := domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	}
	return s.repo.GetPaginatedList(ctx, params)
}

func (s *Service) Update(ctx context.Context, phase *project.Phase) error {
	return s.repo.Update(ctx, phase)
}

func (s *Service) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteByIds(ctx, []uuid.UUID{id})
}
