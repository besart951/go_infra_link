package db

import (
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"gorm.io/gorm"
)

func unifyPermissionModel(db *gorm.DB) error {
	return db.Transaction(func(tx *gorm.DB) error {
		canonicalPermissions := make(map[string]struct{})
		for _, definition := range user.CanonicalPermissionDefinitions() {
			canonicalPermissions[definition.Name] = struct{}{}
			if err := ensureProjectPermissionDefinition(tx, projectPermissionDefinitionFromDomain(definition)); err != nil {
				return err
			}
			if err := ensureProjectRolePermission(tx, user.RoleSuperAdmin, definition.Name); err != nil {
				return err
			}
		}

		if err := migrateEditRoleGrantsToUpdate(tx, canonicalPermissions); err != nil {
			return err
		}
		if err := migrateEditPhaseRulesToUpdate(tx, canonicalPermissions); err != nil {
			return err
		}
		if err := tx.Where("permission LIKE ?", "%.edit").Delete(&user.RolePermission{}).Error; err != nil {
			return err
		}
		return tx.Where("name LIKE ? OR action = ?", "%.edit", "edit").Delete(&user.Permission{}).Error
	})
}

func migrateEditRoleGrantsToUpdate(tx *gorm.DB, canonicalPermissions map[string]struct{}) error {
	var grants []user.RolePermission
	if err := tx.Where("permission LIKE ?", "%.edit").Find(&grants).Error; err != nil {
		return err
	}

	for _, grant := range grants {
		replacement := editPermissionReplacement(grant.Permission)
		if _, ok := canonicalPermissions[replacement]; !ok {
			continue
		}
		if err := ensureProjectRolePermission(tx, grant.Role, replacement); err != nil {
			return err
		}
	}
	return nil
}

func migrateEditPhaseRulesToUpdate(tx *gorm.DB, canonicalPermissions map[string]struct{}) error {
	var rules []project.PhasePermission
	if err := tx.Find(&rules).Error; err != nil {
		return err
	}

	for _, rule := range rules {
		changed := false
		permissions := make([]string, 0, len(rule.Permissions))
		seen := make(map[string]struct{}, len(rule.Permissions))

		for _, permission := range rule.Permissions {
			next := permission
			if strings.HasSuffix(permission, ".edit") {
				changed = true
				replacement := editPermissionReplacement(permission)
				if _, ok := canonicalPermissions[replacement]; !ok {
					continue
				}
				next = replacement
			}
			if !isCanonicalPhaseRulePermission(next, canonicalPermissions) {
				changed = true
				continue
			}
			if _, exists := seen[next]; exists {
				changed = true
				continue
			}
			seen[next] = struct{}{}
			permissions = append(permissions, next)
		}

		if !changed {
			continue
		}
		rule.Permissions = permissions
		if err := tx.Save(&rule).Error; err != nil {
			return err
		}
	}
	return nil
}

func editPermissionReplacement(permission string) string {
	return strings.TrimSuffix(permission, ".edit") + ".update"
}

func isCanonicalPhaseRulePermission(permission string, canonicalPermissions map[string]struct{}) bool {
	if _, ok := canonicalPermissions[permission]; !ok {
		return false
	}
	if !strings.HasPrefix(permission, "project.") {
		return false
	}
	switch permission {
	case user.PermissionProjectCreate, user.PermissionProjectListAll:
		return false
	default:
		return !strings.HasSuffix(permission, ".edit")
	}
}
