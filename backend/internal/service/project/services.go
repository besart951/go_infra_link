package project

import (
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
)

type Dependencies struct {
	Projects                 domainProject.ProjectRepository
	ProjectControlCabinets   domainProject.ProjectControlCabinetRepository
	ProjectSPSControllers    domainProject.ProjectSPSControllerRepository
	ProjectFieldDevices      domainProject.ProjectFieldDeviceRepository
	Users                    domainUser.UserRepository
	RolePermissions          domainUser.RolePermissionRepository
	ObjectData               domainFacility.ObjectDataStore
	BacnetObjects            domainFacility.BacnetObjectStore
	Specifications           domainFacility.SpecificationStore
	ControlCabinets          domainFacility.ControlCabinetRepository
	SPSControllers           domainFacility.SPSControllerRepository
	SPSControllerSystemTypes domainFacility.SPSControllerSystemTypeStore
	FieldDevices             domainFacility.FieldDeviceStore
	HierarchyCopier          *facilityservice.HierarchyCopier
	FieldDeviceCreator       fieldDeviceCreator
}

type Services struct {
	Lifecycle    *ProjectLifecycleService
	AccessPolicy *ProjectAccessPolicyService
	Membership   *ProjectMembershipService
	Workflow     *ProjectWorkflowService
	FacilityLink *ProjectFacilityLinkService
}

func NewServices(deps Dependencies, cfgs ...Config) *Services {
	var cfg Config
	if len(cfgs) > 0 {
		cfg = cfgs[0]
	}
	tx := newTxCoordinator(cfg)

	services := &Services{}
	services.AccessPolicy = &ProjectAccessPolicyService{
		repo:     deps.Projects,
		userRepo: deps.Users,
	}
	services.Lifecycle = &ProjectLifecycleService{
		repo:               deps.Projects,
		userRepo:           deps.Users,
		rolePermissionRepo: deps.RolePermissions,
		objectDataRepo:     deps.ObjectData,
		bacnetObjectRepo:   deps.BacnetObjects,
	}
	services.Lifecycle.bindTransactions(tx)
	services.Membership = &ProjectMembershipService{
		repo:     deps.Projects,
		userRepo: deps.Users,
	}
	services.Workflow = newProjectWorkflowService(services.Lifecycle, services.Membership)
	services.FacilityLink = &ProjectFacilityLinkService{
		projectRepo:               deps.Projects,
		projectControlCabinetRepo: deps.ProjectControlCabinets,
		projectSPSControllerRepo:  deps.ProjectSPSControllers,
		projectFieldDeviceRepo:    deps.ProjectFieldDevices,
		objectDataRepo:            deps.ObjectData,
		bacnetObjectRepo:          deps.BacnetObjects,
		specificationRepo:         deps.Specifications,
		controlCabinetRepo:        deps.ControlCabinets,
		spsControllerRepo:         deps.SPSControllers,
		spsControllerSystemRepo:   deps.SPSControllerSystemTypes,
		fieldDeviceRepo:           deps.FieldDevices,
		hierarchyCopier:           deps.HierarchyCopier,
		fieldDeviceCreator:        deps.FieldDeviceCreator,
	}
	services.FacilityLink.bindTransactions(tx)

	return services
}
