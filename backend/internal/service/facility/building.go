package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetBuildingByIds(ids []uuid.UUID) ([]*domainFacility.Building, error) {
	return s.repo.GetBuildingByIds(ids)
}

func (s *Service) CreateBuilding(entity *domainFacility.Building) error {
	return s.repo.CreateBuilding(entity)
}

func (s *Service) UpdateBuilding(entity *domainFacility.Building) error {
	return s.repo.UpdateBuilding(entity)
}

func (s *Service) DeleteBuildingByIds(ids []uuid.UUID) error {
	return s.repo.DeleteBuildingByIds(ids)
}

func (s *Service) GetPaginatedBuildings(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.Building], error) {
	return s.repo.GetPaginatedBuildings(params)
}
