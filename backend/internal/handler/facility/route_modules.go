package facility

import (
	"github.com/besart951/go_infra_link/backend/internal/handler/facility/alarm"
	"github.com/besart951/go_infra_link/backend/internal/handler/facility/fielddevice"
	"github.com/besart951/go_infra_link/backend/internal/handler/facility/hierarchy"
	"github.com/besart951/go_infra_link/backend/internal/handler/facility/objectdata"
	"github.com/besart951/go_infra_link/backend/internal/handler/facility/reference"
)

func hierarchyHandlers(handlers *Handlers) hierarchy.Handlers {
	return hierarchy.Handlers{
		ValidateBuilding:              handlers.Validation.ValidateBuilding,
		CreateBuilding:                handlers.Building.CreateBuilding,
		GetBuildingsByIDs:             handlers.Building.GetBuildingsByIDs,
		ListBuildings:                 handlers.Building.ListBuildings,
		GetBuilding:                   handlers.Building.GetBuilding,
		UpdateBuilding:                handlers.Building.UpdateBuilding,
		DeleteBuilding:                handlers.Building.DeleteBuilding,
		ValidateControlCabinet:        handlers.Validation.ValidateControlCabinet,
		CreateControlCabinet:          handlers.ControlCabinet.CreateControlCabinet,
		GetControlCabinetsByIDs:       handlers.ControlCabinet.GetControlCabinetsByIDs,
		CopyControlCabinet:            handlers.ControlCabinet.CopyControlCabinet,
		ListControlCabinets:           handlers.ControlCabinet.ListControlCabinets,
		GetControlCabinet:             handlers.ControlCabinet.GetControlCabinet,
		GetControlCabinetDeleteImpact: handlers.ControlCabinet.GetControlCabinetDeleteImpact,
		UpdateControlCabinet:          handlers.ControlCabinet.UpdateControlCabinet,
		DeleteControlCabinet:          handlers.ControlCabinet.DeleteControlCabinet,
		ValidateSPSController:         handlers.Validation.ValidateSPSController,
		CreateSPSController:           handlers.SPSController.CreateSPSController,
		GetSPSControllersByIDs:        handlers.SPSController.GetSPSControllersByIDs,
		CopySPSController:             handlers.SPSController.CopySPSController,
		ListSPSControllers:            handlers.SPSController.ListSPSControllers,
		GetNextAvailableGADevice:      handlers.SPSController.GetNextAvailableGADevice,
		GetSPSController:              handlers.SPSController.GetSPSController,
		UpdateSPSController:           handlers.SPSController.UpdateSPSController,
		DeleteSPSController:           handlers.SPSController.DeleteSPSController,
		CreateSPSControllerSystemType: handlers.SPSControllerSystemType.CreateSPSControllerSystemType,
		ListSPSControllerSystemTypes:  handlers.SPSControllerSystemType.ListSPSControllerSystemTypes,
		GetSPSControllerSystemType:    handlers.SPSControllerSystemType.GetSPSControllerSystemType,
		UpdateSPSControllerSystemType: handlers.SPSControllerSystemType.UpdateSPSControllerSystemType,
		CopySPSControllerSystemType:   handlers.SPSControllerSystemType.CopySPSControllerSystemType,
		DeleteSPSControllerSystemType: handlers.SPSControllerSystemType.DeleteSPSControllerSystemType,
	}
}

func referenceHandlers(handlers *Handlers) reference.Handlers {
	return reference.Handlers{
		CreateSystemType:         handlers.SystemType.CreateSystemType,
		ListSystemTypes:          handlers.SystemType.ListSystemTypes,
		GetSystemType:            handlers.SystemType.GetSystemType,
		UpdateSystemType:         handlers.SystemType.UpdateSystemType,
		DeleteSystemType:         handlers.SystemType.DeleteSystemType,
		CreateSystemPart:         handlers.SystemPart.CreateSystemPart,
		ListSystemParts:          handlers.SystemPart.ListSystemParts,
		GetSystemPart:            handlers.SystemPart.GetSystemPart,
		UpdateSystemPart:         handlers.SystemPart.UpdateSystemPart,
		DeleteSystemPart:         handlers.SystemPart.DeleteSystemPart,
		CreateApparat:            handlers.Apparat.CreateApparat,
		GetApparatsByIDs:         handlers.Apparat.GetApparatsByIDs,
		ListApparats:             handlers.Apparat.ListApparats,
		GetApparat:               handlers.Apparat.GetApparat,
		UpdateApparat:            handlers.Apparat.UpdateApparat,
		DeleteApparat:            handlers.Apparat.DeleteApparat,
		ListStateTexts:           handlers.StateText.ListStateTexts,
		GetStateText:             handlers.StateText.GetStateText,
		CreateStateText:          handlers.StateText.CreateStateText,
		UpdateStateText:          handlers.StateText.UpdateStateText,
		DeleteStateText:          handlers.StateText.DeleteStateText,
		ListNotificationClasses:  handlers.NotificationClass.ListNotificationClasses,
		GetNotificationClass:     handlers.NotificationClass.GetNotificationClass,
		CreateNotificationClass:  handlers.NotificationClass.CreateNotificationClass,
		UpdateNotificationClass:  handlers.NotificationClass.UpdateNotificationClass,
		DeleteNotificationClass:  handlers.NotificationClass.DeleteNotificationClass,
		GetBacnetReferenceUsages: handlers.BacnetReferenceUsage.GetBacnetReferenceUsages,
	}
}

func fieldDeviceHandlers(handlers *Handlers) fielddevice.Handlers {
	return fielddevice.Handlers{
		MultiCreateFieldDevices:        handlers.FieldDevice.MultiCreateFieldDevices,
		GetFieldDeviceOptions:          handlers.FieldDevice.GetFieldDeviceOptions,
		ListAvailableApparatNumbers:    handlers.FieldDevice.ListAvailableApparatNumbers,
		ListFieldDevices:               handlers.FieldDevice.ListFieldDevices,
		GetFieldDevice:                 handlers.FieldDevice.GetFieldDevice,
		CopyFieldDevice:                handlers.FieldDevice.CopyFieldDevice,
		ListFieldDeviceBacnetObjects:   handlers.FieldDevice.ListFieldDeviceBacnetObjects,
		CreateFieldDeviceSpecification: handlers.FieldDevice.CreateFieldDeviceSpecification,
		GetFieldDeviceSpecification:    handlers.FieldDevice.GetFieldDeviceSpecification,
		UpdateFieldDeviceSpecification: handlers.FieldDevice.UpdateFieldDeviceSpecification,
		DeleteFieldDeviceSpecification: handlers.FieldDevice.DeleteFieldDeviceSpecification,
		UpdateFieldDevice:              handlers.FieldDevice.UpdateFieldDevice,
		DeleteFieldDevice:              handlers.FieldDevice.DeleteFieldDevice,
		BulkUpdateFieldDevices:         handlers.FieldDevice.BulkUpdateFieldDevices,
		BulkDeleteFieldDevices:         handlers.FieldDevice.BulkDeleteFieldDevices,
		CreateFieldDeviceExport:        handlers.Export.CreateFieldDeviceExport,
		GetExportStatus:                handlers.Export.GetExportStatus,
		DownloadExport:                 handlers.Export.DownloadExport,
		ImportFieldDevices:             handlers.Import.ImportFieldDevices,
	}
}

func objectDataHandlers(handlers *Handlers) objectdata.Handlers {
	return objectdata.Handlers{
		CreateBacnetObject:         handlers.BacnetObject.CreateBacnetObject,
		GetBacnetObject:            handlers.BacnetObject.GetBacnetObject,
		UpdateBacnetObject:         handlers.BacnetObject.UpdateBacnetObject,
		DeleteBacnetObject:         handlers.BacnetObject.DeleteBacnetObject,
		ListObjectData:             handlers.ObjectData.ListObjectData,
		GetObjectData:              handlers.ObjectData.GetObjectData,
		CopyObjectData:             handlers.ObjectData.CopyObjectData,
		GetObjectDataBacnetObjects: handlers.ObjectData.GetObjectDataBacnetObjects,
		CreateObjectData:           handlers.ObjectData.CreateObjectData,
		UpdateObjectData:           handlers.ObjectData.UpdateObjectData,
		DeleteObjectData:           handlers.ObjectData.DeleteObjectData,
	}
}

func alarmHandlers(handlers *Handlers) alarm.Handlers {
	return alarm.Handlers{
		ListAlarmDefinitions:  handlers.AlarmDefinition.ListAlarmDefinitions,
		GetAlarmDefinition:    handlers.AlarmDefinition.GetAlarmDefinition,
		CreateAlarmDefinition: handlers.AlarmDefinition.CreateAlarmDefinition,
		UpdateAlarmDefinition: handlers.AlarmDefinition.UpdateAlarmDefinition,
		DeleteAlarmDefinition: handlers.AlarmDefinition.DeleteAlarmDefinition,
		ListAlarmTypes:        handlers.AlarmType.ListAlarmTypes,
		CreateAlarmType:       handlers.AlarmType.CreateAlarmType,
		GetAlarmType:          handlers.AlarmType.GetAlarmType,
		UpdateAlarmType:       handlers.AlarmType.UpdateAlarmType,
		DeleteAlarmType:       handlers.AlarmType.DeleteAlarmType,
		GetAlarmTypeFields:    handlers.AlarmType.GetAlarmTypeFields,
		CreateAlarmTypeField:  handlers.AlarmTypeField.CreateAlarmTypeField,
		UpdateAlarmTypeField:  handlers.AlarmTypeField.UpdateAlarmTypeField,
		DeleteAlarmTypeField:  handlers.AlarmTypeField.DeleteAlarmTypeField,
		ListUnits:             handlers.Unit.ListUnits,
		GetUnit:               handlers.Unit.GetUnit,
		CreateUnit:            handlers.Unit.CreateUnit,
		UpdateUnit:            handlers.Unit.UpdateUnit,
		DeleteUnit:            handlers.Unit.DeleteUnit,
		ListAlarmFields:       handlers.AlarmField.ListAlarmFields,
		GetAlarmField:         handlers.AlarmField.GetAlarmField,
		CreateAlarmField:      handlers.AlarmField.CreateAlarmField,
		UpdateAlarmField:      handlers.AlarmField.UpdateAlarmField,
		DeleteAlarmField:      handlers.AlarmField.DeleteAlarmField,
		GetAlarmSchema:        handlers.BacnetAlarm.GetAlarmSchema,
		GetAlarmValues:        handlers.BacnetAlarm.GetAlarmValues,
		PutAlarmValues:        handlers.BacnetAlarm.PutAlarmValues,
	}
}
