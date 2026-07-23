package spscontroller

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
	controller  *domainFacility.SPSController
	systemTypes []domainFacility.SPSControllerSystemType
	history     []string
}

func (s createTransactionState) clone() createTransactionState {
	return createTransactionState{
		controller:  cloneSPSController(s.controller),
		systemTypes: cloneSystemTypes(s.systemTypes),
		history:     append([]string(nil), s.history...),
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
	generatedName  string
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

func (s *createWorkflowStub) CreateWithSystemTypes(
	ctx context.Context,
	controller *domainFacility.SPSController,
	systemTypes []domainFacility.SPSControllerSystemType,
) error {
	s.harness.createCalls++
	if batchID, ok := ctx.Value(createHistoryBatchKey{}).(uuid.UUID); ok {
		s.harness.historyBatchID = &batchID
	}
	created := cloneSPSController(controller)
	created.ID = s.harness.createdID
	created.CreatedAt = s.harness.createdAt
	created.UpdatedAt = s.harness.createdAt
	if s.harness.generatedName != "" {
		created.DeviceName = s.harness.generatedName
	}
	controller.ID = created.ID
	s.state.controller = created
	s.state.systemTypes = cloneSystemTypes(systemTypes)
	s.state.history = append(s.state.history, "sps_controller:create")
	for range systemTypes {
		s.state.history = append(s.state.history, "sps_controller_system_type:create")
	}
	return s.harness.createErr
}

func (s *createWorkflowStub) GetByID(
	_ context.Context,
	id uuid.UUID,
) (*domainFacility.SPSController, error) {
	if s.harness.reloadErr != nil {
		return nil, s.harness.reloadErr
	}
	if s.state.controller == nil || s.state.controller.ID != id {
		return nil, domain.ErrNotFound
	}
	return cloneSPSController(s.state.controller), nil
}

type createProjectLinkReaderStub struct {
	harness *createTransactionHarness
	links   []*domainProject.ProjectSPSController
	err     error
	calls   int
	ids     []uuid.UUID
}

func (s *createProjectLinkReaderStub) GetBySPSControllerIDs(
	_ context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectSPSController, error) {
	s.calls++
	s.ids = append([]uuid.UUID(nil), ids...)
	if s.harness != nil && s.harness.committed.controller == nil {
		return nil, errors.New("project scope resolved before commit")
	}
	return s.links, s.err
}

func TestCreateCommitsControllerAndSystemTypeHistoryBeforeDispatch(t *testing.T) {
	cabinetID := spsTestUUID(1)
	controllerID := spsTestUUID(2)
	systemTypeID := spsTestUUID(3)
	projectOne := spsTestUUID(11)
	projectTwo := spsTestUUID(12)
	actorID := spsTestUUID(21)
	operationID := spsTestUUID(31)
	eventOne := spsTestUUID(32)
	eventTwo := spsTestUUID(33)
	createdAt := time.Date(2026, time.July, 20, 23, 45, 0, 0, time.UTC)
	occurredAt := createdAt.Add(time.Second)
	gaDevice := "AAA"
	description := "Automation controller"
	number := 17

	harness := &createTransactionHarness{
		createdID:     controllerID,
		createdAt:     createdAt,
		generatedName: "BLD_AK01_AAA",
	}
	links := &createProjectLinkReaderStub{
		harness: harness,
		links: []*domainProject.ProjectSPSController{
			{ProjectID: projectTwo, SPSControllerID: controllerID},
			{ProjectID: projectOne, SPSControllerID: controllerID},
			{ProjectID: projectOne, SPSControllerID: controllerID},
			{ProjectID: spsTestUUID(13), SPSControllerID: spsTestUUID(99)},
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
		ControlCabinetID:  cabinetID,
		GADevice:          &gaDevice,
		DeviceName:        "client value",
		DeviceDescription: &description,
		SystemTypes: []domainFacility.SPSControllerSystemType{{
			SystemTypeID: systemTypeID,
			Number:       &number,
		}},
	})
	if err != nil {
		t.Fatalf("execute create: %v", err)
	}
	if harness.runnerCalls != 1 || harness.createCalls != 1 ||
		harness.historyBatchID == nil || *harness.historyBatchID != operationID ||
		len(harness.committed.history) != 2 {
		t.Fatalf("unexpected committed transaction: %+v", harness)
	}
	if len(harness.committed.systemTypes) != 1 ||
		harness.committed.systemTypes[0].SystemTypeID != systemTypeID {
		t.Fatalf("unexpected system types: %+v", harness.committed.systemTypes)
	}
	if outcome.SPSController.ID != controllerID ||
		outcome.SPSController.ControlCabinetID != cabinetID ||
		outcome.SPSController.DeviceName != "BLD_AK01_AAA" {
		t.Fatalf("unexpected committed controller: %+v", outcome.SPSController)
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
	if want := []uuid.UUID{controllerID}; !reflect.DeepEqual(links.ids, want) {
		t.Fatalf("project lookup IDs: got %v, want %v", links.ids, want)
	}
	change := outcome.Mutation.Changes[0]
	if change.EntityID != controllerID || change.ParentID == nil || *change.ParentID != cabinetID ||
		change.Action != domainHistory.ActionCreate || len(change.Before) != 0 {
		t.Fatalf("unexpected create change: %+v", change)
	}
	var snapshot spsControllerSnapshot
	if err := json.Unmarshal(change.After, &snapshot); err != nil {
		t.Fatalf("decode after snapshot: %v", err)
	}
	if snapshot.ID != controllerID || snapshot.DeviceName != "BLD_AK01_AAA" ||
		snapshot.DeviceDescription == nil || *snapshot.DeviceDescription != description {
		t.Fatalf("unexpected after snapshot: %+v", snapshot)
	}
	if len(dispatcher.commands) != 2 {
		t.Fatalf("commands: got %d, want 2", len(dispatcher.commands))
	}
	for index, raw := range dispatcher.commands {
		command, ok := raw.(appcollaboration.SPSControllerCreated)
		if !ok {
			t.Fatalf("command: got %T, want SPSControllerCreated", raw)
		}
		wantProjectID := []uuid.UUID{projectOne, projectTwo}[index]
		if command.ProjectID != wantProjectID || command.SPSController.ID != controllerID ||
			command.SPSController.ControlCabinetID != cabinetID ||
			command.SPSController.DeviceName != "BLD_AK01_AAA" ||
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
				createdID: controllerIDForCreateTest(),
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
			gaDevice := "AAA"

			_, err := handler.Execute(context.Background(), CreateCommand{
				ControlCabinetID: spsTestUUID(1),
				GADevice:         &gaDevice,
				DeviceName:       "BLD_AK01_AAA",
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
			if harness.committed.controller != nil || len(harness.committed.history) != 0 {
				t.Fatalf("failed mutation escaped transaction: %+v", harness.committed)
			}
			if links.calls != 0 || len(dispatcher.commands) != 0 {
				t.Fatalf("post-commit work ran after failure: links=%d commands=%d", links.calls, len(dispatcher.commands))
			}
		})
	}
}

func controllerIDForCreateTest() uuid.UUID {
	return spsTestUUID(2)
}
