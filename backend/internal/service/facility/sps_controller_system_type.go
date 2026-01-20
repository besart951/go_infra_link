package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetSPSControllerSystemTypeByIds(ids []uuid.UUID) ([]*domainFacility.SPSControllerSystemType, error) {
	return s.SPSControllerSystemTypes.GetByIds(ids)
}

func (s *Service) CreateSPSControllerSystemType(entity *domainFacility.SPSControllerSystemType) error {
	return s.SPSControllerSystemTypes.Create(entity)
}

func (s *Service) UpdateSPSControllerSystemType(entity *domainFacility.SPSControllerSystemType) error {
	return s.SPSControllerSystemTypes.Update(entity)
}

func (s *Service) DeleteSPSControllerSystemTypeByIds(ids []uuid.UUID) error {
	return s.SPSControllerSystemTypes.DeleteByIds(ids)
}

func (s *Service) GetPaginatedSPSControllerSystemTypes(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.SPSControllerSystemType], error) {
	return s.SPSControllerSystemTypes.GetPaginatedList(params)
}
