package fielddevice

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
	domainCollaboration "github.com/besart951/go_infra_link/backend/internal/domain/collaboration"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type updateTransactionState struct {
	fieldDevice *domainFacility.FieldDevice
	bacnet      []domainFacility.BacnetObject
	history     []string
}

type updateHistoryBatchKey struct{}

func (s updateTransactionState) clone() updateTransactionState {
	clonedBacnet := make([]domainFacility.BacnetObject, len(s.bacnet))
	for i := range s.bacnet {
		clonedBacnet[i] = *cloneBacnetObject(&s.bacnet[i])
	}
	return updateTransactionState{
		fieldDevice: cloneFieldDevice(s.fieldDevice),
		bacnet:      clonedBacnet,
		history:     append([]string(nil), s.history...),
	}
}

type updateTransactionUnit struct {
	state *updateTransactionState
}

type updateTransactionHarness struct {
	committed             updateTransactionState
	updateErr             error
	commitErr             error
	runnerCalls           int
	listCalls             int
	updatedAt             time.Time
	historyBatchID        *uuid.UUID
	objectDataReplacement []domainFacility.BacnetObject
	receivedObjectDataID  *uuid.UUID
	outbox                domainCollaboration.OutboxStore
}

func (h *updateTransactionHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	if batchID, ok := ctx.Value(updateHistoryBatchKey{}).(uuid.UUID); ok {
		h.historyBatchID = &batchID
	}
	staged := h.committed.clone()
	ctx = domainCollaboration.WithOutboxStore(ctx, h.outbox)
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
) (*domainFacility.FieldDevice, error) {
	if s.state.fieldDevice == nil || s.state.fieldDevice.ID != id {
		return nil, domain.ErrNotFound
	}
	return cloneFieldDevice(s.state.fieldDevice), nil
}

func (s *updateWorkflowStub) ListBacnetObjects(
	_ context.Context,
	fieldDeviceID uuid.UUID,
) ([]domainFacility.BacnetObject, error) {
	s.harness.listCalls++
	objects := make([]domainFacility.BacnetObject, 0, len(s.state.bacnet))
	for i := range s.state.bacnet {
		if s.state.bacnet[i].FieldDeviceID == nil ||
			*s.state.bacnet[i].FieldDeviceID != fieldDeviceID {
			continue
		}
		objects = append(objects, *cloneBacnetObject(&s.state.bacnet[i]))
	}
	return objects, nil
}

func (s *updateWorkflowStub) UpdateWithBacnetObjects(
	ctx context.Context,
	fieldDevice *domainFacility.FieldDevice,
	objectDataID *uuid.UUID,
	bacnetObjects *[]domainFacility.BacnetObject,
) error {
	if objectDataID != nil && bacnetObjects != nil {
		return domain.ErrInvalidArgument
	}
	if batchID, ok := ctx.Value(updateHistoryBatchKey{}).(uuid.UUID); ok {
		s.harness.historyBatchID = &batchID
	}

	updated := cloneFieldDevice(fieldDevice)
	updated.UpdatedAt = s.harness.updatedAt
	s.state.fieldDevice = updated
	s.state.history = append(s.state.history, "field_device:update")

	replacement := bacnetObjects
	if objectDataID != nil {
		s.harness.receivedObjectDataID = clonePointer(objectDataID)
		objects := make([]domainFacility.BacnetObject, len(s.harness.objectDataReplacement))
		for i := range s.harness.objectDataReplacement {
			objects[i] = *cloneBacnetObject(&s.harness.objectDataReplacement[i])
		}
		replacement = &objects
	}
	if replacement != nil {
		for range s.state.bacnet {
			s.state.history = append(s.state.history, "bacnet_object:delete")
		}
		s.state.bacnet = make([]domainFacility.BacnetObject, len(*replacement))
		for i := range *replacement {
			object := cloneBacnetObject(&(*replacement)[i])
			parentID := fieldDevice.ID
			object.FieldDeviceID = &parentID
			object.UpdatedAt = s.harness.updatedAt
			s.state.bacnet[i] = *object
			s.state.history = append(s.state.history, "bacnet_object:create")
		}
	}

	// This simulates a history or later child-write failure after both the
	// facility mutation and history rows have been staged.
	return s.harness.updateErr
}

type updateProjectLinkReaderStub struct {
	links        []*domainProject.ProjectFieldDevice
	err          error
	calls        int
	received     []uuid.UUID
	assertCommit func() error
}

func (s *updateProjectLinkReaderStub) GetByFieldDeviceIDs(
	_ context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectFieldDevice, error) {
	s.calls++
	s.received = append([]uuid.UUID(nil), ids...)
	if s.assertCommit != nil {
		if err := s.assertCommit(); err != nil {
			return nil, err
		}
	}
	return s.links, s.err
}

type updateCommandDispatcherStub struct {
	commands []appcollaboration.Command
	err      error
}

type transactionalUpdateOutboxStub struct {
	*updateWorkflowStub
	links []*domainProject.ProjectFieldDevice
}

func (s *transactionalUpdateOutboxStub) GetByFieldDeviceIDs(
	_ context.Context,
	_ []uuid.UUID,
) ([]*domainProject.ProjectFieldDevice, error) {
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

func TestUpdateWritesVersionTwoOutboxCommandsInsideTheTransaction(t *testing.T) {
	fieldDeviceID := testUUID(901)
	projectID := testUUID(902)
	operationID := testUUID(903)
	outboxEventID := testUUID(904)
	v1EventID := testUUID(905)
	updatedAt := time.Date(2026, time.July, 23, 14, 0, 0, 0, time.UTC)
	harness := &updateTransactionHarness{committed: updateTransactionState{fieldDevice: &domainFacility.FieldDevice{
		Base: domain.Base{ID: fieldDeviceID}, SPSControllerSystemTypeID: testUUID(906), SystemPartID: testUUID(907), ApparatID: testUUID(908),
	}}, updatedAt: updatedAt}
	outboxStore := &updateOutboxStoreStub{}
	harness.outbox = outboxStore
	var outbox *transactionalUpdateOutboxStub
	factory := func(unit apptransaction.UnitOfWork) (UpdateWorkflow, error) {
		typed := unit.(updateTransactionUnit)
		outbox = &transactionalUpdateOutboxStub{
			updateWorkflowStub: &updateWorkflowStub{harness: harness, state: typed.state},
			links:              []*domainProject.ProjectFieldDevice{{ProjectID: projectID, FieldDeviceID: fieldDeviceID}},
		}
		return outbox, nil
	}
	dispatcher := &updateCommandDispatcherStub{}
	ids := []uuid.UUID{operationID, outboxEventID, v1EventID}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner: harness.runner, TransactionWorkflow: factory, Dispatcher: dispatcher,
		NewID: func() uuid.UUID { id := ids[0]; ids = ids[1:]; return id }, Now: func() time.Time { return updatedAt },
	})
	if _, err := handler.Update(context.Background(), UpdateCommand{FieldDeviceID: fieldDeviceID}); err != nil {
		t.Fatalf("update: %v", err)
	}
	if outbox == nil || len(outboxStore.events) != 1 {
		t.Fatalf("expected one transaction-scoped outbox command, got %#v", outboxStore.events)
	}
	decoded, err := appcollaboration.DecodeCommand(appcollaboration.EncodedCommand{
		Type: outboxStore.events[0].EventType, Payload: outboxStore.events[0].Payload,
	})
	if err != nil {
		t.Fatalf("decode queued command: %v", err)
	}
	queued, ok := decoded.(appcollaboration.FieldDeviceUpdated)
	if !ok || queued.SchemaVersion != appcollaboration.SchemaVersionV2 || queued.EventID != outboxEventID || queued.ProjectID != projectID {
		t.Fatalf("unexpected outbox command: %#v", decoded)
	}
	if len(dispatcher.commands) != 1 {
		t.Fatalf("expected one v1 compatibility dispatch, got %#v", dispatcher.commands)
	}
	compatibility, ok := dispatcher.commands[0].(appcollaboration.FieldDeviceUpdated)
	if !ok || compatibility.SchemaVersion != appcollaboration.SchemaVersionV1 || compatibility.EventID != v1EventID || compatibility.ProjectID != projectID {
		t.Fatalf("unexpected v1 command: %#v", dispatcher.commands[0])
	}
}

func (s *updateCommandDispatcherStub) Dispatch(
	_ context.Context,
	command appcollaboration.Command,
) error {
	s.commands = append(s.commands, command)
	return s.err
}

func TestUpdateCommitsFacilityAndHistoryBeforeProjectScopedDispatch(t *testing.T) {
	fieldDeviceID := testUUID(101)
	oldBacnetID := testUUID(102)
	newBacnetID := testUUID(103)
	oldSystemTypeID := testUUID(104)
	newSystemTypeID := testUUID(105)
	systemPartID := testUUID(106)
	apparatID := testUUID(107)
	projectOne := testUUID(111)
	projectTwo := testUUID(112)
	actorID := testUUID(121)
	operationID := testUUID(131)
	eventOne := testUUID(132)
	eventTwo := testUUID(133)
	createdAt := time.Date(2026, time.July, 1, 10, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, time.July, 20, 10, 0, 0, 0, time.UTC)
	occurredAt := updatedAt.Add(time.Second)
	oldBMK := "OLD"
	newBMK := "NEW"

	harness := &updateTransactionHarness{
		committed: updateTransactionState{
			fieldDevice: &domainFacility.FieldDevice{
				Base: domain.Base{
					ID:        fieldDeviceID,
					CreatedAt: createdAt,
					UpdatedAt: createdAt,
				},
				BMK:                       &oldBMK,
				ApparatNr:                 11,
				SPSControllerSystemTypeID: oldSystemTypeID,
				SystemPartID:              systemPartID,
				ApparatID:                 apparatID,
			},
			bacnet: []domainFacility.BacnetObject{{
				Base: domain.Base{
					ID:        oldBacnetID,
					CreatedAt: createdAt,
					UpdatedAt: createdAt,
				},
				TextFix:       "OLD-OBJECT",
				SoftwareType:  domainFacility.BacnetSoftwareTypeAI,
				FieldDeviceID: &fieldDeviceID,
			}},
		},
		updatedAt: updatedAt,
	}
	links := &updateProjectLinkReaderStub{
		links: []*domainProject.ProjectFieldDevice{
			{ProjectID: projectTwo, FieldDeviceID: fieldDeviceID},
			{ProjectID: projectOne, FieldDeviceID: fieldDeviceID},
			{ProjectID: projectOne, FieldDeviceID: fieldDeviceID},
			{ProjectID: testUUID(113), FieldDeviceID: testUUID(999)},
		},
		assertCommit: func() error {
			if harness.committed.fieldDevice.BMK == nil ||
				*harness.committed.fieldDevice.BMK != newBMK {
				return errors.New("project scope resolved before facility commit")
			}
			if want := []string{
				"field_device:update",
				"bacnet_object:delete",
				"bacnet_object:create",
			}; !reflect.DeepEqual(harness.committed.history, want) {
				return errors.New("project scope resolved before history commit")
			}
			return nil
		},
	}
	dispatcher := &updateCommandDispatcherStub{}
	generatedIDs := []uuid.UUID{operationID, eventOne, eventTwo}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, updateHistoryBatchKey{}, batchID)
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
	replacement := []domainFacility.BacnetObject{{
		Base:         domain.Base{ID: newBacnetID, CreatedAt: updatedAt},
		TextFix:      "NEW-OBJECT",
		SoftwareType: domainFacility.BacnetSoftwareTypeAO,
	}}

	outcome, err := handler.Execute(context.Background(), UpdateCommand{
		FieldDeviceID:             fieldDeviceID,
		BMK:                       &newBMK,
		SPSControllerSystemTypeID: &newSystemTypeID,
		BacnetObjects:             &replacement,
	})

	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if harness.runnerCalls != 1 {
		t.Fatalf("expected one transaction, got %d", harness.runnerCalls)
	}
	if harness.historyBatchID == nil || *harness.historyBatchID != operationID {
		t.Fatalf(
			"expected transactional history batch %s, got %v",
			operationID,
			harness.historyBatchID,
		)
	}
	if harness.listCalls != 2 {
		t.Fatalf("expected before/after BACnet reads in the transaction, got %d", harness.listCalls)
	}
	if links.calls != 1 {
		t.Fatalf("expected one batched project query after commit, got %d", links.calls)
	}
	if want := []uuid.UUID{fieldDeviceID}; !reflect.DeepEqual(links.received, want) {
		t.Fatalf("project lookup IDs: got %v, want %v", links.received, want)
	}
	if outcome.FieldDevice.BMK == nil || *outcome.FieldDevice.BMK != newBMK {
		t.Fatalf("expected committed BMK %q, got %v", newBMK, outcome.FieldDevice.BMK)
	}
	if outcome.FieldDevice.SPSControllerSystemTypeID != newSystemTypeID {
		t.Fatalf(
			"expected committed parent %s, got %s",
			newSystemTypeID,
			outcome.FieldDevice.SPSControllerSystemTypeID,
		)
	}

	if outcome.Mutation.OperationID != operationID {
		t.Fatalf("expected operation %s, got %s", operationID, outcome.Mutation.OperationID)
	}
	if outcome.Mutation.BatchID == nil || *outcome.Mutation.BatchID != operationID {
		t.Fatalf("expected replacement batch %s, got %v", operationID, outcome.Mutation.BatchID)
	}
	if want := []uuid.UUID{projectOne, projectTwo}; !reflect.DeepEqual(outcome.Mutation.ProjectIDs, want) {
		t.Fatalf("mutation projects: got %v, want %v", outcome.Mutation.ProjectIDs, want)
	}
	if len(outcome.Mutation.Changes) != 3 {
		t.Fatalf("expected parent update plus delete/create changes, got %d", len(outcome.Mutation.Changes))
	}
	assertFieldDeviceMutationChange(
		t,
		outcome.Mutation.Changes[0],
		fieldDeviceID,
		oldSystemTypeID,
		newSystemTypeID,
		oldBMK,
		newBMK,
	)
	assertBacnetMutationChange(
		t,
		outcome.Mutation.Changes[1],
		oldBacnetID,
		fieldDeviceID,
		domainHistory.ActionDelete,
	)
	assertBacnetMutationChange(
		t,
		outcome.Mutation.Changes[2],
		newBacnetID,
		fieldDeviceID,
		domainHistory.ActionCreate,
	)

	if len(dispatcher.commands) != 2 {
		t.Fatalf("expected one command per linked project, got %d", len(dispatcher.commands))
	}
	for index, command := range dispatcher.commands {
		moved, ok := command.(appcollaboration.FieldDeviceMoved)
		if !ok {
			t.Fatalf("command %d has type %T, want FieldDeviceMoved", index, command)
		}
		wantProjectID := []uuid.UUID{projectOne, projectTwo}[index]
		if moved.ProjectID != wantProjectID {
			t.Fatalf("command %d project: got %s, want %s", index, moved.ProjectID, wantProjectID)
		}
		if moved.FieldDeviceID != fieldDeviceID {
			t.Fatalf("command %d field device: got %s", index, moved.FieldDeviceID)
		}
		if moved.FromSPSControllerSystemTypeID != oldSystemTypeID ||
			moved.ToSPSControllerSystemTypeID != newSystemTypeID {
			t.Fatalf("command %d move: %+v", index, moved)
		}
		if moved.OperationID != operationID || moved.CorrelationID != operationID {
			t.Fatalf("command %d correlation: %+v", index, moved.Envelope)
		}
		if moved.ActorID == nil || *moved.ActorID != actorID {
			t.Fatalf("command %d actor: got %v, want %s", index, moved.ActorID, actorID)
		}
		if moved.OccurredAt != occurredAt {
			t.Fatalf("command %d occurred_at: got %s, want %s", index, moved.OccurredAt, occurredAt)
		}
	}
}

func TestUpdateParentMoveWithoutChildReplacementUsesMoveBatchAndCommand(t *testing.T) {
	fieldDeviceID := testUUID(151)
	oldSystemTypeID := testUUID(152)
	newSystemTypeID := testUUID(153)
	systemPartID := testUUID(154)
	apparatID := testUUID(155)
	projectID := testUUID(156)
	operationID := testUUID(157)
	eventID := testUUID(158)
	harness := &updateTransactionHarness{
		committed: updateTransactionState{
			fieldDevice: &domainFacility.FieldDevice{
				Base:                      domain.Base{ID: fieldDeviceID},
				SPSControllerSystemTypeID: oldSystemTypeID,
				SystemPartID:              systemPartID,
				ApparatID:                 apparatID,
				ApparatNr:                 7,
			},
		},
	}
	dispatcher := &updateCommandDispatcherStub{}
	generatedIDs := []uuid.UUID{operationID, eventID}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, updateHistoryBatchKey{}, batchID)
		},
		ProjectLinks: &updateProjectLinkReaderStub{links: []*domainProject.ProjectFieldDevice{{
			ProjectID:     projectID,
			FieldDeviceID: fieldDeviceID,
		}}},
		Dispatcher: dispatcher,
		NewID: func() uuid.UUID {
			id := generatedIDs[0]
			generatedIDs = generatedIDs[1:]
			return id
		},
	})

	outcome, err := handler.Execute(context.Background(), UpdateCommand{
		FieldDeviceID:             fieldDeviceID,
		SPSControllerSystemTypeID: &newSystemTypeID,
	})
	if err != nil {
		t.Fatalf("move FieldDevice: %v", err)
	}
	if harness.listCalls != 0 {
		t.Fatalf("parent move must not read BACnet children, got %d reads", harness.listCalls)
	}
	if harness.historyBatchID == nil || *harness.historyBatchID != operationID ||
		outcome.Mutation.BatchID == nil || *outcome.Mutation.BatchID != operationID {
		t.Fatalf(
			"move batch: history=%v result=%v, want %s",
			harness.historyBatchID,
			outcome.Mutation.BatchID,
			operationID,
		)
	}
	if got := outcome.FieldDevice.SPSControllerSystemTypeID; got != newSystemTypeID {
		t.Fatalf("parent: got %s, want %s", got, newSystemTypeID)
	}
	if len(outcome.Mutation.Changes) != 1 {
		t.Fatalf("changes: got %d, want 1", len(outcome.Mutation.Changes))
	}
	if want := []mutation.FieldName{mutation.FieldNameSystemType}; !reflect.DeepEqual(
		outcome.Mutation.Changes[0].ChangedFields,
		want,
	) {
		t.Fatalf(
			"changed fields: got %v, want %v",
			outcome.Mutation.Changes[0].ChangedFields,
			want,
		)
	}
	if len(dispatcher.commands) != 1 {
		t.Fatalf("commands: got %d, want 1", len(dispatcher.commands))
	}
	moved, ok := dispatcher.commands[0].(appcollaboration.FieldDeviceMoved)
	if !ok {
		t.Fatalf("command has type %T, want FieldDeviceMoved", dispatcher.commands[0])
	}
	if moved.FromSPSControllerSystemTypeID != oldSystemTypeID ||
		moved.ToSPSControllerSystemTypeID != newSystemTypeID {
		t.Fatalf("unexpected move command: %+v", moved)
	}
}

func TestUpdateRollsBackFacilityAndHistoryAndSkipsDispatchOnWorkflowFailure(t *testing.T) {
	fieldDeviceID := testUUID(201)
	originalBMK := "ORIGINAL"
	replacementBMK := "REPLACEMENT"
	updateErr := errors.New("history write failed")
	harness := &updateTransactionHarness{
		committed: updateTransactionState{
			fieldDevice: &domainFacility.FieldDevice{
				Base: domain.Base{ID: fieldDeviceID},
				BMK:  &originalBMK,
			},
		},
		updateErr: updateErr,
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
		FieldDeviceID: fieldDeviceID,
		BMK:           &replacementBMK,
	})

	if !errors.Is(err, updateErr) {
		t.Fatalf("expected workflow error, got %v", err)
	}
	if harness.committed.fieldDevice.BMK == nil ||
		*harness.committed.fieldDevice.BMK != originalBMK {
		t.Fatalf("failed transaction changed committed facility state: %+v", harness.committed.fieldDevice)
	}
	if len(harness.committed.history) != 0 {
		t.Fatalf("failed transaction left successful history: %v", harness.committed.history)
	}
	if links.calls != 0 || len(dispatcher.commands) != 0 {
		t.Fatalf(
			"failed transaction must not resolve or dispatch, got links=%d commands=%d",
			links.calls,
			len(dispatcher.commands),
		)
	}
}

func TestUpdateSkipsDispatchWhenCommitFailsAfterSuccessfulWorkflow(t *testing.T) {
	fieldDeviceID := testUUID(301)
	originalBMK := "ORIGINAL"
	replacementBMK := "REPLACEMENT"
	commitErr := errors.New("commit failed")
	harness := &updateTransactionHarness{
		committed: updateTransactionState{
			fieldDevice: &domainFacility.FieldDevice{
				Base: domain.Base{ID: fieldDeviceID},
				BMK:  &originalBMK,
			},
		},
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
		FieldDeviceID: fieldDeviceID,
		BMK:           &replacementBMK,
	})

	if !errors.Is(err, commitErr) {
		t.Fatalf("expected commit error, got %v", err)
	}
	if harness.committed.fieldDevice.BMK == nil ||
		*harness.committed.fieldDevice.BMK != originalBMK {
		t.Fatalf("failed commit changed facility state: %+v", harness.committed.fieldDevice)
	}
	if len(harness.committed.history) != 0 {
		t.Fatalf("failed commit left successful history: %v", harness.committed.history)
	}
	if links.calls != 0 || len(dispatcher.commands) != 0 {
		t.Fatalf(
			"failed commit must not resolve or dispatch, got links=%d commands=%d",
			links.calls,
			len(dispatcher.commands),
		)
	}
}

func TestUpdateWithoutBacnetSelectionDoesNotReadOrReplaceChildren(t *testing.T) {
	fieldDeviceID := testUUID(401)
	bacnetObjectID := testUUID(402)
	newDescription := "changed"
	harness := &updateTransactionHarness{
		committed: updateTransactionState{
			fieldDevice: &domainFacility.FieldDevice{
				Base: domain.Base{ID: fieldDeviceID},
			},
			bacnet: []domainFacility.BacnetObject{{
				Base:          domain.Base{ID: bacnetObjectID},
				TextFix:       "PRESERVED",
				FieldDeviceID: &fieldDeviceID,
			}},
		},
	}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
	})

	outcome, err := handler.Execute(context.Background(), UpdateCommand{
		FieldDeviceID: fieldDeviceID,
		Description:   &newDescription,
	})

	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if harness.listCalls != 0 {
		t.Fatalf("expected no BACnet reads without replacement, got %d", harness.listCalls)
	}
	if len(harness.committed.bacnet) != 1 ||
		harness.committed.bacnet[0].ID != bacnetObjectID {
		t.Fatalf("omitted BACnet selection changed children: %+v", harness.committed.bacnet)
	}
	if outcome.Mutation.BatchID != nil {
		t.Fatalf("base-only update must not expose a child batch, got %s", *outcome.Mutation.BatchID)
	}
	if len(outcome.Mutation.Changes) != 1 {
		t.Fatalf("expected one base change, got %d", len(outcome.Mutation.Changes))
	}
	if want := []mutation.FieldName{mutation.FieldNameDescription}; !reflect.DeepEqual(
		outcome.Mutation.Changes[0].ChangedFields,
		want,
	) {
		t.Fatalf(
			"changed fields: got %v, want %v",
			outcome.Mutation.Changes[0].ChangedFields,
			want,
		)
	}
}

func TestUpdateFromObjectDataUsesTransactionalChildReplacement(t *testing.T) {
	fieldDeviceID := testUUID(451)
	oldBacnetID := testUUID(452)
	newBacnetID := testUUID(453)
	objectDataID := testUUID(454)
	operationID := testUUID(455)
	harness := &updateTransactionHarness{
		committed: updateTransactionState{
			fieldDevice: &domainFacility.FieldDevice{
				Base: domain.Base{ID: fieldDeviceID},
			},
			bacnet: []domainFacility.BacnetObject{{
				Base:          domain.Base{ID: oldBacnetID},
				TextFix:       "OLD",
				FieldDeviceID: &fieldDeviceID,
			}},
		},
		objectDataReplacement: []domainFacility.BacnetObject{{
			Base:    domain.Base{ID: newBacnetID},
			TextFix: "FROM-TEMPLATE",
		}},
	}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, updateHistoryBatchKey{}, batchID)
		},
		NewID: func() uuid.UUID { return operationID },
	})

	outcome, err := handler.Execute(context.Background(), UpdateCommand{
		FieldDeviceID: fieldDeviceID,
		ObjectDataID:  &objectDataID,
	})

	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if harness.receivedObjectDataID == nil ||
		*harness.receivedObjectDataID != objectDataID {
		t.Fatalf(
			"expected ObjectData selection %s, got %v",
			objectDataID,
			harness.receivedObjectDataID,
		)
	}
	if len(harness.committed.bacnet) != 1 ||
		harness.committed.bacnet[0].ID != newBacnetID ||
		harness.committed.bacnet[0].FieldDeviceID == nil ||
		*harness.committed.bacnet[0].FieldDeviceID != fieldDeviceID {
		t.Fatalf("unexpected committed template replacement: %+v", harness.committed.bacnet)
	}
	if outcome.Mutation.BatchID == nil ||
		*outcome.Mutation.BatchID != operationID ||
		harness.historyBatchID == nil ||
		*harness.historyBatchID != operationID {
		t.Fatalf(
			"expected shared operation/history batch %s, got result=%v history=%v",
			operationID,
			outcome.Mutation.BatchID,
			harness.historyBatchID,
		)
	}
	if len(outcome.Mutation.Changes) != 3 ||
		outcome.Mutation.Changes[1].EntityID != oldBacnetID ||
		outcome.Mutation.Changes[1].Action != domainHistory.ActionDelete ||
		outcome.Mutation.Changes[2].EntityID != newBacnetID ||
		outcome.Mutation.Changes[2].Action != domainHistory.ActionCreate {
		t.Fatalf("unexpected template replacement changes: %+v", outcome.Mutation.Changes)
	}
}

func TestUpdateReportsDispatchFailureWithoutReplacingCommittedResult(t *testing.T) {
	fieldDeviceID := testUUID(501)
	projectID := testUUID(511)
	bmk := "COMMITTED"
	dispatchErr := errors.New("realtime unavailable")
	reported := make([]error, 0, 1)
	dispatcher := &updateCommandDispatcherStub{err: dispatchErr}
	harness := &updateTransactionHarness{
		committed: updateTransactionState{
			fieldDevice: &domainFacility.FieldDevice{
				Base: domain.Base{ID: fieldDeviceID},
			},
		},
	}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		ProjectLinks: &updateProjectLinkReaderStub{
			links: []*domainProject.ProjectFieldDevice{{
				ProjectID:     projectID,
				FieldDeviceID: fieldDeviceID,
			}},
		},
		Dispatcher: dispatcher,
		ReportError: func(err error) {
			reported = append(reported, err)
		},
	})

	updated, err := handler.Update(context.Background(), UpdateCommand{
		FieldDeviceID: fieldDeviceID,
		BMK:           &bmk,
	})

	if err != nil {
		t.Fatalf("committed update must not fail because dispatch failed: %v", err)
	}
	if updated.BMK == nil || *updated.BMK != bmk {
		t.Fatalf("expected committed result, got %+v", updated)
	}
	if len(reported) != 1 || !errors.Is(reported[0], dispatchErr) {
		t.Fatalf("expected one reported dispatch error, got %v", reported)
	}
	if len(dispatcher.commands) != 1 {
		t.Fatalf("commands: got %d, want 1", len(dispatcher.commands))
	}
	if _, ok := dispatcher.commands[0].(appcollaboration.FieldDeviceUpdated); !ok {
		t.Fatalf("command has type %T, want FieldDeviceUpdated", dispatcher.commands[0])
	}
}

func TestUpdateRejectsMutuallyExclusiveBacnetSelectionsBeforeTransaction(t *testing.T) {
	fieldDeviceID := testUUID(601)
	objectDataID := testUUID(602)
	objects := []domainFacility.BacnetObject{}
	harness := &updateTransactionHarness{}
	handler := NewUpdateHandler(UpdateDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
	})

	_, err := handler.Execute(context.Background(), UpdateCommand{
		FieldDeviceID: fieldDeviceID,
		ObjectDataID:  &objectDataID,
		BacnetObjects: &objects,
	})

	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("expected invalid argument, got %v", err)
	}
	if harness.runnerCalls != 0 {
		t.Fatalf("invalid command must not open a transaction, got %d", harness.runnerCalls)
	}
}

func assertFieldDeviceMutationChange(
	t *testing.T,
	change mutation.EntityChange,
	fieldDeviceID uuid.UUID,
	oldParentID uuid.UUID,
	newParentID uuid.UUID,
	oldBMK string,
	newBMK string,
) {
	t.Helper()

	if change.EntityType != mutation.EntityTypeFieldDevice ||
		change.EntityID != fieldDeviceID ||
		change.Action != domainHistory.ActionUpdate {
		t.Fatalf("unexpected FieldDevice change: %+v", change)
	}
	if change.ParentID == nil || *change.ParentID != newParentID {
		t.Fatalf("expected new parent %s, got %v", newParentID, change.ParentID)
	}
	if want := []mutation.FieldName{
		mutation.FieldNameBMK,
		mutation.FieldNameSystemType,
		mutation.FieldNameBacnetObjects,
	}; !reflect.DeepEqual(change.ChangedFields, want) {
		t.Fatalf("changed fields: got %v, want %v", change.ChangedFields, want)
	}

	var before fieldDeviceSnapshot
	if err := json.Unmarshal(change.Before, &before); err != nil {
		t.Fatalf("decode before snapshot: %v", err)
	}
	var after fieldDeviceSnapshot
	if err := json.Unmarshal(change.After, &after); err != nil {
		t.Fatalf("decode after snapshot: %v", err)
	}
	if before.BMK == nil || *before.BMK != oldBMK ||
		before.SPSControllerSystemTypeID != oldParentID {
		t.Fatalf("unexpected before snapshot: %+v", before)
	}
	if after.BMK == nil || *after.BMK != newBMK ||
		after.SPSControllerSystemTypeID != newParentID {
		t.Fatalf("unexpected after snapshot: %+v", after)
	}
}

func assertBacnetMutationChange(
	t *testing.T,
	change mutation.EntityChange,
	entityID uuid.UUID,
	parentID uuid.UUID,
	action domainHistory.Action,
) {
	t.Helper()

	if change.EntityType != mutation.EntityTypeBacnetObject ||
		change.EntityID != entityID ||
		change.Action != action {
		t.Fatalf("unexpected BACnet change: %+v", change)
	}
	if change.ParentID == nil || *change.ParentID != parentID {
		t.Fatalf("expected BACnet parent %s, got %v", parentID, change.ParentID)
	}
	if action == domainHistory.ActionDelete &&
		(len(change.Before) == 0 || len(change.After) != 0) {
		t.Fatalf("delete must have before only, got before=%s after=%s", change.Before, change.After)
	}
	if action == domainHistory.ActionCreate &&
		(len(change.After) == 0 || len(change.Before) != 0) {
		t.Fatalf("create must have after only, got before=%s after=%s", change.Before, change.After)
	}
}
