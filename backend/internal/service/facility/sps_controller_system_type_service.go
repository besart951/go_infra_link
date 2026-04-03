package facility

import (
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

func (s *SPSControllerSystemTypeService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *SPSControllerSystemTypeService) ListBySPSControllerID(spsControllerID uuid.UUID, page, limit int, search string) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	return s.repo.GetPaginatedListBySPSControllerID(spsControllerID, domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *SPSControllerSystemTypeService) GetByID(id uuid.UUID) (*domainFacility.SPSControllerSystemType, error) {
	return domain.GetByID(s.repo, id)
}

func (s *SPSControllerSystemTypeService) CopyByID(id uuid.UUID) (*domainFacility.SPSControllerSystemType, error) {
	return s.hierarchyCopier.CopySPSControllerSystemTypeByID(id)
}

func (s *SPSControllerSystemTypeService) DeleteByID(id uuid.UUID) error {
	return s.repo.DeleteByIds([]uuid.UUID{id})
}
