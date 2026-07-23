package bacnetobject

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	"github.com/besart951/go_infra_link/backend/internal/application/facility/mutation"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type alarmValueTransactionState struct {
	values  []domainFacility.BacnetObjectAlarmValue
	history []string
}

func (s alarmValueTransactionState) clone() alarmValueTransactionState {
	return alarmValueTransactionState{
		values:  cloneAlarmValues(s.values),
		history: append([]string(nil), s.history...),
	}
}

type alarmValueTransactionUnit struct {
	state *alarmValueTransactionState
}

type alarmValueHistoryBatchKey struct{}

type alarmValueTransactionHarness struct {
	committed      alarmValueTransactionState
	createdIDs     []uuid.UUID
	createdAt      time.Time
	writeErr       error
	reloadErr      error
	commitErr      error
	runnerCalls    int
	putCalls       int
	getCalls       int
	historyBatchID *uuid.UUID
}

func (h *alarmValueTransactionHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	staged := h.committed.clone()
	if err := run(ctx, alarmValueTransactionUnit{state: &staged}); err != nil {
		return err
	}
	if h.commitErr != nil {
		return h.commitErr
	}
	h.committed = staged
	return nil
}

func (h *alarmValueTransactionHarness) factory(
	unit apptransaction.UnitOfWork,
) (ReplaceAlarmValuesWorkflow, error) {
	typed, ok := unit.(alarmValueTransactionUnit)
	if !ok || typed.state == nil {
		return nil, errors.New("unexpected alarm-value transaction unit")
	}
	return &alarmValueWorkflowStub{harness: h, state: typed.state}, nil
}

type alarmValueWorkflowStub struct {
	harness  *alarmValueTransactionHarness
	state    *alarmValueTransactionState
	getCalls int
}

func (s *alarmValueWorkflowStub) GetValues(
	_ context.Context,
	_ uuid.UUID,
) ([]domainFacility.BacnetObjectAlarmValue, error) {
	s.getCalls++
	s.harness.getCalls++
	if s.getCalls > 1 && s.harness.reloadErr != nil {
		return nil, s.harness.reloadErr
	}
	return cloneAlarmValues(s.state.values), nil
}

func (s *alarmValueWorkflowStub) PutValues(
	ctx context.Context,
	bacnetObjectID uuid.UUID,
	values []domainFacility.BacnetObjectAlarmValue,
) error {
	s.harness.putCalls++
	if batchID, ok := ctx.Value(alarmValueHistoryBatchKey{}).(uuid.UUID); ok {
		s.harness.historyBatchID = &batchID
	}
	for range s.state.values {
		s.state.history = append(s.state.history, "alarm_value:delete")
	}
	s.state.values = nil
	for i := range values {
		value := values[i]
		value.BacnetObjectID = bacnetObjectID
		if i < len(s.harness.createdIDs) {
			value.ID = s.harness.createdIDs[i]
		}
		value.CreatedAt = s.harness.createdAt
		value.UpdatedAt = s.harness.createdAt
		s.state.values = append(s.state.values, value)
		s.state.history = append(s.state.history, "alarm_value:create")
	}
	return s.harness.writeErr
}

type alarmValueBacnetObjectReaderStub struct {
	harness        *alarmValueTransactionHarness
	expectedNewID  uuid.UUID
	objects        []*domainFacility.BacnetObject
	err            error
	calls          int
	received       []uuid.UUID
	commitObserved bool
}

func (s *alarmValueBacnetObjectReaderStub) GetByIds(
	_ context.Context,
	ids []uuid.UUID,
) ([]*domainFacility.BacnetObject, error) {
	s.calls++
	s.received = append([]uuid.UUID(nil), ids...)
	if s.harness != nil && s.expectedNewID != uuid.Nil &&
		len(s.harness.committed.values) == 1 &&
		s.harness.committed.values[0].ID == s.expectedNewID &&
		len(s.harness.committed.history) == 2 {
		s.commitObserved = true
	}
	return s.objects, s.err
}

func TestReplaceAlarmValuesCommitsHistoryAndReloadBeforeDirectProjectDispatch(t *testing.T) {
	bacnetObjectID := bacnetTestUUID(1)
	fieldDeviceID := bacnetTestUUID(2)
	alarmFieldID := bacnetTestUUID(3)
	oldValueID := bacnetTestUUID(4)
	newValueID := bacnetTestUUID(5)
	projectID := bacnetTestUUID(6)
	actorID := bacnetTestUUID(7)
	operationID := bacnetTestUUID(8)
	eventID := bacnetTestUUID(9)
	createdAt := time.Date(2026, time.July, 20, 23, 0, 0, 0, time.UTC)
	occurredAt := createdAt.Add(time.Second)
	valueNumber := 42.5

	harness := &alarmValueTransactionHarness{
		committed: alarmValueTransactionState{values: []domainFacility.BacnetObjectAlarmValue{{
			Base:             domain.Base{ID: oldValueID},
			BacnetObjectID:   bacnetObjectID,
			AlarmTypeFieldID: alarmFieldID,
			ValueNumber:      float64Pointer(1.5),
			Source:           domainFacility.AlarmValueSourceDefault,
		}}},
		createdIDs: []uuid.UUID{newValueID},
		createdAt:  createdAt,
	}
	objects := &alarmValueBacnetObjectReaderStub{
		harness:       harness,
		expectedNewID: newValueID,
		objects: []*domainFacility.BacnetObject{{
			Base:          domain.Base{ID: bacnetObjectID},
			FieldDeviceID: &fieldDeviceID,
		}},
	}
	links := &updateProjectLinkReaderStub{links: []*domainProject.ProjectFieldDevice{
		{ProjectID: projectID, FieldDeviceID: fieldDeviceID},
		{ProjectID: bacnetTestUUID(99), FieldDeviceID: bacnetTestUUID(98)},
	}}
	owners := &updateObjectDataOwnerReaderStub{}
	dispatcher := &updateCommandDispatcherStub{}
	ids := []uuid.UUID{operationID, eventID}
	handler := NewReplaceAlarmValuesHandler(ReplaceAlarmValuesDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, alarmValueHistoryBatchKey{}, batchID)
		},
		BacnetObjects:    objects,
		ProjectLinks:     links,
		ObjectDataOwners: owners,
		Dispatcher:       dispatcher,
		Actor:            func(context.Context) *uuid.UUID { return &actorID },
		NewID: func() uuid.UUID {
			id := ids[0]
			ids = ids[1:]
			return id
		},
		Now: func() time.Time { return occurredAt },
	})

	outcome, err := handler.Execute(context.Background(), ReplaceAlarmValuesCommand{
		BacnetObjectID: bacnetObjectID,
		Values: []AlarmValueInput{{
			AlarmTypeFieldID: alarmFieldID,
			ValueNumber:      &valueNumber,
		}},
	})
	if err != nil {
		t.Fatalf("replace alarm values: %v", err)
	}
	if !objects.commitObserved {
		t.Fatal("recipient resolution ran before value/history commit")
	}
	if !reflect.DeepEqual(harness.committed.history, []string{
		"alarm_value:delete",
		"alarm_value:create",
	}) {
		t.Fatalf("committed history: %v", harness.committed.history)
	}
	if harness.historyBatchID == nil || *harness.historyBatchID != operationID ||
		outcome.Mutation.BatchID == nil || *outcome.Mutation.BatchID != operationID {
		t.Fatalf("batch correlation: workflow=%v result=%v", harness.historyBatchID, outcome.Mutation.BatchID)
	}
	if outcome.Mutation.ActorID == nil || *outcome.Mutation.ActorID != actorID ||
		outcome.Mutation.OperationID != operationID ||
		!outcome.Mutation.OccurredAt.Equal(occurredAt) {
		t.Fatalf("mutation envelope: %+v", outcome.Mutation)
	}
	if len(outcome.Values) != 1 || outcome.Values[0].ID != newValueID ||
		outcome.Values[0].Source != domainFacility.AlarmValueSourceUser ||
		outcome.Values[0].ValueNumber == nil || *outcome.Values[0].ValueNumber != valueNumber {
		t.Fatalf("authoritative values: %+v", outcome.Values)
	}
	if len(outcome.Mutation.Changes) != 2 {
		t.Fatalf("changes: got %d, want delete/create", len(outcome.Mutation.Changes))
	}
	assertAlarmValueChange(t, outcome.Mutation.Changes[0], oldValueID, bacnetObjectID, domainHistory.ActionDelete)
	assertAlarmValueChange(t, outcome.Mutation.Changes[1], newValueID, bacnetObjectID, domainHistory.ActionCreate)
	if !reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectID}) ||
		objects.calls != 1 || links.calls != 1 || owners.calls != 1 ||
		len(dispatcher.commands) != 1 {
		t.Fatalf("post-commit scope: projects=%v objectCalls=%d linkCalls=%d ownerCalls=%d commands=%v",
			outcome.Mutation.ProjectIDs,
			objects.calls,
			links.calls,
			owners.calls,
			dispatcher.commands,
		)
	}
	command, ok := dispatcher.commands[0].(appcollaboration.BacnetObjectUpdated)
	if !ok || command.BacnetObjectID != bacnetObjectID ||
		!reflect.DeepEqual(command.FieldDeviceIDs, []uuid.UUID{fieldDeviceID}) ||
		command.ProjectID != projectID || command.OperationID != operationID ||
		command.EventID != eventID || command.ActorID == nil || *command.ActorID != actorID {
		t.Fatalf("collaboration command: %+v", dispatcher.commands[0])
	}
}

func TestReplaceAlarmValuesRollsBackValuesAndHistoryWithoutPostCommitWork(t *testing.T) {
	writeErr := errors.New("replace failed")
	reloadErr := errors.New("reload failed")
	commitErr := errors.New("commit failed")
	tests := []struct {
		name      string
		writeErr  error
		reloadErr error
		commitErr error
		wantErr   error
	}{
		{name: "write", writeErr: writeErr, wantErr: writeErr},
		{name: "reload", reloadErr: reloadErr, wantErr: reloadErr},
		{name: "commit", commitErr: commitErr, wantErr: commitErr},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			bacnetObjectID := bacnetTestUUID(21)
			oldValueID := bacnetTestUUID(22)
			harness := &alarmValueTransactionHarness{
				committed: alarmValueTransactionState{values: []domainFacility.BacnetObjectAlarmValue{{
					Base:           domain.Base{ID: oldValueID},
					BacnetObjectID: bacnetObjectID,
				}}},
				createdIDs: []uuid.UUID{bacnetTestUUID(23)},
				writeErr:   test.writeErr,
				reloadErr:  test.reloadErr,
				commitErr:  test.commitErr,
			}
			objects := &alarmValueBacnetObjectReaderStub{}
			links := &updateProjectLinkReaderStub{}
			owners := &updateObjectDataOwnerReaderStub{}
			dispatcher := &updateCommandDispatcherStub{}
			handler := NewReplaceAlarmValuesHandler(ReplaceAlarmValuesDependencies{
				TransactionRunner:   harness.runner,
				TransactionWorkflow: harness.factory,
				BacnetObjects:       objects,
				ProjectLinks:        links,
				ObjectDataOwners:    owners,
				Dispatcher:          dispatcher,
			})

			_, err := handler.Execute(context.Background(), ReplaceAlarmValuesCommand{
				BacnetObjectID: bacnetObjectID,
				Values:         []AlarmValueInput{{AlarmTypeFieldID: bacnetTestUUID(24)}},
			})
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error: got %v, want %v", err, test.wantErr)
			}
			if test.reloadErr != nil {
				var typed *AlarmValuesReloadError
				if !errors.As(err, &typed) {
					t.Fatalf("reload error is not typed: %v", err)
				}
			}
			if len(harness.committed.values) != 1 || harness.committed.values[0].ID != oldValueID ||
				len(harness.committed.history) != 0 {
				t.Fatalf("failed mutation escaped transaction: %+v", harness.committed)
			}
			if objects.calls != 0 || links.calls != 0 || owners.calls != 0 ||
				len(dispatcher.commands) != 0 {
				t.Fatalf("post-commit work ran after rollback: objects=%d links=%d owners=%d commands=%d",
					objects.calls,
					links.calls,
					owners.calls,
					len(dispatcher.commands),
				)
			}
		})
	}
}

func TestReplaceAlarmValuesUsesOneBroadRefreshForOverlappingObjectDataProject(t *testing.T) {
	bacnetObjectID := bacnetTestUUID(31)
	fieldDeviceID := bacnetTestUUID(32)
	objectDataID := bacnetTestUUID(33)
	projectID := bacnetTestUUID(34)
	globalObjectDataID := bacnetTestUUID(35)
	harness := &alarmValueTransactionHarness{
		createdIDs: []uuid.UUID{bacnetTestUUID(36)},
	}
	objects := &alarmValueBacnetObjectReaderStub{objects: []*domainFacility.BacnetObject{{
		Base:          domain.Base{ID: bacnetObjectID},
		FieldDeviceID: &fieldDeviceID,
	}}}
	links := &updateProjectLinkReaderStub{links: []*domainProject.ProjectFieldDevice{{
		ProjectID:     projectID,
		FieldDeviceID: fieldDeviceID,
	}}}
	owners := &updateObjectDataOwnerReaderStub{owners: []domainObjectData.BacnetObjectOwner{
		{BacnetObjectID: bacnetObjectID, ObjectDataID: objectDataID, ProjectID: &projectID},
		{BacnetObjectID: bacnetObjectID, ObjectDataID: objectDataID, ProjectID: &projectID},
		{BacnetObjectID: bacnetObjectID, ObjectDataID: globalObjectDataID},
		{BacnetObjectID: bacnetTestUUID(99), ObjectDataID: objectDataID, ProjectID: &projectID},
	}}
	dispatcher := &updateCommandDispatcherStub{}
	handler := NewReplaceAlarmValuesHandler(ReplaceAlarmValuesDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		BacnetObjects:       objects,
		ProjectLinks:        links,
		ObjectDataOwners:    owners,
		Dispatcher:          dispatcher,
	})

	outcome, err := handler.Execute(context.Background(), ReplaceAlarmValuesCommand{
		BacnetObjectID: bacnetObjectID,
		Values:         []AlarmValueInput{{AlarmTypeFieldID: bacnetTestUUID(37)}},
	})
	if err != nil {
		t.Fatalf("replace alarm values: %v", err)
	}
	if !reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectID}) ||
		len(dispatcher.commands) != 1 {
		t.Fatalf("deduplicated recipients: projects=%v commands=%v", outcome.Mutation.ProjectIDs, dispatcher.commands)
	}
	command, ok := dispatcher.commands[0].(appcollaboration.FacilityHierarchyRefreshRequired)
	if !ok || command.ProjectID != projectID ||
		command.Scope != appcollaboration.FacilityScopeProject || !command.FullRefresh ||
		len(command.EntityIDs) != 0 {
		t.Fatalf("expected one broad project refresh, got %+v", dispatcher.commands[0])
	}
}

func TestReplaceAlarmValuesKeepsKnownObjectDataRecipientWhenCurrentOwnerReadFails(t *testing.T) {
	bacnetObjectID := bacnetTestUUID(41)
	objectDataID := bacnetTestUUID(42)
	projectID := bacnetTestUUID(43)
	ownerReadErr := errors.New("BACnet owner unavailable")
	harness := &alarmValueTransactionHarness{createdIDs: []uuid.UUID{bacnetTestUUID(44)}}
	objects := &alarmValueBacnetObjectReaderStub{err: ownerReadErr}
	owners := &updateObjectDataOwnerReaderStub{owners: []domainObjectData.BacnetObjectOwner{{
		BacnetObjectID: bacnetObjectID,
		ObjectDataID:   objectDataID,
		ProjectID:      &projectID,
	}}}
	dispatcher := &updateCommandDispatcherStub{}
	handler := NewReplaceAlarmValuesHandler(ReplaceAlarmValuesDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		BacnetObjects:       objects,
		ObjectDataOwners:    owners,
		Dispatcher:          dispatcher,
	})

	outcome, err := handler.Execute(context.Background(), ReplaceAlarmValuesCommand{
		BacnetObjectID: bacnetObjectID,
		Values:         []AlarmValueInput{{AlarmTypeFieldID: bacnetTestUUID(45)}},
	})
	if err != nil {
		t.Fatalf("replace alarm values: %v", err)
	}
	if len(outcome.DispatchErrors) != 1 || !errors.Is(outcome.DispatchErrors[0], ownerReadErr) {
		t.Fatalf("dispatch errors: got %v, want wrapped %v", outcome.DispatchErrors, ownerReadErr)
	}
	if !reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectID}) ||
		len(dispatcher.commands) != 1 {
		t.Fatalf("known ObjectData recipient lost: projects=%v commands=%v", outcome.Mutation.ProjectIDs, dispatcher.commands)
	}
}

func TestReplaceEmptyAlarmValuesOnEmptyCollectionDoesNotDispatch(t *testing.T) {
	harness := &alarmValueTransactionHarness{}
	objects := &alarmValueBacnetObjectReaderStub{}
	owners := &updateObjectDataOwnerReaderStub{}
	dispatcher := &updateCommandDispatcherStub{}
	handler := NewReplaceAlarmValuesHandler(ReplaceAlarmValuesDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		BacnetObjects:       objects,
		ObjectDataOwners:    owners,
		Dispatcher:          dispatcher,
	})

	outcome, err := handler.Execute(context.Background(), ReplaceAlarmValuesCommand{
		BacnetObjectID: bacnetTestUUID(51),
		Values:         []AlarmValueInput{},
	})
	if err != nil {
		t.Fatalf("replace empty alarm values: %v", err)
	}
	if len(outcome.Mutation.Changes) != 0 || objects.calls != 0 || owners.calls != 0 ||
		len(dispatcher.commands) != 0 {
		t.Fatalf("empty no-op dispatched: changes=%v objects=%d owners=%d commands=%v",
			outcome.Mutation.Changes,
			objects.calls,
			owners.calls,
			dispatcher.commands,
		)
	}
}

func assertAlarmValueChange(
	t *testing.T,
	change mutation.EntityChange,
	entityID uuid.UUID,
	parentID uuid.UUID,
	action domainHistory.Action,
) {
	t.Helper()
	if change.EntityType != mutation.EntityTypeBacnetAlarmValue ||
		change.EntityID != entityID || change.Action != action ||
		change.ParentID == nil || *change.ParentID != parentID {
		t.Fatalf("alarm-value change: got %+v", change)
	}
	if action == domainHistory.ActionDelete {
		if len(change.Before) == 0 || len(change.After) != 0 {
			t.Fatalf("delete snapshots: before=%s after=%s", change.Before, change.After)
		}
		assertAlarmValueSnapshot(t, change.Before, entityID)
		return
	}
	if len(change.Before) != 0 || len(change.After) == 0 {
		t.Fatalf("create snapshots: before=%s after=%s", change.Before, change.After)
	}
	assertAlarmValueSnapshot(t, change.After, entityID)
}

func float64Pointer(value float64) *float64 {
	return &value
}

func assertAlarmValueSnapshot(t *testing.T, raw json.RawMessage, entityID uuid.UUID) {
	t.Helper()
	var snapshot alarmValueSnapshot
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatalf("decode alarm-value snapshot: %v", err)
	}
	if snapshot.ID != entityID {
		t.Fatalf("snapshot ID: got %s, want %s", snapshot.ID, entityID)
	}
}
