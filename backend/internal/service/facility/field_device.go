package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

func (s *Service) GetFieldDeviceByIds(ids []uuid.UUID) ([]*domainFacility.FieldDevice, error) {
	return s.repo.GetFieldDeviceByIds(ids)
}

func (s *Service) CreateFieldDevice(entity *domainFacility.FieldDevice) error {
	return s.repo.CreateFieldDevice(entity)
}

func (s *Service) UpdateFieldDevice(entity *domainFacility.FieldDevice) error {
	return s.repo.UpdateFieldDevice(entity)
}

func (s *Service) DeleteFieldDeviceByIds(ids []uuid.UUID) error {
	return s.repo.DeleteFieldDeviceByIds(ids)
}

func (s *Service) GetPaginatedFieldDevices(params domain.PaginationParams) (*domain.PaginatedList[domainFacility.FieldDevice], error) {
	return s.repo.GetPaginatedFieldDevices(params)
}
