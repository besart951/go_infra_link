package project

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type ProjectService interface {
	Create(project *domainProject.Project) error
	GetByID(id uuid.UUID) (*domainProject.Project, error)
	List(requesterID uuid.UUID, page, limit int, search string) (*domain.PaginatedList[domainProject.Project], error)
	CanAccessProject(requesterID, projectID uuid.UUID) (bool, error)
	InviteUser(projectID, userID uuid.UUID) error
	ListUsers(projectID uuid.UUID) ([]domainUser.User, error)
	RemoveUser(projectID, userID uuid.UUID) error
	CreateControlCabinet(projectID, controlCabinetID uuid.UUID) (*domainProject.ProjectControlCabinet, error)
	CopyControlCabinet(projectID, controlCabinetID uuid.UUID) (*domainFacility.ControlCabinet, error)
	UpdateControlCabinet(linkID, projectID, controlCabinetID uuid.UUID) (*domainProject.ProjectControlCabinet, error)
	DeleteControlCabinet(linkID, projectID uuid.UUID) error
	CreateSPSController(projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error)
	CopySPSController(projectID, spsControllerID uuid.UUID) (*domainFacility.SPSController, error)
	CopySPSControllerSystemType(projectID, systemTypeID uuid.UUID) (*domainFacility.SPSControllerSystemType, error)
	UpdateSPSController(linkID, projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error)
	DeleteSPSController(linkID, projectID uuid.UUID) error
	CreateFieldDevice(projectID, fieldDeviceID uuid.UUID) (*domainProject.ProjectFieldDevice, error)
	UpdateFieldDevice(linkID, projectID, fieldDeviceID uuid.UUID) (*domainProject.ProjectFieldDevice, error)
	DeleteFieldDevice(linkID, projectID uuid.UUID) error
	ListControlCabinets(projectID uuid.UUID, page, limit int) (*domain.PaginatedList[domainProject.ProjectControlCabinet], error)
	ListSPSControllers(projectID uuid.UUID, page, limit int) (*domain.PaginatedList[domainProject.ProjectSPSController], error)
	ListFieldDevices(projectID uuid.UUID, page, limit int) (*domain.PaginatedList[domainProject.ProjectFieldDevice], error)
	ListObjectData(projectID uuid.UUID, page, limit int, search string, apparatID, systemPartID *uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error)
	AddObjectData(projectID, objectDataID uuid.UUID) (*domainFacility.ObjectData, error)
	RemoveObjectData(projectID, objectDataID uuid.UUID) (*domainFacility.ObjectData, error)
	Update(project *domainProject.Project) error
	DeleteByID(id uuid.UUID) error
}

type PhaseService interface {
	Create(phase *domainProject.Phase) error
	GetByID(id uuid.UUID) (*domainProject.Phase, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainProject.Phase], error)
	Update(phase *domainProject.Phase) error
	DeleteByID(id uuid.UUID) error
}

type FieldDeviceOptionsService interface {
	GetFieldDeviceOptionsForProject(projectID uuid.UUID) (*domainFacility.FieldDeviceOptions, error)
}
