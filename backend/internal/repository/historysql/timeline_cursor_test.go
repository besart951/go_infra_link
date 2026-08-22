package historysql

import (
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/cursor"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestTimelineCursorTraversesBothDirections(t *testing.T) {
	store := newTimelineTestStore(t)
	events := seedTimelineEvents(t, store.db, 5)

	first, err := store.ListTimelineCursor(t.Context(), domainHistory.TimelineFilter{Limit: 2})
	if err != nil {
		t.Fatalf("first timeline page: %v", err)
	}
	second, err := store.ListTimelineCursor(t.Context(), domainHistory.TimelineFilter{Limit: 2, Cursor: first.NextCursor})
	if err != nil {
		t.Fatalf("second timeline page: %v", err)
	}
	back, err := store.ListTimelineCursor(t.Context(), domainHistory.TimelineFilter{Limit: 2, Cursor: second.PreviousCursor})
	if err != nil {
		t.Fatalf("previous timeline page: %v", err)
	}

	assertTimelineIDs(t, first.Items, []uuid.UUID{events[4].ID, events[3].ID})
	assertTimelineIDs(t, second.Items, []uuid.UUID{events[2].ID, events[1].ID})
	assertTimelineIDs(t, back.Items, []uuid.UUID{events[4].ID, events[3].ID})
}

func TestTimelineCursorRejectsChangedFilter(t *testing.T) {
	store := newTimelineTestStore(t)
	seedTimelineEvents(t, store.db, 3)
	first, err := store.ListTimelineCursor(t.Context(), domainHistory.TimelineFilter{Limit: 1, Actions: []domainHistory.Action{domainHistory.ActionCreate}})
	if err != nil {
		t.Fatalf("first timeline page: %v", err)
	}
	_, err = store.ListTimelineCursor(t.Context(), domainHistory.TimelineFilter{Limit: 1, Actions: []domainHistory.Action{domainHistory.ActionUpdate}, Cursor: first.NextCursor})
	if !errors.Is(err, cursor.ErrInvalid) {
		t.Fatalf("changed-filter error = %v, want %v", err, cursor.ErrInvalid)
	}
}

func newTimelineTestStore(t *testing.T) *Store {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&domainHistory.ChangeEvent{}, &domainHistory.ChangeEventScope{}, &domainHistory.EntityVersion{}); err != nil {
		t.Fatal(err)
	}
	return NewStore(db)
}

func seedTimelineEvents(t *testing.T, db *gorm.DB, count int) []domainHistory.ChangeEvent {
	t.Helper()
	events := make([]domainHistory.ChangeEvent, count)
	for index := range count {
		events[index] = domainHistory.ChangeEvent{
			ID: uuid.New(), OccurredAt: time.Date(2026, 1, 1, 0, index, 0, 0, time.UTC),
			Action: domainHistory.ActionCreate, EntityTable: "field_devices", EntityID: uuid.New(),
		}
		if err := db.Create(&events[index]).Error; err != nil {
			t.Fatal(err)
		}
	}
	return events
}

func assertTimelineIDs(t *testing.T, events []domainHistory.ChangeEvent, want []uuid.UUID) {
	t.Helper()
	got := make([]uuid.UUID, len(events))
	for index := range events {
		got[index] = events[index].ID
	}
	if fmt.Sprint(got) != fmt.Sprint(want) {
		t.Fatalf("timeline ids = %v, want %v", got, want)
	}
}
