package historysql

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

func TestUndoPreflightRejectsNewerChange(t *testing.T) {
	db := historyV2TestDB(t)
	now := time.Now().UTC()
	event := undoTestEvent(now, 2)
	newer := undoTestEvent(now.Add(time.Second), 3)
	newer.EntityID = event.EntityID
	if err := db.Create(&event).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&newer).Error; err != nil {
		t.Fatal(err)
	}

	err := NewStore(db).checkUndoConflict(context.Background(), undoPreflight{
		event: &event, mode: domainHistory.RestoreModeBefore,
		current: domainHistory.JSONB(`{"id":"same","version":3}`),
	})
	var conflict *domainHistory.UndoConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("expected structured undo conflict, got %v", err)
	}
	if conflict.Conflict.ExpectedVersion == nil || *conflict.Conflict.ExpectedVersion != 2 {
		t.Fatalf("expected version 2, got %+v", conflict.Conflict.ExpectedVersion)
	}
	if conflict.Conflict.CurrentVersion == nil || *conflict.Conflict.CurrentVersion != 3 {
		t.Fatalf("current version 3, got %+v", conflict.Conflict.CurrentVersion)
	}
}

func TestUndoPreflightAllowsUnchangedCurrentVersion(t *testing.T) {
	db := historyV2TestDB(t)
	event := undoTestEvent(time.Now().UTC(), 2)
	if err := db.Create(&event).Error; err != nil {
		t.Fatal(err)
	}
	err := NewStore(db).checkUndoConflict(context.Background(), undoPreflight{
		event: &event, mode: domainHistory.RestoreModeBefore,
		current: domainHistory.JSONB(`{"id":"same","version":2}`),
	})
	if err != nil {
		t.Fatalf("expected undo to be allowed, got %v", err)
	}
}

func TestUndoBatchPreflightsEveryEntityBeforeMutation(t *testing.T) {
	db := historyV2TestDB(t)
	if err := db.AutoMigrate(&domainFacility.Unit{}); err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	batchID := uuid.New()
	units := []domainFacility.Unit{
		{Base: domain.Base{ID: uuid.New(), Version: 2}, Code: "u1", Symbol: "1", Name: "one"},
		{Base: domain.Base{ID: uuid.New(), Version: 2}, Code: "u2", Symbol: "2", Name: "two"},
	}
	if err := db.Create(&units).Error; err != nil {
		t.Fatal(err)
	}
	for index := range units {
		event := undoTestEvent(now.Add(time.Duration(index)*time.Millisecond), 2)
		event.EntityID = units[index].ID
		event.BatchID = &batchID
		event.BeforeJSON = unitSnapshot(units[index], 1)
		event.AfterJSON = unitSnapshot(units[index], 2)
		if err := db.Create(&event).Error; err != nil {
			t.Fatal(err)
		}
	}
	newer := undoTestEvent(now.Add(time.Second), 3)
	newer.EntityID = units[1].ID
	if err := db.Create(&newer).Error; err != nil {
		t.Fatal(err)
	}

	_, err := NewStore(db).UndoBatch(t.Context(), batchID)
	var conflict *domainHistory.UndoConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("expected batch conflict, got %v", err)
	}
	var unchanged domainFacility.Unit
	if err := db.First(&unchanged, "id = ?", units[0].ID).Error; err != nil {
		t.Fatal(err)
	}
	if unchanged.Version != 2 {
		t.Fatalf("first entity was mutated before full preflight: version=%d", unchanged.Version)
	}
}

func TestUndoBatchRestoresAllEntitiesAtomically(t *testing.T) {
	db := historyV2TestDB(t)
	if err := db.AutoMigrate(&domainFacility.Unit{}); err != nil {
		t.Fatal(err)
	}
	batchID := uuid.New()
	unit := domainFacility.Unit{
		Base: domain.Base{ID: uuid.New(), Version: 2}, Code: "batch", Symbol: "b", Name: "after",
	}
	if err := db.Create(&unit).Error; err != nil {
		t.Fatal(err)
	}
	event := undoTestEvent(time.Now().UTC(), 2)
	event.EntityID = unit.ID
	event.BatchID = &batchID
	event.BeforeJSON = domainHistory.JSONB([]byte(fmt.Sprintf(
		`{"id":%q,"version":1,"code":"batch","symbol":"b","name":"before"}`, unit.ID,
	)))
	event.AfterJSON = unitSnapshot(unit, 2)
	if err := db.Create(&event).Error; err != nil {
		t.Fatal(err)
	}

	result, err := NewStore(db).UndoBatch(t.Context(), batchID)
	if err != nil {
		t.Fatalf("undo batch: %v", err)
	}
	if result.RestoredCount != 1 {
		t.Fatalf("unexpected result: %+v", result)
	}
	var restored domainFacility.Unit
	if err := db.First(&restored, "id = ?", unit.ID).Error; err != nil {
		t.Fatal(err)
	}
	if restored.Name != "before" || restored.Version != 1 {
		t.Fatalf("unexpected restored unit: %+v", restored)
	}
}

func unitSnapshot(unit domainFacility.Unit, version uint64) domainHistory.JSONB {
	return domainHistory.JSONB([]byte(fmt.Sprintf(
		`{"id":%q,"version":%d,"code":%q,"symbol":%q,"name":%q}`,
		unit.ID, version, unit.Code, unit.Symbol, unit.Name,
	)))
}

func undoTestEvent(occurredAt time.Time, version uint64) domainHistory.ChangeEvent {
	return domainHistory.ChangeEvent{
		ID: uuid.New(), OccurredAt: occurredAt, Action: domainHistory.ActionUpdate,
		EntityTable: "units", EntityID: uuid.New(),
		BeforeJSON: domainHistory.JSONB(`{"id":"same","version":1}`),
		AfterJSON:  domainHistory.JSONB([]byte(fmt.Sprintf(`{"id":"same","version":%d}`, version))),
	}
}
