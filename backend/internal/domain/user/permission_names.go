package user

const (
	PermissionUserCreate      = "user.create"
	PermissionUserRead        = "user.read"
	PermissionUserReadDeleted = "user.read_deleted"
	PermissionUserUpdate      = "user.update"
	PermissionUserDelete      = "user.delete"

	PermissionTeamCreate = "team.create"
	PermissionTeamRead   = "team.read"
	PermissionTeamUpdate = "team.update"
	PermissionTeamDelete = "team.delete"

	PermissionPhaseCreate = "phase.create"
	PermissionPhaseRead   = "phase.read"
	PermissionPhaseUpdate = "phase.update"
	PermissionPhaseDelete = "phase.delete"

	PermissionPhasePermissionManage = "phase_permission.manage"

	PermissionRoleCreate = "role.create"
	PermissionRoleRead   = "role.read"
	PermissionRoleUpdate = "role.update"
	PermissionRoleDelete = "role.delete"

	PermissionPermissionCreate = "permission.create"
	PermissionPermissionRead   = "permission.read"
	PermissionPermissionUpdate = "permission.update"
	PermissionPermissionDelete = "permission.delete"

	PermissionProjectCreate  = "project.create"
	PermissionProjectUpdate  = "project.update"
	PermissionProjectDelete  = "project.delete"
	PermissionProjectListAll = "project.listAll"

	PermissionProjectControlCabinetCreate = "project.controlcabinet.create"
	PermissionProjectControlCabinetRead   = "project.controlcabinet.read"
	PermissionProjectControlCabinetUpdate = "project.controlcabinet.update"
	PermissionProjectControlCabinetDelete = "project.controlcabinet.delete"

	PermissionProjectSPSControllerCreate = "project.spscontroller.create"
	PermissionProjectSPSControllerRead   = "project.spscontroller.read"
	PermissionProjectSPSControllerUpdate = "project.spscontroller.update"
	PermissionProjectSPSControllerDelete = "project.spscontroller.delete"

	PermissionProjectSPSControllerSystemTypeCreate = "project.spscontroller.systemtype.create"
	PermissionProjectSPSControllerSystemTypeRead   = "project.spscontroller.systemtype.read"
	PermissionProjectSPSControllerSystemTypeUpdate = "project.spscontroller.systemtype.update"
	PermissionProjectSPSControllerSystemTypeDelete = "project.spscontroller.systemtype.delete"

	PermissionProjectFieldDeviceCreate = "project.fielddevice.create"
	PermissionProjectFieldDeviceRead   = "project.fielddevice.read"
	PermissionProjectFieldDeviceUpdate = "project.fielddevice.update"
	PermissionProjectFieldDeviceDelete = "project.fielddevice.delete"

	PermissionProjectFieldDeviceSpecificationCreate = "project.fielddevice_specification.create"
	PermissionProjectFieldDeviceSpecificationRead   = "project.fielddevice_specification.read"
	PermissionProjectFieldDeviceSpecificationUpdate = "project.fielddevice_specification.update"
	PermissionProjectFieldDeviceSpecificationDelete = "project.fielddevice_specification.delete"

	PermissionProjectFieldDeviceBacnetObjectsCreate = "project.fielddevice.bacnetobjects.create"
	PermissionProjectFieldDeviceBacnetObjectsRead   = "project.fielddevice.bacnetobjects.read"
	PermissionProjectFieldDeviceBacnetObjectsUpdate = "project.fielddevice.bacnetobjects.update"
	PermissionProjectFieldDeviceBacnetObjectsDelete = "project.fielddevice.bacnetobjects.delete"

	PermissionBuildingCreate = "building.create"
	PermissionBuildingRead   = "building.read"
	PermissionBuildingUpdate = "building.update"
	PermissionBuildingDelete = "building.delete"

	PermissionControlCabinetCreate = "controlcabinet.create"
	PermissionControlCabinetRead   = "controlcabinet.read"
	PermissionControlCabinetUpdate = "controlcabinet.update"
	PermissionControlCabinetDelete = "controlcabinet.delete"

	PermissionSPSControllerCreate = "spscontroller.create"
	PermissionSPSControllerRead   = "spscontroller.read"
	PermissionSPSControllerUpdate = "spscontroller.update"
	PermissionSPSControllerDelete = "spscontroller.delete"

	PermissionSPSControllerSystemTypeCreate = "spscontrollersystemtype.create"
	PermissionSPSControllerSystemTypeRead   = "spscontrollersystemtype.read"
	PermissionSPSControllerSystemTypeUpdate = "spscontrollersystemtype.update"
	PermissionSPSControllerSystemTypeDelete = "spscontrollersystemtype.delete"

	PermissionFieldDeviceCreate = "fielddevice.create"
	PermissionFieldDeviceRead   = "fielddevice.read"
	PermissionFieldDeviceUpdate = "fielddevice.update"
	PermissionFieldDeviceDelete = "fielddevice.delete"

	PermissionBacnetObjectCreate = "bacnetobject.create"
	PermissionBacnetObjectRead   = "bacnetobject.read"
	PermissionBacnetObjectUpdate = "bacnetobject.update"
	PermissionBacnetObjectDelete = "bacnetobject.delete"

	PermissionSystemPartCreate = "systempart.create"
	PermissionSystemPartRead   = "systempart.read"
	PermissionSystemPartUpdate = "systempart.update"
	PermissionSystemPartDelete = "systempart.delete"

	PermissionSystemTypeCreate = "systemtype.create"
	PermissionSystemTypeRead   = "systemtype.read"
	PermissionSystemTypeUpdate = "systemtype.update"
	PermissionSystemTypeDelete = "systemtype.delete"

	PermissionSpecificationCreate = "specification.create"
	PermissionSpecificationRead   = "specification.read"
	PermissionSpecificationUpdate = "specification.update"
	PermissionSpecificationDelete = "specification.delete"

	PermissionApparatCreate = "apparat.create"
	PermissionApparatRead   = "apparat.read"
	PermissionApparatUpdate = "apparat.update"
	PermissionApparatDelete = "apparat.delete"

	PermissionNotificationClassCreate = "notificationclass.create"
	PermissionNotificationClassRead   = "notificationclass.read"
	PermissionNotificationClassUpdate = "notificationclass.update"
	PermissionNotificationClassDelete = "notificationclass.delete"

	PermissionStateTextCreate = "statetext.create"
	PermissionStateTextRead   = "statetext.read"
	PermissionStateTextUpdate = "statetext.update"
	PermissionStateTextDelete = "statetext.delete"

	PermissionObjectDataCreate = "objectdata.create"
	PermissionObjectDataRead   = "objectdata.read"
	PermissionObjectDataUpdate = "objectdata.update"
	PermissionObjectDataDelete = "objectdata.delete"

	PermissionAlarmDefinitionCreate = "alarmdefinition.create"
	PermissionAlarmDefinitionRead   = "alarmdefinition.read"
	PermissionAlarmDefinitionUpdate = "alarmdefinition.update"
	PermissionAlarmDefinitionDelete = "alarmdefinition.delete"

	PermissionAlarmTypeCreate = "alarmtype.create"
	PermissionAlarmTypeRead   = "alarmtype.read"
	PermissionAlarmTypeUpdate = "alarmtype.update"
	PermissionAlarmTypeDelete = "alarmtype.delete"

	PermissionAlarmFieldCreate = "alarmfield.create"
	PermissionAlarmFieldRead   = "alarmfield.read"
	PermissionAlarmFieldUpdate = "alarmfield.update"
	PermissionAlarmFieldDelete = "alarmfield.delete"

	PermissionUnitCreate = "unit.create"
	PermissionUnitRead   = "unit.read"
	PermissionUnitUpdate = "unit.update"
	PermissionUnitDelete = "unit.delete"

	PermissionNotificationSMTPManage = "notification.smtp.manage"

	PermissionTimelineRead    = "timeline.read"
	PermissionTimelineRestore = "timeline.restore"
)
