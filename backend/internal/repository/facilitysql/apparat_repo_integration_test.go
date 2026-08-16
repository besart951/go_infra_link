package facilitysql

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestApparatRepositoryUpdatePersistsClearedScalarsAndSystemPartRelations(t *testing.T) {
	ctx := context.Background()
	db := newApparatRepoTestDB(t)
	repo := NewApparatRepository(db)

	firstPart := seedApparatSystemPart(t, db, "AIR", "Air")
	secondPart := seedApparatSystemPart(t, db, "WAT", "Water")
	description := "original description"
	entity := &domainFacility.Apparat{
		ShortName:   "PMP",
		Name:        "Pump",
		Description: &description,
		SystemParts: []*domainFacility.SystemPart{firstPart},
	}
	if err := repo.Create(ctx, entity); err != nil {
		t.Fatalf("create apparat: %v", err)
	}
	initialVersion := entity.Version

	entity.Name = "Pump replacement"
	entity.Description = nil
	entity.SystemParts = []*domainFacility.SystemPart{secondPart}
	if err := repo.Update(ctx, entity); err != nil {
		t.Fatalf("update apparat: %v", err)
	}
	if entity.Version != initialVersion+1 {
		t.Fatalf("version = %d, want %d", entity.Version, initialVersion+1)
	}

	items, err := repo.GetByIds(ctx, []uuid.UUID{entity.ID})
	if err != nil {
		t.Fatalf("reload apparat: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("reloaded %d apparats, want 1", len(items))
	}
	stored := items[0]
	if stored.Name != "Pump replacement" || stored.Description != nil {
		t.Fatalf("scalar update = %+v, want name change and nil description", stored)
	}
	if len(stored.SystemParts) != 1 || stored.SystemParts[0].ID != secondPart.ID {
		t.Fatalf("system part relation = %+v, want only %s", stored.SystemParts, secondPart.ID)
	}
}

func newApparatRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.NewReplacer("/", "_", " ", "_", "#", "_").Replace(t.Name()))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{DisableForeignKeyConstraintWhenMigrating: true})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("sqlite handle: %v", err)
	}
	sqlDB.SetMaxOpenConns(1)
	t.Cleanup(func() { _ = sqlDB.Close() })
	if err := db.AutoMigrate(&domainFacility.SystemPart{}, &domainFacility.Apparat{}); err != nil {
		t.Fatalf("migrate apparatus records: %v", err)
	}
	return db
}

func seedApparatSystemPart(t *testing.T, db *gorm.DB, shortName, name string) *domainFacility.SystemPart {
	t.Helper()
	part := &domainFacility.SystemPart{ShortName: shortName, Name: name}
	if err := part.InitForCreate(time.Now().UTC()); err != nil {
		t.Fatalf("initialize system part: %v", err)
	}
	if err := db.Create(part).Error; err != nil {
		t.Fatalf("create system part: %v", err)
	}
	return part
}
