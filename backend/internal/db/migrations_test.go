package db

import (
	"slices"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestProductionMigrationsAreBlueGreenCompatible(t *testing.T) {
	for _, migration := range migrations {
		if !migration.blueGreenCompatible {
			t.Fatalf(
				"migration %s (%s) is not blue-green compatible; use a maintenance migration path instead",
				migration.version,
				migration.description,
			)
		}
	}
}

func TestUnifyPermissionModelMigratesRemovedEditPermissions(t *testing.T) {
	db := openMigrationTestDB(t)
	if err := db.AutoMigrate(&domainUser.Permission{}, &domainUser.RolePermission{}, &domainProject.PhasePermission{}); err != nil {
		t.Fatalf("expected schema migration to succeed, got %v", err)
	}

	createPermission(t, db, "project.fielddevice.edit", "project.fielddevice", "edit")
	createRolePermission(t, db, domainUser.RoleAdminPlaner, "project.fielddevice.edit")
	createRolePermission(t, db, domainUser.RoleSuperAdmin, "project.fielddevice.edit")

	phaseID := uuid.New()
	rule := domainProject.PhasePermission{
		Base:        domain.Base{ID: uuid.New(), CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()},
		PhaseID:     phaseID,
		Role:        domainUser.RoleAdminPlaner,
		Permissions: []string{"project.fielddevice.edit", domainUser.PermissionProjectListAll, domainUser.PermissionProjectFieldDeviceUpdate},
	}
	if err := db.Create(&rule).Error; err != nil {
		t.Fatalf("expected phase rule setup to succeed, got %v", err)
	}

	if err := unifyPermissionModel(db); err != nil {
		t.Fatalf("expected permission unification migration to succeed, got %v", err)
	}

	var removedPermissionCount int64
	if err := db.Model(&domainUser.Permission{}).
		Where("name LIKE ? OR action = ?", "%.edit", "edit").
		Count(&removedPermissionCount).Error; err != nil {
		t.Fatalf("expected removed permission count to succeed, got %v", err)
	}
	if removedPermissionCount != 0 {
		t.Fatalf("expected removed edit permission rows to be deleted, got %d", removedPermissionCount)
	}

	adminPermissions := rolePermissions(t, db, domainUser.RoleAdminPlaner)
	if !slices.Contains(adminPermissions, domainUser.PermissionProjectFieldDeviceUpdate) {
		t.Fatalf("expected removed admin edit grant to migrate to update, got %v", adminPermissions)
	}
	if slices.Contains(adminPermissions, "project.fielddevice.edit") {
		t.Fatalf("expected removed admin edit grant to be removed, got %v", adminPermissions)
	}

	superadminPermissions := rolePermissions(t, db, domainUser.RoleSuperAdmin)
	if !slices.Contains(superadminPermissions, domainUser.PermissionAlarmTypeCreate) ||
		!slices.Contains(superadminPermissions, domainUser.PermissionUnitRead) ||
		!slices.Contains(superadminPermissions, domainUser.PermissionProjectListAll) {
		t.Fatalf("expected superadmin to be synced to canonical permissions, got %v", superadminPermissions)
	}

	var migratedRule domainProject.PhasePermission
	if err := db.First(&migratedRule, "id = ?", rule.ID).Error; err != nil {
		t.Fatalf("expected migrated phase rule to load, got %v", err)
	}
	if len(migratedRule.Permissions) != 1 || migratedRule.Permissions[0] != domainUser.PermissionProjectFieldDeviceUpdate {
		t.Fatalf("expected phase rule to keep only canonical phase-scoped update permission, got %v", migratedRule.Permissions)
	}
}

func openMigrationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatalf("expected sqlite db to open, got %v", err)
	}
	return db
}

func createPermission(t *testing.T, db *gorm.DB, name string, resource string, action string) {
	t.Helper()
	now := time.Now().UTC()
	permission := domainUser.Permission{
		Name:        name,
		Resource:    resource,
		Action:      action,
		Description: name,
	}
	if err := permission.InitForCreate(now); err != nil {
		t.Fatalf("expected permission init to succeed, got %v", err)
	}
	if err := db.Create(&permission).Error; err != nil {
		t.Fatalf("expected permission setup to succeed, got %v", err)
	}
}

func createRolePermission(t *testing.T, db *gorm.DB, role domainUser.Role, permission string) {
	t.Helper()
	now := time.Now().UTC()
	grant := domainUser.RolePermission{Role: role, Permission: permission}
	if err := grant.InitForCreate(now); err != nil {
		t.Fatalf("expected role permission init to succeed, got %v", err)
	}
	if err := db.Create(&grant).Error; err != nil {
		t.Fatalf("expected role permission setup to succeed, got %v", err)
	}
}

func rolePermissions(t *testing.T, db *gorm.DB, role domainUser.Role) []string {
	t.Helper()
	var rows []domainUser.RolePermission
	if err := db.Where("role = ?", role).Find(&rows).Error; err != nil {
		t.Fatalf("expected role permissions to load, got %v", err)
	}
	permissions := make([]string, len(rows))
	for i, row := range rows {
		permissions[i] = row.Permission
	}
	return permissions
}
