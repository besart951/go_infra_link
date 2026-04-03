package facility

import "github.com/gin-gonic/gin"

func RegisterRoutes(protectedV1 *gin.RouterGroup, handlers *Handlers) {
	facility := protectedV1.Group("/facility")
	{
		facility.POST("/buildings/validate", handlers.Validation.ValidateBuilding)
		facility.POST("/buildings", handlers.Building.CreateBuilding)
		facility.POST("/buildings/bulk", handlers.Building.GetBuildingsByIDs)
		facility.GET("/buildings", handlers.Building.ListBuildings)
		facility.GET("/buildings/:id", handlers.Building.GetBuilding)
		facility.PUT("/buildings/:id", handlers.Building.UpdateBuilding)
		facility.DELETE("/buildings/:id", handlers.Building.DeleteBuilding)

		facility.POST("/system-types", handlers.SystemType.CreateSystemType)
		facility.GET("/system-types", handlers.SystemType.ListSystemTypes)
		facility.GET("/system-types/:id", handlers.SystemType.GetSystemType)
		facility.PUT("/system-types/:id", handlers.SystemType.UpdateSystemType)
		facility.DELETE("/system-types/:id", handlers.SystemType.DeleteSystemType)

		facility.POST("/system-parts", handlers.SystemPart.CreateSystemPart)
		facility.GET("/system-parts", handlers.SystemPart.ListSystemParts)
		facility.GET("/system-parts/:id", handlers.SystemPart.GetSystemPart)
		facility.PUT("/system-parts/:id", handlers.SystemPart.UpdateSystemPart)
		facility.DELETE("/system-parts/:id", handlers.SystemPart.DeleteSystemPart)

		facility.POST("/apparats", handlers.Apparat.CreateApparat)
		facility.POST("/apparats/bulk", handlers.Apparat.GetApparatsByIDs)
		facility.GET("/apparats", handlers.Apparat.ListApparats)
		facility.GET("/apparats/:id", handlers.Apparat.GetApparat)
		facility.PUT("/apparats/:id", handlers.Apparat.UpdateApparat)
		facility.DELETE("/apparats/:id", handlers.Apparat.DeleteApparat)

		facility.POST("/control-cabinets/validate", handlers.Validation.ValidateControlCabinet)
		facility.POST("/control-cabinets", handlers.ControlCabinet.CreateControlCabinet)
		facility.POST("/control-cabinets/bulk", handlers.ControlCabinet.GetControlCabinetsByIDs)
		facility.POST("/control-cabinets/:id/copy", handlers.ControlCabinet.CopyControlCabinet)
		facility.GET("/control-cabinets", handlers.ControlCabinet.ListControlCabinets)
		facility.GET("/control-cabinets/:id", handlers.ControlCabinet.GetControlCabinet)
		facility.GET("/control-cabinets/:id/delete-impact", handlers.ControlCabinet.GetControlCabinetDeleteImpact)
		facility.PUT("/control-cabinets/:id", handlers.ControlCabinet.UpdateControlCabinet)
		facility.DELETE("/control-cabinets/:id", handlers.ControlCabinet.DeleteControlCabinet)

		facility.POST("/field-devices/multi-create", handlers.FieldDevice.MultiCreateFieldDevices)
		facility.GET("/field-devices/options", handlers.FieldDevice.GetFieldDeviceOptions)
		facility.GET("/field-devices/available-apparat-nr", handlers.FieldDevice.ListAvailableApparatNumbers)
		facility.GET("/field-devices", handlers.FieldDevice.ListFieldDevices)
		facility.GET("/field-devices/:id", handlers.FieldDevice.GetFieldDevice)
		facility.GET("/field-devices/:id/bacnet-objects", handlers.FieldDevice.ListFieldDeviceBacnetObjects)
		facility.POST("/field-devices/:id/specification", handlers.FieldDevice.CreateFieldDeviceSpecification)
		facility.PUT("/field-devices/:id/specification", handlers.FieldDevice.UpdateFieldDeviceSpecification)
		facility.PUT("/field-devices/:id", handlers.FieldDevice.UpdateFieldDevice)
		facility.DELETE("/field-devices/:id", handlers.FieldDevice.DeleteFieldDevice)
		facility.PATCH("/field-devices/bulk-update", handlers.FieldDevice.BulkUpdateFieldDevices)
		facility.DELETE("/field-devices/bulk-delete", handlers.FieldDevice.BulkDeleteFieldDevices)

		facility.POST("/bacnet-objects", handlers.BacnetObject.CreateBacnetObject)
		facility.PUT("/bacnet-objects/:id", handlers.BacnetObject.UpdateBacnetObject)

		facility.POST("/sps-controllers/validate", handlers.Validation.ValidateSPSController)
		facility.POST("/sps-controllers", handlers.SPSController.CreateSPSController)
		facility.POST("/sps-controllers/bulk", handlers.SPSController.GetSPSControllersByIDs)
		facility.POST("/sps-controllers/:id/copy", handlers.SPSController.CopySPSController)
		facility.GET("/sps-controllers", handlers.SPSController.ListSPSControllers)
		facility.GET("/sps-controllers/next-ga-device", handlers.SPSController.GetNextAvailableGADevice)
		facility.GET("/sps-controllers/:id", handlers.SPSController.GetSPSController)
		facility.PUT("/sps-controllers/:id", handlers.SPSController.UpdateSPSController)
		facility.DELETE("/sps-controllers/:id", handlers.SPSController.DeleteSPSController)

		facility.GET("/state-texts", handlers.StateText.ListStateTexts)
		facility.GET("/state-texts/:id", handlers.StateText.GetStateText)
		facility.POST("/state-texts", handlers.StateText.CreateStateText)
		facility.PUT("/state-texts/:id", handlers.StateText.UpdateStateText)
		facility.DELETE("/state-texts/:id", handlers.StateText.DeleteStateText)

		facility.GET("/notification-classes", handlers.NotificationClass.ListNotificationClasses)
		facility.GET("/notification-classes/:id", handlers.NotificationClass.GetNotificationClass)
		facility.POST("/notification-classes", handlers.NotificationClass.CreateNotificationClass)
		facility.PUT("/notification-classes/:id", handlers.NotificationClass.UpdateNotificationClass)
		facility.DELETE("/notification-classes/:id", handlers.NotificationClass.DeleteNotificationClass)

		facility.GET("/alarm-definitions", handlers.AlarmDefinition.ListAlarmDefinitions)
		facility.GET("/alarm-definitions/:id", handlers.AlarmDefinition.GetAlarmDefinition)
		facility.POST("/alarm-definitions", handlers.AlarmDefinition.CreateAlarmDefinition)
		facility.PUT("/alarm-definitions/:id", handlers.AlarmDefinition.UpdateAlarmDefinition)
		facility.DELETE("/alarm-definitions/:id", handlers.AlarmDefinition.DeleteAlarmDefinition)

		facility.GET("/object-data", handlers.ObjectData.ListObjectData)
		facility.GET("/object-data/:id", handlers.ObjectData.GetObjectData)
		facility.GET("/object-data/:id/bacnet-objects", handlers.ObjectData.GetObjectDataBacnetObjects)
		facility.POST("/object-data", handlers.ObjectData.CreateObjectData)
		facility.PUT("/object-data/:id", handlers.ObjectData.UpdateObjectData)
		facility.DELETE("/object-data/:id", handlers.ObjectData.DeleteObjectData)

		facility.GET("/sps-controller-system-types", handlers.SPSControllerSystemType.ListSPSControllerSystemTypes)
		facility.GET("/sps-controller-system-types/:id", handlers.SPSControllerSystemType.GetSPSControllerSystemType)
		facility.POST("/sps-controller-system-types/:id/copy", handlers.SPSControllerSystemType.CopySPSControllerSystemType)
		facility.DELETE("/sps-controller-system-types/:id", handlers.SPSControllerSystemType.DeleteSPSControllerSystemType)

		facility.GET("/alarm-types", handlers.AlarmType.ListAlarmTypes)
		facility.POST("/alarm-types", handlers.AlarmType.CreateAlarmType)
		facility.GET("/alarm-types/:id", handlers.AlarmType.GetAlarmType)
		facility.PUT("/alarm-types/:id", handlers.AlarmType.UpdateAlarmType)
		facility.DELETE("/alarm-types/:id", handlers.AlarmType.DeleteAlarmType)
		facility.GET("/alarm-types/:id/fields", handlers.AlarmType.GetAlarmTypeFields)
		facility.POST("/alarm-types/:id/fields", handlers.AlarmTypeField.CreateAlarmTypeField)

		facility.PUT("/alarm-type-fields/:id", handlers.AlarmTypeField.UpdateAlarmTypeField)
		facility.DELETE("/alarm-type-fields/:id", handlers.AlarmTypeField.DeleteAlarmTypeField)

		facility.GET("/alarm-units", handlers.Unit.ListUnits)
		facility.GET("/alarm-units/:id", handlers.Unit.GetUnit)
		facility.POST("/alarm-units", handlers.Unit.CreateUnit)
		facility.PUT("/alarm-units/:id", handlers.Unit.UpdateUnit)
		facility.DELETE("/alarm-units/:id", handlers.Unit.DeleteUnit)

		facility.GET("/alarm-fields", handlers.AlarmField.ListAlarmFields)
		facility.GET("/alarm-fields/:id", handlers.AlarmField.GetAlarmField)
		facility.POST("/alarm-fields", handlers.AlarmField.CreateAlarmField)
		facility.PUT("/alarm-fields/:id", handlers.AlarmField.UpdateAlarmField)
		facility.DELETE("/alarm-fields/:id", handlers.AlarmField.DeleteAlarmField)

		facility.GET("/bacnet-objects/:id/alarm-schema", handlers.BacnetAlarm.GetAlarmSchema)
		facility.GET("/bacnet-objects/:id/alarm-values", handlers.BacnetAlarm.GetAlarmValues)
		facility.PUT("/bacnet-objects/:id/alarm-values", handlers.BacnetAlarm.PutAlarmValues)

		facility.POST("/exports/field-devices", handlers.Export.CreateFieldDeviceExport)
		facility.GET("/exports/jobs/:jobId", handlers.Export.GetExportStatus)
		facility.GET("/exports/jobs/:jobId/download", handlers.Export.DownloadExport)
	}
}
