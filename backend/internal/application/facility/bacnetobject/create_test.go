package bacnetobject

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
	object  *domainFacility.BacnetObject
	history []string
}

func (s createTransactionState) clone() createTransactionState {
	return createTransactionState{
		object:  cloneBacnetObject(s.object),
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
	objectData     *domainFacility.ObjectData
	createErr      error
	reloadErr      error
	ownerReloadErr error
	commitErr      error
	runnerCalls    int
	createCalls    int
	historyBatchID *uuid.UUID
	fieldDeviceID  *uuid.UUID
	objectDataID   *uuid.UUID
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

func (s *createWorkflowStub) CreateWithParent(
	ctx context.Context,
	object *domainFacility.BacnetObject,
	fieldDeviceID *uuid.UUID,
	objectDataID *uuid.UUID,
) error {
	s.harness.createCalls++
	if batchID, ok := ctx.Value(createHistoryBatchKey{}).(uuid.UUID); ok {
		s.harness.historyBatchID = &batchID
	}
	s.harness.fieldDeviceID = clonePointer(fieldDeviceID)
	s.harness.objectDataID = clonePointer(objectDataID)
	created := cloneBacnetObject(object)
	created.ID = s.harness.createdID
	created.CreatedAt = s.harness.createdAt
	created.UpdatedAt = s.harness.createdAt
	created.FieldDeviceID = clonePointer(fieldDeviceID)
	object.ID = created.ID
	s.state.object = created
	s.state.history = append(s.state.history, "bacnet_object:create")
	return s.harness.createErr
}

func (s *createWorkflowStub) GetByID(
	_ context.Context,
	id uuid.UUID,
) (*domainFacility.BacnetObject, error) {
	if s.harness.reloadErr != nil {
		return nil, s.harness.reloadErr
	}
	if s.state.object == nil || s.state.object.ID != id {
		return nil, domain.ErrNotFound
	}
	return cloneBacnetObject(s.state.object), nil
}

func (s *createWorkflowStub) GetObjectDataByID(
	_ context.Context,
	id uuid.UUID,
) (*domainFacility.ObjectData, error) {
	if s.harness.ownerReloadErr != nil {
		return nil, s.harness.ownerReloadErr
	}
	if s.harness.objectData == nil || s.harness.objectData.ID != id {
		return nil, domain.ErrNotFound
	}
	clone := *s.harness.objectData
	clone.ProjectID = clonePointer(s.harness.objectData.ProjectID)
	return &clone, nil
}

type createProjectLinkReaderStub struct {
	harness *createTransactionHarness
	links   []*domainProject.ProjectFieldDevice
	err     error
	calls   int
	ids     []uuid.UUID
}

func (s *createProjectLinkReaderStub) GetByFieldDeviceIDs(
	_ context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectFieldDevice, error) {
	s.calls++
	s.ids = append([]uuid.UUID(nil), ids...)
	if s.harness != nil && s.harness.committed.object == nil {
		return nil, errors.New("project scope resolved before commit")
	}
	return s.links, s.err
}

func TestCreateForFieldDeviceCommitsHistoryAndDispatchesLinkedProjects(t *testing.T) {
	fieldDeviceID := bacnetTestUUID(1)
	objectID := bacnetTestUUID(2)
	projectOne := bacnetTestUUID(11)
	projectTwo := bacnetTestUUID(12)
	actorID := bacnetTestUUID(21)
	operationID := bacnetTestUUID(31)
	eventOne := bacnetTestUUID(32)
	eventTwo := bacnetTestUUID(33)
	createdAt := time.Date(2026, time.July, 21, 0, 0, 0, 0, time.UTC)
	occurredAt := createdAt.Add(time.Second)
	description := "Room temperature"
	alarmTypeID := bacnetTestUUID(41)

	harness := &createTransactionHarness{createdID: objectID, createdAt: createdAt}
	links := &createProjectLinkReaderStub{
		harness: harness,
		links: []*domainProject.ProjectFieldDevice{
			{ProjectID: projectTwo, FieldDeviceID: fieldDeviceID},
			{ProjectID: projectOne, FieldDeviceID: fieldDeviceID},
			{ProjectID: projectOne, FieldDeviceID: fieldDeviceID},
			{ProjectID: bacnetTestUUID(13), FieldDeviceID: bacnetTestUUID(99)},
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

	outcome, err := handler.ExecuteForFieldDevice(context.Background(), CreateForFieldDeviceCommand{
		FieldDeviceID: fieldDeviceID,
		Input: CreateInput{
			TextFix:          "AI",
			Description:      &description,
			GMSVisible:       true,
			SoftwareType:     domainFacility.BacnetSoftwareTypeAI,
			SoftwareNumber:   7,
			HardwareType:     domainFacility.BacnetHardwareTypeAI,
			HardwareQuantity: 1,
			AlarmTypeID:      &alarmTypeID,
		},
	})
	if err != nil {
		t.Fatalf("execute create: %v", err)
	}
	if harness.runnerCalls != 1 || harness.createCalls != 1 ||
		harness.fieldDeviceID == nil || *harness.fieldDeviceID != fieldDeviceID ||
		harness.objectDataID != nil {
		t.Fatalf("unexpected transaction call: %+v", harness)
	}
	if harness.historyBatchID == nil || *harness.historyBatchID != operationID ||
		len(harness.committed.history) != 1 {
		t.Fatalf("unexpected committed history: batch=%v history=%v", harness.historyBatchID, harness.committed.history)
	}
	if outcome.BacnetObject.ID != objectID ||
		outcome.BacnetObject.FieldDeviceID == nil || *outcome.BacnetObject.FieldDeviceID != fieldDeviceID {
		t.Fatalf("unexpected created object: %+v", outcome.BacnetObject)
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
	if want := []uuid.UUID{fieldDeviceID}; !reflect.DeepEqual(links.ids, want) {
		t.Fatalf("project lookup IDs: got %v, want %v", links.ids, want)
	}
	if len(outcome.Mutation.Changes) != 1 {
		t.Fatalf("changes: got %d, want 1", len(outcome.Mutation.Changes))
	}
	change := outcome.Mutation.Changes[0]
	if change.EntityID != objectID || change.ParentID == nil || *change.ParentID != fieldDeviceID ||
		change.Action != domainHistory.ActionCreate || len(change.Before) != 0 {
		t.Fatalf("unexpected create change: %+v", change)
	}
	var snapshot bacnetObjectSnapshot
	if err := json.Unmarshal(change.After, &snapshot); err != nil {
		t.Fatalf("decode after snapshot: %v", err)
	}
	if snapshot.ID != objectID || snapshot.TextFix != "AI" ||
		snapshot.AlarmTypeID == nil || *snapshot.AlarmTypeID != alarmTypeID {
		t.Fatalf("unexpected after snapshot: %+v", snapshot)
	}
	if len(dispatcher.commands) != 2 {
		t.Fatalf("commands: got %d, want 2", len(dispatcher.commands))
	}
	for index, raw := range dispatcher.commands {
		command, ok := raw.(appcollaboration.BacnetObjectCreated)
		if !ok {
			t.Fatalf("command: got %T, want BacnetObjectCreated", raw)
		}
		wantProjectID := []uuid.UUID{projectOne, projectTwo}[index]
		if command.ProjectID != wantProjectID || command.BacnetObjectID != objectID ||
			command.FieldDeviceID != fieldDeviceID || command.OperationID != operationID ||
			command.CorrelationID != operationID || command.SchemaVersion != appcollaboration.SchemaVersionV1 {
			t.Fatalf("unexpected collaboration command: %+v", command)
		}
	}
}

func TestCreateForFieldDeviceFailureDoesNotDispatchOrEscapeTransaction(t *testing.T) {
	createErr := errors.New("create failed")
	commitErr := errors.New("commit failed")
	for _, test := range []struct {
		name      string
		createErr error
		reloadErr error
		commitErr error
	}{
		{name: "write", createErr: createErr},
		{name: "reload", reloadErr: errors.New("reload failed")},
		{name: "commit", commitErr: commitErr},
	} {
		t.Run(test.name, func(t *testing.T) {
			harness := &createTransactionHarness{
				createdID: bacnetTestUUID(2),
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

			_, err := handler.ExecuteForFieldDevice(context.Background(), CreateForFieldDeviceCommand{
				FieldDeviceID: bacnetTestUUID(1),
				Input: CreateInput{
					TextFix:        "AI",
					SoftwareType:   domainFacility.BacnetSoftwareTypeAI,
					SoftwareNumber: 1,
				},
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
			if harness.committed.object != nil || len(harness.committed.history) != 0 {
				t.Fatalf("failed mutation escaped transaction: %+v", harness.committed)
			}
			if links.calls != 0 || len(dispatcher.commands) != 0 {
				t.Fatalf("post-commit work ran after failure: links=%d commands=%d", links.calls, len(dispatcher.commands))
			}
		})
	}
}

func TestCreateForFieldDeviceReportsPostCommitScopeFailureWithoutFailingCreate(t *testing.T) {
	fieldDeviceID := bacnetTestUUID(1)
	objectID := bacnetTestUUID(2)
	scopeErr := errors.New("scope lookup failed")
	harness := &createTransactionHarness{createdID: objectID}
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

	created, err := handler.CreateForFieldDevice(context.Background(), CreateForFieldDeviceCommand{
		FieldDeviceID: fieldDeviceID,
		Input: CreateInput{
			TextFix:        "AI",
			SoftwareType:   domainFacility.BacnetSoftwareTypeAI,
			SoftwareNumber: 1,
		},
	})
	if err != nil || created == nil || created.ID != objectID {
		t.Fatalf("committed create: object=%+v err=%v", created, err)
	}
	if len(reported) != 1 || !errors.Is(reported[0], scopeErr) {
		t.Fatalf("reported errors: got %v, want wrapped %v", reported, scopeErr)
	}
	if len(dispatcher.commands) != 0 || harness.committed.object == nil {
		t.Fatalf("unexpected post-commit state: commands=%d committed=%+v", len(dispatcher.commands), harness.committed)
	}
}

func TestCreateForFieldDeviceRejectsMissingParentBeforeTransaction(t *testing.T) {
	harness := &createTransactionHarness{}
	handler := NewCreateHandler(CreateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
	})

	_, err := handler.ExecuteForFieldDevice(context.Background(), CreateForFieldDeviceCommand{})
	if !errors.Is(err, domain.ErrInvalidArgument) || harness.runnerCalls != 0 {
		t.Fatalf("missing parent: err=%v runnerCalls=%d", err, harness.runnerCalls)
	}
}

func TestCreateForObjectDataCommitsWithServerResolvedProjectAndDispatchesFullRefresh(t *testing.T) {
	objectDataID := bacnetTestUUID(51)
	objectID := bacnetTestUUID(52)
	projectID := bacnetTestUUID(53)
	actorID := bacnetTestUUID(54)
	operationID := bacnetTestUUID(55)
	eventID := bacnetTestUUID(56)
	createdAt := time.Date(2026, time.July, 21, 1, 0, 0, 0, time.UTC)
	occurredAt := createdAt.Add(time.Second)

	harness := &createTransactionHarness{
		createdID: objectID,
		createdAt: createdAt,
		objectData: &domainFacility.ObjectData{
			Base:      domain.Base{ID: objectDataID},
			ProjectID: &projectID,
		},
	}
	dispatcher := &updateCommandDispatcherStub{}
	ids := []uuid.UUID{operationID, eventID}
	handler := NewCreateHandler(CreateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, createHistoryBatchKey{}, batchID)
		},
		Dispatcher: dispatcher,
		Actor:      func(context.Context) *uuid.UUID { return &actorID },
		NewID: func() uuid.UUID {
			id := ids[0]
			ids = ids[1:]
			return id
		},
		Now: func() time.Time { return occurredAt },
	})

	outcome, err := handler.ExecuteForObjectData(context.Background(), CreateForObjectDataCommand{
		ObjectDataID: objectDataID,
		Input: CreateInput{
			TextFix:        "AI",
			SoftwareType:   domainFacility.BacnetSoftwareTypeAI,
			SoftwareNumber: 4,
		},
	})
	if err != nil {
		t.Fatalf("execute create: %v", err)
	}
	if harness.runnerCalls != 1 || harness.createCalls != 1 ||
		harness.fieldDeviceID != nil || harness.objectDataID == nil ||
		*harness.objectDataID != objectDataID {
		t.Fatalf("unexpected transaction call: %+v", harness)
	}
	if harness.historyBatchID == nil || *harness.historyBatchID != operationID ||
		len(harness.committed.history) != 1 {
		t.Fatalf("unexpected committed history: batch=%v history=%v", harness.historyBatchID, harness.committed.history)
	}
	if outcome.BacnetObject.ID != objectID || outcome.BacnetObject.FieldDeviceID != nil {
		t.Fatalf("unexpected created object: %+v", outcome.BacnetObject)
	}
	if outcome.Mutation.OperationID != operationID ||
		outcome.Mutation.BatchID == nil || *outcome.Mutation.BatchID != operationID ||
		outcome.Mutation.ActorID == nil || *outcome.Mutation.ActorID != actorID ||
		!reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectID}) {
		t.Fatalf("unexpected mutation: %+v", outcome.Mutation)
	}
	change := outcome.Mutation.Changes[0]
	if change.EntityID != objectID || change.ParentID == nil || *change.ParentID != objectDataID ||
		change.Action != domainHistory.ActionCreate || len(change.Before) != 0 {
		t.Fatalf("unexpected create change: %+v", change)
	}
	if len(dispatcher.commands) != 1 {
		t.Fatalf("commands: got %d, want 1", len(dispatcher.commands))
	}
	command, ok := dispatcher.commands[0].(appcollaboration.FacilityHierarchyRefreshRequired)
	if !ok {
		t.Fatalf("command: got %T, want FacilityHierarchyRefreshRequired", dispatcher.commands[0])
	}
	if command.ProjectID != projectID || command.Scope != appcollaboration.FacilityScopeProject ||
		!command.FullRefresh || len(command.EntityIDs) != 0 ||
		command.OperationID != operationID || command.EventID != eventID ||
		command.CorrelationID != operationID || command.SchemaVersion != appcollaboration.SchemaVersionV1 {
		t.Fatalf("unexpected collaboration command: %+v", command)
	}
}

func TestCreateForObjectDataWithoutProjectCommitsWithoutDispatch(t *testing.T) {
	objectDataID := bacnetTestUUID(61)
	harness := &createTransactionHarness{
		createdID:  bacnetTestUUID(62),
		objectData: &domainFacility.ObjectData{Base: domain.Base{ID: objectDataID}},
	}
	dispatcher := &updateCommandDispatcherStub{}
	handler := NewCreateHandler(CreateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Dispatcher:          dispatcher,
	})

	outcome, err := handler.ExecuteForObjectData(context.Background(), CreateForObjectDataCommand{
		ObjectDataID: objectDataID,
		Input: CreateInput{
			TextFix:        "AI",
			SoftwareType:   domainFacility.BacnetSoftwareTypeAI,
			SoftwareNumber: 1,
		},
	})
	if err != nil || outcome.BacnetObject == nil {
		t.Fatalf("committed create: outcome=%+v err=%v", outcome, err)
	}
	if len(outcome.Mutation.ProjectIDs) != 0 || len(dispatcher.commands) != 0 {
		t.Fatalf("unexpected global-template dispatch: projects=%v commands=%v", outcome.Mutation.ProjectIDs, dispatcher.commands)
	}
}

func TestCreateForObjectDataReportsPostCommitDispatchFailureWithoutFailingCreate(t *testing.T) {
	objectDataID := bacnetTestUUID(63)
	projectID := bacnetTestUUID(64)
	dispatchErr := errors.New("dispatch failed")
	harness := &createTransactionHarness{
		createdID: bacnetTestUUID(65),
		objectData: &domainFacility.ObjectData{
			Base:      domain.Base{ID: objectDataID},
			ProjectID: &projectID,
		},
	}
	dispatcher := &updateCommandDispatcherStub{err: dispatchErr}
	var reported []error
	handler := NewCreateHandler(CreateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Dispatcher:          dispatcher,
		ReportError: func(err error) {
			reported = append(reported, err)
		},
	})

	created, err := handler.CreateForObjectData(context.Background(), CreateForObjectDataCommand{
		ObjectDataID: objectDataID,
		Input: CreateInput{
			TextFix:        "AI",
			SoftwareType:   domainFacility.BacnetSoftwareTypeAI,
			SoftwareNumber: 1,
		},
	})
	if err != nil || created == nil || created.ID != harness.createdID {
		t.Fatalf("committed create: object=%+v err=%v", created, err)
	}
	if len(reported) != 1 || !errors.Is(reported[0], dispatchErr) {
		t.Fatalf("reported errors: got %v, want wrapped %v", reported, dispatchErr)
	}
	if harness.committed.object == nil || len(dispatcher.commands) != 1 {
		t.Fatalf("unexpected post-commit state: committed=%+v commands=%v", harness.committed, dispatcher.commands)
	}
}

func TestCreateForObjectDataFailureDoesNotDispatchOrEscapeTransaction(t *testing.T) {
	createErr := errors.New("create failed")
	commitErr := errors.New("commit failed")
	ownerErr := errors.New("owner reload failed")
	for _, test := range []struct {
		name           string
		createErr      error
		ownerReloadErr error
		commitErr      error
	}{
		{name: "write", createErr: createErr},
		{name: "owner reload", ownerReloadErr: ownerErr},
		{name: "commit", commitErr: commitErr},
	} {
		t.Run(test.name, func(t *testing.T) {
			objectDataID := bacnetTestUUID(71)
			harness := &createTransactionHarness{
				createdID:      bacnetTestUUID(72),
				objectData:     &domainFacility.ObjectData{Base: domain.Base{ID: objectDataID}},
				createErr:      test.createErr,
				ownerReloadErr: test.ownerReloadErr,
				commitErr:      test.commitErr,
			}
			dispatcher := &updateCommandDispatcherStub{}
			handler := NewCreateHandler(CreateDependencies{
				TransactionRunner:   harness.runner,
				TransactionWorkflow: harness.factory,
				Dispatcher:          dispatcher,
			})

			_, err := handler.ExecuteForObjectData(context.Background(), CreateForObjectDataCommand{
				ObjectDataID: objectDataID,
				Input: CreateInput{
					TextFix:        "AI",
					SoftwareType:   domainFacility.BacnetSoftwareTypeAI,
					SoftwareNumber: 1,
				},
			})
			wantErr := test.createErr
			if wantErr == nil {
				wantErr = test.ownerReloadErr
			}
			if wantErr == nil {
				wantErr = test.commitErr
			}
			if !errors.Is(err, wantErr) {
				t.Fatalf("error: got %v, want %v", err, wantErr)
			}
			if harness.committed.object != nil || len(harness.committed.history) != 0 {
				t.Fatalf("failed mutation escaped transaction: %+v", harness.committed)
			}
			if len(dispatcher.commands) != 0 {
				t.Fatalf("post-commit command ran after failure: %v", dispatcher.commands)
			}
		})
	}
}

func TestCreateForObjectDataRejectsMissingParentBeforeTransaction(t *testing.T) {
	harness := &createTransactionHarness{}
	handler := NewCreateHandler(CreateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
	})

	_, err := handler.ExecuteForObjectData(context.Background(), CreateForObjectDataCommand{})
	if !errors.Is(err, domain.ErrInvalidArgument) || harness.runnerCalls != 0 {
		t.Fatalf("missing parent: err=%v runnerCalls=%d", err, harness.runnerCalls)
	}
}
