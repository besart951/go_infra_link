package db

import (
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"gorm.io/gorm"
)

func migratePhasePermissions(db *gorm.DB) error {
	legacyReplacements := map[string][]string{
		user.PermissionProjectControlCabinetEdit: {
			user.PermissionProjectControlCabinetCreate,
			user.PermissionProjectControlCabinetRead,
			user.PermissionProjectControlCabinetUpdate,
			user.PermissionProjectControlCabinetDelete,
		},
		user.PermissionProjectSPSControllerEdit: {
			user.PermissionProjectSPSControllerCreate,
			user.PermissionProjectSPSControllerRead,
			user.PermissionProjectSPSControllerUpdate,
			user.PermissionProjectSPSControllerDelete,
			user.PermissionProjectSPSControllerSystemTypeCreate,
			user.PermissionProjectSPSControllerSystemTypeRead,
			user.PermissionProjectSPSControllerSystemTypeUpdate,
			user.PermissionProjectSPSControllerSystemTypeDelete,
		},
		user.PermissionProjectSPSControllerSystemTypeEdit: {
			user.PermissionProjectSPSControllerSystemTypeCreate,
			user.PermissionProjectSPSControllerSystemTypeRead,
			user.PermissionProjectSPSControllerSystemTypeUpdate,
			user.PermissionProjectSPSControllerSystemTypeDelete,
		},
		user.PermissionProjectFieldDeviceEdit: {
			user.PermissionProjectFieldDeviceCreate,
			user.PermissionProjectFieldDeviceRead,
			user.PermissionProjectFieldDeviceUpdate,
			user.PermissionProjectFieldDeviceDelete,
			user.PermissionProjectFieldDeviceSpecificationCreate,
			user.PermissionProjectFieldDeviceSpecificationRead,
			user.PermissionProjectFieldDeviceSpecificationUpdate,
			user.PermissionProjectFieldDeviceSpecificationDelete,
			user.PermissionProjectFieldDeviceBacnetObjectsCreate,
			user.PermissionProjectFieldDeviceBacnetObjectsRead,
			user.PermissionProjectFieldDeviceBacnetObjectsUpdate,
			user.PermissionProjectFieldDeviceBacnetObjectsDelete,
		},
		user.PermissionProjectFieldDeviceSpecificationEdit: {
			user.PermissionProjectFieldDeviceSpecificationCreate,
			user.PermissionProjectFieldDeviceSpecificationRead,
			user.PermissionProjectFieldDeviceSpecificationUpdate,
			user.PermissionProjectFieldDeviceSpecificationDelete,
		},
		user.PermissionProjectFieldDeviceBacnetObjectsEdit: {
			user.PermissionProjectFieldDeviceBacnetObjectsCreate,
			user.PermissionProjectFieldDeviceBacnetObjectsRead,
			user.PermissionProjectFieldDeviceBacnetObjectsUpdate,
			user.PermissionProjectFieldDeviceBacnetObjectsDelete,
		},
	}

	return db.Transaction(func(tx *gorm.DB) error {
		if err := tx.AutoMigrate(&project.PhasePermission{}); err != nil {
			return err
		}

		if err := ensureProjectPermissionDefinition(tx, projectPermissionDefinition{
			name:        user.PermissionPhasePermissionManage,
			resource:    "phase_permission",
			action:      "manage",
			description: "Manage phase-based project permission rules",
		}); err != nil {
			return err
		}
		for _, role := range []user.Role{user.RoleSuperAdmin, user.RoleAdminFZAG} {
			if err := ensureProjectRolePermission(tx, role, user.PermissionPhasePermissionManage); err != nil {
				return err
			}
		}

		legacyPermissions := make([]string, 0, len(legacyReplacements))
		for legacyPermission, replacements := range legacyReplacements {
			legacyPermissions = append(legacyPermissions, legacyPermission)
			var grants []user.RolePermission
			if err := tx.Where("permission = ?", legacyPermission).Find(&grants).Error; err != nil {
				return err
			}
			for _, grant := range grants {
				for _, replacement := range replacements {
					if err := ensureProjectRolePermission(tx, grant.Role, replacement); err != nil {
						return err
					}
				}
			}
		}

		if err := tx.Where("permission IN ?", legacyPermissions).Delete(&user.RolePermission{}).Error; err != nil {
			return err
		}
		return tx.Where("name IN ?", legacyPermissions).Delete(&user.Permission{}).Error
	})
}
