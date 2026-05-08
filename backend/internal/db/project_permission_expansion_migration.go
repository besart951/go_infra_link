package db

import (
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"gorm.io/gorm"
)

func expandProjectSubresourcePermissions(db *gorm.DB) error {
	projectResources := map[string]struct{}{
		"project.controlcabinet":            {},
		"project.spscontroller":             {},
		"project.spscontroller.systemtype":  {},
		"project.fielddevice":               {},
		"project.fielddevice_specification": {},
		"project.fielddevice.bacnetobjects": {},
	}

	return db.Transaction(func(tx *gorm.DB) error {
		for _, definition := range user.CanonicalPermissionDefinitions() {
			if _, ok := projectResources[definition.Resource]; !ok {
				continue
			}
			if err := ensureProjectPermissionDefinition(tx, projectPermissionDefinitionFromDomain(definition)); err != nil {
				return err
			}
		}
		return nil
	})
}
