package project

import (
	controlcabinethandler "github.com/besart951/go_infra_link/backend/internal/handler/project/controlcabinet"
	fielddevicehandler "github.com/besart951/go_infra_link/backend/internal/handler/project/fielddevice"
	membershiphandler "github.com/besart951/go_infra_link/backend/internal/handler/project/membership"
	objectdatahandler "github.com/besart951/go_infra_link/backend/internal/handler/project/objectdata"
	phasehandler "github.com/besart951/go_infra_link/backend/internal/handler/project/phase"
	spscontrollerhandler "github.com/besart951/go_infra_link/backend/internal/handler/project/spscontroller"
)

type Handlers struct {
	Project            *ProjectHandler
	Membership         *membershiphandler.Handler
	ControlCabinet     *controlcabinethandler.Handler
	SPSController      *spscontrollerhandler.Handler
	FieldDevice        *fielddevicehandler.Handler
	ObjectData         *objectdatahandler.Handler
	Phase              *phasehandler.Handler
	FieldDeviceOptions *fielddevicehandler.OptionsHandler
	RefreshBroadcaster *FacilityRefreshBroadcaster
}

type ServiceDeps struct {
	Lifecycle          ProjectLifecycleService
	AccessPolicy       ProjectAccessPolicyService
	Membership         ProjectMembershipService
	Workflow           ProjectWorkflowService
	FacilityLink       ProjectFacilityLinkService
	Phase              PhaseService
	FieldDeviceOptions FieldDeviceOptionsService
}

func NewHandlers(deps ServiceDeps) *Handlers {
	collaboration := NewProjectCollaborationHub()
	workflow := deps.Workflow
	if workflow == nil {
		workflow = newWorkflowFromServices(deps.Lifecycle, deps.Membership)
	}
	projectHandler := newProjectHandler(deps.Lifecycle, deps.AccessPolicy, deps.Membership, workflow, deps.FacilityLink, collaboration)
	return &Handlers{
		Project:            projectHandler,
		Membership:         membershiphandler.NewHandler(deps.AccessPolicy, workflow, projectHandler.notifyProjectChange),
		ControlCabinet:     controlcabinethandler.NewHandler(deps.AccessPolicy, deps.FacilityLink, projectHandler.notifyProjectChange, projectHandler.notifyProjectControlCabinetDelta),
		SPSController:      spscontrollerhandler.NewHandler(deps.AccessPolicy, deps.FacilityLink, projectHandler.notifyProjectChange, projectHandler.notifyProjectSPSControllerDelta),
		FieldDevice:        fielddevicehandler.NewHandler(deps.AccessPolicy, deps.FacilityLink, projectHandler.notifyProjectChange, projectHandler.notifyProjectFieldDeviceDelta),
		ObjectData:         objectdatahandler.NewHandler(deps.AccessPolicy, deps.FacilityLink, projectHandler.notifyProjectChange),
		Phase:              phasehandler.NewHandler(deps.Phase),
		FieldDeviceOptions: fielddevicehandler.NewOptionsHandler(deps.AccessPolicy, deps.FieldDeviceOptions),
		RefreshBroadcaster: NewFacilityRefreshBroadcaster(deps.FacilityLink, collaboration),
	}
}
