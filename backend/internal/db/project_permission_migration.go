package db

import (
	"errors"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"gorm.io/gorm"
)

type projectPermissionDefinition struct {
	name        string
	resource    string
	action      string
	description string
}

func migrateProjectPermissions(db *gorm.DB) error {
	definitions := []projectPermissionDefinition{
		{
			name:        user.PermissionProjectListAll,
			resource:    "project",
			action:      "listAll",
			description: "List all projects",
		},
	}

	grantMigrations := map[string][]string{
		user.PermissionProjectListAll: {"project.read"},
		user.PermissionProjectSPSControllerSystemTypeCreate: {
			"project.spscontrollersystemtype.create",
			"project.systemtype.create",
		},
		user.PermissionProjectSPSControllerSystemTypeRead: {
			"project.spscontrollersystemtype.read",
			"project.systemtype.read",
		},
		user.PermissionProjectSPSControllerSystemTypeUpdate: {
			"project.spscontrollersystemtype.update",
			"project.systemtype.update",
		},
		user.PermissionProjectSPSControllerSystemTypeDelete: {
			"project.spscontrollersystemtype.delete",
			"project.systemtype.delete",
		},
		user.PermissionProjectFieldDeviceBacnetObjectsCreate: {"project.bacnetobject.create"},
		user.PermissionProjectFieldDeviceBacnetObjectsRead:   {"project.bacnetobject.read"},
		user.PermissionProjectFieldDeviceBacnetObjectsUpdate: {"project.bacnetobject.update"},
		user.PermissionProjectFieldDeviceBacnetObjectsDelete: {"project.bacnetobject.delete"},
	}

	obsoletePermissions := []string{
		"project.read",
		"project.spscontrollersystemtype.create",
		"project.spscontrollersystemtype.read",
		"project.spscontrollersystemtype.update",
		"project.spscontrollersystemtype.delete",
		"project.bacnetobject.create",
		"project.bacnetobject.read",
		"project.bacnetobject.update",
		"project.bacnetobject.delete",
		"project.systemtype.create",
		"project.systemtype.read",
		"project.systemtype.update",
		"project.systemtype.delete",
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

		if err := tx.Where("permission IN ?", obsoletePermissions).Delete(&user.RolePermission{}).Error; err != nil {
			return err
		}
		return tx.Where("name IN ?", obsoletePermissions).Delete(&user.Permission{}).Error
	})
}

func ensureProjectPermissionDefinition(tx *gorm.DB, definition projectPermissionDefinition) error {
	var permission user.Permission
	err := tx.Where("name = ?", definition.name).First(&permission).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		now := time.Now().UTC()
		permission = user.Permission{
			Name:        definition.name,
			Resource:    definition.resource,
			Action:      definition.action,
			Description: definition.description,
		}
		if err := permission.InitForCreate(now); err != nil {
			return err
		}
		return tx.Create(&permission).Error
	}
	if err != nil {
		return err
	}

	updates := map[string]any{}
	if permission.Resource != definition.resource {
		updates["resource"] = definition.resource
	}
	if permission.Action != definition.action {
		updates["action"] = definition.action
	}
	if permission.Description != definition.description {
		updates["description"] = definition.description
	}
	if len(updates) == 0 {
		return nil
	}
	return tx.Model(&permission).Updates(updates).Error
}

func projectPermissionDefinitionFromDomain(definition user.PermissionDefinition) projectPermissionDefinition {
	return projectPermissionDefinition{
		name:        definition.Name,
		resource:    definition.Resource,
		action:      definition.Action,
		description: definition.Description,
	}
}

func migrateProjectPermissionGrants(tx *gorm.DB, targetPermission string, sourcePermissions []string) error {
	var sourceGrants []user.RolePermission
	if err := tx.Where("permission IN ?", sourcePermissions).Find(&sourceGrants).Error; err != nil {
		return err
	}

	roles := make(map[user.Role]struct{}, len(sourceGrants))
	for _, grant := range sourceGrants {
		roles[grant.Role] = struct{}{}
	}

	for role := range roles {
		if err := ensureProjectRolePermission(tx, role, targetPermission); err != nil {
			return err
		}
	}
	return nil
}

func ensureProjectRolePermission(tx *gorm.DB, role user.Role, permission string) error {
	var existing user.RolePermission
	err := tx.Where("role = ? AND permission = ?", role, permission).First(&existing).Error
	if err == nil {
		return nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return err
	}

	now := time.Now().UTC()
	grant := user.RolePermission{Role: role, Permission: permission}
	if err := grant.InitForCreate(now); err != nil {
		return err
	}
	return tx.Create(&grant).Error
}
