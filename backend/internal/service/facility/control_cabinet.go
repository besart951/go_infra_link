package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetControlCabinetByIds(ids []uuid.UUID) ([]*domainFacility.ControlCabinet, error) {
	return s.ControlCabinets.GetByIds(ids)
}

func (s *Service) CreateControlCabinet(entity *domainFacility.ControlCabinet) error {
	return s.ControlCabinets.Create(entity)
}

func (s *Service) UpdateControlCabinet(entity *domainFacility.ControlCabinet) error {
	return s.ControlCabinets.Update(entity)
}

func (s *Service) DeleteControlCabinetByIds(ids []uuid.UUID) error {
	return s.ControlCabinets.DeleteByIds(ids)
}

func (s *Service) GetPaginatedControlCabinets(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.ControlCabinet], error) {
	return s.ControlCabinets.GetPaginatedList(params)
}
