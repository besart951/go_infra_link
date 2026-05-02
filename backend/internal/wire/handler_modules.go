package wire

import (
	facilityhandler "github.com/besart951/go_infra_link/backend/internal/handler/facility"
	projecthandler "github.com/besart951/go_infra_link/backend/internal/handler/project"
	userhandler "github.com/besart951/go_infra_link/backend/internal/handler/user"
)

func newProjectHandlers(services *Services, runtime *RuntimeAdapters) *projecthandler.Handlers {
	return projecthandler.NewHandlers(projecthandler.ServiceDeps{
		Lifecycle:          services.Project.Lifecycle,
		AccessPolicy:       services.Project.AccessPolicy,
		Membership:         services.Project.Membership,
		Workflow:           services.Project.Workflow,
		FacilityLink:       services.Project.FacilityLink,
		Phase:              services.Phase,
		PhasePermission:    services.PhasePermission,
		FieldDeviceOptions: services.Facility.FieldDevice,
		Notifications:      services.Notification,
		Collaboration:      runtime.ProjectCollaboration,
	})
}

func newFacilityHandlers(services *Services, collaboration facilityhandler.ProjectRefreshBroadcaster) *facilityhandler.Handlers {
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
		Collaboration:           collaboration,
	})
}

func newUserHandlers(services *Services) *userhandler.Handlers {
	return userhandler.NewHandlers(
		services.User,
		services.Admin,
		services.RBAC,
		services.UserDirectory,
		services.RBAC,
		services.RBAC,
	)
}
