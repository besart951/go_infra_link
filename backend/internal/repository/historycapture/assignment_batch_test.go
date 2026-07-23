package historycapture

import (
	"context"
	"reflect"
	"testing"

	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainFieldDevice "github.com/besart951/go_infra_link/backend/internal/domain/facility/fielddevice"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

type assignmentChangeStore struct {
	ChangeStore
	before             map[uuid.UUID]domainHistory.JSONB
	order              []string
	loadedTable        string
	loadedIDs          []uuid.UUID
	recordedTable      string
	recordedBeforeRows map[uuid.UUID]domainHistory.JSONB
	loadCalls          int
	recordCalls        int
}

func (s *assignmentChangeStore) LoadRows(
	_ context.Context,
	table string,
	ids []uuid.UUID,
) (map[uuid.UUID]domainHistory.JSONB, error) {
	s.order = append(s.order, "history:load")
	s.loadCalls++
	s.loadedTable = table
	s.loadedIDs = append([]uuid.UUID(nil), ids...)
	return cloneHistoryRows(s.before), nil
}

func (s *assignmentChangeStore) RecordUpdates(
	_ context.Context,
	table string,
	beforeRows map[uuid.UUID]domainHistory.JSONB,
) error {
	s.order = append(s.order, "history:record")
	s.recordCalls++
	s.recordedTable = table
	s.recordedBeforeRows = cloneHistoryRows(beforeRows)
	return nil
}

type fieldDeviceAssignmentStore struct {
	domainFieldDevice.FieldDeviceStore
	order       *[]string
	assignments map[uuid.UUID]uuid.UUID
	calls       int
}

func (s *fieldDeviceAssignmentStore) AssignSpecificationIDs(
	_ context.Context,
	assignments map[uuid.UUID]uuid.UUID,
) error {
	*s.order = append(*s.order, "persistence:assign")
	s.calls++
	s.assignments = cloneUUIDMap(assignments)
	return nil
}

type bacnetReferenceAssignmentStore struct {
	domainObjectData.BacnetObjectStore
	order       *[]string
	assignments map[uuid.UUID]uuid.UUID
	calls       int
}

func (s *bacnetReferenceAssignmentStore) AssignSoftwareReferenceIDs(
	_ context.Context,
	assignments map[uuid.UUID]uuid.UUID,
) error {
	*s.order = append(*s.order, "persistence:assign")
	s.calls++
	s.assignments = cloneUUIDMap(assignments)
	return nil
}

func TestFieldDeviceSpecificationAssignmentsUseOneBeforeAfterHistoryBatch(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	assignments := map[uuid.UUID]uuid.UUID{
		firstID:  uuid.New(),
		secondID: uuid.New(),
	}
	before := map[uuid.UUID]domainHistory.JSONB{
		firstID:  domainHistory.JSONB(`{"id":"first","specification_id":null}`),
		secondID: domainHistory.JSONB(`{"id":"second","specification_id":null}`),
	}
	changes := &assignmentChangeStore{before: before}
	persistence := &fieldDeviceAssignmentStore{order: &changes.order}
	wrapped := WrapFieldDevice(persistence, changes)

	if err := wrapped.AssignSpecificationIDs(context.Background(), assignments); err != nil {
		t.Fatalf("assign specifications: %v", err)
	}

	assertAssignmentBatch(t, changes, persistence.calls, persistence.assignments, assignments, before, "field_devices")
}

func TestBacnetSoftwareReferenceAssignmentsUseOneBeforeAfterHistoryBatch(t *testing.T) {
	firstID := uuid.New()
	secondID := uuid.New()
	assignments := map[uuid.UUID]uuid.UUID{
		firstID:  uuid.New(),
		secondID: uuid.New(),
	}
	before := map[uuid.UUID]domainHistory.JSONB{
		firstID:  domainHistory.JSONB(`{"id":"first","software_reference_id":null}`),
		secondID: domainHistory.JSONB(`{"id":"second","software_reference_id":null}`),
	}
	changes := &assignmentChangeStore{before: before}
	persistence := &bacnetReferenceAssignmentStore{order: &changes.order}
	wrapped := WrapBacnetObject(persistence, changes)

	if err := wrapped.AssignSoftwareReferenceIDs(context.Background(), assignments); err != nil {
		t.Fatalf("assign software references: %v", err)
	}

	assertAssignmentBatch(t, changes, persistence.calls, persistence.assignments, assignments, before, "bacnet_objects")
}

func assertAssignmentBatch(
	t *testing.T,
	changes *assignmentChangeStore,
	persistenceCalls int,
	persistedAssignments map[uuid.UUID]uuid.UUID,
	wantAssignments map[uuid.UUID]uuid.UUID,
	wantBefore map[uuid.UUID]domainHistory.JSONB,
	wantTable string,
) {
	t.Helper()

	if changes.loadCalls != 1 || persistenceCalls != 1 || changes.recordCalls != 1 {
		t.Fatalf(
			"batch calls: history load=%d persistence=%d history record=%d",
			changes.loadCalls,
			persistenceCalls,
			changes.recordCalls,
		)
	}
	if want := []string{"history:load", "persistence:assign", "history:record"}; !reflect.DeepEqual(changes.order, want) {
		t.Fatalf("call order: got %v, want %v", changes.order, want)
	}
	if changes.loadedTable != wantTable || changes.recordedTable != wantTable {
		t.Fatalf("history tables: loaded=%q recorded=%q want=%q", changes.loadedTable, changes.recordedTable, wantTable)
	}
	if !reflect.DeepEqual(persistedAssignments, wantAssignments) {
		t.Fatalf("persisted assignments: got %v, want %v", persistedAssignments, wantAssignments)
	}
	if !reflect.DeepEqual(changes.recordedBeforeRows, wantBefore) {
		t.Fatalf("recorded before rows: got %v, want %v", changes.recordedBeforeRows, wantBefore)
	}
	if !reflect.DeepEqual(uuidSet(changes.loadedIDs), uuidSet(mapKeys(wantAssignments))) {
		t.Fatalf("loaded IDs: got %v, want keys of %v", changes.loadedIDs, wantAssignments)
	}
}

func cloneUUIDMap(source map[uuid.UUID]uuid.UUID) map[uuid.UUID]uuid.UUID {
	clone := make(map[uuid.UUID]uuid.UUID, len(source))
	for id, targetID := range source {
		clone[id] = targetID
	}
	return clone
}

func uuidSet(ids []uuid.UUID) map[uuid.UUID]struct{} {
	set := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		set[id] = struct{}{}
	}
	return set
}

var _ domainFieldDevice.FieldDeviceStore = (*fieldDeviceAssignmentStore)(nil)
var _ domainObjectData.BacnetObjectStore = (*bacnetReferenceAssignmentStore)(nil)
var _ domainFacility.BacnetObjectRepository = (*bacnetReferenceAssignmentStore)(nil)
