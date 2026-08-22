package historysql

import (
	"testing"
	"time"

	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestBackfillV2IsCursorBasedAndIdempotent(t *testing.T) {
	store := newBackfillTestStore(t)
	now := time.Now().UTC().Truncate(time.Millisecond)
	events := []domainHistory.ChangeEvent{
		{ID: uuid.New(), OccurredAt: now, Action: domainHistory.ActionCreate, EntityTable: "field_devices", EntityID: uuid.New()},
		{ID: uuid.New(), OccurredAt: now.Add(time.Second), Action: domainHistory.ActionUpdate, EntityTable: "field_devices", EntityID: uuid.New()},
	}
	if err := store.db.Create(&events).Error; err != nil {
		t.Fatalf("seed events: %v", err)
	}

	first, err := store.BackfillV2(t.Context(), V2BackfillRequest{Limit: 1})
	if err != nil || first.Processed != 1 || first.Done {
		t.Fatalf("unexpected first page: result=%+v error=%v", first, err)
	}
	second, err := store.BackfillV2(t.Context(), V2BackfillRequest{
		AfterOccurredAt: first.NextOccurredAt, AfterID: first.NextID, Limit: 1,
	})
	if err != nil || second.Processed != 1 {
		t.Fatalf("unexpected second page: result=%+v error=%v", second, err)
	}
	if _, err := store.BackfillV2(t.Context(), V2BackfillRequest{Limit: 500}); err != nil {
		t.Fatalf("idempotent replay: %v", err)
	}

	report, err := store.VerifyV2Month(t.Context(), now)
	if err != nil {
		t.Fatalf("verify: %v", err)
	}
	if !report["events"].Matches() {
		t.Fatalf("event verification mismatch: %+v", report["events"])
	}
	if err := store.db.Model(&changeEventV2Record{}).
		Where("id = ?", events[0].ID).
		Update("after_json", domainHistory.JSONB(`{"corrupted":true}`)).Error; err != nil {
		t.Fatalf("corrupt v2 payload: %v", err)
	}
	report, err = store.VerifyV2Month(t.Context(), now)
	if err != nil {
		t.Fatalf("verify corrupted target: %v", err)
	}
	if report["events"].Matches() {
		t.Fatal("expected content checksum to detect corrupted v2 event")
	}
}

func newBackfillTestStore(t *testing.T) *Store {
	t.Helper()
	db, err := gorm.Open(sqlite.Open("file:history-backfill?mode=memory&cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("migrate history: %v", err)
	}
	return NewStore(db)
}
