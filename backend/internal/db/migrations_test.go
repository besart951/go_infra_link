package db

import (
	"context"
	"path/filepath"
	"slices"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	projectrepo "github.com/besart951/go_infra_link/backend/internal/repository/project"
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

func TestProjectPhaseMigrationBackfillsAlreadyBaselinedDatabase(t *testing.T) {
	db := openMigrationTestDB(t)
	if err := db.AutoMigrate(&schemaMigration{}); err != nil {
		t.Fatalf("expected migration table setup to succeed, got %v", err)
	}

	for _, migration := range migrations {
		if migration.version == "202605080003" {
			continue
		}
		if err := db.Create(&schemaMigration{
			Version:     migration.version,
			Description: migration.description,
			AppliedAt:   time.Now().UTC(),
		}).Error; err != nil {
			t.Fatalf("expected migration %s to be marked applied, got %v", migration.version, err)
		}
	}

	if err := db.Exec(`
		CREATE TABLE projects (
			id text PRIMARY KEY,
			created_at datetime NOT NULL,
			updated_at datetime NOT NULL,
			name text NOT NULL,
			description text,
			status text NOT NULL,
			start_date datetime,
			creator_id text NOT NULL
		)
	`).Error; err != nil {
		t.Fatalf("expected old projects table setup to succeed, got %v", err)
	}

	projectID := uuid.New()
	creatorID := uuid.New()
	now := time.Now().UTC()
	if err := db.Exec(
		`INSERT INTO projects (id, created_at, updated_at, name, description, status, creator_id) VALUES (?, ?, ?, ?, ?, ?, ?)`,
		projectID.String(), now, now, "Legacy Project", "created before phases", string(domainProject.StatusPlanned), creatorID.String(),
	).Error; err != nil {
		t.Fatalf("expected old project setup to succeed, got %v", err)
	}

	if err := ApplyMigrations(db); err != nil {
		t.Fatalf("expected migrations to backfill project phases, got %v", err)
	}

	if !db.Migrator().HasTable(&domainProject.Phase{}) {
		t.Fatal("expected phases table to be created for already-baselined databases")
	}
	if !db.Migrator().HasColumn(&projectrepo.ProjectRecord{}, "phase_id") {
		t.Fatal("expected projects.phase_id to be added for already-baselined databases")
	}

	repo := projectrepo.NewProjectRepository(db)
	projects, err := repo.GetPaginatedList(context.Background(), domain.PaginationParams{Page: 1, Limit: 10})
	if err != nil {
		t.Fatalf("expected project list to work after migration, got %v", err)
	}
	if len(projects.Items) != 1 {
		t.Fatalf("expected one legacy project, got %d", len(projects.Items))
	}

	expectedPhaseID := uuid.MustParse("019c780c-f7eb-709a-93dc-5e7458cf4466")
	if projects.Items[0].PhaseID != expectedPhaseID {
		t.Fatalf("expected legacy project to be assigned default phase %s, got %s", expectedPhaseID, projects.Items[0].PhaseID)
	}
}

func TestApplyMigrationsEnsuresUserReadDeletedPermission(t *testing.T) {
	db := openMigrationTestDB(t)

	if err := ApplyMigrations(db); err != nil {
		t.Fatalf("expected migrations to succeed, got %v", err)
	}

	var permission domainUser.Permission
	if err := db.Where("name = ?", domainUser.PermissionUserReadDeleted).First(&permission).Error; err != nil {
		t.Fatalf("expected %s permission to exist, got %v", domainUser.PermissionUserReadDeleted, err)
	}
	if permission.Resource != "user" {
		t.Fatalf("expected resource user, got %s", permission.Resource)
	}
	if permission.Action != "read_deleted" {
		t.Fatalf("expected action read_deleted, got %s", permission.Action)
	}

	var grant domainUser.RolePermission
	if err := db.Where("role = ? AND permission = ?", domainUser.RoleSuperAdmin, domainUser.PermissionUserReadDeleted).First(&grant).Error; err != nil {
		t.Fatalf("expected superadmin grant for %s, got %v", domainUser.PermissionUserReadDeleted, err)
	}
}

func openMigrationTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "migration.db")), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatalf("expected sqlite db to open, got %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("expected sqlite db handle, got %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})
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
