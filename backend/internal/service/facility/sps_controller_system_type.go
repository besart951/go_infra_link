package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetSPSControllerSystemTypeByIds(ids []uuid.UUID) ([]*domainFacility.SPSControllerSystemType, error) {
	return s.repo.GetSPSControllerSystemTypeByIds(ids)
}

func (s *Service) CreateSPSControllerSystemType(entity *domainFacility.SPSControllerSystemType) error {
	return s.repo.CreateSPSControllerSystemType(entity)
}

func (s *Service) UpdateSPSControllerSystemType(entity *domainFacility.SPSControllerSystemType) error {
	return s.repo.UpdateSPSControllerSystemType(entity)
}

func (s *Service) DeleteSPSControllerSystemTypeByIds(ids []uuid.UUID) error {
	return s.repo.DeleteSPSControllerSystemTypeByIds(ids)
}

func (s *Service) GetPaginatedSPSControllerSystemTypes(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	return s.repo.GetPaginatedSPSControllerSystemTypes(params)
}
