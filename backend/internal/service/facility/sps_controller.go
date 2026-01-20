package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetSPSControllerByIds(ids []uuid.UUID) ([]*domainFacility.SPSController, error) {
	return s.repo.GetSPSControllerByIds(ids)
}

func (s *Service) CreateSPSController(entity *domainFacility.SPSController) error {
	return s.repo.CreateSPSController(entity)
}

func (s *Service) UpdateSPSController(entity *domainFacility.SPSController) error {
	return s.repo.UpdateSPSController(entity)
}

func (s *Service) DeleteSPSControllerByIds(ids []uuid.UUID) error {
	return s.repo.DeleteSPSControllerByIds(ids)
}

func (s *Service) GetPaginatedSPSControllers(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSController], error) {
	return s.repo.GetPaginatedSPSControllers(params)
}
