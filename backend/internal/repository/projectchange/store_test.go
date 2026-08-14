package projectchange

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestStoreAllocatesMonotonicProjectRevisionsAndPages(t *testing.T) {
	store := openStore(t)
	projectID := uuid.New()
	for _, action := range []domainProject.ChangeAction{domainProject.ChangeCreated, domainProject.ChangeUpdated, domainProject.ChangeDeleted} {
		change, err := store.Append(context.Background(), domainProject.NewChange{
			ProjectID: projectID, AggregateType: "field_device", Action: action,
		})
		if err != nil {
			t.Fatalf("append %s: %v", action, err)
		}
		if want := uint64(len(actionRevisionsBefore(action)) + 1); change.Revision != want {
			t.Fatalf("revision = %d, want %d", change.Revision, want)
		}
	}

	page, err := store.ListAfter(context.Background(), projectID, 0, 2)
	if err != nil {
		t.Fatalf("list first page: %v", err)
	}
	if page.CurrentRevision != 3 || len(page.Events) != 2 || !page.HasMore || page.ResetRequired {
		t.Fatalf("unexpected first page: %+v", page)
	}
	if page.Events[0].Revision != 1 || page.Events[1].Revision != 2 {
		t.Fatalf("events are not revision ordered: %+v", page.Events)
	}

	page, err = store.ListAfter(context.Background(), projectID, 2, 2)
	if err != nil {
		t.Fatalf("list second page: %v", err)
	}
	if len(page.Events) != 1 || page.Events[0].Revision != 3 || page.HasMore {
		t.Fatalf("unexpected second page: %+v", page)
	}
}

func TestStoreRequiresResetWhenRetentionCreatesRevisionGap(t *testing.T) {
	store := openStore(t)
	now := time.Date(2026, 8, 13, 12, 0, 0, 0, time.UTC)
	store.now = func() time.Time { return now }
	projectID := uuid.New()

	if _, err := store.Append(context.Background(), domainProject.NewChange{
		ProjectID: projectID, AggregateType: "project", Action: domainProject.ChangeCreated,
		OccurredAt: now.Add(-Retention - time.Hour),
	}); err != nil {
		t.Fatalf("append old event: %v", err)
	}
	if _, err := store.Append(context.Background(), domainProject.NewChange{
		ProjectID: projectID, AggregateType: "project", Action: domainProject.ChangeUpdated,
		OccurredAt: now,
	}); err != nil {
		t.Fatalf("append current event: %v", err)
	}

	page, err := store.ListAfter(context.Background(), projectID, 0, 100)
	if err != nil {
		t.Fatalf("list retained events: %v", err)
	}
	if !page.ResetRequired || len(page.Events) != 0 || page.CurrentRevision != 2 {
		t.Fatalf("expected reset at revision 2, got %+v", page)
	}
}

func TestStoreRequiresResetForFutureRevision(t *testing.T) {
	store := openStore(t)
	projectID := uuid.New()
	if _, err := store.Append(context.Background(), domainProject.NewChange{
		ProjectID: projectID, AggregateType: "project", Action: domainProject.ChangeCreated,
	}); err != nil {
		t.Fatalf("append event: %v", err)
	}

	page, err := store.ListAfter(context.Background(), projectID, 99, 100)
	if err != nil {
		t.Fatalf("list future revision: %v", err)
	}
	if !page.ResetRequired || page.CurrentRevision != 1 {
		t.Fatalf("expected future cursor reset, got %+v", page)
	}
}

func TestStoreAppendBatchAllocatesConsecutiveRevisions(t *testing.T) {
	store := openStore(t)
	projectID := uuid.New()
	changes, err := store.AppendBatch(context.Background(), []domainProject.NewChange{
		{ProjectID: projectID, AggregateType: "field_device", AggregateID: ptr(uuid.New()), Action: domainProject.ChangeCreated},
		{ProjectID: projectID, AggregateType: "field_device", AggregateID: ptr(uuid.New()), Action: domainProject.ChangeCreated},
	})
	if err != nil {
		t.Fatalf("append batch: %v", err)
	}
	if len(changes) != 2 || changes[0].Revision != 1 || changes[1].Revision != 2 {
		t.Fatalf("unexpected revisions: %+v", changes)
	}
	page, err := store.ListAfter(context.Background(), projectID, 0, 10)
	if err != nil {
		t.Fatal(err)
	}
	if page.CurrentRevision != 2 || len(page.Events) != 2 {
		t.Fatalf("unexpected page: %+v", page)
	}
}

func ptr[T any](value T) *T { return &value }

func openStore(t *testing.T) *Store {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(filepath.Join(t.TempDir(), "changes.db")), &gorm.Config{})
	if err != nil {
		t.Fatalf("open sqlite: %v", err)
	}
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("migrate change store: %v", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		t.Fatalf("get sqlite handle: %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return NewStore(db)
}

func actionRevisionsBefore(action domainProject.ChangeAction) []domainProject.ChangeAction {
	switch action {
	case domainProject.ChangeCreated:
		return nil
	case domainProject.ChangeUpdated:
		return []domainProject.ChangeAction{domainProject.ChangeCreated}
	default:
		return []domainProject.ChangeAction{domainProject.ChangeCreated, domainProject.ChangeUpdated}
	}
}
