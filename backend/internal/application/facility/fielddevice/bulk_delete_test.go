package fielddevice

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainCollaboration "github.com/besart951/go_infra_link/backend/internal/domain/collaboration"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type bulkDeleteHistoryBatchKey struct{}

type bulkDeleteTransactionState struct {
	fieldDevices map[uuid.UUID]*domainFacility.FieldDevice
	history      []uuid.UUID
	lastAttempt  uuid.UUID
}

func (s bulkDeleteTransactionState) clone() bulkDeleteTransactionState {
	fieldDevices := make(map[uuid.UUID]*domainFacility.FieldDevice, len(s.fieldDevices))
	for id, fieldDevice := range s.fieldDevices {
		fieldDevices[id] = cloneFieldDevice(fieldDevice)
	}
	return bulkDeleteTransactionState{
		fieldDevices: fieldDevices,
		history:      append([]uuid.UUID(nil), s.history...),
		lastAttempt:  s.lastAttempt,
	}
}

type bulkDeleteTransactionUnit struct {
	state *bulkDeleteTransactionState
}

type bulkDeleteTransactionHarness struct {
	committed       bulkDeleteTransactionState
	deleteErrs      map[uuid.UUID]error
	commitErrs      map[uuid.UUID]error
	runnerCalls     int
	deleteCalls     []uuid.UUID
	historyBatchIDs []uuid.UUID
	outbox          domainCollaboration.OutboxStore
}

func (h *bulkDeleteTransactionHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	staged := h.committed.clone()
	txCtx := ctx
	if h.outbox != nil {
		txCtx = domainCollaboration.WithOutboxStore(ctx, h.outbox)
	}
	if err := run(txCtx, bulkDeleteTransactionUnit{state: &staged}); err != nil {
		return err
	}
	if commitErr := h.commitErrs[staged.lastAttempt]; commitErr != nil {
		return commitErr
	}
	h.committed = staged
	return nil
}

func (h *bulkDeleteTransactionHarness) factory(
	unit apptransaction.UnitOfWork,
) (BulkDeleteWorkflow, error) {
	typed, ok := unit.(bulkDeleteTransactionUnit)
	if !ok || typed.state == nil {
		return nil, errors.New("unexpected bulk-delete transaction unit")
	}
	return &bulkDeleteWorkflowStub{harness: h, state: typed.state}, nil
}

type bulkDeleteWorkflowStub struct {
	harness *bulkDeleteTransactionHarness
	state   *bulkDeleteTransactionState
}

type bulkDeleteOutboxWorkflowStub struct {
	*bulkDeleteWorkflowStub
	links []*domainProject.ProjectFieldDevice
}

func (s *bulkDeleteOutboxWorkflowStub) GetByFieldDeviceIDs(
	context.Context,
	[]uuid.UUID,
) ([]*domainProject.ProjectFieldDevice, error) {
	return s.links, nil
}

func (s *bulkDeleteWorkflowStub) DeleteByID(ctx context.Context, id uuid.UUID) error {
	s.harness.deleteCalls = append(s.harness.deleteCalls, id)
	s.state.lastAttempt = id
	if batchID, ok := ctx.Value(bulkDeleteHistoryBatchKey{}).(uuid.UUID); ok {
		s.harness.historyBatchIDs = append(s.harness.historyBatchIDs, batchID)
	}
	if fieldDevice := s.state.fieldDevices[id]; fieldDevice != nil {
		delete(s.state.fieldDevices, id)
		s.state.history = append(s.state.history, id)
	}
	return s.harness.deleteErrs[id]
}

type bulkDeleteSnapshotReaderStub struct {
	harness                 *bulkDeleteTransactionHarness
	err                     error
	calls                   int
	received                []uuid.UUID
	calledBeforeTransaction bool
}

func (s *bulkDeleteSnapshotReaderStub) GetByIds(
	_ context.Context,
	ids []uuid.UUID,
) ([]*domainFacility.FieldDevice, error) {
	s.calls++
	s.received = append([]uuid.UUID(nil), ids...)
	s.calledBeforeTransaction = s.harness.runnerCalls == 0
	if s.err != nil {
		return nil, s.err
	}
	items := make([]*domainFacility.FieldDevice, 0, len(ids))
	for _, id := range ids {
		if fieldDevice := s.harness.committed.fieldDevices[id]; fieldDevice != nil {
			items = append(items, cloneFieldDevice(fieldDevice))
		}
	}
	return items, nil
}

type bulkDeleteProjectLinkReaderStub struct {
	harness                 *bulkDeleteTransactionHarness
	links                   []*domainProject.ProjectFieldDevice
	err                     error
	calls                   int
	received                []uuid.UUID
	calledBeforeTransaction bool
}

func (s *bulkDeleteProjectLinkReaderStub) GetByFieldDeviceIDs(
	_ context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectFieldDevice, error) {
	s.calls++
	s.received = append([]uuid.UUID(nil), ids...)
	s.calledBeforeTransaction = s.harness.runnerCalls == 0
	return s.links, s.err
}

type bulkDeleteDispatcherStub struct {
	harness           *bulkDeleteTransactionHarness
	wantDeleted       []uuid.UUID
	wantRemaining     []uuid.UUID
	commands          []appcollaboration.Command
	err               error
	calledAfterCommit bool
}

func (s *bulkDeleteDispatcherStub) Dispatch(
	_ context.Context,
	command appcollaboration.Command,
) error {
	s.commands = append(s.commands, command)
	committed := true
	for _, id := range s.wantDeleted {
		_, exists := s.harness.committed.fieldDevices[id]
		committed = committed && !exists
	}
	for _, id := range s.wantRemaining {
		committed = committed && s.harness.committed.fieldDevices[id] != nil
	}
	s.calledAfterCommit = committed
	return s.err
}

func TestBulkDeletePreservesPartialResultsAndDispatchesOneRefreshPerProjectAfterCommits(t *testing.T) {
	fieldDeviceA := testUUID(401)
	fieldDeviceB := testUUID(402)
	missingFieldDevice := testUUID(403)
	fieldDeviceD := testUUID(404)
	parentA := testUUID(405)
	projectOne := testUUID(406)
	projectTwo := testUUID(407)
	unrelatedProject := testUUID(408)
	actorID := testUUID(409)
	operationID := testUUID(410)
	eventOne := testUUID(411)
	eventTwo := testUUID(412)
	deleteErr := errors.New("delete failed")
	commitErr := errors.New("commit failed")
	createdAt := time.Date(2026, time.July, 21, 12, 0, 0, 0, time.UTC)
	occurredAt := createdAt.Add(time.Hour)
	bmk := "M01"
	harness := &bulkDeleteTransactionHarness{
		committed: bulkDeleteTransactionState{
			fieldDevices: map[uuid.UUID]*domainFacility.FieldDevice{
				fieldDeviceA: {
					Base:                      domain.Base{ID: fieldDeviceA, CreatedAt: createdAt, UpdatedAt: createdAt},
					BMK:                       &bmk,
					ApparatNr:                 1,
					SPSControllerSystemTypeID: parentA,
					SystemPartID:              testUUID(413),
					ApparatID:                 testUUID(414),
				},
				fieldDeviceB: {Base: domain.Base{ID: fieldDeviceB}},
				fieldDeviceD: {Base: domain.Base{ID: fieldDeviceD}},
			},
		},
		deleteErrs: map[uuid.UUID]error{fieldDeviceB: deleteErr},
		commitErrs: map[uuid.UUID]error{fieldDeviceD: commitErr},
	}
	snapshots := &bulkDeleteSnapshotReaderStub{harness: harness}
	links := &bulkDeleteProjectLinkReaderStub{
		harness: harness,
		links: []*domainProject.ProjectFieldDevice{
			{ProjectID: projectTwo, FieldDeviceID: fieldDeviceA},
			{ProjectID: projectOne, FieldDeviceID: fieldDeviceA},
			{ProjectID: projectOne, FieldDeviceID: fieldDeviceB},
			{ProjectID: projectTwo, FieldDeviceID: fieldDeviceD},
			{ProjectID: unrelatedProject, FieldDeviceID: testUUID(499)},
		},
	}
	dispatcher := &bulkDeleteDispatcherStub{
		harness:       harness,
		wantDeleted:   []uuid.UUID{fieldDeviceA},
		wantRemaining: []uuid.UUID{fieldDeviceB, fieldDeviceD},
	}
	generatedIDs := []uuid.UUID{operationID, eventOne, eventTwo}
	handler := NewBulkDeleteHandler(BulkDeleteDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Snapshots:           snapshots,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, bulkDeleteHistoryBatchKey{}, batchID)
		},
		ProjectLinks: links,
		Dispatcher:   dispatcher,
		Actor:        func(context.Context) *uuid.UUID { return &actorID },
		NewID: func() uuid.UUID {
			id := generatedIDs[0]
			generatedIDs = generatedIDs[1:]
			return id
		},
		Now: func() time.Time { return occurredAt },
	})
	requestIDs := []uuid.UUID{
		fieldDeviceA,
		fieldDeviceB,
		missingFieldDevice,
		fieldDeviceA,
		fieldDeviceD,
		uuid.Nil,
	}

	outcome := handler.Execute(context.Background(), BulkDeleteCommand{FieldDeviceIDs: requestIDs})

	if snapshots.calls != 1 || links.calls != 1 || !snapshots.calledBeforeTransaction ||
		!links.calledBeforeTransaction || harness.runnerCalls != 4 ||
		!reflect.DeepEqual(harness.deleteCalls, []uuid.UUID{
			fieldDeviceA,
			fieldDeviceB,
			fieldDeviceA,
			fieldDeviceD,
		}) {
		t.Fatalf("set-based reads/transactions: snapshots=%+v links=%+v runner=%d deletes=%v",
			snapshots,
			links,
			harness.runnerCalls,
			harness.deleteCalls,
		)
	}
	wantCandidates := []uuid.UUID{fieldDeviceA, fieldDeviceB, missingFieldDevice, fieldDeviceD}
	if !reflect.DeepEqual(snapshots.received, wantCandidates) ||
		!reflect.DeepEqual(links.received, wantCandidates) {
		t.Fatalf("candidate IDs: snapshots=%v links=%v want=%v",
			snapshots.received,
			links.received,
			wantCandidates,
		)
	}
	if outcome.Result.TotalCount != len(requestIDs) || outcome.Result.SuccessCount != 2 ||
		outcome.Result.FailureCount != 4 || len(outcome.Result.Results) != len(requestIDs) {
		t.Fatalf("partial result counts: %+v", outcome.Result)
	}
	wantSuccess := []bool{true, false, false, true, false, false}
	for index, item := range outcome.Result.Results {
		if item.ID != requestIDs[index] || item.Success != wantSuccess[index] {
			t.Fatalf("result %d: %+v", index, item)
		}
	}
	if outcome.Result.Results[1].Error != deleteErr.Error() ||
		outcome.Result.Results[4].Error != commitErr.Error() {
		t.Fatalf("failure errors changed: %+v", outcome.Result.Results)
	}
	for _, index := range []int{2, 5} {
		item := outcome.Result.Results[index]
		if item.ErrorCode != itemErrorCodeNotFound ||
			item.ErrorField != "fielddevice.id" ||
			item.Reason != "field device not found" {
			t.Fatalf("missing result %d: %+v", index, item)
		}
	}
	if harness.committed.fieldDevices[fieldDeviceA] != nil ||
		harness.committed.fieldDevices[fieldDeviceB] == nil ||
		harness.committed.fieldDevices[fieldDeviceD] == nil ||
		!reflect.DeepEqual(harness.committed.history, []uuid.UUID{fieldDeviceA}) {
		t.Fatalf("partial transaction state: %+v", harness.committed)
	}
	if len(harness.historyBatchIDs) != harness.runnerCalls {
		t.Fatalf("history batch calls: %v", harness.historyBatchIDs)
	}
	for _, batchID := range harness.historyBatchIDs {
		if batchID != operationID {
			t.Fatalf("history batch ID: got %s, want %s", batchID, operationID)
		}
	}
	if outcome.Mutation.OperationID != operationID ||
		outcome.Mutation.BatchID == nil || *outcome.Mutation.BatchID != operationID ||
		outcome.Mutation.ActorID == nil || *outcome.Mutation.ActorID != actorID ||
		!outcome.Mutation.OccurredAt.Equal(occurredAt) ||
		!reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectOne, projectTwo}) ||
		!reflect.DeepEqual(outcome.ReconciliationIDs, []uuid.UUID{fieldDeviceA}) ||
		len(outcome.Mutation.Changes) != 1 {
		t.Fatalf("mutation outcome: %+v", outcome)
	}
	change := outcome.Mutation.Changes[0]
	if change.EntityID != fieldDeviceA || change.ParentID == nil || *change.ParentID != parentA ||
		change.Action != domainHistory.ActionDelete || len(change.After) != 0 {
		t.Fatalf("delete change: %+v", change)
	}
	var snapshot fieldDeviceSnapshot
	if err := json.Unmarshal(change.Before, &snapshot); err != nil {
		t.Fatalf("decode delete snapshot: %v", err)
	}
	if snapshot.ID != fieldDeviceA || snapshot.BMK == nil || *snapshot.BMK != bmk {
		t.Fatalf("delete snapshot: %+v", snapshot)
	}
	if len(dispatcher.commands) != 2 || !dispatcher.calledAfterCommit {
		t.Fatalf("dispatch: commands=%v afterCommit=%t", dispatcher.commands, dispatcher.calledAfterCommit)
	}
	for index, raw := range dispatcher.commands {
		command, ok := raw.(appcollaboration.FacilityHierarchyRefreshRequired)
		if !ok {
			t.Fatalf("command %d: got %T", index, raw)
		}
		wantProjectID := []uuid.UUID{projectOne, projectTwo}[index]
		wantEventID := []uuid.UUID{eventOne, eventTwo}[index]
		if command.ProjectID != wantProjectID || command.EventID != wantEventID ||
			command.OperationID != operationID || command.CorrelationID != operationID ||
			command.ActorID == nil || *command.ActorID != actorID ||
			command.Scope != appcollaboration.FacilityScopeFieldDevice || command.FullRefresh ||
			!reflect.DeepEqual(command.EntityIDs, []uuid.UUID{fieldDeviceA}) {
			t.Fatalf("command %d: %+v", index, command)
		}
	}
}

func TestBulkDeleteWritesPerItemVersionTwoOutboxInsideTransaction(t *testing.T) {
	fieldDeviceID := testUUID(801)
	projectID := testUUID(802)
	operationID := testUUID(803)
	outboxEventID := testUUID(804)
	compatibilityEventID := testUUID(805)
	occurredAt := time.Date(2026, time.July, 23, 17, 30, 0, 0, time.UTC)
	harness := &bulkDeleteTransactionHarness{
		committed: bulkDeleteTransactionState{fieldDevices: map[uuid.UUID]*domainFacility.FieldDevice{
			fieldDeviceID: {Base: domain.Base{ID: fieldDeviceID}},
		}},
	}
	store := &updateOutboxStoreStub{}
	harness.outbox = store
	factory := func(unit apptransaction.UnitOfWork) (BulkDeleteWorkflow, error) {
		typed := unit.(bulkDeleteTransactionUnit)
		return &bulkDeleteOutboxWorkflowStub{
			bulkDeleteWorkflowStub: &bulkDeleteWorkflowStub{harness: harness, state: typed.state},
			links: []*domainProject.ProjectFieldDevice{{
				ProjectID: projectID, FieldDeviceID: fieldDeviceID,
			}},
		}, nil
	}
	links := &bulkDeleteProjectLinkReaderStub{
		harness: harness,
		links: []*domainProject.ProjectFieldDevice{{
			ProjectID: projectID, FieldDeviceID: fieldDeviceID,
		}},
	}
	dispatcher := &bulkDeleteDispatcherStub{harness: harness, wantDeleted: []uuid.UUID{fieldDeviceID}}
	ids := []uuid.UUID{operationID, outboxEventID, compatibilityEventID}
	handler := NewBulkDeleteHandler(BulkDeleteDependencies{
		TransactionRunner: harness.runner, TransactionWorkflow: factory,
		Snapshots:    &bulkDeleteSnapshotReaderStub{harness: harness},
		ProjectLinks: links, Dispatcher: dispatcher,
		NewID: func() uuid.UUID { id := ids[0]; ids = ids[1:]; return id },
		Now:   func() time.Time { return occurredAt },
	})

	outcome := handler.Execute(context.Background(), BulkDeleteCommand{
		FieldDeviceIDs: []uuid.UUID{fieldDeviceID},
	})
	if outcome.Result.SuccessCount != 1 || len(store.events) != 1 {
		t.Fatalf("bulk delete outcome=%+v events=%d", outcome.Result, len(store.events))
	}
	decoded, err := appcollaboration.DecodeCommand(appcollaboration.EncodedCommand{
		Type: store.events[0].EventType, Payload: store.events[0].Payload,
	})
	if err != nil {
		t.Fatalf("decode queued command: %v", err)
	}
	refresh, ok := decoded.(appcollaboration.FacilityHierarchyRefreshRequired)
	if !ok || refresh.SchemaVersion != appcollaboration.SchemaVersionV2 ||
		refresh.EventID != outboxEventID || refresh.ProjectID != projectID ||
		!reflect.DeepEqual(refresh.EntityIDs, []uuid.UUID{fieldDeviceID}) {
		t.Fatalf("unexpected queued command: %#v", decoded)
	}
}

func TestBulkDeleteSnapshotFailureReturnsIndexAlignedFailuresWithoutWrites(t *testing.T) {
	readErr := errors.New("snapshot read failed")
	ids := []uuid.UUID{testUUID(421), testUUID(422)}
	harness := &bulkDeleteTransactionHarness{
		committed: bulkDeleteTransactionState{fieldDevices: map[uuid.UUID]*domainFacility.FieldDevice{}},
	}
	snapshots := &bulkDeleteSnapshotReaderStub{harness: harness, err: readErr}
	links := &bulkDeleteProjectLinkReaderStub{harness: harness}
	dispatcher := &bulkDeleteDispatcherStub{harness: harness}
	handler := NewBulkDeleteHandler(BulkDeleteDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Snapshots:           snapshots,
		ProjectLinks:        links,
		Dispatcher:          dispatcher,
	})

	outcome := handler.Execute(context.Background(), BulkDeleteCommand{FieldDeviceIDs: ids})

	if outcome.Result.TotalCount != 2 || outcome.Result.FailureCount != 2 ||
		outcome.Result.SuccessCount != 0 || harness.runnerCalls != 0 || links.calls != 0 ||
		len(dispatcher.commands) != 0 {
		t.Fatalf("snapshot failure outcome: %+v harness=%+v links=%+v", outcome, harness, links)
	}
	for index, item := range outcome.Result.Results {
		if item.ID != ids[index] || item.Success || item.Error != readErr.Error() {
			t.Fatalf("failure item %d: %+v", index, item)
		}
	}
}

func TestBulkDeleteScopeFailureIsReportedAfterCommittedPartialResult(t *testing.T) {
	fieldDeviceID := testUUID(431)
	scopeErr := errors.New("scope unavailable")
	harness := &bulkDeleteTransactionHarness{
		committed: bulkDeleteTransactionState{fieldDevices: map[uuid.UUID]*domainFacility.FieldDevice{
			fieldDeviceID: {Base: domain.Base{ID: fieldDeviceID}},
		}},
	}
	snapshots := &bulkDeleteSnapshotReaderStub{harness: harness}
	links := &bulkDeleteProjectLinkReaderStub{harness: harness, err: scopeErr}
	dispatcher := &bulkDeleteDispatcherStub{harness: harness}
	var reported []error
	handler := NewBulkDeleteHandler(BulkDeleteDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Snapshots:           snapshots,
		ProjectLinks:        links,
		Dispatcher:          dispatcher,
		ReportError:         func(err error) { reported = append(reported, err) },
	})

	result := handler.BulkDelete(
		context.Background(),
		BulkDeleteCommand{FieldDeviceIDs: []uuid.UUID{fieldDeviceID}},
	)

	if result.SuccessCount != 1 || harness.committed.fieldDevices[fieldDeviceID] != nil ||
		len(dispatcher.commands) != 0 || len(reported) != 1 ||
		!errors.Is(reported[0], scopeErr) {
		t.Fatalf("scope failure changed committed result: result=%+v state=%+v reported=%v",
			result,
			harness.committed,
			reported,
		)
	}
}

func TestBulkDeleteLargeProjectDeltaFallsBackToOneFullRefresh(t *testing.T) {
	projectID := testUUID(441)
	operationID := testUUID(442)
	ids := make([]uuid.UUID, defaultMaxTargetedRefreshIDs+1)
	fieldDevices := make(map[uuid.UUID]*domainFacility.FieldDevice, len(ids))
	linksForProject := make([]*domainProject.ProjectFieldDevice, len(ids))
	for index := range ids {
		ids[index] = testUUID(500 + index)
		fieldDevices[ids[index]] = &domainFacility.FieldDevice{Base: domain.Base{ID: ids[index]}}
		linksForProject[index] = &domainProject.ProjectFieldDevice{
			ProjectID:     projectID,
			FieldDeviceID: ids[index],
		}
	}
	harness := &bulkDeleteTransactionHarness{
		committed: bulkDeleteTransactionState{fieldDevices: fieldDevices},
	}
	snapshots := &bulkDeleteSnapshotReaderStub{harness: harness}
	links := &bulkDeleteProjectLinkReaderStub{harness: harness, links: linksForProject}
	dispatcher := &bulkDeleteDispatcherStub{harness: harness, wantDeleted: ids}
	generatedIDs := []uuid.UUID{operationID, testUUID(443)}
	handler := NewBulkDeleteHandler(BulkDeleteDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Snapshots:           snapshots,
		ProjectLinks:        links,
		Dispatcher:          dispatcher,
		NewID: func() uuid.UUID {
			id := generatedIDs[0]
			generatedIDs = generatedIDs[1:]
			return id
		},
	})

	outcome := handler.Execute(context.Background(), BulkDeleteCommand{FieldDeviceIDs: ids})

	if outcome.Result.SuccessCount != len(ids) || len(dispatcher.commands) != 1 ||
		!dispatcher.calledAfterCommit {
		t.Fatalf("large delete outcome: result=%+v commands=%v", outcome.Result, dispatcher.commands)
	}
	command, ok := dispatcher.commands[0].(appcollaboration.FacilityHierarchyRefreshRequired)
	if !ok || command.ProjectID != projectID || !command.FullRefresh || len(command.EntityIDs) != 0 {
		t.Fatalf("large delete command: %+v (%T)", dispatcher.commands[0], dispatcher.commands[0])
	}
}

func TestBulkDeleteMissingConfigurationReturnsIndexAlignedFailures(t *testing.T) {
	ids := []uuid.UUID{testUUID(451), testUUID(452)}
	handler := NewBulkDeleteHandler(BulkDeleteDependencies{})

	result := handler.BulkDelete(
		context.Background(),
		BulkDeleteCommand{FieldDeviceIDs: ids},
	)

	if result.TotalCount != 2 || result.SuccessCount != 0 || result.FailureCount != 2 ||
		len(result.Results) != 2 {
		t.Fatalf("configuration result: %+v", result)
	}
	for index, item := range result.Results {
		if item.ID != ids[index] || item.Success || item.Error == "" {
			t.Fatalf("configuration item %d: %+v", index, item)
		}
	}
}
