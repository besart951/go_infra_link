package wire

import (
	facilityhandler "github.com/besart951/go_infra_link/backend/internal/handler/facility"
	projecthandler "github.com/besart951/go_infra_link/backend/internal/handler/project"
	projectcontrolcabinethandler "github.com/besart951/go_infra_link/backend/internal/handler/project/controlcabinet"
	projectfielddevicehandler "github.com/besart951/go_infra_link/backend/internal/handler/project/fielddevice"
	projectobjectdatahandler "github.com/besart951/go_infra_link/backend/internal/handler/project/objectdata"
	projectspscontrollerhandler "github.com/besart951/go_infra_link/backend/internal/handler/project/spscontroller"
	userhandler "github.com/besart951/go_infra_link/backend/internal/handler/user"
)

func newProjectHandlers(services *Services, runtime *RuntimeAdapters) *projecthandler.Handlers {
	var controlCabinetCloner projectcontrolcabinethandler.ProjectControlCabinetCloner
	var controlCabinetAssigner projectcontrolcabinethandler.ProjectControlCabinetAssigner
	var controlCabinetReassigner projectcontrolcabinethandler.ProjectControlCabinetReassigner
	var spsControllerCloner projectspscontrollerhandler.ProjectSPSControllerCloner
	var spsControllerSystemTypeCloner projectspscontrollerhandler.ProjectSPSControllerSystemTypeCloner
	var spsControllerAssigner projectspscontrollerhandler.ProjectSPSControllerAssigner
	var spsControllerReassigner projectspscontrollerhandler.ProjectSPSControllerReassigner
	var fieldDeviceMultiCreator projectfielddevicehandler.ProjectFieldDeviceMultiCreator
	var fieldDeviceAssigner projectfielddevicehandler.ProjectFieldDeviceAssigner
	var fieldDeviceBulkAssigner projectfielddevicehandler.ProjectFieldDeviceBulkAssigner
	var fieldDeviceReassigner projectfielddevicehandler.ProjectFieldDeviceReassigner
	var objectDataAttacher projectobjectdatahandler.ProjectObjectDataAttacher
	var objectDataDeactivator projectobjectdatahandler.ProjectObjectDataDeactivator
	if services.FacilityApplication != nil && services.FacilityApplication.ControlCabinet != nil {
		controlCabinetCloner = services.FacilityApplication.ControlCabinet.CloneForProject
		controlCabinetAssigner = services.FacilityApplication.ControlCabinet.AssignToProject
		controlCabinetReassigner = services.FacilityApplication.ControlCabinet.ReassignProjectLink
	}
	if services.FacilityApplication != nil && services.FacilityApplication.SPSController != nil {
		spsControllerCloner = services.FacilityApplication.SPSController.CloneForProject
		spsControllerSystemTypeCloner = services.FacilityApplication.SPSController.CloneSystemTypeForProject
		spsControllerAssigner = services.FacilityApplication.SPSController.AssignToProject
		spsControllerReassigner = services.FacilityApplication.SPSController.ReassignProjectLink
	}
	if services.FacilityApplication != nil && services.FacilityApplication.FieldDevice != nil {
		fieldDeviceMultiCreator = services.FacilityApplication.FieldDevice.MultiCreateForProject
		fieldDeviceAssigner = services.FacilityApplication.FieldDevice.AssignToProject
		fieldDeviceBulkAssigner = services.FacilityApplication.FieldDevice.BulkAssignToProject
		fieldDeviceReassigner = services.FacilityApplication.FieldDevice.ReassignProjectLink
	}
	if services.FacilityApplication != nil && services.FacilityApplication.ObjectData != nil {
		objectDataAttacher = services.FacilityApplication.ObjectData.ProjectAssociation
		objectDataDeactivator = services.FacilityApplication.ObjectData.ProjectAssociation
	}
	return projecthandler.NewHandlers(projecthandler.ServiceDeps{
		Lifecycle:                     services.Project.Lifecycle,
		AccessPolicy:                  services.Project.AccessPolicy,
		Membership:                    services.Project.Membership,
		Workflow:                      services.Project.Workflow,
		FacilityLink:                  services.Project.FacilityLink,
		ControlCabinetCloner:          controlCabinetCloner,
		ControlCabinetAssigner:        controlCabinetAssigner,
		ControlCabinetReassigner:      controlCabinetReassigner,
		SPSControllerCloner:           spsControllerCloner,
		SPSControllerSystemTypeCloner: spsControllerSystemTypeCloner,
		SPSControllerAssigner:         spsControllerAssigner,
		SPSControllerReassigner:       spsControllerReassigner,
		FieldDeviceMultiCreator:       fieldDeviceMultiCreator,
		FieldDeviceAssigner:           fieldDeviceAssigner,
		FieldDeviceBulkAssigner:       fieldDeviceBulkAssigner,
		FieldDeviceReassigner:         fieldDeviceReassigner,
		ObjectDataAttacher:            objectDataAttacher,
		ObjectDataDeactivator:         objectDataDeactivator,
		Phase:                         services.Phase,
		PhasePermission:               services.PhasePermission,
		FieldDeviceOptions:            services.Facility.FieldDevice,
		Notifications:                 services.Notification,
		Collaboration:                 runtime.ProjectCollaboration,
	})
}

func newFacilityHandlers(services *Services) *facilityhandler.Handlers {
	fieldDeviceBulkUpdater := facilityhandler.FieldDeviceBulkUpdater(services.Facility.FieldDevice)
	var controlCabinetCreator facilityhandler.ControlCabinetCreator
	var controlCabinetCloner facilityhandler.ControlCabinetCloner
	var controlCabinetUpdater facilityhandler.ControlCabinetUpdater
	var controlCabinetDeleter facilityhandler.ControlCabinetDeleter
	var bacnetObjectCreator facilityhandler.BacnetObjectCreator
	var bacnetObjectUpdater facilityhandler.BacnetObjectUpdater
	var bacnetAlarmValueReplacer facilityhandler.BacnetAlarmValueReplacer
	var fieldDeviceUpdater facilityhandler.FieldDeviceUpdater
	var fieldDeviceDeleter facilityhandler.FieldDeviceDeleter
	var fieldDeviceMultiCreator facilityhandler.FieldDeviceMultiCreator
	var fieldDeviceBulkDeleter facilityhandler.FieldDeviceBulkDeleter
	var spsControllerCreator facilityhandler.SPSControllerCreator
	var spsControllerCloner facilityhandler.SPSControllerCloner
	var spsControllerUpdater facilityhandler.SPSControllerUpdater
	var spsControllerDeleter facilityhandler.SPSControllerDeleter
	var spsControllerSystemTypeCloner facilityhandler.SPSControllerSystemTypeCloner
	var spsControllerSystemTypeDeleter facilityhandler.SPSControllerSystemTypeDeleter
	if services.FacilityApplication != nil && services.FacilityApplication.FieldDevice != nil {
		fieldDeviceUpdater = services.FacilityApplication.FieldDevice.Update
		fieldDeviceDeleter = services.FacilityApplication.FieldDevice.Delete
		fieldDeviceMultiCreator = services.FacilityApplication.FieldDevice.MultiCreate
		fieldDeviceBulkUpdater = services.FacilityApplication.FieldDevice.BulkUpdate
		fieldDeviceBulkDeleter = services.FacilityApplication.FieldDevice.BulkDelete
	}
	if services.FacilityApplication != nil && services.FacilityApplication.ControlCabinet != nil {
		controlCabinetCreator = services.FacilityApplication.ControlCabinet.Create
		controlCabinetCloner = services.FacilityApplication.ControlCabinet.Clone
		controlCabinetUpdater = services.FacilityApplication.ControlCabinet.Update
		controlCabinetDeleter = services.FacilityApplication.ControlCabinet.Delete
	}
	if services.FacilityApplication != nil && services.FacilityApplication.SPSController != nil {
		spsControllerCreator = services.FacilityApplication.SPSController.Create
		spsControllerCloner = services.FacilityApplication.SPSController.Clone
		spsControllerUpdater = services.FacilityApplication.SPSController.Update
		spsControllerDeleter = services.FacilityApplication.SPSController.Delete
		spsControllerSystemTypeCloner = services.FacilityApplication.SPSController.CloneSystemType
		spsControllerSystemTypeDeleter = services.FacilityApplication.SPSController.DeleteSystemType
	}
	if services.FacilityApplication != nil && services.FacilityApplication.BacnetObject != nil {
		bacnetObjectCreator = services.FacilityApplication.BacnetObject.Create
		bacnetObjectUpdater = services.FacilityApplication.BacnetObject.Update
		bacnetAlarmValueReplacer = services.FacilityApplication.BacnetObject.ReplaceAlarmValues
	}

	return facilityhandler.NewHandlers(facilityhandler.ServiceDeps{
		Building:                       services.Facility.Building,
		SystemType:                     services.Facility.SystemType,
		SystemPart:                     services.Facility.SystemPart,
		Apparat:                        services.Facility.Apparat,
		ControlCabinet:                 services.Facility.ControlCabinet,
		ControlCabinetCreator:          controlCabinetCreator,
		ControlCabinetCloner:           controlCabinetCloner,
		ControlCabinetUpdater:          controlCabinetUpdater,
		ControlCabinetDeleter:          controlCabinetDeleter,
		FieldDevice:                    services.Facility.FieldDevice,
		FieldDeviceMultiCreator:        fieldDeviceMultiCreator,
		FieldDeviceUpdater:             fieldDeviceUpdater,
		FieldDeviceDeleter:             fieldDeviceDeleter,
		FieldDeviceBulkUpdater:         fieldDeviceBulkUpdater,
		FieldDeviceBulkDeleter:         fieldDeviceBulkDeleter,
		BacnetObject:                   services.Facility.BacnetObject,
		BacnetObjectCreator:            bacnetObjectCreator,
		BacnetObjectUpdater:            bacnetObjectUpdater,
		SPSController:                  services.Facility.SPSController,
		SPSControllerCreator:           spsControllerCreator,
		SPSControllerCloner:            spsControllerCloner,
		SPSControllerUpdater:           spsControllerUpdater,
		SPSControllerDeleter:           spsControllerDeleter,
		SPSControllerSystemTypeCloner:  spsControllerSystemTypeCloner,
		SPSControllerSystemTypeDeleter: spsControllerSystemTypeDeleter,
		StateText:                      services.Facility.StateText,
		NotificationClass:              services.Facility.NotificationClass,
		AlarmDefinition:                services.Facility.AlarmDefinition,
		ObjectData:                     services.Facility.ObjectData,
		SPSControllerSystemType:        services.Facility.SPSControllerSystemType,
		Export:                         services.Export,
		AlarmType:                      services.Facility.AlarmType,
		Unit:                           services.Facility.Unit,
		AlarmField:                     services.Facility.AlarmField,
		AlarmTypeField:                 services.Facility.AlarmTypeField,
		BacnetAlarm:                    services.Facility.BacnetAlarmValue,
		BacnetAlarmReplacer:            bacnetAlarmValueReplacer,
		BacnetReferenceUsage:           services.Facility.BacnetReferenceUsage,
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
