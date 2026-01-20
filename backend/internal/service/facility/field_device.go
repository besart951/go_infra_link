package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetFieldDeviceByIds(ids []uuid.UUID) ([]*domainFacility.FieldDevice, error) {
	return s.FieldDevices.GetByIds(ids)
}

func (s *Service) CreateFieldDevice(entity *domainFacility.FieldDevice) error {
	return s.FieldDevices.Create(entity)
}

func (s *Service) UpdateFieldDevice(entity *domainFacility.FieldDevice) error {
	return s.FieldDevices.Update(entity)
}

func (s *Service) DeleteFieldDeviceByIds(ids []uuid.UUID) error {
	return s.FieldDevices.DeleteByIds(ids)
}

func (s *Service) GetPaginatedFieldDevices(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	return s.FieldDevices.GetPaginatedList(params)
}
