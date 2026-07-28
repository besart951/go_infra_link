package historysql

import (
	"context"
	"os"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// This opt-in test exercises PostgreSQL to_jsonb snapshots, constraint
// conflicts, and restore rollback against a disposable, fully migrated
// database.
func TestPostgresRestoreEntityUsesSnapshotAndRollsBackUniqueConflict(
	t *testing.T,
) {
	dsn := os.Getenv("FACILITY_TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("FACILITY_TEST_DATABASE_URL is not set")
	}

	database, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("connect test database: %v", err)
	}
	tx := database.Begin()
	if tx.Error != nil {
		t.Fatalf("begin test transaction: %v", tx.Error)
	}
	t.Cleanup(func() {
		_ = tx.Rollback().Error
	})

	ctx := context.Background()
	store := NewStore(tx)
	entityID := uuid.New()
	suffix := entityID.String()
	oldShortName := "old-" + suffix
	oldName := "Old system part " + suffix
	newShortName := "new-" + suffix
	newName := "New system part " + suffix
	entity := &domainFacility.SystemPart{
		Base:      domain.Base{ID: entityID},
		ShortName: oldShortName,
		Name:      oldName,
	}
	if err := tx.WithContext(ctx).Create(entity).Error; err != nil {
		t.Fatalf("create SystemPart: %v", err)
	}

	before, exists, err := store.LoadRow(ctx, "system_parts", entityID)
	if err != nil {
		t.Fatalf("load PostgreSQL snapshot: %v", err)
	}
	if !exists || len(before) == 0 {
		t.Fatal("expected PostgreSQL to_jsonb snapshot")
	}
	if err := tx.WithContext(ctx).
		Model(&domainFacility.SystemPart{}).
		Where("id = ?", entityID).
		Updates(map[string]any{
			"short_name": newShortName,
			"name":       newName,
		}).Error; err != nil {
		t.Fatalf("update SystemPart: %v", err)
	}
	if err := store.RecordUpdate(ctx, "system_parts", entityID, before); err != nil {
		t.Fatalf("record update: %v", err)
	}

	var updateEvent domainHistory.ChangeEvent
	if err := tx.WithContext(ctx).
		Where(
			"entity_table = ? AND entity_id = ? AND action = ?",
			"system_parts",
			entityID,
			domainHistory.ActionUpdate,
		).
		First(&updateEvent).Error; err != nil {
		t.Fatalf("load update event: %v", err)
	}

	conflict := &domainFacility.SystemPart{
		Base:      domain.Base{ID: uuid.New()},
		ShortName: oldShortName,
		Name:      oldName,
	}
	if err := tx.WithContext(ctx).Create(conflict).Error; err != nil {
		t.Fatalf("create restore conflict: %v", err)
	}

	if _, err := store.RestoreEntityToEvent(
		ctx,
		updateEvent.ID,
		domainHistory.RestoreModeBefore,
	); err == nil {
		t.Fatal("expected unique conflict during restore")
	}
	assertPostgresSystemPartNames(
		t,
		tx,
		entityID,
		newShortName,
		newName,
	)
	var restoreCount int64
	if err := tx.WithContext(ctx).
		Model(&domainHistory.ChangeEvent{}).
		Where(
			"entity_table = ? AND entity_id = ? AND action = ?",
			"system_parts",
			entityID,
			domainHistory.ActionRestore,
		).
		Count(&restoreCount).Error; err != nil {
		t.Fatalf("count rolled-back restore events: %v", err)
	}
	if restoreCount != 0 {
		t.Fatalf("restore events after conflict: got %d, want 0", restoreCount)
	}

	if err := tx.WithContext(ctx).Delete(conflict).Error; err != nil {
		t.Fatalf("remove restore conflict: %v", err)
	}
	result, err := store.RestoreEntityToEvent(
		ctx,
		updateEvent.ID,
		domainHistory.RestoreModeBefore,
	)
	if err != nil {
		t.Fatalf("restore SystemPart: %v", err)
	}
	if result.RestoredCount != 1 || result.DeletedCount != 0 {
		t.Fatalf("restore result: %+v", result)
	}
	assertPostgresSystemPartNames(t, tx, entityID, oldShortName, oldName)
}

func assertPostgresSystemPartNames(
	t *testing.T,
	db *gorm.DB,
	id uuid.UUID,
	wantShortName string,
	wantName string,
) {
	t.Helper()
	var entity domainFacility.SystemPart
	if err := db.Where("id = ?", id).First(&entity).Error; err != nil {
		t.Fatalf("load SystemPart %s: %v", id, err)
	}
	if entity.ShortName != wantShortName || entity.Name != wantName {
		t.Fatalf(
			"SystemPart names: got %q/%q, want %q/%q",
			entity.ShortName,
			entity.Name,
			wantShortName,
			wantName,
		)
	}
}
