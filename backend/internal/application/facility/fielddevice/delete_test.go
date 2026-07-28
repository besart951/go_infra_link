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
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type deleteHistoryBatchKey struct{}

type deleteTransactionState struct {
	fieldDevice *domainFacility.FieldDevice
	history     []string
}

func (s deleteTransactionState) clone() deleteTransactionState {
	return deleteTransactionState{
		fieldDevice: cloneFieldDevice(s.fieldDevice),
		history:     append([]string(nil), s.history...),
	}
}

type deleteTransactionUnit struct {
	state *deleteTransactionState
}

type deleteTransactionHarness struct {
	committed      deleteTransactionState
	deleteErr      error
	commitErr      error
	runnerCalls    int
	deleteCalls    int
	historyBatchID *uuid.UUID
}

func (h *deleteTransactionHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	staged := h.committed.clone()
	if err := run(ctx, deleteTransactionUnit{state: &staged}); err != nil {
		return err
	}
	if h.commitErr != nil {
		return h.commitErr
	}
	h.committed = staged
	return nil
}

func (h *deleteTransactionHarness) factory(
	unit apptransaction.UnitOfWork,
) (DeleteWorkflow, error) {
	typed, ok := unit.(deleteTransactionUnit)
	if !ok || typed.state == nil {
		return nil, errors.New("unexpected transaction unit")
	}
	return &deleteWorkflowStub{harness: h, state: typed.state}, nil
}

type deleteWorkflowStub struct {
	harness *deleteTransactionHarness
	state   *deleteTransactionState
}

func (s *deleteWorkflowStub) GetByID(
	_ context.Context,
	id uuid.UUID,
) (*domainFacility.FieldDevice, error) {
	if s.state.fieldDevice == nil || s.state.fieldDevice.ID != id {
		return nil, domain.ErrNotFound
	}
	return cloneFieldDevice(s.state.fieldDevice), nil
}

func (s *deleteWorkflowStub) DeleteByID(ctx context.Context, id uuid.UUID) error {
	s.harness.deleteCalls++
	if batchID, ok := ctx.Value(deleteHistoryBatchKey{}).(uuid.UUID); ok {
		s.harness.historyBatchID = &batchID
	}
	if s.state.fieldDevice != nil && s.state.fieldDevice.ID == id {
		s.state.fieldDevice = nil
		s.state.history = append(s.state.history, "field_device:delete")
	}
	return s.harness.deleteErr
}

type deleteProjectLinkReaderStub struct {
	harness                 *deleteTransactionHarness
	links                   []*domainProject.ProjectFieldDevice
	err                     error
	calls                   int
	received                []uuid.UUID
	calledBeforeTransaction bool
}

func (s *deleteProjectLinkReaderStub) GetByFieldDeviceIDs(
	_ context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectFieldDevice, error) {
	s.calls++
	s.received = append([]uuid.UUID(nil), ids...)
	s.calledBeforeTransaction = s.harness == nil ||
		(s.harness.runnerCalls == 0 && s.harness.committed.fieldDevice != nil)
	return s.links, s.err
}

func TestDeleteCapturesRecipientsAndHistoryBeforeCommittedDispatch(t *testing.T) {
	fieldDeviceID := testUUID(201)
	parentID := testUUID(202)
	projectOne := testUUID(211)
	projectTwo := testUUID(212)
	actorID := testUUID(221)
	operationID := testUUID(231)
	eventOne := testUUID(232)
	eventTwo := testUUID(233)
	createdAt := time.Date(2026, time.July, 20, 10, 0, 0, 0, time.UTC)
	occurredAt := createdAt.Add(time.Hour)
	bmk := "FD-01"

	harness := &deleteTransactionHarness{
		committed: deleteTransactionState{
			fieldDevice: &domainFacility.FieldDevice{
				Base:                      domain.Base{ID: fieldDeviceID, CreatedAt: createdAt, UpdatedAt: createdAt},
				BMK:                       &bmk,
				ApparatNr:                 7,
				SPSControllerSystemTypeID: parentID,
				SystemPartID:              testUUID(203),
				ApparatID:                 testUUID(204),
			},
		},
	}
	links := &deleteProjectLinkReaderStub{
		harness: harness,
		links: []*domainProject.ProjectFieldDevice{
			{ProjectID: projectTwo, FieldDeviceID: fieldDeviceID},
			{ProjectID: projectOne, FieldDeviceID: fieldDeviceID},
			{ProjectID: projectOne, FieldDeviceID: fieldDeviceID},
			{ProjectID: testUUID(213), FieldDeviceID: testUUID(299)},
		},
	}
	dispatcher := &updateCommandDispatcherStub{}
	generatedIDs := []uuid.UUID{operationID, eventOne, eventTwo}
	handler := NewDeleteHandler(DeleteDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, deleteHistoryBatchKey{}, batchID)
		},
		ProjectLinks: links,
		Dispatcher:   dispatcher,
		Actor: func(context.Context) *uuid.UUID {
			return &actorID
		},
		NewID: func() uuid.UUID {
			id := generatedIDs[0]
			generatedIDs = generatedIDs[1:]
			return id
		},
		Now: func() time.Time { return occurredAt },
	})

	outcome, err := handler.Execute(
		context.Background(),
		DeleteCommand{FieldDeviceID: fieldDeviceID},
	)
	if err != nil {
		t.Fatalf("execute delete: %v", err)
	}
	if harness.runnerCalls != 1 || harness.deleteCalls != 1 || links.calls != 1 {
		t.Fatalf("unexpected workflow calls: %+v", harness)
	}
	if !links.calledBeforeTransaction {
		t.Fatal("project recipients were not captured before deletion")
	}
	if want := []uuid.UUID{fieldDeviceID}; !reflect.DeepEqual(links.received, want) {
		t.Fatalf("recipient lookup IDs: got %v, want %v", links.received, want)
	}
	if harness.committed.fieldDevice != nil || !reflect.DeepEqual(harness.committed.history, []string{"field_device:delete"}) {
		t.Fatalf("unexpected committed state: %+v", harness.committed)
	}
	if harness.historyBatchID == nil || *harness.historyBatchID != operationID {
		t.Fatalf("history batch: got %v, want %s", harness.historyBatchID, operationID)
	}
	if !outcome.Existed || outcome.Mutation.OperationID != operationID ||
		outcome.Mutation.BatchID == nil || *outcome.Mutation.BatchID != operationID ||
		outcome.Mutation.ActorID == nil || *outcome.Mutation.ActorID != actorID ||
		!outcome.Mutation.OccurredAt.Equal(occurredAt) {
		t.Fatalf("unexpected outcome: %+v", outcome)
	}
	if want := []uuid.UUID{projectOne, projectTwo}; !reflect.DeepEqual(outcome.Mutation.ProjectIDs, want) {
		t.Fatalf("project IDs: got %v, want %v", outcome.Mutation.ProjectIDs, want)
	}
	if len(outcome.Mutation.Changes) != 1 {
		t.Fatalf("changes: got %d, want 1", len(outcome.Mutation.Changes))
	}
	change := outcome.Mutation.Changes[0]
	if change.EntityID != fieldDeviceID || change.ParentID == nil || *change.ParentID != parentID ||
		change.Action != domainHistory.ActionDelete || len(change.After) != 0 {
		t.Fatalf("unexpected delete change: %+v", change)
	}
	var snapshot fieldDeviceSnapshot
	if err := json.Unmarshal(change.Before, &snapshot); err != nil {
		t.Fatalf("decode before snapshot: %v", err)
	}
	if snapshot.ID != fieldDeviceID || snapshot.SPSControllerSystemTypeID != parentID ||
		snapshot.BMK == nil || *snapshot.BMK != bmk {
		t.Fatalf("unexpected before snapshot: %+v", snapshot)
	}
	if len(dispatcher.commands) != 2 {
		t.Fatalf("commands: got %d, want 2", len(dispatcher.commands))
	}
	for index, raw := range dispatcher.commands {
		command, ok := raw.(appcollaboration.FieldDeviceDeleted)
		if !ok {
			t.Fatalf("command: got %T, want FieldDeviceDeleted", raw)
		}
		wantProjectID := []uuid.UUID{projectOne, projectTwo}[index]
		if command.ProjectID != wantProjectID || command.FieldDeviceID != fieldDeviceID ||
			command.SPSControllerSystemTypeID != parentID || command.OperationID != operationID ||
			command.CorrelationID != operationID || command.SchemaVersion != appcollaboration.SchemaVersionV2 {
			t.Fatalf("unexpected collaboration command: %+v", command)
		}
	}
}

func TestDeleteFailureOrCommitFailureDoesNotDispatchOrEscapeTransaction(t *testing.T) {
	deleteErr := errors.New("delete failed")
	commitErr := errors.New("commit failed")
	for _, test := range []struct {
		name      string
		deleteErr error
		commitErr error
	}{
		{name: "delete", deleteErr: deleteErr},
		{name: "commit", commitErr: commitErr},
	} {
		t.Run(test.name, func(t *testing.T) {
			fieldDeviceID := testUUID(241)
			harness := &deleteTransactionHarness{
				committed: deleteTransactionState{
					fieldDevice: &domainFacility.FieldDevice{
						Base:                      domain.Base{ID: fieldDeviceID},
						SPSControllerSystemTypeID: testUUID(242),
					},
				},
				deleteErr: test.deleteErr,
				commitErr: test.commitErr,
			}
			links := &deleteProjectLinkReaderStub{
				harness: harness,
				links: []*domainProject.ProjectFieldDevice{{
					ProjectID:     testUUID(243),
					FieldDeviceID: fieldDeviceID,
				}},
			}
			dispatcher := &updateCommandDispatcherStub{}
			handler := NewDeleteHandler(DeleteDependencies{
				TransactionRunner:   harness.runner,
				TransactionWorkflow: harness.factory,
				ProjectLinks:        links,
				Dispatcher:          dispatcher,
			})

			_, err := handler.Execute(context.Background(), DeleteCommand{FieldDeviceID: fieldDeviceID})
			wantErr := test.deleteErr
			if wantErr == nil {
				wantErr = test.commitErr
			}
			if !errors.Is(err, wantErr) {
				t.Fatalf("error: got %v, want %v", err, wantErr)
			}
			if harness.committed.fieldDevice == nil || len(harness.committed.history) != 0 {
				t.Fatalf("failed transaction escaped: %+v", harness.committed)
			}
			if len(dispatcher.commands) != 0 {
				t.Fatalf("dispatched after rollback: %+v", dispatcher.commands)
			}
		})
	}
}

func TestDeleteMissingFieldDevicePreservesIdempotentSuccess(t *testing.T) {
	harness := &deleteTransactionHarness{}
	links := &deleteProjectLinkReaderStub{harness: harness}
	dispatcher := &updateCommandDispatcherStub{}
	handler := NewDeleteHandler(DeleteDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		ProjectLinks:        links,
		Dispatcher:          dispatcher,
	})

	outcome, err := handler.Execute(
		context.Background(),
		DeleteCommand{FieldDeviceID: testUUID(251)},
	)
	if err != nil {
		t.Fatalf("missing delete: %v", err)
	}
	if outcome.Existed || len(outcome.Mutation.Changes) != 0 || harness.deleteCalls != 1 ||
		links.calls != 1 || len(dispatcher.commands) != 0 {
		t.Fatalf("unexpected missing-row outcome: outcome=%+v harness=%+v", outcome, harness)
	}
}

func TestDeleteRecipientLookupFailureIsBestEffortAfterCommit(t *testing.T) {
	fieldDeviceID := testUUID(261)
	scopeErr := errors.New("scope lookup failed")
	harness := &deleteTransactionHarness{
		committed: deleteTransactionState{
			fieldDevice: &domainFacility.FieldDevice{
				Base:                      domain.Base{ID: fieldDeviceID},
				SPSControllerSystemTypeID: testUUID(262),
			},
		},
	}
	links := &deleteProjectLinkReaderStub{
		harness: harness,
		err:     scopeErr,
	}
	dispatcher := &updateCommandDispatcherStub{}
	handler := NewDeleteHandler(DeleteDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		ProjectLinks:        links,
		Dispatcher:          dispatcher,
	})

	outcome, err := handler.Execute(
		context.Background(),
		DeleteCommand{FieldDeviceID: fieldDeviceID},
	)
	if err != nil {
		t.Fatalf("delete with scope failure: %v", err)
	}
	if harness.committed.fieldDevice != nil || !outcome.Existed {
		t.Fatalf("delete did not commit: outcome=%+v state=%+v", outcome, harness.committed)
	}
	if len(outcome.DispatchErrors) != 1 || !errors.Is(outcome.DispatchErrors[0], scopeErr) {
		t.Fatalf("dispatch errors: got %v, want wrapped %v", outcome.DispatchErrors, scopeErr)
	}
	if len(dispatcher.commands) != 0 {
		t.Fatalf("unexpected commands after scope failure: %+v", dispatcher.commands)
	}
}
