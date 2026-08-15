package project

import (
	"context"

	changeshandler "github.com/besart951/go_infra_link/backend/internal/handler/project/changes"
	controlcabinethandler "github.com/besart951/go_infra_link/backend/internal/handler/project/controlcabinet"
	fielddevicehandler "github.com/besart951/go_infra_link/backend/internal/handler/project/fielddevice"
	membershiphandler "github.com/besart951/go_infra_link/backend/internal/handler/project/membership"
	objectdatahandler "github.com/besart951/go_infra_link/backend/internal/handler/project/objectdata"
	phasehandler "github.com/besart951/go_infra_link/backend/internal/handler/project/phase"
	phasepermissionhandler "github.com/besart951/go_infra_link/backend/internal/handler/project/phasepermission"
	spscontrollerhandler "github.com/besart951/go_infra_link/backend/internal/handler/project/spscontroller"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
	"github.com/google/uuid"
)

type Handlers struct {
	Project            *ProjectHandler
	Changes            *changeshandler.Handler
	Membership         *membershiphandler.Handler
	ControlCabinet     *controlcabinethandler.Handler
	SPSController      *spscontrollerhandler.Handler
	FieldDevice        *fielddevicehandler.Handler
	ObjectData         *objectdatahandler.Handler
	Phase              *phasehandler.Handler
	PhasePermission    *phasepermissionhandler.Handler
	FieldDeviceOptions *fielddevicehandler.OptionsHandler
	RefreshBroadcaster *FacilityRefreshBroadcaster
}

type ServiceDeps struct {
	Lifecycle          ProjectLifecycleService
	Changes            ProjectChangeService
	AccessPolicy       ProjectAccessPolicyService
	Membership         ProjectMembershipService
	Workflow           ProjectWorkflowService
	FacilityLink       ProjectFacilityLinkService
	Phase              PhaseService
	PhasePermission    PhasePermissionService
	FieldDeviceOptions FieldDeviceOptionsService
	Notifications      NotificationEventDispatcher
	Collaboration      *ProjectCollaborationHub
	CopyJobs           *facilityservice.CopyJobManager
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
	projectHandler := newProjectHandler(deps.Lifecycle, deps.AccessPolicy, deps.Membership, workflow, deps.FacilityLink, collaboration, deps.Notifications, deps.Changes)
	controlCabinetHandler := controlcabinethandler.NewHandler(deps.AccessPolicy, deps.FacilityLink, projectHandler.notifyProjectChange, projectHandler.notifyProjectControlCabinetDelta)
	controlCabinetHandler.ConfigureCopyJobs(deps.CopyJobs, func(ctx context.Context, actorID *uuid.UUID, projectID uuid.UUID, eventType, entityID string) {
		projectHandler.notifyProjectChangeForActor(ctx, actorID, projectID, eventType, entityID)
	})
	spsControllerHandler := spscontrollerhandler.NewHandler(deps.AccessPolicy, deps.FacilityLink, projectHandler.notifyProjectChange, projectHandler.notifyProjectSPSControllerDelta)
	spsControllerHandler.ConfigureCopyJobs(deps.CopyJobs, func(ctx context.Context, actorID *uuid.UUID, projectID uuid.UUID, eventType, entityID string) {
		projectHandler.notifyProjectChangeForActor(ctx, actorID, projectID, eventType, entityID)
	})

	return &Handlers{
		Project:            projectHandler,
		Changes:            changeshandler.NewHandler(deps.AccessPolicy, deps.Changes),
		Membership:         membershiphandler.NewHandler(deps.AccessPolicy, workflow, projectHandler.notifyProjectChange),
		ControlCabinet:     controlCabinetHandler,
		SPSController:      spsControllerHandler,
		FieldDevice:        fielddevicehandler.NewHandler(deps.AccessPolicy, deps.FacilityLink, projectHandler.notifyProjectChange, projectHandler.notifyProjectFieldDeviceDelta),
		ObjectData:         objectdatahandler.NewHandler(deps.AccessPolicy, deps.FacilityLink, projectHandler.notifyProjectChange),
		Phase:              phasehandler.NewHandler(deps.Phase),
		PhasePermission:    phasepermissionhandler.NewHandler(deps.PhasePermission),
		FieldDeviceOptions: fielddevicehandler.NewOptionsHandler(deps.AccessPolicy, deps.FieldDeviceOptions),
		RefreshBroadcaster: NewFacilityRefreshBroadcaster(deps.FacilityLink, collaboration, deps.Changes),
	}
}
