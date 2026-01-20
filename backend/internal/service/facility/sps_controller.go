package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetSPSControllerByIds(ids []uuid.UUID) ([]*domainFacility.SPSController, error) {
	return s.SPSControllers.GetByIds(ids)
}

func (s *Service) CreateSPSController(entity *domainFacility.SPSController) error {
	return s.SPSControllers.Create(entity)
}

func (s *Service) UpdateSPSController(entity *domainFacility.SPSController) error {
	return s.SPSControllers.Update(entity)
}

func (s *Service) DeleteSPSControllerByIds(ids []uuid.UUID) error {
	return s.SPSControllers.DeleteByIds(ids)
}

func (s *Service) GetPaginatedSPSControllers(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSController], error) {
	return s.SPSControllers.GetPaginatedList(params)
}
