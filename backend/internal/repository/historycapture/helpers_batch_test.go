package historycapture

import (
	"context"
	"reflect"
	"testing"

	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

type batchCaptureStore struct {
	order         []string
	createdIDs    []uuid.UUID
	deletedRows   map[uuid.UUID]domainHistory.JSONB
	singleCreates int
	singleDeletes int
}

func (*batchCaptureStore) LoadRow(
	context.Context,
	string,
	uuid.UUID,
) (domainHistory.JSONB, bool, error) {
	return nil, false, nil
}

func (*batchCaptureStore) LoadRows(
	context.Context,
	string,
	[]uuid.UUID,
) (map[uuid.UUID]domainHistory.JSONB, error) {
	return nil, nil
}

func (*batchCaptureStore) LoadRowsWhere(
	context.Context,
	string,
	string,
	...any,
) (map[uuid.UUID]domainHistory.JSONB, error) {
	return nil, nil
}

func (s *batchCaptureStore) RecordCreate(context.Context, string, uuid.UUID) error {
	s.singleCreates++
	return nil
}

func (s *batchCaptureStore) RecordCreates(
	_ context.Context,
	_ string,
	ids []uuid.UUID,
) error {
	s.order = append(s.order, "record")
	s.createdIDs = append([]uuid.UUID(nil), ids...)
	return nil
}

func (*batchCaptureStore) RecordUpdate(
	context.Context,
	string,
	uuid.UUID,
	domainHistory.JSONB,
) error {
	return nil
}

func (*batchCaptureStore) RecordUpdates(
	context.Context,
	string,
	map[uuid.UUID]domainHistory.JSONB,
) error {
	return nil
}

func (s *batchCaptureStore) RecordDelete(
	context.Context,
	string,
	uuid.UUID,
	domainHistory.JSONB,
) error {
	s.singleDeletes++
	return nil
}

func (s *batchCaptureStore) RecordDeletes(
	_ context.Context,
	_ string,
	rows map[uuid.UUID]domainHistory.JSONB,
) error {
	s.order = append(s.order, "record")
	s.deletedRows = cloneHistoryRows(rows)
	return nil
}

func TestAuditBulkCreateUsesOneBatchHistoryCallAfterPersistence(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	store := &batchCaptureStore{}
	audit := newAudit[struct{}]("field_devices", store)

	err := audit.bulkCreate(
		context.Background(),
		func(context.Context) error {
			store.order = append(store.order, "persist")
			return nil
		},
		func() []uuid.UUID { return []uuid.UUID{firstID, secondID} },
	)
	if err != nil {
		t.Fatalf("bulk create: %v", err)
	}

	if want := []string{"persist", "record"}; !reflect.DeepEqual(store.order, want) {
		t.Fatalf("order: got %v, want %v", store.order, want)
	}
	if want := []uuid.UUID{firstID, secondID}; !reflect.DeepEqual(store.createdIDs, want) {
		t.Fatalf("created IDs: got %v, want %v", store.createdIDs, want)
	}
	if store.singleCreates != 0 {
		t.Fatalf("expected no per-row create history calls, got %d", store.singleCreates)
	}
}

func TestAuditDeleteRowsUsesOneBatchHistoryCallAfterPersistence(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	rows := map[uuid.UUID]domainHistory.JSONB{
		firstID:  domainHistory.JSONB(`{"id":"first"}`),
		secondID: domainHistory.JSONB(`{"id":"second"}`),
	}
	store := &batchCaptureStore{}
	audit := newAudit[struct{}]("field_devices", store)

	err := audit.deleteRows(
		context.Background(),
		func(context.Context) (map[uuid.UUID]domainHistory.JSONB, error) {
			store.order = append(store.order, "load")
			return rows, nil
		},
		func(context.Context) error {
			store.order = append(store.order, "persist")
			return nil
		},
	)
	if err != nil {
		t.Fatalf("delete rows: %v", err)
	}

	if want := []string{"load", "persist", "record"}; !reflect.DeepEqual(store.order, want) {
		t.Fatalf("order: got %v, want %v", store.order, want)
	}
	if !reflect.DeepEqual(store.deletedRows, rows) {
		t.Fatalf("deleted rows: got %v, want %v", store.deletedRows, rows)
	}
	if store.singleDeletes != 0 {
		t.Fatalf("expected no per-row delete history calls, got %d", store.singleDeletes)
	}
}

func cloneHistoryRows(
	rows map[uuid.UUID]domainHistory.JSONB,
) map[uuid.UUID]domainHistory.JSONB {
	cloned := make(map[uuid.UUID]domainHistory.JSONB, len(rows))
	for id, snapshot := range rows {
		cloned[id] = append(domainHistory.JSONB(nil), snapshot...)
	}
	return cloned
}
