package facility

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type SPSControllerSystemTypeService struct {
	repo            domainFacility.SPSControllerSystemTypeStore
	hierarchyCopier *HierarchyCopier
}

func NewSPSControllerSystemTypeService(
	repo domainFacility.SPSControllerSystemTypeStore,
	hierarchyCopier *HierarchyCopier,
) *SPSControllerSystemTypeService {
	return &SPSControllerSystemTypeService{
		repo:            repo,
		hierarchyCopier: hierarchyCopier,
	}
}

func (s *SPSControllerSystemTypeService) List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	return s.repo.GetPaginatedList(ctx, domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *SPSControllerSystemTypeService) ListBySPSControllerID(ctx context.Context, spsControllerID uuid.UUID, page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	return s.repo.GetPaginatedListBySPSControllerID(ctx, spsControllerID, domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *SPSControllerSystemTypeService) ListByProjectID(ctx context.Context, projectID uuid.UUID, page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	return s.repo.GetPaginatedListByProjectID(ctx, projectID, domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *SPSControllerSystemTypeService) GetByID(ctx context.Context, id uuid.UUID) (*domainFacility.SPSControllerSystemType, error) {
	return domain.GetByID(ctx, s.repo, id)
}

func (s *SPSControllerSystemTypeService) CopyByID(ctx context.Context, id uuid.UUID) (*domainFacility.SPSControllerSystemType, error) {
	return s.hierarchyCopier.CopySPSControllerSystemTypeByID(ctx, id)
}

func (s *SPSControllerSystemTypeService) DeleteByID(ctx context.Context, id uuid.UUID) error {
	return s.repo.DeleteByIds(ctx, []uuid.UUID{id})
}
