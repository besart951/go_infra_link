package user

const (
	PermissionActionCreate  = "create"
	PermissionActionRead    = "read"
	PermissionActionUpdate  = "update"
	PermissionActionDelete  = "delete"
	PermissionActionListAll = "listAll"
	PermissionActionManage  = "manage"
	PermissionActionRestore = "restore"
)

type PermissionDefinition struct {
	Name        string
	Resource    string
	Action      string
	Description string
}

func CanonicalPermissionDefinitions() []PermissionDefinition {
	definitions := make([]PermissionDefinition, 0, 120)
	definitions = append(definitions,
		crudPermissionDefinitions("user", "users", PermissionUserCreate, PermissionUserRead, PermissionUserUpdate, PermissionUserDelete)...,
	)
	definitions = append(definitions, PermissionDefinition{
		Name:        PermissionUserReadDeleted,
		Resource:    "user",
		Action:      "read_deleted",
		Description: "Read deleted users",
	})
	definitions = append(definitions,
		crudPermissionDefinitions("team", "teams", PermissionTeamCreate, PermissionTeamRead, PermissionTeamUpdate, PermissionTeamDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("phase", "phases", PermissionPhaseCreate, PermissionPhaseRead, PermissionPhaseUpdate, PermissionPhaseDelete)...,
	)
	definitions = append(definitions, PermissionDefinition{
		Name:        PermissionPhasePermissionManage,
		Resource:    "phase_permission",
		Action:      PermissionActionManage,
		Description: "Manage phase-based project permission rules",
	})
	definitions = append(definitions,
		crudPermissionDefinitions("role", "roles", PermissionRoleCreate, PermissionRoleRead, PermissionRoleUpdate, PermissionRoleDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("permission", "permissions", PermissionPermissionCreate, PermissionPermissionRead, PermissionPermissionUpdate, PermissionPermissionDelete)...,
	)

	definitions = append(definitions, PermissionDefinition{
		Name:        PermissionProjectCreate,
		Resource:    "project",
		Action:      PermissionActionCreate,
		Description: "Create projects",
	}, PermissionDefinition{
		Name:        PermissionProjectUpdate,
		Resource:    "project",
		Action:      PermissionActionUpdate,
		Description: "Update projects",
	}, PermissionDefinition{
		Name:        PermissionProjectDelete,
		Resource:    "project",
		Action:      PermissionActionDelete,
		Description: "Delete projects",
	}, PermissionDefinition{
		Name:        PermissionProjectListAll,
		Resource:    "project",
		Action:      PermissionActionListAll,
		Description: "List all projects",
	})
	definitions = append(definitions,
		crudPermissionDefinitions("project.controlcabinet", "project control cabinets", PermissionProjectControlCabinetCreate, PermissionProjectControlCabinetRead, PermissionProjectControlCabinetUpdate, PermissionProjectControlCabinetDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("project.spscontroller", "project SPS controllers", PermissionProjectSPSControllerCreate, PermissionProjectSPSControllerRead, PermissionProjectSPSControllerUpdate, PermissionProjectSPSControllerDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("project.spscontroller.systemtype", "project SPS controller system types", PermissionProjectSPSControllerSystemTypeCreate, PermissionProjectSPSControllerSystemTypeRead, PermissionProjectSPSControllerSystemTypeUpdate, PermissionProjectSPSControllerSystemTypeDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("project.fielddevice", "project field devices", PermissionProjectFieldDeviceCreate, PermissionProjectFieldDeviceRead, PermissionProjectFieldDeviceUpdate, PermissionProjectFieldDeviceDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("project.fielddevice_specification", "project field device specifications", PermissionProjectFieldDeviceSpecificationCreate, PermissionProjectFieldDeviceSpecificationRead, PermissionProjectFieldDeviceSpecificationUpdate, PermissionProjectFieldDeviceSpecificationDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("project.fielddevice.bacnetobjects", "project field device BACnet objects", PermissionProjectFieldDeviceBacnetObjectsCreate, PermissionProjectFieldDeviceBacnetObjectsRead, PermissionProjectFieldDeviceBacnetObjectsUpdate, PermissionProjectFieldDeviceBacnetObjectsDelete)...,
	)

	definitions = append(definitions,
		crudPermissionDefinitions("building", "buildings", PermissionBuildingCreate, PermissionBuildingRead, PermissionBuildingUpdate, PermissionBuildingDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("controlcabinet", "control cabinets", PermissionControlCabinetCreate, PermissionControlCabinetRead, PermissionControlCabinetUpdate, PermissionControlCabinetDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("spscontroller", "SPS controllers", PermissionSPSControllerCreate, PermissionSPSControllerRead, PermissionSPSControllerUpdate, PermissionSPSControllerDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("spscontrollersystemtype", "SPS controller system types", PermissionSPSControllerSystemTypeCreate, PermissionSPSControllerSystemTypeRead, PermissionSPSControllerSystemTypeUpdate, PermissionSPSControllerSystemTypeDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("fielddevice", "field devices", PermissionFieldDeviceCreate, PermissionFieldDeviceRead, PermissionFieldDeviceUpdate, PermissionFieldDeviceDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("bacnetobject", "BACnet objects", PermissionBacnetObjectCreate, PermissionBacnetObjectRead, PermissionBacnetObjectUpdate, PermissionBacnetObjectDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("systempart", "system parts", PermissionSystemPartCreate, PermissionSystemPartRead, PermissionSystemPartUpdate, PermissionSystemPartDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("systemtype", "system types", PermissionSystemTypeCreate, PermissionSystemTypeRead, PermissionSystemTypeUpdate, PermissionSystemTypeDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("specification", "specifications", PermissionSpecificationCreate, PermissionSpecificationRead, PermissionSpecificationUpdate, PermissionSpecificationDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("apparat", "apparats", PermissionApparatCreate, PermissionApparatRead, PermissionApparatUpdate, PermissionApparatDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("notificationclass", "notification classes", PermissionNotificationClassCreate, PermissionNotificationClassRead, PermissionNotificationClassUpdate, PermissionNotificationClassDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("statetext", "state texts", PermissionStateTextCreate, PermissionStateTextRead, PermissionStateTextUpdate, PermissionStateTextDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("objectdata", "object data", PermissionObjectDataCreate, PermissionObjectDataRead, PermissionObjectDataUpdate, PermissionObjectDataDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("alarmdefinition", "alarm definitions", PermissionAlarmDefinitionCreate, PermissionAlarmDefinitionRead, PermissionAlarmDefinitionUpdate, PermissionAlarmDefinitionDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("alarmtype", "alarm types", PermissionAlarmTypeCreate, PermissionAlarmTypeRead, PermissionAlarmTypeUpdate, PermissionAlarmTypeDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("alarmfield", "alarm fields", PermissionAlarmFieldCreate, PermissionAlarmFieldRead, PermissionAlarmFieldUpdate, PermissionAlarmFieldDelete)...,
	)
	definitions = append(definitions,
		crudPermissionDefinitions("unit", "units", PermissionUnitCreate, PermissionUnitRead, PermissionUnitUpdate, PermissionUnitDelete)...,
	)

	definitions = append(definitions, PermissionDefinition{
		Name:        PermissionNotificationSMTPManage,
		Resource:    "notification.smtp",
		Action:      PermissionActionManage,
		Description: "Manage SMTP notification settings",
	}, PermissionDefinition{
		Name:        PermissionTimelineRead,
		Resource:    "timeline",
		Action:      PermissionActionRead,
		Description: "Read timeline entries",
	}, PermissionDefinition{
		Name:        PermissionTimelineRestore,
		Resource:    "timeline",
		Action:      PermissionActionRestore,
		Description: "Restore timeline entries",
	})

	return definitions
}

func crudPermissionDefinitions(resource string, label string, createName string, readName string, updateName string, deleteName string) []PermissionDefinition {
	return []PermissionDefinition{
		{
			Name:        createName,
			Resource:    resource,
			Action:      PermissionActionCreate,
			Description: "Create " + label,
		},
		{
			Name:        readName,
			Resource:    resource,
			Action:      PermissionActionRead,
			Description: "Read " + label,
		},
		{
			Name:        updateName,
			Resource:    resource,
			Action:      PermissionActionUpdate,
			Description: "Update " + label,
		},
		{
			Name:        deleteName,
			Resource:    resource,
			Action:      PermissionActionDelete,
			Description: "Delete " + label,
		},
	}
}
