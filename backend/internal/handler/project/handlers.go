package project

type Handlers struct {
	Project            *ProjectHandler
	Phase              *PhaseHandler
	FieldDeviceOptions *FieldDeviceOptionsHandler
	RefreshBroadcaster *FacilityRefreshBroadcaster
}

type ServiceDeps struct {
	Lifecycle          ProjectLifecycleService
	AccessPolicy       ProjectAccessPolicyService
	Membership         ProjectMembershipService
	FacilityLink       ProjectFacilityLinkService
	Phase              PhaseService
	FieldDeviceOptions FieldDeviceOptionsService
}

func NewHandlers(deps ServiceDeps) *Handlers {
	events := NewProjectEventHub()
	collaboration := NewProjectCollaborationHub()
	projectHandler := newProjectHandler(deps.Lifecycle, deps.AccessPolicy, deps.Membership, deps.FacilityLink, events, collaboration)
	return &Handlers{
		Project:            projectHandler,
		Phase:              NewPhaseHandler(deps.Phase),
		FieldDeviceOptions: NewFieldDeviceOptionsHandler(deps.AccessPolicy, deps.FieldDeviceOptions),
		RefreshBroadcaster: NewFacilityRefreshBroadcaster(deps.FacilityLink, collaboration),
	}
}
