package project

import (
	"context"

	appprojectlink "github.com/besart951/go_infra_link/backend/internal/application/facility/projectlink"
	appproject "github.com/besart951/go_infra_link/backend/internal/application/project"
	controlcabinethandler "github.com/besart951/go_infra_link/backend/internal/handler/project/controlcabinet"
	fielddevicehandler "github.com/besart951/go_infra_link/backend/internal/handler/project/fielddevice"
	membershiphandler "github.com/besart951/go_infra_link/backend/internal/handler/project/membership"
	objectdatahandler "github.com/besart951/go_infra_link/backend/internal/handler/project/objectdata"
	phasehandler "github.com/besart951/go_infra_link/backend/internal/handler/project/phase"
	phasepermissionhandler "github.com/besart951/go_infra_link/backend/internal/handler/project/phasepermission"
	spscontrollerhandler "github.com/besart951/go_infra_link/backend/internal/handler/project/spscontroller"
	"github.com/google/uuid"
)

type Handlers struct {
	Project            *ProjectHandler
	Membership         *membershiphandler.Handler
	ControlCabinet     *controlcabinethandler.Handler
	SPSController      *spscontrollerhandler.Handler
	FieldDevice        *fielddevicehandler.Handler
	ObjectData         *objectdatahandler.Handler
	Phase              *phasehandler.Handler
	PhasePermission    *phasepermissionhandler.Handler
	FieldDeviceOptions *fielddevicehandler.OptionsHandler
}

type ServiceDeps struct {
	Lifecycle                     ProjectLifecycleService
	AccessPolicy                  ProjectAccessPolicyService
	Membership                    ProjectMembershipService
	Workflow                      ProjectWorkflowService
	FacilityLink                  ProjectFacilityLinkService
	FacilityUnlinker              ProjectFacilityUnlinker
	ProjectDeleter                ProjectDeleter
	ControlCabinetCloner          controlcabinethandler.ProjectControlCabinetCloner
	ControlCabinetAssigner        controlcabinethandler.ProjectControlCabinetAssigner
	ControlCabinetReassigner      controlcabinethandler.ProjectControlCabinetReassigner
	SPSControllerCloner           spscontrollerhandler.ProjectSPSControllerCloner
	SPSControllerSystemTypeCloner spscontrollerhandler.ProjectSPSControllerSystemTypeCloner
	SPSControllerAssigner         spscontrollerhandler.ProjectSPSControllerAssigner
	SPSControllerReassigner       spscontrollerhandler.ProjectSPSControllerReassigner
	FieldDeviceMultiCreator       fielddevicehandler.ProjectFieldDeviceMultiCreator
	FieldDeviceAssigner           fielddevicehandler.ProjectFieldDeviceAssigner
	FieldDeviceBulkAssigner       fielddevicehandler.ProjectFieldDeviceBulkAssigner
	FieldDeviceReassigner         fielddevicehandler.ProjectFieldDeviceReassigner
	ObjectDataAttacher            objectdatahandler.ProjectObjectDataAttacher
	ObjectDataDeactivator         objectdatahandler.ProjectObjectDataDeactivator
	Phase                         PhaseService
	PhasePermission               PhasePermissionService
	FieldDeviceOptions            FieldDeviceOptionsService
	Notifications                 NotificationEventDispatcher
	Collaboration                 *ProjectCollaborationHub
}

type ProjectFacilityUnlinker interface {
	Unlink(context.Context, appprojectlink.Command) error
}

type ProjectDeleter interface {
	Delete(context.Context, appproject.DeleteCommand) error
}

type applicationProjectFacilityLink struct {
	ProjectFacilityLinkService
	unlinker ProjectFacilityUnlinker
}

func (service *applicationProjectFacilityLink) DeleteControlCabinet(
	ctx context.Context,
	linkID, projectID uuid.UUID,
) error {
	return service.unlinker.Unlink(ctx, appprojectlink.Command{
		Kind: appprojectlink.KindControlCabinet, ProjectID: projectID, LinkID: linkID,
	})
}

func (service *applicationProjectFacilityLink) DeleteSPSController(
	ctx context.Context,
	linkID, projectID uuid.UUID,
) error {
	return service.unlinker.Unlink(ctx, appprojectlink.Command{
		Kind: appprojectlink.KindSPSController, ProjectID: projectID, LinkID: linkID,
	})
}

func (service *applicationProjectFacilityLink) DeleteFieldDevice(
	ctx context.Context,
	linkID, projectID uuid.UUID,
) error {
	return service.unlinker.Unlink(ctx, appprojectlink.Command{
		Kind: appprojectlink.KindFieldDevice, ProjectID: projectID, LinkID: linkID,
	})
}

func NewHandlers(deps ServiceDeps) *Handlers {
	collaboration := deps.Collaboration
	if collaboration == nil {
		collaboration = NewProjectCollaborationHub()
	}
	workflow := deps.Workflow
	if workflow == nil {
		workflow = newWorkflowFromServices(deps.Lifecycle, deps.Membership)
	}
	facilityLink := deps.FacilityLink
	if deps.FacilityUnlinker != nil {
		facilityLink = &applicationProjectFacilityLink{
			ProjectFacilityLinkService: deps.FacilityLink,
			unlinker:                   deps.FacilityUnlinker,
		}
	}
	projectHandler := newProjectHandler(
		deps.Lifecycle,
		deps.AccessPolicy,
		deps.Membership,
		workflow,
		facilityLink,
		deps.ProjectDeleter,
		collaboration,
		deps.Notifications,
	)
	return &Handlers{
		Project:        projectHandler,
		Membership:     membershiphandler.NewHandler(deps.AccessPolicy, workflow, projectHandler.notifyProjectChange),
		ControlCabinet: controlcabinethandler.NewHandler(deps.AccessPolicy, facilityLink, deps.ControlCabinetCloner, deps.ControlCabinetAssigner, deps.ControlCabinetReassigner, projectHandler.notifyProjectChange, projectHandler.notifyProjectEvent),
		SPSController:  spscontrollerhandler.NewHandler(deps.AccessPolicy, facilityLink, deps.SPSControllerCloner, deps.SPSControllerSystemTypeCloner, deps.SPSControllerAssigner, deps.SPSControllerReassigner, projectHandler.notifyProjectChange, projectHandler.notifyProjectEvent),
		FieldDevice:    fielddevicehandler.NewHandler(deps.AccessPolicy, facilityLink, deps.FieldDeviceMultiCreator, deps.FieldDeviceAssigner, deps.FieldDeviceBulkAssigner, deps.FieldDeviceReassigner, projectHandler.notifyProjectChange, projectHandler.notifyProjectEvent),
		ObjectData: objectdatahandler.NewHandler(
			deps.AccessPolicy,
			facilityLink,
			deps.ObjectDataAttacher,
			deps.ObjectDataDeactivator,
			projectHandler.notifyProjectEvent,
		),
		Phase:              phasehandler.NewHandler(deps.Phase),
		PhasePermission:    phasepermissionhandler.NewHandler(deps.PhasePermission),
		FieldDeviceOptions: fielddevicehandler.NewOptionsHandler(deps.AccessPolicy, deps.FieldDeviceOptions),
	}
}
