package db

import (
	"errors"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"gorm.io/gorm"
)

type timelinePermissionDefinition struct {
	name        string
	action      string
	description string
	grantRoles  []user.Role
}

func ensureTimelinePermissions(db *gorm.DB) error {
	definitions := []timelinePermissionDefinition{
		{
			name:        user.PermissionTimelineRead,
			action:      "read",
			description: "View the change timeline",
			grantRoles:  []user.Role{user.RoleSuperAdmin},
		},
		{
			name:        user.PermissionTimelineRestore,
			action:      "restore",
			description: "Restore entities from the change timeline",
			grantRoles:  []user.Role{user.RoleSuperAdmin},
		},
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, definition := range definitions {
			if err := ensureTimelinePermissionDefinition(tx, definition); err != nil {
				return err
			}
			for _, role := range definition.grantRoles {
				if err := ensureTimelineRolePermission(tx, role, definition.name); err != nil {
					return err
				}
			}
		}
		return nil
	})
}

func ensureTimelinePermissionDefinition(tx *gorm.DB, definition timelinePermissionDefinition) error {
	var permission user.Permission
	err := tx.Where("name = ?", definition.name).First(&permission).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		now := time.Now().UTC()
		permission = user.Permission{
			Name:        definition.name,
			Resource:    "timeline",
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
	if permission.Resource != "timeline" {
		updates["resource"] = "timeline"
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

func ensureTimelineRolePermission(tx *gorm.DB, role user.Role, permission string) error {
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
