package db

import (
	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"gorm.io/gorm"
)

func migratePhasePermissions(db *gorm.DB) error {
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

		return ensureProjectRolePermission(tx, user.RoleSuperAdmin, user.PermissionPhasePermissionManage)
	})
}
