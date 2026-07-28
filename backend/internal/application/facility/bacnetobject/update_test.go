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
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type updateTransactionState struct {
	object          *domainFacility.BacnetObject
	history         []string
	objectDataLinks []uuid.UUID
}

func (s updateTransactionState) clone() updateTransactionState {
	return updateTransactionState{
		object:          cloneBacnetObject(s.object),
		history:         append([]string(nil), s.history...),
		objectDataLinks: append([]uuid.UUID(nil), s.objectDataLinks...),
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
	runnerCalls    int
	updateCalls    int
	historyBatchID *uuid.UUID
	objectDataID   *uuid.UUID
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
) (*domainFacility.BacnetObject, error) {
	if s.state.object == nil || s.state.object.ID != id {
		return nil, domain.ErrNotFound
	}
	return cloneBacnetObject(s.state.object), nil
}

func (s *updateWorkflowStub) Update(
	ctx context.Context,
	object *domainFacility.BacnetObject,
	objectDataID *uuid.UUID,
) error {
	s.harness.updateCalls++
	if batchID, ok := ctx.Value(updateHistoryBatchKey{}).(uuid.UUID); ok {
		s.harness.historyBatchID = &batchID
	}
	s.harness.objectDataID = clonePointer(objectDataID)
	updated := cloneBacnetObject(object)
	updated.UpdatedAt = s.harness.updatedAt
	s.state.object = updated
	s.state.history = append(s.state.history, "bacnet_object:update")
	if objectDataID != nil {
		s.state.objectDataLinks = append(s.state.objectDataLinks, *objectDataID)
	}
	return s.harness.updateErr
}

type updateProjectLinkReaderStub struct {
	harness               *updateTransactionHarness
	expectedFieldDeviceID *uuid.UUID
	links                 []*domainProject.ProjectFieldDevice
	err                   error
	calls                 int
	received              []uuid.UUID
}

func (s *updateProjectLinkReaderStub) GetByFieldDeviceIDs(
	_ context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectFieldDevice, error) {
	s.calls++
	s.received = append([]uuid.UUID(nil), ids...)
	if s.harness != nil && s.expectedFieldDeviceID != nil &&
		(s.harness.committed.object.FieldDeviceID == nil ||
			*s.harness.committed.object.FieldDeviceID != *s.expectedFieldDeviceID) {
		return nil, errors.New("project scope resolved before commit")
	}
	return s.links, s.err
}

type updateObjectDataOwnerReaderStub struct {
	harness       *updateTransactionHarness
	committedText string
	owners        []domainObjectData.BacnetObjectOwner
	err           error
	calls         int
	received      []uuid.UUID
}

func (s *updateObjectDataOwnerReaderStub) GetByBacnetObjectIDs(
	_ context.Context,
	ids []uuid.UUID,
) ([]domainObjectData.BacnetObjectOwner, error) {
	s.calls++
	s.received = append([]uuid.UUID(nil), ids...)
	if s.harness != nil && s.committedText != "" &&
		(s.harness.committed.object == nil ||
			s.harness.committed.object.TextFix != s.committedText) {
		return nil, errors.New("ObjectData owners resolved before commit")
	}
	return append([]domainObjectData.BacnetObjectOwner(nil), s.owners...), s.err
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

func TestUpdateLoadsAuthoritativeStateCommitsHistoryAndDispatchesFilteredOldAndNewOwners(t *testing.T) {
	objectID := bacnetTestUUID(1)
	oldFieldDeviceID := bacnetTestUUID(2)
	newFieldDeviceID := bacnetTestUUID(3)
	projectOld := bacnetTestUUID(11)
	projectNew := bacnetTestUUID(12)
	projectBoth := bacnetTestUUID(13)
	actorID := bacnetTestUUID(21)
	operationID := bacnetTestUUID(31)
	eventOld := bacnetTestUUID(32)
	eventNew := bacnetTestUUID(33)
	eventBoth := bacnetTestUUID(34)
	updatedAt := time.Date(2026, time.July, 20, 22, 0, 0, 0, time.UTC)
	occurredAt := updatedAt.Add(time.Second)
	newTextFix := "NEW"

	harness := &updateTransactionHarness{
		committed: updateTransactionState{object: &domainFacility.BacnetObject{
			Base:           domain.Base{ID: objectID},
			TextFix:        "DATABASE_VALUE",
			SoftwareType:   domainFacility.BacnetSoftwareTypeAI,
			SoftwareNumber: 1,
			FieldDeviceID:  &oldFieldDeviceID,
		}},
		updatedAt: updatedAt,
	}
	links := &updateProjectLinkReaderStub{
		harness:               harness,
		expectedFieldDeviceID: &newFieldDeviceID,
		links: []*domainProject.ProjectFieldDevice{
			{ProjectID: projectOld, FieldDeviceID: oldFieldDeviceID},
			{ProjectID: projectNew, FieldDeviceID: newFieldDeviceID},
			{ProjectID: projectBoth, FieldDeviceID: newFieldDeviceID},
			{ProjectID: projectBoth, FieldDeviceID: oldFieldDeviceID},
			{ProjectID: projectBoth, FieldDeviceID: oldFieldDeviceID},
			{ProjectID: bacnetTestUUID(14), FieldDeviceID: bacnetTestUUID(99)},
		},
	}
	dispatcher := &updateCommandDispatcherStub{}
	ids := []uuid.UUID{operationID, eventOld, eventNew, eventBoth}
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
		BacnetObjectID: objectID,
		FieldDeviceID:  &newFieldDeviceID,
		Patch: domainFacility.BacnetObjectPatch{
			TextFix: &newTextFix,
		},
	})
	if err != nil {
		t.Fatalf("execute update: %v", err)
	}
	if harness.runnerCalls != 1 || harness.updateCalls != 1 {
		t.Fatalf("transaction/update calls: runner=%d update=%d", harness.runnerCalls, harness.updateCalls)
	}
	if harness.committed.object.TextFix != newTextFix ||
		harness.committed.object.FieldDeviceID == nil ||
		*harness.committed.object.FieldDeviceID != newFieldDeviceID {
		t.Fatalf("unexpected committed object: %+v", harness.committed.object)
	}
	if harness.historyBatchID == nil || *harness.historyBatchID != operationID {
		t.Fatalf("history batch: got %v, want %s", harness.historyBatchID, operationID)
	}
	if outcome.Mutation.BatchID == nil || *outcome.Mutation.BatchID != operationID ||
		outcome.Mutation.ActorID == nil || *outcome.Mutation.ActorID != actorID {
		t.Fatalf("unexpected mutation envelope: %+v", outcome.Mutation)
	}
	if want := []mutation.FieldName{mutation.FieldNameTextFix, mutation.FieldNameFieldDevice}; !reflect.DeepEqual(outcome.Mutation.Changes[0].ChangedFields, want) {
		t.Fatalf("changed fields: got %v, want %v", outcome.Mutation.Changes[0].ChangedFields, want)
	}
	var beforeSnapshot bacnetObjectSnapshot
	if err := json.Unmarshal(outcome.Mutation.Changes[0].Before, &beforeSnapshot); err != nil {
		t.Fatalf("decode before snapshot: %v", err)
	}
	if beforeSnapshot.TextFix != "DATABASE_VALUE" {
		t.Fatalf("expected authoritative before snapshot, got %+v", beforeSnapshot)
	}
	if want := []uuid.UUID{oldFieldDeviceID, newFieldDeviceID}; !reflect.DeepEqual(links.received, want) {
		t.Fatalf("project lookup IDs: got %v, want %v", links.received, want)
	}
	if want := []uuid.UUID{projectOld, projectNew, projectBoth}; !reflect.DeepEqual(outcome.Mutation.ProjectIDs, want) {
		t.Fatalf("project IDs: got %v, want %v", outcome.Mutation.ProjectIDs, want)
	}
	if len(dispatcher.commands) != 3 {
		t.Fatalf("commands: got %d, want 3", len(dispatcher.commands))
	}
	wantDevices := map[uuid.UUID][]uuid.UUID{
		projectOld:  {oldFieldDeviceID},
		projectNew:  {newFieldDeviceID},
		projectBoth: {oldFieldDeviceID, newFieldDeviceID},
	}
	for _, raw := range dispatcher.commands {
		command, ok := raw.(appcollaboration.BacnetObjectUpdated)
		if !ok {
			t.Fatalf("command: got %T, want BacnetObjectUpdated", raw)
		}
		if command.BacnetObjectID != objectID || command.OperationID != operationID ||
			!reflect.DeepEqual(command.FieldDeviceIDs, wantDevices[command.ProjectID]) {
			t.Fatalf("unexpected collaboration command: %+v", command)
		}
	}
}

func TestUpdateFailureAndCommitFailureDoNotDispatchOrEscapeState(t *testing.T) {
	for _, test := range []struct {
		name      string
		updateErr error
		commitErr error
	}{
		{name: "write", updateErr: errors.New("write failed")},
		{name: "commit", commitErr: errors.New("commit failed")},
	} {
		t.Run(test.name, func(t *testing.T) {
			objectID := bacnetTestUUID(1)
			fieldDeviceID := bacnetTestUUID(2)
			newTextFix := "NEW"
			harness := &updateTransactionHarness{
				committed: updateTransactionState{object: &domainFacility.BacnetObject{
					Base:          domain.Base{ID: objectID},
					TextFix:       "OLD",
					SoftwareType:  domainFacility.BacnetSoftwareTypeAI,
					FieldDeviceID: &fieldDeviceID,
				}},
				updateErr: test.updateErr,
				commitErr: test.commitErr,
			}
			links := &updateProjectLinkReaderStub{}
			owners := &updateObjectDataOwnerReaderStub{}
			dispatcher := &updateCommandDispatcherStub{}
			handler := NewUpdateHandler(UpdateDependencies{
				TransactionRunner:   harness.runner,
				TransactionWorkflow: harness.factory,
				ProjectLinks:        links,
				ObjectDataOwners:    owners,
				Dispatcher:          dispatcher,
			})

			_, err := handler.Execute(context.Background(), UpdateCommand{
				BacnetObjectID: objectID,
				Patch:          domainFacility.BacnetObjectPatch{TextFix: &newTextFix},
			})
			wantErr := test.updateErr
			if wantErr == nil {
				wantErr = test.commitErr
			}
			if !errors.Is(err, wantErr) {
				t.Fatalf("error: got %v, want %v", err, wantErr)
			}
			if harness.committed.object.TextFix != "OLD" || len(harness.committed.history) != 0 {
				t.Fatalf("failed mutation escaped transaction: %+v", harness.committed)
			}
			if links.calls != 0 || owners.calls != 0 || len(dispatcher.commands) != 0 {
				t.Fatalf("post-commit work ran after failure: links=%d owners=%d commands=%d", links.calls, owners.calls, len(dispatcher.commands))
			}
		})
	}
}

func TestObjectDataOnlyUpdateUsesServerOwnedProjectRefreshAfterCommit(t *testing.T) {
	objectID := bacnetTestUUID(1)
	objectDataID := bacnetTestUUID(4)
	softwareReferenceID := bacnetTestUUID(5)
	projectID := bacnetTestUUID(6)
	harness := &updateTransactionHarness{committed: updateTransactionState{object: &domainFacility.BacnetObject{
		Base:         domain.Base{ID: objectID},
		TextFix:      "AI",
		SoftwareType: domainFacility.BacnetSoftwareTypeAI,
	}}}
	links := &updateProjectLinkReaderStub{}
	otherProjectID := bacnetTestUUID(7)
	owners := &updateObjectDataOwnerReaderStub{
		harness:       harness,
		committedText: "AI",
		owners: []domainObjectData.BacnetObjectOwner{
			{BacnetObjectID: objectID, ObjectDataID: objectDataID, ProjectID: &projectID},
			{BacnetObjectID: objectID, ObjectDataID: objectDataID, ProjectID: &projectID},
			{BacnetObjectID: bacnetTestUUID(99), ObjectDataID: objectDataID, ProjectID: &otherProjectID},
		},
	}
	dispatcher := &updateCommandDispatcherStub{}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		ProjectLinks:        links,
		ObjectDataOwners:    owners,
		Dispatcher:          dispatcher,
	})

	outcome, err := handler.Execute(context.Background(), UpdateCommand{
		BacnetObjectID: objectID,
		ObjectDataID:   &objectDataID,
		Patch: domainFacility.BacnetObjectPatch{
			SoftwareReferenceID: &softwareReferenceID,
		},
	})
	if err != nil {
		t.Fatalf("execute ObjectData update: %v", err)
	}
	if harness.objectDataID == nil || *harness.objectDataID != objectDataID ||
		len(harness.committed.objectDataLinks) != 1 {
		t.Fatalf("ObjectData compatibility path not preserved: %+v", harness)
	}
	if outcome.Mutation.Changes[0].ParentID == nil || *outcome.Mutation.Changes[0].ParentID != objectDataID {
		t.Fatalf("mutation parent: got %v, want %s", outcome.Mutation.Changes[0].ParentID, objectDataID)
	}
	if links.calls != 0 || owners.calls != 1 ||
		!reflect.DeepEqual(owners.received, []uuid.UUID{objectID}) {
		t.Fatalf("unexpected scope resolution: links=%d owners=%d ids=%v", links.calls, owners.calls, owners.received)
	}
	if !reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectID}) || len(dispatcher.commands) != 1 {
		t.Fatalf("unexpected recipients/commands: projects=%v commands=%v", outcome.Mutation.ProjectIDs, dispatcher.commands)
	}
	command, ok := dispatcher.commands[0].(appcollaboration.FacilityHierarchyRefreshRequired)
	if !ok || command.ProjectID != projectID || command.Scope != appcollaboration.FacilityScopeProject ||
		!command.FullRefresh || len(command.EntityIDs) != 0 {
		t.Fatalf("unexpected ObjectData refresh command: %+v", dispatcher.commands[0])
	}
}

func TestObjectDataGlobalOwnerDoesNotCreateProjectRecipient(t *testing.T) {
	objectID := bacnetTestUUID(11)
	objectDataID := bacnetTestUUID(12)
	newText := "NEW"
	harness := &updateTransactionHarness{committed: updateTransactionState{object: &domainFacility.BacnetObject{
		Base:         domain.Base{ID: objectID},
		TextFix:      "OLD",
		SoftwareType: domainFacility.BacnetSoftwareTypeAI,
	}}}
	owners := &updateObjectDataOwnerReaderStub{
		owners: []domainObjectData.BacnetObjectOwner{{
			BacnetObjectID: objectID,
			ObjectDataID:   objectDataID,
		}},
	}
	dispatcher := &updateCommandDispatcherStub{}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		ObjectDataOwners:    owners,
		Dispatcher:          dispatcher,
	})

	outcome, err := handler.Execute(context.Background(), UpdateCommand{
		BacnetObjectID: objectID,
		Patch:          domainFacility.BacnetObjectPatch{TextFix: &newText},
	})
	if err != nil {
		t.Fatalf("execute global-template update: %v", err)
	}
	if owners.calls != 1 || len(outcome.Mutation.ProjectIDs) != 0 || len(dispatcher.commands) != 0 {
		t.Fatalf("unexpected global-template recipients: calls=%d projects=%v commands=%v",
			owners.calls,
			outcome.Mutation.ProjectIDs,
			dispatcher.commands,
		)
	}
}

func TestUpdateCanExplicitlyDetachDirectFieldDeviceWithJSONNullSemantics(t *testing.T) {
	objectID := bacnetTestUUID(101)
	fieldDeviceID := bacnetTestUUID(102)
	harness := &updateTransactionHarness{
		committed: updateTransactionState{object: &domainFacility.BacnetObject{
			Base:           domain.Base{ID: objectID},
			TextFix:        "AI",
			SoftwareType:   domainFacility.BacnetSoftwareTypeAI,
			SoftwareNumber: 1,
			FieldDeviceID:  &fieldDeviceID,
		}},
	}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
	})

	outcome, err := handler.Execute(context.Background(), UpdateCommand{
		BacnetObjectID: objectID,
		FieldDeviceSet: true,
		FieldDeviceID:  nil,
	})
	if err != nil {
		t.Fatalf("detach FieldDevice: %v", err)
	}
	if harness.committed.object.FieldDeviceID != nil {
		t.Fatalf("FieldDevice attachment survived detach: %+v", harness.committed.object)
	}
	if !reflect.DeepEqual(outcome.Mutation.Changes[0].ChangedFields, []mutation.FieldName{
		mutation.FieldNameFieldDevice,
	}) {
		t.Fatalf("detach change fields: %+v", outcome.Mutation.Changes[0])
	}
}

func TestUpdatePrefersOneProjectRefreshWhenFieldDeviceAndObjectDataShareProject(t *testing.T) {
	objectID := bacnetTestUUID(21)
	fieldDeviceID := bacnetTestUUID(22)
	objectDataID := bacnetTestUUID(23)
	projectID := bacnetTestUUID(24)
	newText := "NEW"
	harness := &updateTransactionHarness{committed: updateTransactionState{object: &domainFacility.BacnetObject{
		Base:          domain.Base{ID: objectID},
		TextFix:       "OLD",
		SoftwareType:  domainFacility.BacnetSoftwareTypeAI,
		FieldDeviceID: &fieldDeviceID,
	}}}
	links := &updateProjectLinkReaderStub{links: []*domainProject.ProjectFieldDevice{{
		ProjectID:     projectID,
		FieldDeviceID: fieldDeviceID,
	}}}
	owners := &updateObjectDataOwnerReaderStub{owners: []domainObjectData.BacnetObjectOwner{{
		BacnetObjectID: objectID,
		ObjectDataID:   objectDataID,
		ProjectID:      &projectID,
	}}}
	dispatcher := &updateCommandDispatcherStub{}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		ProjectLinks:        links,
		ObjectDataOwners:    owners,
		Dispatcher:          dispatcher,
	})

	outcome, err := handler.Execute(context.Background(), UpdateCommand{
		BacnetObjectID: objectID,
		Patch:          domainFacility.BacnetObjectPatch{TextFix: &newText},
	})
	if err != nil {
		t.Fatalf("execute dual-owner update: %v", err)
	}
	if !reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectID}) ||
		len(dispatcher.commands) != 1 {
		t.Fatalf("unexpected deduplicated result: projects=%v commands=%v", outcome.Mutation.ProjectIDs, dispatcher.commands)
	}
	command, ok := dispatcher.commands[0].(appcollaboration.FacilityHierarchyRefreshRequired)
	if !ok || command.ProjectID != projectID || command.Scope != appcollaboration.FacilityScopeProject ||
		!command.FullRefresh {
		t.Fatalf("expected one broad project refresh, got %+v", dispatcher.commands[0])
	}
}

func TestObjectDataScopeFailureDoesNotSuppressKnownFieldDeviceRecipient(t *testing.T) {
	objectID := bacnetTestUUID(31)
	fieldDeviceID := bacnetTestUUID(32)
	projectID := bacnetTestUUID(33)
	ownerErr := errors.New("owner scope failed")
	newText := "NEW"
	harness := &updateTransactionHarness{committed: updateTransactionState{object: &domainFacility.BacnetObject{
		Base:          domain.Base{ID: objectID},
		TextFix:       "OLD",
		SoftwareType:  domainFacility.BacnetSoftwareTypeAI,
		FieldDeviceID: &fieldDeviceID,
	}}}
	links := &updateProjectLinkReaderStub{links: []*domainProject.ProjectFieldDevice{{
		ProjectID:     projectID,
		FieldDeviceID: fieldDeviceID,
	}}}
	owners := &updateObjectDataOwnerReaderStub{err: ownerErr}
	dispatcher := &updateCommandDispatcherStub{}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		ProjectLinks:        links,
		ObjectDataOwners:    owners,
		Dispatcher:          dispatcher,
	})

	outcome, err := handler.Execute(context.Background(), UpdateCommand{
		BacnetObjectID: objectID,
		Patch:          domainFacility.BacnetObjectPatch{TextFix: &newText},
	})
	if err != nil {
		t.Fatalf("execute update: %v", err)
	}
	if len(outcome.DispatchErrors) != 1 || !errors.Is(outcome.DispatchErrors[0], ownerErr) {
		t.Fatalf("dispatch errors: got %v, want wrapped %v", outcome.DispatchErrors, ownerErr)
	}
	if !reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectID}) ||
		len(dispatcher.commands) != 1 {
		t.Fatalf("known recipient lost: projects=%v commands=%v", outcome.Mutation.ProjectIDs, dispatcher.commands)
	}
	if _, ok := dispatcher.commands[0].(appcollaboration.BacnetObjectUpdated); !ok {
		t.Fatalf("command: got %T, want BacnetObjectUpdated", dispatcher.commands[0])
	}
}

func TestUpdateLoadErrorIsTypedAndDoesNotWrite(t *testing.T) {
	harness := &updateTransactionHarness{}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
	})

	_, err := handler.Execute(context.Background(), UpdateCommand{BacnetObjectID: uuid.New()})
	var loadErr *LoadError
	if !errors.As(err, &loadErr) || !errors.Is(loadErr.Err, domain.ErrNotFound) {
		t.Fatalf("expected typed not-found load error, got %v", err)
	}
	if harness.updateCalls != 0 {
		t.Fatalf("unexpected write after load failure: %d", harness.updateCalls)
	}
}

func TestUpdateRejectsAmbiguousParentBeforeTransaction(t *testing.T) {
	fieldDeviceID := uuid.New()
	objectDataID := uuid.New()
	harness := &updateTransactionHarness{}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
	})

	_, err := handler.Execute(context.Background(), UpdateCommand{
		BacnetObjectID: uuid.New(),
		FieldDeviceID:  &fieldDeviceID,
		ObjectDataID:   &objectDataID,
	})
	if !errors.Is(err, domain.ErrInvalidArgument) || harness.runnerCalls != 0 {
		t.Fatalf("ambiguous parent: err=%v runnerCalls=%d", err, harness.runnerCalls)
	}
}

func bacnetTestUUID(value byte) uuid.UUID {
	var id uuid.UUID
	id[15] = value
	return id
}
