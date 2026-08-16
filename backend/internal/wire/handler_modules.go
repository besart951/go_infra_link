package wire

import (
	facilityhandler "github.com/besart951/go_infra_link/backend/internal/handler/facility"
	projecthandler "github.com/besart951/go_infra_link/backend/internal/handler/project"
	userhandler "github.com/besart951/go_infra_link/backend/internal/handler/user"
	facilityservice "github.com/besart951/go_infra_link/backend/internal/service/facility"
)

func newProjectHandlers(services *Services, runtime *RuntimeAdapters, copyJobs *facilityservice.CopyJobManager) *projecthandler.Handlers {
	return projecthandler.NewHandlers(projecthandler.ServiceDeps{
		Lifecycle:          services.Project.Lifecycle,
		Changes:            services.Project.Changes,
		AccessPolicy:       services.Project.AccessPolicy,
		Membership:         services.Project.Membership,
		Workflow:           services.Project.Workflow,
		FacilityLink:       services.Project.FacilityLink,
		Phase:              services.Phase,
		PhasePermission:    services.PhasePermission,
		FieldDeviceOptions: services.Facility.FieldDevice,
		FacilityDetail: projecthandler.FacilityDetailServices{
			Building:                services.Facility.Building,
			ControlCabinet:          services.Facility.ControlCabinet,
			SPSController:           services.Facility.SPSController,
			SPSControllerSystemType: services.Facility.SPSControllerSystemType,
			FieldDevice:             services.Facility.FieldDevice,
			Apparat:                 services.Facility.Apparat,
			SystemPart:              services.Facility.SystemPart,
		},
		Authorization: services.RBAC,
		Notifications: services.Notification,
		Collaboration: runtime.ProjectCollaboration,
		CopyJobs:      copyJobs,
	})
}

func newFacilityHandlers(services *Services, collaboration facilityhandler.ProjectRefreshBroadcaster, referenceData facilityhandler.FacilityReferenceDataRealtime, copyJobs *facilityservice.CopyJobManager) *facilityhandler.Handlers {
	return facilityhandler.NewHandlers(facilityhandler.ServiceDeps{
		Building:                services.Facility.Building,
		SystemType:              services.Facility.SystemType,
		SystemPart:              services.Facility.SystemPart,
		Apparat:                 services.Facility.Apparat,
		ControlCabinet:          services.Facility.ControlCabinet,
		FieldDevice:             services.Facility.FieldDevice,
		BacnetObject:            services.Facility.BacnetObject,
		SPSController:           services.Facility.SPSController,
		StateText:               services.Facility.StateText,
		NotificationClass:       services.Facility.NotificationClass,
		AlarmDefinition:         services.Facility.AlarmDefinition,
		ObjectData:              services.Facility.ObjectData,
		SPSControllerSystemType: services.Facility.SPSControllerSystemType,
		Export:                  services.Export,
		AlarmType:               services.Facility.AlarmType,
		Unit:                    services.Facility.Unit,
		AlarmField:              services.Facility.AlarmField,
		AlarmTypeField:          services.Facility.AlarmTypeField,
		BacnetAlarm:             services.Facility.BacnetAlarmValue,
		BacnetReferenceUsage:    services.Facility.BacnetReferenceUsage,
		DeleteImpact:            services.Facility.DeleteImpact,
		CopyJobs:                copyJobs,
		Collaboration:           collaboration,
		ReferenceData:           referenceData,
	})
}

func newUserHandlers(services *Services) *userhandler.Handlers {
	return userhandler.NewHandlers(userhandler.Dependencies{
		Users:                 services.User,
		Admin:                 services.Admin,
		RoleService:           services.RBAC,
		DirectoryService:      services.UserDirectory,
		RegistrationService:   services.UserRegistration,
		PermissionService:     services.RBAC,
		RolePermissionService: services.RBAC,
	})
}
