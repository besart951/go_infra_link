package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetBuildingByIds(ids []uuid.UUID) ([]*domainFacility.Building, error) {
	return s.Buildings.GetByIds(ids)
}

func (s *Service) CreateBuilding(entity *domainFacility.Building) error {
	return s.Buildings.Create(entity)
}

func (s *Service) UpdateBuilding(entity *domainFacility.Building) error {
	return s.Buildings.Update(entity)
}

func (s *Service) DeleteBuildingByIds(ids []uuid.UUID) error {
	return s.Buildings.DeleteByIds(ids)
}

func (s *Service) GetPaginatedBuildings(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Building], error) {
	return s.Buildings.GetPaginatedList(params)
}
