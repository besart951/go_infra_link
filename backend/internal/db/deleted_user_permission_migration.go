package db

import (
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"gorm.io/gorm"
)

func ensureDeletedUserReadPermission(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		for _, definition := range user.CanonicalPermissionDefinitions() {
			if err := ensureProjectPermissionDefinition(tx, projectPermissionDefinitionFromDomain(definition)); err != nil {
				return err
			}
			if err := ensureProjectRolePermission(tx, user.RoleSuperAdmin, definition.Name); err != nil {
				return err
			}
		}
		return nil
	})
}
