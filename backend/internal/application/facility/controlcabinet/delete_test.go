package controlcabinet

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
	cabinet     *domainFacility.ControlCabinet
	descendants []string
	history     []string
}

func (s deleteTransactionState) clone() deleteTransactionState {
	return deleteTransactionState{
		cabinet:     cloneControlCabinet(s.cabinet),
		descendants: append([]string(nil), s.descendants...),
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
		return nil, errors.New("unexpected cabinet delete transaction unit")
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
) (*domainFacility.ControlCabinet, error) {
	if s.state.cabinet == nil || s.state.cabinet.ID != id {
		return nil, domain.ErrNotFound
	}
	return cloneControlCabinet(s.state.cabinet), nil
}

func (s *deleteWorkflowStub) DeleteByID(ctx context.Context, id uuid.UUID) error {
	s.harness.deleteCalls++
	if batchID, ok := ctx.Value(deleteHistoryBatchKey{}).(uuid.UUID); ok {
		s.harness.historyBatchID = &batchID
	}
	if s.state.cabinet != nil && s.state.cabinet.ID == id {
		s.state.cabinet = nil
		s.state.descendants = nil
		s.state.history = append(s.state.history, "control_cabinet:delete")
	}
	return s.harness.deleteErr
}

type deleteProjectLinkReaderStub struct {
	harness                 *deleteTransactionHarness
	links                   []*domainProject.ProjectControlCabinet
	err                     error
	calls                   int
	received                []uuid.UUID
	calledBeforeTransaction bool
}

func (s *deleteProjectLinkReaderStub) GetByControlCabinetIDs(
	_ context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectControlCabinet, error) {
	s.calls++
	s.received = append([]uuid.UUID(nil), ids...)
	s.calledBeforeTransaction = s.harness == nil || s.harness.runnerCalls == 0
	return s.links, s.err
}

func TestDeleteCapturesDirectRecipientsAndCommitsRootHistoryBeforeDispatch(t *testing.T) {
	cabinetID := cabinetTestUUID(601)
	buildingID := cabinetTestUUID(602)
	projectOne := cabinetTestUUID(611)
	projectTwo := cabinetTestUUID(612)
	actorID := cabinetTestUUID(621)
	operationID := cabinetTestUUID(631)
	eventOne := cabinetTestUUID(632)
	eventTwo := cabinetTestUUID(633)
	createdAt := time.Date(2026, time.July, 20, 18, 0, 0, 0, time.UTC)
	occurredAt := createdAt.Add(time.Hour)
	number := "AK05"

	harness := &deleteTransactionHarness{committed: deleteTransactionState{
		cabinet: &domainFacility.ControlCabinet{
			Base:             domain.Base{ID: cabinetID, CreatedAt: createdAt, UpdatedAt: createdAt},
			BuildingID:       buildingID,
			ControlCabinetNr: &number,
		},
		descendants: []string{"sps_controller", "field_device", "bacnet_object"},
	}}
	links := &deleteProjectLinkReaderStub{
		harness: harness,
		links: []*domainProject.ProjectControlCabinet{
			{ProjectID: projectTwo, ControlCabinetID: cabinetID},
			{ProjectID: projectOne, ControlCabinetID: cabinetID},
			{ProjectID: projectOne, ControlCabinetID: cabinetID},
			{ProjectID: cabinetTestUUID(613), ControlCabinetID: cabinetTestUUID(699)},
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
		Actor:        func(context.Context) *uuid.UUID { return &actorID },
		NewID: func() uuid.UUID {
			id := generatedIDs[0]
			generatedIDs = generatedIDs[1:]
			return id
		},
		Now: func() time.Time { return occurredAt },
	})

	outcome, err := handler.Execute(context.Background(), DeleteCommand{
		ControlCabinetID: cabinetID,
	})
	if err != nil {
		t.Fatalf("execute cabinet delete: %v", err)
	}
	if !links.calledBeforeTransaction || links.calls != 1 ||
		!reflect.DeepEqual(links.received, []uuid.UUID{cabinetID}) {
		t.Fatalf("pre-delete scope: before=%t calls=%d ids=%v",
			links.calledBeforeTransaction,
			links.calls,
			links.received,
		)
	}
	if harness.committed.cabinet != nil || len(harness.committed.descendants) != 0 ||
		!reflect.DeepEqual(harness.committed.history, []string{"control_cabinet:delete"}) {
		t.Fatalf("committed state: %+v", harness.committed)
	}
	if harness.historyBatchID == nil || *harness.historyBatchID != operationID ||
		outcome.Mutation.BatchID == nil || *outcome.Mutation.BatchID != operationID {
		t.Fatalf("history batch: workflow=%v result=%v", harness.historyBatchID, outcome.Mutation.BatchID)
	}
	if !outcome.Existed || outcome.Mutation.OperationID != operationID ||
		outcome.Mutation.ActorID == nil || *outcome.Mutation.ActorID != actorID ||
		!outcome.Mutation.OccurredAt.Equal(occurredAt) ||
		!reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectOne, projectTwo}) {
		t.Fatalf("outcome: %+v", outcome)
	}
	if len(outcome.Mutation.Changes) != 1 {
		t.Fatalf("changes: got %d, want 1", len(outcome.Mutation.Changes))
	}
	change := outcome.Mutation.Changes[0]
	if change.EntityID != cabinetID || change.ParentID == nil || *change.ParentID != buildingID ||
		change.Action != domainHistory.ActionDelete || len(change.After) != 0 {
		t.Fatalf("delete change: %+v", change)
	}
	var snapshot controlCabinetSnapshot
	if err := json.Unmarshal(change.Before, &snapshot); err != nil {
		t.Fatalf("decode snapshot: %v", err)
	}
	if snapshot.ID != cabinetID || snapshot.BuildingID != buildingID ||
		snapshot.ControlCabinetNr == nil || *snapshot.ControlCabinetNr != number {
		t.Fatalf("snapshot: %+v", snapshot)
	}
	if len(dispatcher.commands) != 2 {
		t.Fatalf("commands: got %d, want 2", len(dispatcher.commands))
	}
	for index, raw := range dispatcher.commands {
		command, ok := raw.(appcollaboration.ControlCabinetDeleted)
		if !ok {
			t.Fatalf("command: got %T, want ControlCabinetDeleted", raw)
		}
		wantProjectID := []uuid.UUID{projectOne, projectTwo}[index]
		if command.ProjectID != wantProjectID || command.ControlCabinetID != cabinetID ||
			command.BuildingID != buildingID || command.OperationID != operationID ||
			command.CorrelationID != operationID ||
			command.SchemaVersion != appcollaboration.SchemaVersionV2 {
			t.Fatalf("command: %+v", command)
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
		wantErr   error
	}{
		{name: "delete", deleteErr: deleteErr, wantErr: deleteErr},
		{name: "commit", commitErr: commitErr, wantErr: commitErr},
	} {
		t.Run(test.name, func(t *testing.T) {
			cabinetID := cabinetTestUUID(641)
			harness := &deleteTransactionHarness{
				committed: deleteTransactionState{
					cabinet: &domainFacility.ControlCabinet{
						Base:       domain.Base{ID: cabinetID},
						BuildingID: cabinetTestUUID(642),
					},
					descendants: []string{"sps_controller"},
				},
				deleteErr: test.deleteErr,
				commitErr: test.commitErr,
			}
			links := &deleteProjectLinkReaderStub{harness: harness, links: []*domainProject.ProjectControlCabinet{{
				ProjectID:        cabinetTestUUID(643),
				ControlCabinetID: cabinetID,
			}}}
			dispatcher := &updateCommandDispatcherStub{}
			handler := NewDeleteHandler(DeleteDependencies{
				TransactionRunner:   harness.runner,
				TransactionWorkflow: harness.factory,
				ProjectLinks:        links,
				Dispatcher:          dispatcher,
			})

			_, err := handler.Execute(context.Background(), DeleteCommand{ControlCabinetID: cabinetID})
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error: got %v, want %v", err, test.wantErr)
			}
			if harness.committed.cabinet == nil ||
				!reflect.DeepEqual(harness.committed.descendants, []string{"sps_controller"}) ||
				len(harness.committed.history) != 0 {
				t.Fatalf("failed transaction escaped: %+v", harness.committed)
			}
			if len(dispatcher.commands) != 0 {
				t.Fatalf("dispatched after rollback: %+v", dispatcher.commands)
			}
		})
	}
}

func TestDeleteMissingControlCabinetPreservesIdempotentSuccess(t *testing.T) {
	harness := &deleteTransactionHarness{}
	links := &deleteProjectLinkReaderStub{harness: harness}
	dispatcher := &updateCommandDispatcherStub{}
	handler := NewDeleteHandler(DeleteDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		ProjectLinks:        links,
		Dispatcher:          dispatcher,
	})

	outcome, err := handler.Execute(context.Background(), DeleteCommand{
		ControlCabinetID: cabinetTestUUID(651),
	})
	if err != nil {
		t.Fatalf("missing delete: %v", err)
	}
	if outcome.Existed || len(outcome.Mutation.Changes) != 0 || harness.deleteCalls != 1 ||
		links.calls != 1 || len(dispatcher.commands) != 0 {
		t.Fatalf("missing-row outcome: outcome=%+v harness=%+v", outcome, harness)
	}
}

func TestDeleteRecipientLookupFailureIsBestEffortAfterCommit(t *testing.T) {
	cabinetID := cabinetTestUUID(661)
	scopeErr := errors.New("scope lookup failed")
	harness := &deleteTransactionHarness{committed: deleteTransactionState{
		cabinet: &domainFacility.ControlCabinet{
			Base:       domain.Base{ID: cabinetID},
			BuildingID: cabinetTestUUID(662),
		},
	}}
	links := &deleteProjectLinkReaderStub{harness: harness, err: scopeErr}
	dispatcher := &updateCommandDispatcherStub{}
	handler := NewDeleteHandler(DeleteDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		ProjectLinks:        links,
		Dispatcher:          dispatcher,
	})

	outcome, err := handler.Execute(context.Background(), DeleteCommand{ControlCabinetID: cabinetID})
	if err != nil {
		t.Fatalf("delete with scope failure: %v", err)
	}
	if harness.committed.cabinet != nil || !outcome.Existed {
		t.Fatalf("delete did not commit: outcome=%+v state=%+v", outcome, harness.committed)
	}
	if len(outcome.DispatchErrors) != 1 || !errors.Is(outcome.DispatchErrors[0], scopeErr) {
		t.Fatalf("dispatch errors: got %v, want wrapped %v", outcome.DispatchErrors, scopeErr)
	}
	if len(dispatcher.commands) != 0 {
		t.Fatalf("unexpected command after scope failure: %+v", dispatcher.commands)
	}
}

func TestDeleteReportsDispatchFailureWithoutChangingCommittedResult(t *testing.T) {
	cabinetID := cabinetTestUUID(671)
	projectID := cabinetTestUUID(672)
	dispatchErr := errors.New("transport unavailable")
	harness := &deleteTransactionHarness{committed: deleteTransactionState{
		cabinet: &domainFacility.ControlCabinet{
			Base:       domain.Base{ID: cabinetID},
			BuildingID: cabinetTestUUID(673),
		},
	}}
	links := &deleteProjectLinkReaderStub{harness: harness, links: []*domainProject.ProjectControlCabinet{{
		ProjectID:        projectID,
		ControlCabinetID: cabinetID,
	}}}
	dispatcher := &updateCommandDispatcherStub{err: dispatchErr}
	var reported []error
	handler := NewDeleteHandler(DeleteDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		ProjectLinks:        links,
		Dispatcher:          dispatcher,
		ReportError: func(err error) {
			reported = append(reported, err)
		},
	})

	err := handler.Delete(context.Background(), DeleteCommand{ControlCabinetID: cabinetID})
	if err != nil {
		t.Fatalf("best-effort delete returned error: %v", err)
	}
	if harness.committed.cabinet != nil ||
		len(reported) != 1 || !errors.Is(reported[0], dispatchErr) {
		t.Fatalf("commit/report state: committed=%+v reported=%v", harness.committed, reported)
	}
}

func TestDeleteRejectsNilIDBeforeScopeLookupOrTransaction(t *testing.T) {
	harness := &deleteTransactionHarness{}
	links := &deleteProjectLinkReaderStub{harness: harness}
	handler := NewDeleteHandler(DeleteDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		ProjectLinks:        links,
		Dispatcher:          &updateCommandDispatcherStub{},
	})

	_, err := handler.Execute(context.Background(), DeleteCommand{})
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("delete error: got %v, want %v", err, domain.ErrInvalidArgument)
	}
	if harness.runnerCalls != 0 || links.calls != 0 {
		t.Fatalf("invalid delete advanced: runner=%d links=%d", harness.runnerCalls, links.calls)
	}
}
