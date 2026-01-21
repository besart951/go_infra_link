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

// LinkSPSControllerToProject links an SPS controller to a project
func (s *Service) LinkSPSControllerToProject(projectID uuid.UUID, spsControllerID uuid.UUID) error {
	return s.ProjectSPSControllers.Link(projectID, spsControllerID)
}

// UnlinkSPSControllerFromProject unlinks an SPS controller from a project
func (s *Service) UnlinkSPSControllerFromProject(projectID uuid.UUID, spsControllerID uuid.UUID) error {
	return s.ProjectSPSControllers.Unlink(projectID, spsControllerID)
}

// GetProjectIDsBySPSController returns all project IDs associated with an SPS controller
func (s *Service) GetProjectIDsBySPSController(spsControllerID uuid.UUID) ([]uuid.UUID, error) {
	return s.ProjectSPSControllers.GetProjectIDsBySPSController(spsControllerID)
}

// GetSPSControllerIDsByProject returns all SPS controller IDs associated with a project
func (s *Service) GetSPSControllerIDsByProject(projectID uuid.UUID) ([]uuid.UUID, error) {
	return s.ProjectSPSControllers.GetSPSControllerIDsByProject(projectID)
}
