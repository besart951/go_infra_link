package historycapture

import (
	"context"
	"errors"
	"reflect"
	"testing"

	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type exactProjectSPSSetCreatorStub struct {
	domainProject.ProjectSPSControllerRepository
	insertedIDs []uuid.UUID
	err         error
	projectID   uuid.UUID
	targetIDs   []uuid.UUID
	calls       int
}

func (s *exactProjectSPSSetCreatorStub) BulkCreateBySPSControllerIDsReturningIDs(
	_ context.Context,
	projectID uuid.UUID,
	targetIDs []uuid.UUID,
) ([]uuid.UUID, error) {
	s.calls++
	s.projectID = projectID
	s.targetIDs = append([]uuid.UUID(nil), targetIDs...)
	return append([]uuid.UUID(nil), s.insertedIDs...), s.err
}

type exactProjectFieldDeviceSetCreatorStub struct {
	domainProject.ProjectFieldDeviceRepository
	insertedFieldDeviceIDs []uuid.UUID
	insertedSystemTypeIDs  []uuid.UUID
	err                    error
	projectID              uuid.UUID
	targetIDs              []uuid.UUID
	fieldDeviceCalls       int
	systemTypeCalls        int
}

func (s *exactProjectFieldDeviceSetCreatorStub) BulkCreateByFieldDeviceIDsReturningIDs(
	_ context.Context,
	projectID uuid.UUID,
	targetIDs []uuid.UUID,
) ([]uuid.UUID, error) {
	s.fieldDeviceCalls++
	s.projectID = projectID
	s.targetIDs = append([]uuid.UUID(nil), targetIDs...)
	return append([]uuid.UUID(nil), s.insertedFieldDeviceIDs...), s.err
}

func (s *exactProjectFieldDeviceSetCreatorStub) BulkCreateBySPSControllerSystemTypeIDsReturningIDs(
	_ context.Context,
	projectID uuid.UUID,
	targetIDs []uuid.UUID,
) ([]uuid.UUID, error) {
	s.systemTypeCalls++
	s.projectID = projectID
	s.targetIDs = append([]uuid.UUID(nil), targetIDs...)
	return append([]uuid.UUID(nil), s.insertedSystemTypeIDs...), s.err
}

type projectSetCreateHistoryStore struct {
	loadRowsWhereCalls int
	records            []projectSetCreateHistoryRecord
}

type projectSetCreateHistoryRecord struct {
	table string
	ids   []uuid.UUID
}

func (*projectSetCreateHistoryStore) LoadRow(
	context.Context,
	string,
	uuid.UUID,
) (domainHistory.JSONB, bool, error) {
	return nil, false, nil
}

func (*projectSetCreateHistoryStore) LoadRows(
	context.Context,
	string,
	[]uuid.UUID,
) (map[uuid.UUID]domainHistory.JSONB, error) {
	return nil, nil
}

func (s *projectSetCreateHistoryStore) LoadRowsWhere(
	context.Context,
	string,
	string,
	...any,
) (map[uuid.UUID]domainHistory.JSONB, error) {
	s.loadRowsWhereCalls++
	return nil, nil
}

func (*projectSetCreateHistoryStore) RecordCreate(
	context.Context,
	string,
	uuid.UUID,
) error {
	return nil
}

func (s *projectSetCreateHistoryStore) RecordCreates(
	_ context.Context,
	table string,
	ids []uuid.UUID,
) error {
	s.records = append(s.records, projectSetCreateHistoryRecord{
		table: table,
		ids:   append([]uuid.UUID(nil), ids...),
	})
	return nil
}

func (*projectSetCreateHistoryStore) RecordUpdate(
	context.Context,
	string,
	uuid.UUID,
	domainHistory.JSONB,
) error {
	return nil
}

func (*projectSetCreateHistoryStore) RecordUpdates(
	context.Context,
	string,
	map[uuid.UUID]domainHistory.JSONB,
) error {
	return nil
}

func (*projectSetCreateHistoryStore) RecordDelete(
	context.Context,
	string,
	uuid.UUID,
	domainHistory.JSONB,
) error {
	return nil
}

func (*projectSetCreateHistoryStore) RecordDeletes(
	context.Context,
	string,
	map[uuid.UUID]domainHistory.JSONB,
) error {
	return nil
}

func TestProjectSPSControllerSetCreateRecordsOnlyReturnedInsertIDs(t *testing.T) {
	projectID := uuid.New()
	insertedLinkID := uuid.New()
	targetIDs := []uuid.UUID{uuid.New(), uuid.New()}
	repository := &exactProjectSPSSetCreatorStub{
		insertedIDs: []uuid.UUID{insertedLinkID},
	}
	store := &projectSetCreateHistoryStore{}
	wrapped := WrapProjectSPSController(repository, store)
	creator, ok := wrapped.(interface {
		BulkCreateBySPSControllerIDs(context.Context, uuid.UUID, []uuid.UUID) error
	})
	if !ok {
		t.Fatal("wrapped SPS repository lost set-create capability")
	}

	err := creator.BulkCreateBySPSControllerIDs(
		context.Background(),
		projectID,
		targetIDs,
	)

	if err != nil {
		t.Fatalf("bulk create SPS project links: %v", err)
	}
	if repository.calls != 1 || repository.projectID != projectID ||
		!reflect.DeepEqual(repository.targetIDs, targetIDs) {
		t.Fatalf("exact creator call: %+v", repository)
	}
	if store.loadRowsWhereCalls != 0 {
		t.Fatalf("post-insert broad reloads: got %d, want 0", store.loadRowsWhereCalls)
	}
	if len(store.records) != 1 || store.records[0].table != "project_sps_controllers" ||
		!reflect.DeepEqual(store.records[0].ids, []uuid.UUID{insertedLinkID}) {
		t.Fatalf("recorded create IDs: %+v", store.records)
	}
}

func TestProjectFieldDeviceSetCreatesRecordOnlyReturnedInsertIDs(t *testing.T) {
	projectID := uuid.New()
	fieldDeviceTargetIDs := []uuid.UUID{uuid.New(), uuid.New()}
	systemTypeTargetIDs := []uuid.UUID{uuid.New()}
	fieldDeviceLinkID := uuid.New()
	systemTypeLinkID := uuid.New()
	repository := &exactProjectFieldDeviceSetCreatorStub{
		insertedFieldDeviceIDs: []uuid.UUID{fieldDeviceLinkID},
		insertedSystemTypeIDs:  []uuid.UUID{systemTypeLinkID},
	}
	store := &projectSetCreateHistoryStore{}
	wrapped := WrapProjectFieldDevice(repository, store)
	fieldDeviceCreator, ok := wrapped.(interface {
		BulkCreateByFieldDeviceIDs(context.Context, uuid.UUID, []uuid.UUID) error
	})
	if !ok {
		t.Fatal("wrapped FieldDevice repository lost ID set-create capability")
	}
	systemTypeCreator, ok := wrapped.(interface {
		BulkCreateBySPSControllerSystemTypeIDs(context.Context, uuid.UUID, []uuid.UUID) error
	})
	if !ok {
		t.Fatal("wrapped FieldDevice repository lost system-type set-create capability")
	}

	if err := fieldDeviceCreator.BulkCreateByFieldDeviceIDs(
		context.Background(),
		projectID,
		fieldDeviceTargetIDs,
	); err != nil {
		t.Fatalf("bulk create FieldDevice project links: %v", err)
	}
	if err := systemTypeCreator.BulkCreateBySPSControllerSystemTypeIDs(
		context.Background(),
		projectID,
		systemTypeTargetIDs,
	); err != nil {
		t.Fatalf("bulk create system-type FieldDevice project links: %v", err)
	}

	if repository.fieldDeviceCalls != 1 || repository.systemTypeCalls != 1 ||
		repository.projectID != projectID ||
		!reflect.DeepEqual(repository.targetIDs, systemTypeTargetIDs) {
		t.Fatalf("exact creator calls: %+v", repository)
	}
	if store.loadRowsWhereCalls != 0 {
		t.Fatalf("post-insert broad reloads: got %d, want 0", store.loadRowsWhereCalls)
	}
	wantRecords := []projectSetCreateHistoryRecord{
		{table: "project_field_devices", ids: []uuid.UUID{fieldDeviceLinkID}},
		{table: "project_field_devices", ids: []uuid.UUID{systemTypeLinkID}},
	}
	if !reflect.DeepEqual(store.records, wantRecords) {
		t.Fatalf("recorded create IDs: got %+v want %+v", store.records, wantRecords)
	}
}

func TestProjectSetCreateFailureDoesNotRecordHistory(t *testing.T) {
	insertErr := errors.New("insert failed")
	repository := &exactProjectSPSSetCreatorStub{err: insertErr}
	store := &projectSetCreateHistoryStore{}
	wrapped := WrapProjectSPSController(repository, store)
	creator := wrapped.(interface {
		BulkCreateBySPSControllerIDs(context.Context, uuid.UUID, []uuid.UUID) error
	})

	err := creator.BulkCreateBySPSControllerIDs(
		context.Background(),
		uuid.New(),
		[]uuid.UUID{uuid.New()},
	)

	if !errors.Is(err, insertErr) || len(store.records) != 0 ||
		store.loadRowsWhereCalls != 0 {
		t.Fatalf("failed exact insert: err=%v records=%+v loads=%d",
			err,
			store.records,
			store.loadRowsWhereCalls,
		)
	}
}

func TestProjectSetCreateConflictNoOpDoesNotRecordFalseCreateHistory(t *testing.T) {
	repository := &exactProjectSPSSetCreatorStub{}
	store := &projectSetCreateHistoryStore{}
	wrapped := WrapProjectSPSController(repository, store)
	creator := wrapped.(interface {
		BulkCreateBySPSControllerIDs(context.Context, uuid.UUID, []uuid.UUID) error
	})

	err := creator.BulkCreateBySPSControllerIDs(
		context.Background(),
		uuid.New(),
		[]uuid.UUID{uuid.New()},
	)

	if err != nil || repository.calls != 1 || len(store.records) != 0 ||
		store.loadRowsWhereCalls != 0 {
		t.Fatalf("conflict no-op history: err=%v calls=%d records=%+v loads=%d",
			err,
			repository.calls,
			store.records,
			store.loadRowsWhereCalls,
		)
	}
}
