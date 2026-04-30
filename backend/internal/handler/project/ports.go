package project

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type ProjectLifecycleService interface {
	Create(ctx context.Context, project *domainProject.Project) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainProject.Project, error)
	List(ctx context.Context, requesterID uuid.UUID, page, limit int, search string, status *domainProject.ProjectStatus) (*domain.PaginatedList[domainProject.Project], error)
	Update(ctx context.Context, project *domainProject.Project) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type ProjectAccessPolicyService interface {
	CanAccessProject(ctx context.Context, requesterID, projectID uuid.UUID, requesterRole *domainUser.Role) (bool, error)
	CanUseProjectPermission(ctx context.Context, requesterID uuid.UUID, requesterRole *domainUser.Role, permission string) (bool, error)
	CanUseProjectPermissionForProject(ctx context.Context, requesterID, projectID uuid.UUID, requesterRole *domainUser.Role, permission string) (bool, error)
}

type ProjectMembershipService interface {
	InviteUser(ctx context.Context, projectID, userID uuid.UUID) error
	ListUsers(ctx context.Context, projectID uuid.UUID) ([]domainUser.User, error)
	RemoveUser(ctx context.Context, projectID, userID uuid.UUID) error
}

type ProjectWorkflowService interface {
	CreateProject(ctx context.Context, project *domainProject.Project) error
	InviteUser(ctx context.Context, projectID, userID uuid.UUID) error
	ListUsers(ctx context.Context, projectID uuid.UUID) ([]domainUser.User, error)
	RemoveUser(ctx context.Context, projectID, userID uuid.UUID) error
}

type ProjectFacilityLinkService interface {
	CreateControlCabinet(ctx context.Context, projectID, controlCabinetID uuid.UUID) (*domainProject.ProjectControlCabinet, error)
	CopyControlCabinet(ctx context.Context, projectID, controlCabinetID uuid.UUID) (*domainFacility.ControlCabinet, error)
	UpdateControlCabinet(ctx context.Context, linkID, projectID, controlCabinetID uuid.UUID) (*domainProject.ProjectControlCabinet, error)
	DeleteControlCabinet(ctx context.Context, linkID, projectID uuid.UUID) error
	CreateSPSController(ctx context.Context, projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error)
	CopySPSController(ctx context.Context, projectID, spsControllerID uuid.UUID) (*domainFacility.SPSController, error)
	CopySPSControllerSystemType(ctx context.Context, projectID, systemTypeID uuid.UUID) (*domainFacility.SPSControllerSystemType, error)
	UpdateSPSController(ctx context.Context, linkID, projectID, spsControllerID uuid.UUID) (*domainProject.ProjectSPSController, error)
	DeleteSPSController(ctx context.Context, linkID, projectID uuid.UUID) error
	CreateFieldDevice(ctx context.Context, projectID, fieldDeviceID uuid.UUID) (*domainProject.ProjectFieldDevice, error)
	UpdateFieldDevice(ctx context.Context, linkID, projectID, fieldDeviceID uuid.UUID) (*domainProject.ProjectFieldDevice, error)
	DeleteFieldDevice(ctx context.Context, linkID, projectID uuid.UUID) error
	MultiCreateFieldDevices(ctx context.Context, projectID uuid.UUID, fieldDeviceIDs []uuid.UUID) ([]uuid.UUID, []string)
	MultiCreateAndAssignFieldDevices(ctx context.Context, projectID uuid.UUID, items []domainFacility.FieldDeviceCreateItem) (*domainFacility.FieldDeviceMultiCreateResult, error)
	ListControlCabinets(ctx context.Context, projectID uuid.UUID, page, limit int) (*domain.PaginatedList[domainProject.ProjectControlCabinet], error)
	ListSPSControllers(ctx context.Context, projectID uuid.UUID, page, limit int) (*domain.PaginatedList[domainProject.ProjectSPSController], error)
	ListFieldDevices(ctx context.Context, projectID uuid.UUID, page, limit int) (*domain.PaginatedList[domainProject.ProjectFieldDevice], error)
	ListObjectData(ctx context.Context, projectID uuid.UUID, page, limit int, search string, apparatID, systemPartID *uuid.UUID) (*domain.PaginatedList[domainFacility.ObjectData], error)
	AddObjectData(ctx context.Context, projectID, objectDataID uuid.UUID) (*domainFacility.ObjectData, error)
	RemoveObjectData(ctx context.Context, projectID, objectDataID uuid.UUID) (*domainFacility.ObjectData, error)
	ListProjectIDsByControlCabinetID(ctx context.Context, controlCabinetID uuid.UUID) ([]uuid.UUID, error)
	ListProjectIDsBySPSControllerID(ctx context.Context, spsControllerID uuid.UUID) ([]uuid.UUID, error)
}

type PhaseService interface {
	Create(ctx context.Context, phase *domainProject.Phase) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainProject.Phase, error)
	List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainProject.Phase], error)
	Update(ctx context.Context, phase *domainProject.Phase) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type PhasePermissionService interface {
	Create(ctx context.Context, rule *domainProject.PhasePermission) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainProject.PhasePermission, error)
	List(ctx context.Context, phaseID *uuid.UUID) ([]domainProject.PhasePermission, error)
	Update(ctx context.Context, rule *domainProject.PhasePermission) error
	DeleteByID(ctx context.Context, id uuid.UUID) error
}

type FieldDeviceOptionsService interface {
	GetFieldDeviceOptionsForProject(ctx context.Context, projectID uuid.UUID) (*domainFacility.FieldDeviceOptions, error)
}

type NotificationEventDispatcher interface {
	DispatchEvent(ctx context.Context, input domainNotification.DispatchEventInput) error
}
