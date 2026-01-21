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

// LinkFieldDeviceToProject links a field device to a project
func (s *Service) LinkFieldDeviceToProject(projectID uuid.UUID, fieldDeviceID uuid.UUID) error {
	return s.ProjectFieldDevices.Link(projectID, fieldDeviceID)
}

// UnlinkFieldDeviceFromProject unlinks a field device from a project
func (s *Service) UnlinkFieldDeviceFromProject(projectID uuid.UUID, fieldDeviceID uuid.UUID) error {
	return s.ProjectFieldDevices.Unlink(projectID, fieldDeviceID)
}

// GetProjectIDsByFieldDevice returns all project IDs associated with a field device
func (s *Service) GetProjectIDsByFieldDevice(fieldDeviceID uuid.UUID) ([]uuid.UUID, error) {
	return s.ProjectFieldDevices.GetProjectIDsByFieldDevice(fieldDeviceID)
}

// GetFieldDeviceIDsByProject returns all field device IDs associated with a project
func (s *Service) GetFieldDeviceIDsByProject(projectID uuid.UUID) ([]uuid.UUID, error) {
	return s.ProjectFieldDevices.GetFieldDeviceIDsByProject(projectID)
}
