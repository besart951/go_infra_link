package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

type BuildingService struct {
	repo domainFacility.BuildingRepository
}

func NewBuildingService(repo domainFacility.BuildingRepository) *BuildingService {
	return &BuildingService{repo: repo}
}

func (s *BuildingService) Create(building *domainFacility.Building) error {
	return s.repo.Create(building)
}

func (s *BuildingService) GetByID(id uuid.UUID) (*domainFacility.Building, error) {
	buildings, err := s.repo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(buildings) == 0 {
		return nil, domain.ErrNotFound
	}
	return buildings[0], nil
}

func (s *BuildingService) List(page, limit int, search string) (*domain.PaginatedList[domainFacility.Building], error) {
	page, limit = domain.NormalizePagination(page, limit, 10)
	return s.repo.GetPaginatedList(domain.PaginationParams{
		Page:   page,
		Limit:  limit,
		Search: search,
	})
}

func (s *BuildingService) Update(building *domainFacility.Building) error {
	return s.repo.Update(building)
}

func (s *BuildingService) DeleteByIds(ids []uuid.UUID) error {
	return s.repo.DeleteByIds(ids)
}
