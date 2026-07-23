package historycapture_test

import (
	"context"
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/besart951/go_infra_link/backend/internal/repository/historycapture"
	"github.com/besart951/go_infra_link/backend/internal/repository/historysql"
	"github.com/besart951/go_infra_link/backend/internal/service/auditctx"
	"github.com/google/uuid"
)

var _ historycapture.ChangeStore = (*historysql.Store)(nil)

type captureEntity struct {
	domain.Base
	Name string
}

type captureRepository struct {
	items       map[uuid.UUID]*captureEntity
	updateCalls int
}

func (r *captureRepository) GetByIds(
	_ context.Context,
	ids []uuid.UUID,
) ([]*captureEntity, error) {
	items := make([]*captureEntity, 0, len(ids))
	for _, id := range ids {
		if entity := r.items[id]; entity != nil {
			clone := *entity
			items = append(items, &clone)
		}
	}
	return items, nil
}

func (r *captureRepository) Create(_ context.Context, entity *captureEntity) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *captureRepository) Update(_ context.Context, entity *captureEntity) error {
	r.updateCalls++
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *captureRepository) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *captureRepository) GetPaginatedList(
	context.Context,
	domain.PaginationParams,
) (*domain.PaginatedList[captureEntity], error) {
	return &domain.PaginatedList[captureEntity]{}, nil
}

type updateRecord struct {
	table   string
	id      uuid.UUID
	before  domainHistory.JSONB
	batchID *uuid.UUID
}

type captureStore struct {
	before            domainHistory.JSONB
	loadRowCalls      int
	updateRecords     []updateRecord
	recordUpdateError error
}

func (s *captureStore) LoadRow(
	context.Context,
	string,
	uuid.UUID,
) (domainHistory.JSONB, bool, error) {
	s.loadRowCalls++
	return append(domainHistory.JSONB(nil), s.before...), true, nil
}

func (*captureStore) LoadRows(
	context.Context,
	string,
	[]uuid.UUID,
) (map[uuid.UUID]domainHistory.JSONB, error) {
	return nil, nil
}

func (*captureStore) LoadRowsWhere(
	context.Context,
	string,
	string,
	...any,
) (map[uuid.UUID]domainHistory.JSONB, error) {
	return nil, nil
}

func (*captureStore) RecordCreate(context.Context, string, uuid.UUID) error {
	return nil
}

func (*captureStore) RecordCreates(context.Context, string, []uuid.UUID) error {
	return nil
}

func (s *captureStore) RecordUpdate(
	ctx context.Context,
	table string,
	id uuid.UUID,
	before domainHistory.JSONB,
) error {
	batchID, _ := auditctx.BatchID(ctx)
	s.updateRecords = append(s.updateRecords, updateRecord{
		table:   table,
		id:      id,
		before:  append(domainHistory.JSONB(nil), before...),
		batchID: batchID,
	})
	return s.recordUpdateError
}

func (*captureStore) RecordUpdates(
	context.Context,
	string,
	map[uuid.UUID]domainHistory.JSONB,
) error {
	return nil
}

func (*captureStore) RecordDelete(
	context.Context,
	string,
	uuid.UUID,
	domainHistory.JSONB,
) error {
	return nil
}

func (*captureStore) RecordDeletes(
	context.Context,
	string,
	map[uuid.UUID]domainHistory.JSONB,
) error {
	return nil
}

func TestRepositoryUpdateRecordsExactlyOneTransactionalHistoryMutation(t *testing.T) {
	entityID := uuid.New()
	batchID := uuid.New()
	repository := &captureRepository{items: map[uuid.UUID]*captureEntity{
		entityID: {
			Base: domain.Base{ID: entityID},
			Name: "before",
		},
	}}
	store := &captureStore{before: domainHistory.JSONB(`{"name":"before"}`)}
	wrapped := historycapture.WrapRepository("capture_entities", repository, store)

	err := wrapped.Update(
		auditctx.WithBatchID(context.Background(), batchID),
		&captureEntity{
			Base: domain.Base{ID: entityID},
			Name: "after",
		},
	)
	if err != nil {
		t.Fatalf("update: %v", err)
	}

	if repository.updateCalls != 1 {
		t.Fatalf("repository updates: got %d, want 1", repository.updateCalls)
	}
	if store.loadRowCalls != 1 {
		t.Fatalf("before-snapshot loads: got %d, want 1", store.loadRowCalls)
	}
	if len(store.updateRecords) != 1 {
		t.Fatalf("history updates: got %+v, want exactly one", store.updateRecords)
	}
	record := store.updateRecords[0]
	if record.table != "capture_entities" || record.id != entityID {
		t.Fatalf("history identity: got %+v", record)
	}
	if string(record.before) != `{"name":"before"}` {
		t.Fatalf("before snapshot: got %s", record.before)
	}
	if record.batchID == nil || *record.batchID != batchID {
		t.Fatalf("history batch: got %v, want %s", record.batchID, batchID)
	}
	if got := repository.items[entityID].Name; got != "after" {
		t.Fatalf("persisted entity: got %q, want after", got)
	}
}

func TestRepositoryUpdatePropagatesHistoryFailure(t *testing.T) {
	entityID := uuid.New()
	historyErr := errors.New("history unavailable")
	repository := &captureRepository{items: map[uuid.UUID]*captureEntity{
		entityID: {Base: domain.Base{ID: entityID}},
	}}
	store := &captureStore{recordUpdateError: historyErr}
	wrapped := historycapture.WrapRepository("capture_entities", repository, store)

	err := wrapped.Update(context.Background(), &captureEntity{
		Base: domain.Base{ID: entityID},
		Name: "after",
	})

	if !errors.Is(err, historyErr) {
		t.Fatalf("expected history error %v, got %v", historyErr, err)
	}
}
