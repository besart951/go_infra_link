package spscontroller

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	"github.com/besart951/go_infra_link/backend/internal/application/facility/mutation"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainCollaboration "github.com/besart951/go_infra_link/backend/internal/domain/collaboration"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type updateTransactionState struct {
	controller  *domainFacility.SPSController
	systemTypes []domainFacility.SPSControllerSystemType
	history     []string
}

func (s updateTransactionState) clone() updateTransactionState {
	return updateTransactionState{
		controller:  cloneSPSController(s.controller),
		systemTypes: cloneSystemTypes(s.systemTypes),
		history:     append([]string(nil), s.history...),
	}
}

type updateTransactionUnit struct {
	state *updateTransactionState
}

type updateHistoryBatchKey struct{}

type updateTransactionHarness struct {
	committed       updateTransactionState
	updateErr       error
	commitErr       error
	updatedAt       time.Time
	generatedName   string
	runnerCalls     int
	updateCalls     int
	updateWithCalls int
	historyBatchID  *uuid.UUID
	outbox          domainCollaboration.OutboxStore
}

func (h *updateTransactionHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	staged := h.committed.clone()
	txCtx := ctx
	if h.outbox != nil {
		txCtx = domainCollaboration.WithOutboxStore(ctx, h.outbox)
	}
	if err := run(txCtx, updateTransactionUnit{state: &staged}); err != nil {
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
) (*domainFacility.SPSController, error) {
	if s.state.controller == nil || s.state.controller.ID != id {
		return nil, domain.ErrNotFound
	}
	return cloneSPSController(s.state.controller), nil
}

func (s *updateWorkflowStub) Update(
	ctx context.Context,
	controller *domainFacility.SPSController,
) error {
	s.harness.updateCalls++
	return s.persist(ctx, controller, nil)
}

func (s *updateWorkflowStub) UpdateWithSystemTypes(
	ctx context.Context,
	controller *domainFacility.SPSController,
	systemTypes []domainFacility.SPSControllerSystemType,
) error {
	s.harness.updateWithCalls++
	return s.persist(ctx, controller, &systemTypes)
}

func (s *updateWorkflowStub) persist(
	ctx context.Context,
	controller *domainFacility.SPSController,
	systemTypes *[]domainFacility.SPSControllerSystemType,
) error {
	if batchID, ok := ctx.Value(updateHistoryBatchKey{}).(uuid.UUID); ok {
		s.harness.historyBatchID = &batchID
	}
	updated := cloneSPSController(controller)
	if s.harness.generatedName != "" {
		updated.DeviceName = s.harness.generatedName
	}
	updated.UpdatedAt = s.harness.updatedAt
	s.state.controller = updated
	s.state.history = append(s.state.history, "sps_controller:update")
	if systemTypes != nil {
		s.state.systemTypes = cloneSystemTypes(*systemTypes)
		s.state.history = append(s.state.history, "sps_controller_system_types:replace")
	}
	return s.harness.updateErr
}

type updateProjectLinkReaderStub struct {
	harness        *updateTransactionHarness
	links          []*domainProject.ProjectSPSController
	err            error
	calls          int
	received       []uuid.UUID
	expectedParent uuid.UUID
}

func (s *updateProjectLinkReaderStub) GetBySPSControllerIDs(
	_ context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectSPSController, error) {
	s.calls++
	s.received = append([]uuid.UUID(nil), ids...)
	if s.harness != nil && s.expectedParent != uuid.Nil &&
		s.harness.committed.controller.ControlCabinetID != s.expectedParent {
		return nil, errors.New("project scope resolved before commit")
	}
	return s.links, s.err
}

type updateCommandDispatcherStub struct {
	commands []appcollaboration.Command
	err      error
}

type transactionalUpdateOutboxStub struct {
	*updateWorkflowStub
	links []*domainProject.ProjectSPSController
}

func (s *transactionalUpdateOutboxStub) GetBySPSControllerIDs(
	_ context.Context,
	_ []uuid.UUID,
) ([]*domainProject.ProjectSPSController, error) {
	return s.links, nil
}

type updateOutboxStoreStub struct {
	events []*domainCollaboration.OutboxEvent
}

func (s *updateOutboxStoreStub) Enqueue(_ context.Context, event *domainCollaboration.OutboxEvent) error {
	s.events = append(s.events, event)
	return nil
}
func (*updateOutboxStoreStub) ClaimDue(context.Context, time.Time, int) ([]domainCollaboration.OutboxEvent, error) {
	return nil, nil
}
func (*updateOutboxStoreStub) WasProcessed(context.Context, string, uuid.UUID) (bool, error) {
	return false, nil
}
func (*updateOutboxStoreStub) MarkDelivered(context.Context, string, domainCollaboration.OutboxEvent, time.Time) error {
	return nil
}
func (*updateOutboxStoreStub) MarkFailed(context.Context, domainCollaboration.OutboxEvent, string, time.Time, time.Time) error {
	return nil
}

func TestUpdateWritesVersionTwoOutboxCommandInsideTransaction(t *testing.T) {
	controllerID := spsTestUUID(901)
	cabinetID := spsTestUUID(902)
	projectID := spsTestUUID(903)
	operationID := spsTestUUID(904)
	outboxEventID := spsTestUUID(905)
	compatibilityEventID := spsTestUUID(906)
	description := "updated"
	occurredAt := time.Date(2026, time.July, 23, 16, 0, 0, 0, time.UTC)
	harness := &updateTransactionHarness{committed: updateTransactionState{
		controller: &domainFacility.SPSController{
			Base: domain.Base{ID: controllerID}, ControlCabinetID: cabinetID, DeviceName: "SPS",
		},
	}}
	outboxStore := &updateOutboxStoreStub{}
	harness.outbox = outboxStore
	factory := func(unit apptransaction.UnitOfWork) (UpdateWorkflow, error) {
		typed := unit.(updateTransactionUnit)
		return &transactionalUpdateOutboxStub{
			updateWorkflowStub: &updateWorkflowStub{harness: harness, state: typed.state},
			links: []*domainProject.ProjectSPSController{{
				ProjectID: projectID, SPSControllerID: controllerID,
			}},
		}, nil
	}
	dispatcher := &updateCommandDispatcherStub{}
	ids := []uuid.UUID{operationID, outboxEventID, compatibilityEventID}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner: harness.runner, TransactionWorkflow: factory, Dispatcher: dispatcher,
		NewID: func() uuid.UUID { id := ids[0]; ids = ids[1:]; return id },
		Now:   func() time.Time { return occurredAt },
	})

	if _, err := handler.Execute(context.Background(), UpdateCommand{
		SPSControllerID: controllerID, DeviceDescription: &description,
	}); err != nil {
		t.Fatalf("update SPS controller: %v", err)
	}
	if len(outboxStore.events) != 1 {
		t.Fatalf("outbox events: got %d, want 1", len(outboxStore.events))
	}
	decoded, err := appcollaboration.DecodeCommand(appcollaboration.EncodedCommand{
		Type: outboxStore.events[0].EventType, Payload: outboxStore.events[0].Payload,
	})
	if err != nil {
		t.Fatalf("decode queued command: %v", err)
	}
	queued, ok := decoded.(appcollaboration.SPSControllerUpdated)
	if !ok || queued.SchemaVersion != appcollaboration.SchemaVersionV2 ||
		queued.EventID != outboxEventID || queued.ProjectID != projectID {
		t.Fatalf("unexpected queued command: %#v", decoded)
	}
	if len(dispatcher.commands) != 1 {
		t.Fatalf("compatibility commands: got %d, want 1", len(dispatcher.commands))
	}
}

func (s *updateCommandDispatcherStub) Dispatch(
	_ context.Context,
	command appcollaboration.Command,
) error {
	s.commands = append(s.commands, command)
	return s.err
}

func TestUpdateMoveCommitsWithSystemTypesBeforeProjectScopedDispatch(t *testing.T) {
	controllerID := spsTestUUID(1)
	oldCabinetID := spsTestUUID(2)
	newCabinetID := spsTestUUID(3)
	systemTypeID := spsTestUUID(4)
	projectOne := spsTestUUID(11)
	projectTwo := spsTestUUID(12)
	actorID := spsTestUUID(21)
	operationID := spsTestUUID(31)
	eventOne := spsTestUUID(32)
	eventTwo := spsTestUUID(33)
	updatedAt := time.Date(2026, time.July, 20, 20, 0, 0, 0, time.UTC)
	occurredAt := updatedAt.Add(time.Second)
	description := "Moved controller"
	number := 17
	systemTypes := []domainFacility.SPSControllerSystemType{{
		SystemTypeID: systemTypeID,
		Number:       &number,
	}}

	harness := &updateTransactionHarness{
		committed: updateTransactionState{controller: &domainFacility.SPSController{
			Base:             domain.Base{ID: controllerID},
			ControlCabinetID: oldCabinetID,
			GADevice:         new("AAA"),
			DeviceName:       "OLD_AAA",
		}},
		updatedAt:     updatedAt,
		generatedName: "NEW_AAA",
	}
	links := &updateProjectLinkReaderStub{
		harness:        harness,
		expectedParent: newCabinetID,
		links: []*domainProject.ProjectSPSController{
			{ProjectID: projectTwo, SPSControllerID: controllerID},
			{ProjectID: projectOne, SPSControllerID: controllerID},
			{ProjectID: projectOne, SPSControllerID: controllerID},
			{ProjectID: spsTestUUID(13), SPSControllerID: spsTestUUID(99)},
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
		SPSControllerID:   controllerID,
		ControlCabinetID:  &newCabinetID,
		DeviceDescription: &description,
		SystemTypes:       &systemTypes,
	})
	if err != nil {
		t.Fatalf("execute move: %v", err)
	}

	if harness.runnerCalls != 1 || harness.updateCalls != 0 || harness.updateWithCalls != 1 {
		t.Fatalf("unexpected transaction/update calls: runner=%d update=%d updateWith=%d", harness.runnerCalls, harness.updateCalls, harness.updateWithCalls)
	}
	if harness.historyBatchID == nil || *harness.historyBatchID != operationID {
		t.Fatalf("history batch: got %v, want %s", harness.historyBatchID, operationID)
	}
	if outcome.SPSController.ControlCabinetID != newCabinetID ||
		outcome.SPSController.DeviceName != "NEW_AAA" ||
		outcome.SPSController.DeviceDescription == nil ||
		*outcome.SPSController.DeviceDescription != description {
		t.Fatalf("unexpected committed controller: %+v", outcome.SPSController)
	}
	if len(harness.committed.systemTypes) != 1 ||
		harness.committed.systemTypes[0].SystemTypeID != systemTypeID {
		t.Fatalf("unexpected committed system types: %+v", harness.committed.systemTypes)
	}
	if outcome.Mutation.BatchID == nil || *outcome.Mutation.BatchID != operationID {
		t.Fatalf("mutation batch: got %v, want %s", outcome.Mutation.BatchID, operationID)
	}
	wantFields := []mutation.FieldName{
		mutation.FieldNameControlCabinet,
		mutation.FieldNameDeviceName,
		mutation.FieldNameDescription,
		mutation.FieldNameSystemTypes,
	}
	if got := outcome.Mutation.Changes[0].ChangedFields; !reflect.DeepEqual(got, wantFields) {
		t.Fatalf("changed fields: got %v, want %v", got, wantFields)
	}
	if links.calls != 1 || !reflect.DeepEqual(links.received, []uuid.UUID{controllerID}) {
		t.Fatalf("project link lookup: calls=%d ids=%v", links.calls, links.received)
	}
	if want := []uuid.UUID{projectOne, projectTwo}; !reflect.DeepEqual(outcome.Mutation.ProjectIDs, want) {
		t.Fatalf("project IDs: got %v, want %v", outcome.Mutation.ProjectIDs, want)
	}
	if len(dispatcher.commands) != 2 {
		t.Fatalf("commands: got %d, want 2", len(dispatcher.commands))
	}
	for i, raw := range dispatcher.commands {
		command, ok := raw.(appcollaboration.SPSControllerMoved)
		if !ok {
			t.Fatalf("command %d: got %T, want SPSControllerMoved", i, raw)
		}
		if command.FromControlCabinetID != oldCabinetID ||
			command.ToControlCabinetID != newCabinetID ||
			command.OperationID != operationID {
			t.Fatalf("unexpected move command: %+v", command)
		}
	}
}

func TestUpdateWithoutSystemTypesStillUsesTransactionAndUpdatedCommand(t *testing.T) {
	controllerID := spsTestUUID(1)
	cabinetID := spsTestUUID(2)
	projectID := spsTestUUID(11)
	description := "Updated"
	harness := &updateTransactionHarness{
		committed: updateTransactionState{controller: &domainFacility.SPSController{
			Base:             domain.Base{ID: controllerID},
			ControlCabinetID: cabinetID,
			GADevice:         new("AAA"),
			DeviceName:       "CAB_AAA",
		}},
	}
	dispatcher := &updateCommandDispatcherStub{}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		ProjectLinks: &updateProjectLinkReaderStub{links: []*domainProject.ProjectSPSController{
			{ProjectID: projectID, SPSControllerID: controllerID},
		}},
		Dispatcher: dispatcher,
	})

	_, err := handler.Execute(context.Background(), UpdateCommand{
		SPSControllerID:   controllerID,
		DeviceDescription: &description,
	})
	if err != nil {
		t.Fatalf("execute update: %v", err)
	}
	if harness.runnerCalls != 1 || harness.updateCalls != 1 || harness.updateWithCalls != 0 {
		t.Fatalf("unexpected transaction/update calls: runner=%d update=%d updateWith=%d", harness.runnerCalls, harness.updateCalls, harness.updateWithCalls)
	}
	if len(dispatcher.commands) != 1 {
		t.Fatalf("commands: got %d, want 1", len(dispatcher.commands))
	}
	if _, ok := dispatcher.commands[0].(appcollaboration.SPSControllerUpdated); !ok {
		t.Fatalf("command: got %T, want SPSControllerUpdated", dispatcher.commands[0])
	}
}

func TestUpdateRollbackDoesNotResolveProjectsOrDispatch(t *testing.T) {
	controllerID := spsTestUUID(1)
	oldCabinetID := spsTestUUID(2)
	newCabinetID := spsTestUUID(3)
	writeErr := errors.New("system type write failed")
	harness := &updateTransactionHarness{
		committed: updateTransactionState{controller: &domainFacility.SPSController{
			Base:             domain.Base{ID: controllerID},
			ControlCabinetID: oldCabinetID,
			GADevice:         new("AAA"),
			DeviceName:       "OLD_AAA",
		}},
		updateErr: writeErr,
	}
	links := &updateProjectLinkReaderStub{}
	dispatcher := &updateCommandDispatcherStub{}
	emptySystemTypes := []domainFacility.SPSControllerSystemType{}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		ProjectLinks:        links,
		Dispatcher:          dispatcher,
	})

	_, err := handler.Execute(context.Background(), UpdateCommand{
		SPSControllerID:  controllerID,
		ControlCabinetID: &newCabinetID,
		SystemTypes:      &emptySystemTypes,
	})
	if !errors.Is(err, writeErr) {
		t.Fatalf("error: got %v, want %v", err, writeErr)
	}
	if harness.committed.controller.ControlCabinetID != oldCabinetID ||
		len(harness.committed.history) != 0 {
		t.Fatalf("rollback leaked staged state: %+v", harness.committed)
	}
	if links.calls != 0 || len(dispatcher.commands) != 0 {
		t.Fatalf("post-commit work ran after rollback: links=%d commands=%d", links.calls, len(dispatcher.commands))
	}
}

func TestUpdateCommitFailureDoesNotResolveProjectsOrDispatch(t *testing.T) {
	controllerID := spsTestUUID(1)
	cabinetID := spsTestUUID(2)
	commitErr := errors.New("commit failed")
	description := "staged only"
	harness := &updateTransactionHarness{
		committed: updateTransactionState{controller: &domainFacility.SPSController{
			Base:             domain.Base{ID: controllerID},
			ControlCabinetID: cabinetID,
			GADevice:         new("AAA"),
			DeviceName:       "CAB_AAA",
		}},
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
		SPSControllerID:   controllerID,
		DeviceDescription: &description,
	})
	if !errors.Is(err, commitErr) {
		t.Fatalf("error: got %v, want %v", err, commitErr)
	}
	if harness.committed.controller.DeviceDescription != nil ||
		len(harness.committed.history) != 0 {
		t.Fatalf("commit failure leaked staged state: %+v", harness.committed)
	}
	if links.calls != 0 || len(dispatcher.commands) != 0 {
		t.Fatalf("post-commit work ran after commit failure: links=%d commands=%d", links.calls, len(dispatcher.commands))
	}
}

func TestUpdateReportsDispatchFailureWithoutFailingCommittedHTTPResult(t *testing.T) {
	controllerID := spsTestUUID(1)
	cabinetID := spsTestUUID(2)
	projectID := spsTestUUID(11)
	dispatchErr := errors.New("realtime unavailable")
	description := "Committed"
	harness := &updateTransactionHarness{committed: updateTransactionState{
		controller: &domainFacility.SPSController{
			Base:             domain.Base{ID: controllerID},
			ControlCabinetID: cabinetID,
			GADevice:         new("AAA"),
			DeviceName:       "CAB_AAA",
		},
	}}
	dispatcher := &updateCommandDispatcherStub{err: dispatchErr}
	reported := make([]error, 0)
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		ProjectLinks: &updateProjectLinkReaderStub{links: []*domainProject.ProjectSPSController{
			{ProjectID: projectID, SPSControllerID: controllerID},
		}},
		Dispatcher:  dispatcher,
		ReportError: func(err error) { reported = append(reported, err) },
	})

	updated, err := handler.Update(context.Background(), UpdateCommand{
		SPSControllerID:   controllerID,
		DeviceDescription: &description,
	})
	if err != nil {
		t.Fatalf("HTTP-compatible update: %v", err)
	}
	if updated.DeviceDescription == nil || *updated.DeviceDescription != description {
		t.Fatalf("unexpected committed controller: %+v", updated)
	}
	if len(reported) != 1 || !errors.Is(reported[0], dispatchErr) {
		t.Fatalf("reported errors: got %v, want dispatch error", reported)
	}
}

func TestUpdateWrapsInitialLoadFailure(t *testing.T) {
	harness := &updateTransactionHarness{committed: updateTransactionState{}}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
	})

	_, err := handler.Execute(context.Background(), UpdateCommand{
		SPSControllerID: spsTestUUID(1),
	})
	var loadErr *LoadError
	if !errors.As(err, &loadErr) || !errors.Is(loadErr.Err, domain.ErrNotFound) {
		t.Fatalf("error: got %v, want wrapped load not found", err)
	}
}

func spsTestUUID(last int) uuid.UUID {
	return uuid.UUID{14: byte(last / 256), 15: byte(last % 256)}
}
