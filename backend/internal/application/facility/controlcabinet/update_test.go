package controlcabinet

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"testing"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	"github.com/besart951/go_infra_link/backend/internal/application/facility/mutation"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type updateTransactionState struct {
	cabinet        *domainFacility.ControlCabinet
	descendantName string
	history        []string
}

func (s updateTransactionState) clone() updateTransactionState {
	return updateTransactionState{
		cabinet:        cloneControlCabinet(s.cabinet),
		descendantName: s.descendantName,
		history:        append([]string(nil), s.history...),
	}
}

type updateTransactionUnit struct {
	state *updateTransactionState
}

type updateHistoryBatchKey struct{}

type updateTransactionHarness struct {
	committed      updateTransactionState
	updateErr      error
	commitErr      error
	updatedAt      time.Time
	generatedName  string
	runnerCalls    int
	updateCalls    int
	historyBatchID *uuid.UUID
}

func (h *updateTransactionHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	staged := h.committed.clone()
	if err := run(ctx, updateTransactionUnit{state: &staged}); err != nil {
		return err
	}
	if h.commitErr != nil {
		return h.commitErr
	}
	h.committed = staged
	return nil
}

func (h *updateTransactionHarness) factory(
	unit apptransaction.UnitOfWork,
) (UpdateWorkflow, error) {
	typed, ok := unit.(updateTransactionUnit)
	if !ok || typed.state == nil {
		return nil, errors.New("unexpected transaction unit")
	}
	return &updateWorkflowStub{harness: h, state: typed.state}, nil
}

type updateWorkflowStub struct {
	harness *updateTransactionHarness
	state   *updateTransactionState
}

func (s *updateWorkflowStub) GetByID(
	_ context.Context,
	id uuid.UUID,
) (*domainFacility.ControlCabinet, error) {
	if s.state.cabinet == nil || s.state.cabinet.ID != id {
		return nil, domain.ErrNotFound
	}
	return cloneControlCabinet(s.state.cabinet), nil
}

func (s *updateWorkflowStub) Update(
	ctx context.Context,
	cabinet *domainFacility.ControlCabinet,
) error {
	s.harness.updateCalls++
	if batchID, ok := ctx.Value(updateHistoryBatchKey{}).(uuid.UUID); ok {
		s.harness.historyBatchID = &batchID
	}
	updated := cloneControlCabinet(cabinet)
	updated.UpdatedAt = s.harness.updatedAt
	s.state.cabinet = updated
	s.state.descendantName = s.harness.generatedName
	s.state.history = append(s.state.history, "control_cabinet:update", "sps_controller:names")
	return s.harness.updateErr
}

type updateProjectLinkReaderStub struct {
	harness          *updateTransactionHarness
	links            []*domainProject.ProjectControlCabinet
	err              error
	calls            int
	received         []uuid.UUID
	expectedBuilding uuid.UUID
}

func (s *updateProjectLinkReaderStub) GetByControlCabinetIDs(
	_ context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectControlCabinet, error) {
	s.calls++
	s.received = append([]uuid.UUID(nil), ids...)
	if s.harness != nil && s.expectedBuilding != uuid.Nil &&
		s.harness.committed.cabinet.BuildingID != s.expectedBuilding {
		return nil, errors.New("project scope resolved before commit")
	}
	return s.links, s.err
}

type updateCommandDispatcherStub struct {
	commands []appcollaboration.Command
	err      error
}

func (s *updateCommandDispatcherStub) Dispatch(
	_ context.Context,
	command appcollaboration.Command,
) error {
	s.commands = append(s.commands, command)
	return s.err
}

func TestUpdateMoveCommitsDescendantRenameBeforeProjectScopedDispatch(t *testing.T) {
	cabinetID := cabinetTestUUID(1)
	oldBuildingID := cabinetTestUUID(2)
	newBuildingID := cabinetTestUUID(3)
	projectOne := cabinetTestUUID(11)
	projectTwo := cabinetTestUUID(12)
	actorID := cabinetTestUUID(21)
	operationID := cabinetTestUUID(31)
	eventOne := cabinetTestUUID(32)
	eventTwo := cabinetTestUUID(33)
	updatedAt := time.Date(2026, time.July, 20, 21, 0, 0, 0, time.UTC)
	occurredAt := updatedAt.Add(time.Second)
	oldNumber := "AK01"
	newNumber := "AK02"

	harness := &updateTransactionHarness{
		committed: updateTransactionState{cabinet: &domainFacility.ControlCabinet{
			Base:             domain.Base{ID: cabinetID},
			BuildingID:       oldBuildingID,
			ControlCabinetNr: &oldNumber,
		}},
		updatedAt:     updatedAt,
		generatedName: "NEW_AK02_AAA",
	}
	links := &updateProjectLinkReaderStub{
		harness:          harness,
		expectedBuilding: newBuildingID,
		links: []*domainProject.ProjectControlCabinet{
			{ProjectID: projectTwo, ControlCabinetID: cabinetID},
			{ProjectID: projectOne, ControlCabinetID: cabinetID},
			{ProjectID: projectOne, ControlCabinetID: cabinetID},
			{ProjectID: cabinetTestUUID(13), ControlCabinetID: cabinetTestUUID(99)},
		},
	}
	dispatcher := &updateCommandDispatcherStub{}
	ids := []uuid.UUID{operationID, eventOne, eventTwo}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, updateHistoryBatchKey{}, batchID)
		},
		ProjectLinks: links,
		Dispatcher:   dispatcher,
		Actor:        func(context.Context) *uuid.UUID { return &actorID },
		NewID: func() uuid.UUID {
			id := ids[0]
			ids = ids[1:]
			return id
		},
		Now: func() time.Time { return occurredAt },
	})

	outcome, err := handler.Execute(context.Background(), UpdateCommand{
		ControlCabinetID: cabinetID,
		BuildingID:       &newBuildingID,
		ControlCabinetNr: &newNumber,
	})
	if err != nil {
		t.Fatalf("execute move: %v", err)
	}
	if harness.runnerCalls != 1 || harness.updateCalls != 1 {
		t.Fatalf("unexpected transaction/update calls: runner=%d update=%d", harness.runnerCalls, harness.updateCalls)
	}
	if harness.committed.descendantName != "NEW_AK02_AAA" {
		t.Fatalf("descendant rename did not commit: %+v", harness.committed)
	}
	if harness.historyBatchID == nil || *harness.historyBatchID != operationID {
		t.Fatalf("history batch: got %v, want %s", harness.historyBatchID, operationID)
	}
	if outcome.ControlCabinet.BuildingID != newBuildingID ||
		outcome.ControlCabinet.ControlCabinetNr == nil ||
		*outcome.ControlCabinet.ControlCabinetNr != newNumber {
		t.Fatalf("unexpected committed cabinet: %+v", outcome.ControlCabinet)
	}
	if outcome.Mutation.BatchID == nil || *outcome.Mutation.BatchID != operationID {
		t.Fatalf("mutation batch: got %v, want %s", outcome.Mutation.BatchID, operationID)
	}
	wantFields := []mutation.FieldName{
		mutation.FieldNameBuilding,
		mutation.FieldNameCabinetNumber,
	}
	if got := outcome.Mutation.Changes[0].ChangedFields; !reflect.DeepEqual(got, wantFields) {
		t.Fatalf("changed fields: got %v, want %v", got, wantFields)
	}
	if links.calls != 1 || !reflect.DeepEqual(links.received, []uuid.UUID{cabinetID}) {
		t.Fatalf("project link lookup: calls=%d ids=%v", links.calls, links.received)
	}
	if want := []uuid.UUID{projectOne, projectTwo}; !reflect.DeepEqual(outcome.Mutation.ProjectIDs, want) {
		t.Fatalf("project IDs: got %v, want %v", outcome.Mutation.ProjectIDs, want)
	}
	if len(dispatcher.commands) != 2 {
		t.Fatalf("commands: got %d, want 2", len(dispatcher.commands))
	}
	for i, raw := range dispatcher.commands {
		command, ok := raw.(appcollaboration.ControlCabinetMoved)
		if !ok {
			t.Fatalf("command %d: got %T, want ControlCabinetMoved", i, raw)
		}
		if command.FromBuildingID != oldBuildingID ||
			command.ToBuildingID != newBuildingID ||
			command.OperationID != operationID ||
			command.ControlCabinet.ID != cabinetID ||
			command.ControlCabinet.BuildingID != newBuildingID {
			t.Fatalf("unexpected move command: %+v", command)
		}
	}
}

func TestUpdateWithinBuildingUsesUpdatedCommand(t *testing.T) {
	cabinetID := cabinetTestUUID(1)
	buildingID := cabinetTestUUID(2)
	projectID := cabinetTestUUID(11)
	oldNumber := "AK01"
	newNumber := "AK02"
	harness := &updateTransactionHarness{committed: updateTransactionState{
		cabinet: &domainFacility.ControlCabinet{
			Base:             domain.Base{ID: cabinetID},
			BuildingID:       buildingID,
			ControlCabinetNr: &oldNumber,
		},
	}}
	dispatcher := &updateCommandDispatcherStub{}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		ProjectLinks: &updateProjectLinkReaderStub{links: []*domainProject.ProjectControlCabinet{
			{ProjectID: projectID, ControlCabinetID: cabinetID},
		}},
		Dispatcher: dispatcher,
	})

	_, err := handler.Execute(context.Background(), UpdateCommand{
		ControlCabinetID: cabinetID,
		ControlCabinetNr: &newNumber,
	})
	if err != nil {
		t.Fatalf("execute update: %v", err)
	}
	if len(dispatcher.commands) != 1 {
		t.Fatalf("commands: got %d, want 1", len(dispatcher.commands))
	}
	if _, ok := dispatcher.commands[0].(appcollaboration.ControlCabinetUpdated); !ok {
		t.Fatalf("command: got %T, want ControlCabinetUpdated", dispatcher.commands[0])
	}
}

func TestUpdateWriteFailureRollsBackAndDoesNotDispatch(t *testing.T) {
	testUpdateFailureDoesNotDispatch(t, errors.New("descendant rename failed"), nil)
}

func TestUpdateCommitFailureDoesNotDispatch(t *testing.T) {
	testUpdateFailureDoesNotDispatch(t, nil, errors.New("commit failed"))
}

func testUpdateFailureDoesNotDispatch(t *testing.T, updateErr, commitErr error) {
	t.Helper()
	cabinetID := cabinetTestUUID(1)
	oldBuildingID := cabinetTestUUID(2)
	newBuildingID := cabinetTestUUID(3)
	number := "AK01"
	harness := &updateTransactionHarness{
		committed: updateTransactionState{cabinet: &domainFacility.ControlCabinet{
			Base:             domain.Base{ID: cabinetID},
			BuildingID:       oldBuildingID,
			ControlCabinetNr: &number,
		}},
		updateErr: updateErr,
		commitErr: commitErr,
	}
	links := &updateProjectLinkReaderStub{}
	dispatcher := &updateCommandDispatcherStub{}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		ProjectLinks:        links,
		Dispatcher:          dispatcher,
	})

	_, err := handler.Execute(context.Background(), UpdateCommand{
		ControlCabinetID: cabinetID,
		BuildingID:       &newBuildingID,
	})
	wantErr := updateErr
	if wantErr == nil {
		wantErr = commitErr
	}
	if !errors.Is(err, wantErr) {
		t.Fatalf("error: got %v, want %v", err, wantErr)
	}
	if harness.committed.cabinet.BuildingID != oldBuildingID ||
		harness.committed.descendantName != "" || len(harness.committed.history) != 0 {
		t.Fatalf("failed transaction leaked staged state: %+v", harness.committed)
	}
	if links.calls != 0 || len(dispatcher.commands) != 0 {
		t.Fatalf("post-commit work ran after failure: links=%d commands=%d", links.calls, len(dispatcher.commands))
	}
}

func cabinetTestUUID(value int) uuid.UUID {
	return uuid.MustParse(fmt.Sprintf("00000000-0000-0000-0000-%012d", value))
}
