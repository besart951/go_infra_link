package project

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPhasePermissionRepo_CreatesRuleWithEmptyPermissions(t *testing.T) {
	ctx := context.Background()
	db := newPhasePermissionRepoTestDB(t)
	repo := NewPhasePermissionRepository(db)

	rule := &domainProject.PhasePermission{
		PhaseID:     uuid.New(),
		Role:        domainUser.RoleAdminFZAG,
		Permissions: []string{},
	}

	if err := repo.Create(ctx, rule); err != nil {
		t.Fatalf("expected empty phase permission rule to be created, got %v", err)
	}

	got, err := repo.GetByPhaseAndRole(ctx, rule.PhaseID, rule.Role)
	if err != nil {
		t.Fatalf("expected empty phase permission rule lookup to succeed, got %v", err)
	}
	if got.Permissions == nil {
		t.Fatal("expected empty permission list to round-trip as an empty slice, got nil")
	}
	if len(got.Permissions) != 0 {
		t.Fatalf("expected empty permission list, got %+v", got.Permissions)
	}

	duplicate := &domainProject.PhasePermission{
		PhaseID:     rule.PhaseID,
		Role:        rule.Role,
		Permissions: []string{},
	}
	if err := repo.Create(ctx, duplicate); !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected duplicate empty phase permission rule to map to conflict, got %v", err)
	}
}

func newPhasePermissionRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.NewReplacer("/", "_", " ", "_", "#", "_").Replace(t.Name()))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatalf("expected sqlite db to open, got %v", err)
	}

	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("expected sql db handle, got %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	if err := db.AutoMigrate(&domainProject.PhasePermission{}); err != nil {
		t.Fatalf("expected phase permission repo tables to migrate, got %v", err)
	}

	return db
}
