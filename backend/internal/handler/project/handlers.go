package project

type Handlers struct {
	Project            *ProjectHandler
	Phase              *PhaseHandler
	FieldDeviceOptions *FieldDeviceOptionsHandler
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
	projectHandler := NewProjectHandler(deps.Lifecycle, deps.AccessPolicy, deps.Membership, deps.FacilityLink, nil)
	return &Handlers{
		Project:            projectHandler,
		Phase:              NewPhaseHandler(deps.Phase),
		FieldDeviceOptions: NewFieldDeviceOptionsHandler(deps.AccessPolicy, deps.FieldDeviceOptions),
	}
}
