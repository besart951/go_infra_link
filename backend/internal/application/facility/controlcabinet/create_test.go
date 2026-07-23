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

type createTransactionState struct {
	cabinet *domainFacility.ControlCabinet
	history []string
}

func (s createTransactionState) clone() createTransactionState {
	return createTransactionState{
		cabinet: cloneControlCabinet(s.cabinet),
		history: append([]string(nil), s.history...),
	}
}

type createTransactionUnit struct {
	state *createTransactionState
}

type createHistoryBatchKey struct{}

type createTransactionHarness struct {
	committed      createTransactionState
	createdID      uuid.UUID
	createdAt      time.Time
	createErr      error
	reloadErr      error
	commitErr      error
	runnerCalls    int
	createCalls    int
	historyBatchID *uuid.UUID
}

func (h *createTransactionHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	staged := h.committed.clone()
	if err := run(ctx, createTransactionUnit{state: &staged}); err != nil {
		return err
	}
	if h.commitErr != nil {
		return h.commitErr
	}
	h.committed = staged
	return nil
}

func (h *createTransactionHarness) factory(
	unit apptransaction.UnitOfWork,
) (CreateWorkflow, error) {
	typed, ok := unit.(createTransactionUnit)
	if !ok || typed.state == nil {
		return nil, errors.New("unexpected transaction unit")
	}
	return &createWorkflowStub{harness: h, state: typed.state}, nil
}

type createWorkflowStub struct {
	harness *createTransactionHarness
	state   *createTransactionState
}

func (s *createWorkflowStub) Create(
	ctx context.Context,
	cabinet *domainFacility.ControlCabinet,
) error {
	s.harness.createCalls++
	if batchID, ok := ctx.Value(createHistoryBatchKey{}).(uuid.UUID); ok {
		s.harness.historyBatchID = &batchID
	}
	created := cloneControlCabinet(cabinet)
	created.ID = s.harness.createdID
	created.CreatedAt = s.harness.createdAt
	created.UpdatedAt = s.harness.createdAt
	cabinet.ID = created.ID
	s.state.cabinet = created
	s.state.history = append(s.state.history, "control_cabinet:create")
	return s.harness.createErr
}

func (s *createWorkflowStub) GetByID(
	_ context.Context,
	id uuid.UUID,
) (*domainFacility.ControlCabinet, error) {
	if s.harness.reloadErr != nil {
		return nil, s.harness.reloadErr
	}
	if s.state.cabinet == nil || s.state.cabinet.ID != id {
		return nil, domain.ErrNotFound
	}
	return cloneControlCabinet(s.state.cabinet), nil
}

type createProjectLinkReaderStub struct {
	harness *createTransactionHarness
	links   []*domainProject.ProjectControlCabinet
	err     error
	calls   int
	ids     []uuid.UUID
}

func (s *createProjectLinkReaderStub) GetByControlCabinetIDs(
	_ context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectControlCabinet, error) {
	s.calls++
	s.ids = append([]uuid.UUID(nil), ids...)
	if s.harness != nil && s.harness.committed.cabinet == nil {
		return nil, errors.New("project scope resolved before commit")
	}
	return s.links, s.err
}

func TestCreateCommitsHistoryBeforeProjectScopedDispatch(t *testing.T) {
	buildingID := cabinetTestUUID(1)
	cabinetID := cabinetTestUUID(2)
	projectOne := cabinetTestUUID(11)
	projectTwo := cabinetTestUUID(12)
	actorID := cabinetTestUUID(21)
	operationID := cabinetTestUUID(31)
	eventOne := cabinetTestUUID(32)
	eventTwo := cabinetTestUUID(33)
	createdAt := time.Date(2026, time.July, 20, 23, 0, 0, 0, time.UTC)
	occurredAt := createdAt.Add(time.Second)
	number := "AK01"

	harness := &createTransactionHarness{createdID: cabinetID, createdAt: createdAt}
	links := &createProjectLinkReaderStub{
		harness: harness,
		links: []*domainProject.ProjectControlCabinet{
			{ProjectID: projectTwo, ControlCabinetID: cabinetID},
			{ProjectID: projectOne, ControlCabinetID: cabinetID},
			{ProjectID: projectOne, ControlCabinetID: cabinetID},
			{ProjectID: cabinetTestUUID(13), ControlCabinetID: cabinetTestUUID(99)},
		},
	}
	dispatcher := &updateCommandDispatcherStub{}
	ids := []uuid.UUID{operationID, eventOne, eventTwo}
	handler := NewCreateHandler(CreateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, createHistoryBatchKey{}, batchID)
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

	outcome, err := handler.Execute(context.Background(), CreateCommand{
		BuildingID:       buildingID,
		ControlCabinetNr: &number,
	})
	if err != nil {
		t.Fatalf("execute create: %v", err)
	}
	if harness.runnerCalls != 1 || harness.createCalls != 1 ||
		harness.historyBatchID == nil || *harness.historyBatchID != operationID ||
		len(harness.committed.history) != 1 {
		t.Fatalf("unexpected committed transaction: %+v", harness)
	}
	if outcome.ControlCabinet.ID != cabinetID ||
		outcome.ControlCabinet.BuildingID != buildingID ||
		outcome.ControlCabinet.ControlCabinetNr == nil ||
		*outcome.ControlCabinet.ControlCabinetNr != number {
		t.Fatalf("unexpected committed cabinet: %+v", outcome.ControlCabinet)
	}
	if outcome.Mutation.OperationID != operationID ||
		outcome.Mutation.BatchID == nil || *outcome.Mutation.BatchID != operationID ||
		outcome.Mutation.ActorID == nil || *outcome.Mutation.ActorID != actorID ||
		!outcome.Mutation.OccurredAt.Equal(occurredAt) {
		t.Fatalf("unexpected mutation envelope: %+v", outcome.Mutation)
	}
	if want := []uuid.UUID{projectOne, projectTwo}; !reflect.DeepEqual(outcome.Mutation.ProjectIDs, want) {
		t.Fatalf("project IDs: got %v, want %v", outcome.Mutation.ProjectIDs, want)
	}
	if want := []uuid.UUID{cabinetID}; !reflect.DeepEqual(links.ids, want) {
		t.Fatalf("project lookup IDs: got %v, want %v", links.ids, want)
	}
	change := outcome.Mutation.Changes[0]
	if change.EntityID != cabinetID || change.ParentID == nil || *change.ParentID != buildingID ||
		change.Action != domainHistory.ActionCreate || len(change.Before) != 0 {
		t.Fatalf("unexpected create change: %+v", change)
	}
	var snapshot controlCabinetSnapshot
	if err := json.Unmarshal(change.After, &snapshot); err != nil {
		t.Fatalf("decode after snapshot: %v", err)
	}
	if snapshot.ID != cabinetID || snapshot.BuildingID != buildingID ||
		snapshot.ControlCabinetNr == nil || *snapshot.ControlCabinetNr != number {
		t.Fatalf("unexpected after snapshot: %+v", snapshot)
	}
	if len(dispatcher.commands) != 2 {
		t.Fatalf("commands: got %d, want 2", len(dispatcher.commands))
	}
	for index, raw := range dispatcher.commands {
		command, ok := raw.(appcollaboration.ControlCabinetCreated)
		if !ok {
			t.Fatalf("command: got %T, want ControlCabinetCreated", raw)
		}
		wantProjectID := []uuid.UUID{projectOne, projectTwo}[index]
		if command.ProjectID != wantProjectID || command.ControlCabinet.ID != cabinetID ||
			command.OperationID != operationID || command.CorrelationID != operationID ||
			command.SchemaVersion != appcollaboration.SchemaVersionV1 {
			t.Fatalf("unexpected collaboration command: %+v", command)
		}
	}
}

func TestCreateFailureDoesNotDispatchOrEscapeTransaction(t *testing.T) {
	createErr := errors.New("create failed")
	reloadErr := errors.New("reload failed")
	commitErr := errors.New("commit failed")
	for _, test := range []struct {
		name      string
		createErr error
		reloadErr error
		commitErr error
	}{
		{name: "write", createErr: createErr},
		{name: "reload", reloadErr: reloadErr},
		{name: "commit", commitErr: commitErr},
	} {
		t.Run(test.name, func(t *testing.T) {
			harness := &createTransactionHarness{
				createdID: cabinetTestUUID(2),
				createErr: test.createErr,
				reloadErr: test.reloadErr,
				commitErr: test.commitErr,
			}
			links := &createProjectLinkReaderStub{}
			dispatcher := &updateCommandDispatcherStub{}
			handler := NewCreateHandler(CreateDependencies{
				TransactionRunner:   harness.runner,
				TransactionWorkflow: harness.factory,
				ProjectLinks:        links,
				Dispatcher:          dispatcher,
			})
			number := "AK01"

			_, err := handler.Execute(context.Background(), CreateCommand{
				BuildingID:       cabinetTestUUID(1),
				ControlCabinetNr: &number,
			})
			wantErr := test.createErr
			if wantErr == nil {
				wantErr = test.reloadErr
			}
			if wantErr == nil {
				wantErr = test.commitErr
			}
			if !errors.Is(err, wantErr) {
				t.Fatalf("error: got %v, want %v", err, wantErr)
			}
			if harness.committed.cabinet != nil || len(harness.committed.history) != 0 {
				t.Fatalf("failed mutation escaped transaction: %+v", harness.committed)
			}
			if links.calls != 0 || len(dispatcher.commands) != 0 {
				t.Fatalf("post-commit work ran after failure: links=%d commands=%d", links.calls, len(dispatcher.commands))
			}
		})
	}
}

func TestCreateReportsPostCommitScopeFailureWithoutFailingMutation(t *testing.T) {
	cabinetID := cabinetTestUUID(2)
	scopeErr := errors.New("scope lookup failed")
	harness := &createTransactionHarness{createdID: cabinetID}
	links := &createProjectLinkReaderStub{harness: harness, err: scopeErr}
	dispatcher := &updateCommandDispatcherStub{}
	var reported []error
	handler := NewCreateHandler(CreateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		ProjectLinks:        links,
		Dispatcher:          dispatcher,
		ReportError: func(err error) {
			reported = append(reported, err)
		},
	})
	number := "AK01"

	created, err := handler.Create(context.Background(), CreateCommand{
		BuildingID:       cabinetTestUUID(1),
		ControlCabinetNr: &number,
	})
	if err != nil || created == nil || created.ID != cabinetID {
		t.Fatalf("committed create: cabinet=%+v err=%v", created, err)
	}
	if len(reported) != 1 || !errors.Is(reported[0], scopeErr) {
		t.Fatalf("reported errors: got %v, want wrapped %v", reported, scopeErr)
	}
	if len(dispatcher.commands) != 0 || harness.committed.cabinet == nil {
		t.Fatalf("unexpected post-commit state: commands=%d committed=%+v", len(dispatcher.commands), harness.committed)
	}
}
