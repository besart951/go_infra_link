package db

import (
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"gorm.io/gorm"
)

func expandProjectSubresourcePermissions(db *gorm.DB) error {
	definitions := []projectPermissionDefinition{
		{
			name:        user.PermissionProjectControlCabinetCreate,
			resource:    "project.controlcabinet",
			action:      "create",
			description: "Create project control cabinets",
		},
		{
			name:        user.PermissionProjectControlCabinetRead,
			resource:    "project.controlcabinet",
			action:      "read",
			description: "Read project control cabinets",
		},
		{
			name:        user.PermissionProjectControlCabinetUpdate,
			resource:    "project.controlcabinet",
			action:      "update",
			description: "Update project control cabinets",
		},
		{
			name:        user.PermissionProjectControlCabinetDelete,
			resource:    "project.controlcabinet",
			action:      "delete",
			description: "Delete project control cabinets",
		},
		{
			name:        user.PermissionProjectControlCabinetEdit,
			resource:    "project.controlcabinet",
			action:      "edit",
			description: "Edit project control cabinets",
		},
		{
			name:        user.PermissionProjectSPSControllerCreate,
			resource:    "project.spscontroller",
			action:      "create",
			description: "Create project SPS controllers",
		},
		{
			name:        user.PermissionProjectSPSControllerRead,
			resource:    "project.spscontroller",
			action:      "read",
			description: "Read project SPS controllers",
		},
		{
			name:        user.PermissionProjectSPSControllerUpdate,
			resource:    "project.spscontroller",
			action:      "update",
			description: "Update project SPS controllers",
		},
		{
			name:        user.PermissionProjectSPSControllerDelete,
			resource:    "project.spscontroller",
			action:      "delete",
			description: "Delete project SPS controllers",
		},
		{
			name:        user.PermissionProjectSPSControllerEdit,
			resource:    "project.spscontroller",
			action:      "edit",
			description: "Edit project SPS controllers",
		},
		{
			name:        user.PermissionProjectSPSControllerSystemTypeCreate,
			resource:    "project.spscontroller.systemtype",
			action:      "create",
			description: "Create project SPS controller system types",
		},
		{
			name:        user.PermissionProjectSPSControllerSystemTypeRead,
			resource:    "project.spscontroller.systemtype",
			action:      "read",
			description: "Read project SPS controller system types",
		},
		{
			name:        user.PermissionProjectSPSControllerSystemTypeUpdate,
			resource:    "project.spscontroller.systemtype",
			action:      "update",
			description: "Update project SPS controller system types",
		},
		{
			name:        user.PermissionProjectSPSControllerSystemTypeDelete,
			resource:    "project.spscontroller.systemtype",
			action:      "delete",
			description: "Delete project SPS controller system types",
		},
		{
			name:        user.PermissionProjectSPSControllerSystemTypeEdit,
			resource:    "project.spscontroller.systemtype",
			action:      "edit",
			description: "Edit project SPS controller system types",
		},
		{
			name:        user.PermissionProjectFieldDeviceCreate,
			resource:    "project.fielddevice",
			action:      "create",
			description: "Create project field devices",
		},
		{
			name:        user.PermissionProjectFieldDeviceRead,
			resource:    "project.fielddevice",
			action:      "read",
			description: "Read project field devices",
		},
		{
			name:        user.PermissionProjectFieldDeviceUpdate,
			resource:    "project.fielddevice",
			action:      "update",
			description: "Update project field devices",
		},
		{
			name:        user.PermissionProjectFieldDeviceDelete,
			resource:    "project.fielddevice",
			action:      "delete",
			description: "Delete project field devices",
		},
		{
			name:        user.PermissionProjectFieldDeviceEdit,
			resource:    "project.fielddevice",
			action:      "edit",
			description: "Edit project field devices",
		},
		{
			name:        user.PermissionProjectFieldDeviceSpecificationCreate,
			resource:    "project.fielddevice_specification",
			action:      "create",
			description: "Create project field device specifications",
		},
		{
			name:        user.PermissionProjectFieldDeviceSpecificationRead,
			resource:    "project.fielddevice_specification",
			action:      "read",
			description: "Read project field device specifications",
		},
		{
			name:        user.PermissionProjectFieldDeviceSpecificationUpdate,
			resource:    "project.fielddevice_specification",
			action:      "update",
			description: "Update project field device specifications",
		},
		{
			name:        user.PermissionProjectFieldDeviceSpecificationDelete,
			resource:    "project.fielddevice_specification",
			action:      "delete",
			description: "Delete project field device specifications",
		},
		{
			name:        user.PermissionProjectFieldDeviceSpecificationEdit,
			resource:    "project.fielddevice_specification",
			action:      "edit",
			description: "Edit project field device specifications",
		},
		{
			name:        user.PermissionProjectFieldDeviceBacnetObjectsCreate,
			resource:    "project.fielddevice.bacnetobjects",
			action:      "create",
			description: "Create project field device BACnet objects",
		},
		{
			name:        user.PermissionProjectFieldDeviceBacnetObjectsRead,
			resource:    "project.fielddevice.bacnetobjects",
			action:      "read",
			description: "Read project field device BACnet objects",
		},
		{
			name:        user.PermissionProjectFieldDeviceBacnetObjectsUpdate,
			resource:    "project.fielddevice.bacnetobjects",
			action:      "update",
			description: "Update project field device BACnet objects",
		},
		{
			name:        user.PermissionProjectFieldDeviceBacnetObjectsDelete,
			resource:    "project.fielddevice.bacnetobjects",
			action:      "delete",
			description: "Delete project field device BACnet objects",
		},
		{
			name:        user.PermissionProjectFieldDeviceBacnetObjectsEdit,
			resource:    "project.fielddevice.bacnetobjects",
			action:      "edit",
			description: "Edit project field device BACnet objects",
		},
	}

	grantMigrations := map[string][]string{
		user.PermissionProjectControlCabinetCreate: {user.PermissionProjectControlCabinetEdit},
		user.PermissionProjectControlCabinetRead:   {user.PermissionProjectControlCabinetEdit},
		user.PermissionProjectControlCabinetUpdate: {user.PermissionProjectControlCabinetEdit},
		user.PermissionProjectControlCabinetDelete: {user.PermissionProjectControlCabinetEdit},

		user.PermissionProjectSPSControllerCreate: {user.PermissionProjectSPSControllerEdit},
		user.PermissionProjectSPSControllerRead:   {user.PermissionProjectSPSControllerEdit},
		user.PermissionProjectSPSControllerUpdate: {user.PermissionProjectSPSControllerEdit},
		user.PermissionProjectSPSControllerDelete: {user.PermissionProjectSPSControllerEdit},

		user.PermissionProjectSPSControllerSystemTypeCreate: {user.PermissionProjectSPSControllerEdit},
		user.PermissionProjectSPSControllerSystemTypeRead:   {user.PermissionProjectSPSControllerEdit},
		user.PermissionProjectSPSControllerSystemTypeUpdate: {user.PermissionProjectSPSControllerEdit},
		user.PermissionProjectSPSControllerSystemTypeDelete: {user.PermissionProjectSPSControllerEdit},
		user.PermissionProjectSPSControllerSystemTypeEdit:   {user.PermissionProjectSPSControllerEdit},

		user.PermissionProjectFieldDeviceCreate: {user.PermissionProjectFieldDeviceEdit},
		user.PermissionProjectFieldDeviceRead:   {user.PermissionProjectFieldDeviceEdit},
		user.PermissionProjectFieldDeviceUpdate: {user.PermissionProjectFieldDeviceEdit},
		user.PermissionProjectFieldDeviceDelete: {user.PermissionProjectFieldDeviceEdit},

		user.PermissionProjectFieldDeviceSpecificationCreate: {user.PermissionProjectFieldDeviceEdit},
		user.PermissionProjectFieldDeviceSpecificationRead:   {user.PermissionProjectFieldDeviceEdit},
		user.PermissionProjectFieldDeviceSpecificationUpdate: {user.PermissionProjectFieldDeviceEdit},
		user.PermissionProjectFieldDeviceSpecificationDelete: {user.PermissionProjectFieldDeviceEdit},
		user.PermissionProjectFieldDeviceSpecificationEdit:   {user.PermissionProjectFieldDeviceEdit},

		user.PermissionProjectFieldDeviceBacnetObjectsCreate: {user.PermissionProjectFieldDeviceEdit},
		user.PermissionProjectFieldDeviceBacnetObjectsRead:   {user.PermissionProjectFieldDeviceEdit},
		user.PermissionProjectFieldDeviceBacnetObjectsUpdate: {user.PermissionProjectFieldDeviceEdit},
		user.PermissionProjectFieldDeviceBacnetObjectsDelete: {user.PermissionProjectFieldDeviceEdit},
		user.PermissionProjectFieldDeviceBacnetObjectsEdit:   {user.PermissionProjectFieldDeviceEdit},
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, definition := range definitions {
			if err := ensureProjectPermissionDefinition(tx, definition); err != nil {
				return err
			}
		}

		for targetPermission, sourcePermissions := range grantMigrations {
			if err := migrateProjectPermissionGrants(tx, targetPermission, sourcePermissions); err != nil {
				return err
			}
		}

		return nil
	})
}
