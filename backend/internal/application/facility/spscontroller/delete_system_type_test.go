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
	domainCollaboration "github.com/besart951/go_infra_link/backend/internal/domain/collaboration"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

type globalSystemTypeDeleteBatchKey struct{}

type globalSystemTypeDeleteState struct {
	systemType   *domainFacility.SPSControllerSystemType
	fieldDevices []uuid.UUID
	history      []string
}

func (s globalSystemTypeDeleteState) clone() globalSystemTypeDeleteState {
	return globalSystemTypeDeleteState{
		systemType:   cloneSPSControllerSystemType(s.systemType),
		fieldDevices: append([]uuid.UUID(nil), s.fieldDevices...),
		history:      append([]string(nil), s.history...),
	}
}

type globalSystemTypeDeleteUnit struct {
	state *globalSystemTypeDeleteState
}

type globalSystemTypeDeleteHarness struct {
	committed           globalSystemTypeDeleteState
	getErr              error
	deleteErr           error
	deleteErrAfterWrite bool
	commitErr           error
	runnerCalls         int
	getCalls            int
	deleteCalls         int
	deletedID           uuid.UUID
	historyBatchID      *uuid.UUID
	outbox              domainCollaboration.OutboxStore
}

func (h *globalSystemTypeDeleteHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	staged := h.committed.clone()
	txCtx := ctx
	if h.outbox != nil {
		txCtx = domainCollaboration.WithOutboxStore(txCtx, h.outbox)
	}
	if err := run(txCtx, globalSystemTypeDeleteUnit{state: &staged}); err != nil {
		return err
	}
	if h.commitErr != nil {
		return h.commitErr
	}
	h.committed = staged
	return nil
}

type scopedGlobalSystemTypeDeleteWorkflowStub struct {
	*globalSystemTypeDeleteWorkflowStub
	ownerID    uuid.UUID
	projectIDs []uuid.UUID
	scopeErr   error
}

func (s *scopedGlobalSystemTypeDeleteWorkflowStub) GetDeleteProjectScope(
	_ context.Context,
	_ uuid.UUID,
) (uuid.UUID, []uuid.UUID, error) {
	return s.ownerID, append([]uuid.UUID(nil), s.projectIDs...), s.scopeErr
}

func (h *globalSystemTypeDeleteHarness) factory(
	unit apptransaction.UnitOfWork,
) (DeleteSystemTypeWorkflow, error) {
	typed, ok := unit.(globalSystemTypeDeleteUnit)
	if !ok || typed.state == nil {
		return nil, errors.New("unexpected global SPS system-type delete transaction unit")
	}
	return &globalSystemTypeDeleteWorkflowStub{harness: h, state: typed.state}, nil
}

type globalSystemTypeDeleteWorkflowStub struct {
	harness *globalSystemTypeDeleteHarness
	state   *globalSystemTypeDeleteState
}

func (s *globalSystemTypeDeleteWorkflowStub) GetByID(
	_ context.Context,
	id uuid.UUID,
) (*domainFacility.SPSControllerSystemType, error) {
	s.harness.getCalls++
	if s.harness.getErr != nil {
		return nil, s.harness.getErr
	}
	if s.state.systemType == nil || s.state.systemType.ID != id {
		return nil, domain.ErrNotFound
	}
	return cloneSPSControllerSystemType(s.state.systemType), nil
}

func (s *globalSystemTypeDeleteWorkflowStub) DeleteByID(
	ctx context.Context,
	id uuid.UUID,
) error {
	s.harness.deleteCalls++
	s.harness.deletedID = id
	if batchID, ok := ctx.Value(globalSystemTypeDeleteBatchKey{}).(uuid.UUID); ok {
		batch := batchID
		s.harness.historyBatchID = &batch
	}
	if s.harness.deleteErr != nil && !s.harness.deleteErrAfterWrite {
		return s.harness.deleteErr
	}
	if s.state.systemType != nil && s.state.systemType.ID == id {
		s.state.systemType = nil
		// Mirrors the existing database-owned field-device cascade. The
		// application deliberately does not enumerate these descendants.
		s.state.fieldDevices = nil
		s.state.history = append(s.state.history, "sps_controller_system_type:delete")
	}
	if s.harness.deleteErr != nil {
		return s.harness.deleteErr
	}
	return nil
}

func TestDeleteSystemTypeCommitsRootHistoryAndCanonicalResultTogether(t *testing.T) {
	systemTypeID := spsTestUUID(451)
	spsControllerID := spsTestUUID(452)
	definitionID := spsTestUUID(453)
	actorID := spsTestUUID(454)
	operationID := spsTestUUID(455)
	createdAt := time.Date(2026, time.July, 20, 20, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	occurredAt := updatedAt.Add(time.Second)
	number := 7
	documentName := "HLK-07.pdf"
	harness := &globalSystemTypeDeleteHarness{
		committed: globalSystemTypeDeleteState{
			systemType: &domainFacility.SPSControllerSystemType{
				Base: domain.Base{
					ID:        systemTypeID,
					CreatedAt: createdAt,
					UpdatedAt: updatedAt,
				},
				Number:          &number,
				DocumentName:    &documentName,
				SPSControllerID: spsControllerID,
				SystemTypeID:    definitionID,
			},
			fieldDevices: []uuid.UUID{spsTestUUID(456), spsTestUUID(457)},
		},
	}
	handler := NewDeleteSystemTypeHandler(DeleteSystemTypeDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, globalSystemTypeDeleteBatchKey{}, batchID)
		},
		Actor: func(context.Context) *uuid.UUID { return &actorID },
		NewID: func() uuid.UUID { return operationID },
		Now:   func() time.Time { return occurredAt },
	})

	outcome, err := handler.Execute(context.Background(), DeleteSystemTypeCommand{
		SPSControllerSystemTypeID: systemTypeID,
	})
	if err != nil {
		t.Fatalf("execute global system-type delete: %v", err)
	}
	if harness.runnerCalls != 1 || harness.getCalls != 1 || harness.deleteCalls != 1 ||
		harness.deletedID != systemTypeID {
		t.Fatalf("workflow calls: runner=%d get=%d delete=%d id=%s",
			harness.runnerCalls,
			harness.getCalls,
			harness.deleteCalls,
			harness.deletedID,
		)
	}
	if harness.committed.systemType != nil || len(harness.committed.fieldDevices) != 0 ||
		!reflect.DeepEqual(harness.committed.history, []string{"sps_controller_system_type:delete"}) {
		t.Fatalf("committed state: %+v", harness.committed)
	}
	if harness.historyBatchID == nil || *harness.historyBatchID != operationID ||
		outcome.Mutation.BatchID == nil || *outcome.Mutation.BatchID != operationID {
		t.Fatalf("history batch: workflow=%v result=%v", harness.historyBatchID, outcome.Mutation.BatchID)
	}
	if !outcome.Existed || outcome.Mutation.OperationID != operationID ||
		outcome.Mutation.ActorID == nil || *outcome.Mutation.ActorID != actorID ||
		!outcome.Mutation.OccurredAt.Equal(occurredAt) || len(outcome.Mutation.ProjectIDs) != 0 {
		t.Fatalf("delete outcome: %+v", outcome)
	}
	if len(outcome.Mutation.Changes) != 1 {
		t.Fatalf("changes: got %d, want 1", len(outcome.Mutation.Changes))
	}
	change := outcome.Mutation.Changes[0]
	if change.EntityType != "sps_controller_system_type" || change.EntityID != systemTypeID ||
		change.ParentID == nil || *change.ParentID != spsControllerID ||
		change.Action != domainHistory.ActionDelete || len(change.After) != 0 {
		t.Fatalf("delete change: %+v", change)
	}
	var snapshot spsControllerSystemTypeSnapshot
	if err := json.Unmarshal(change.Before, &snapshot); err != nil {
		t.Fatalf("decode delete snapshot: %v", err)
	}
	if snapshot.ID != systemTypeID || snapshot.SPSControllerID != spsControllerID ||
		snapshot.SystemTypeID != definitionID || snapshot.Number == nil || *snapshot.Number != number ||
		snapshot.DocumentName == nil || *snapshot.DocumentName != documentName ||
		!snapshot.CreatedAt.Equal(createdAt) || !snapshot.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("delete snapshot: %+v", snapshot)
	}
}

func TestDeleteSystemTypePersistsProjectRefreshBeforeCommitAndDispatchesAfter(
	t *testing.T,
) {
	systemTypeID := spsTestUUID(481)
	controllerID := spsTestUUID(482)
	projectID := spsTestUUID(483)
	operationID := spsTestUUID(484)
	eventID := spsTestUUID(485)
	occurredAt := time.Date(2026, time.July, 23, 23, 0, 0, 0, time.UTC)
	harness := &globalSystemTypeDeleteHarness{
		committed: globalSystemTypeDeleteState{
			systemType: &domainFacility.SPSControllerSystemType{
				Base:            domain.Base{ID: systemTypeID},
				SPSControllerID: controllerID,
				SystemTypeID:    spsTestUUID(486),
			},
		},
	}
	outbox := &updateOutboxStoreStub{}
	harness.outbox = outbox
	factory := func(unit apptransaction.UnitOfWork) (DeleteSystemTypeWorkflow, error) {
		typed := unit.(globalSystemTypeDeleteUnit)
		return &scopedGlobalSystemTypeDeleteWorkflowStub{
			globalSystemTypeDeleteWorkflowStub: &globalSystemTypeDeleteWorkflowStub{
				harness: harness,
				state:   typed.state,
			},
			ownerID:    controllerID,
			projectIDs: []uuid.UUID{projectID},
		}, nil
	}
	dispatcher := &updateCommandDispatcherStub{}
	ids := []uuid.UUID{operationID, eventID}
	handler := NewDeleteSystemTypeHandler(DeleteSystemTypeDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: factory,
		Dispatcher:          dispatcher,
		NewID: func() uuid.UUID {
			id := ids[0]
			ids = ids[1:]
			return id
		},
		Now: func() time.Time { return occurredAt },
	})

	outcome, err := handler.Execute(context.Background(), DeleteSystemTypeCommand{
		SPSControllerSystemTypeID: systemTypeID,
	})
	if err != nil {
		t.Fatalf("delete system type: %v", err)
	}
	if !reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectID}) {
		t.Fatalf("project IDs: %v", outcome.Mutation.ProjectIDs)
	}
	if len(outbox.events) != 1 {
		t.Fatalf("outbox events: got %d, want 1", len(outbox.events))
	}
	decoded, err := appcollaboration.DecodeCommand(appcollaboration.EncodedCommand{
		Type:    outbox.events[0].EventType,
		Payload: outbox.events[0].Payload,
	})
	if err != nil {
		t.Fatalf("decode outbox event: %v", err)
	}
	event, ok := decoded.(appcollaboration.FacilityHierarchyRefreshRequired)
	if !ok ||
		event.SchemaVersion != appcollaboration.SchemaVersionV2 ||
		event.ProjectID != projectID ||
		event.Scope != appcollaboration.FacilityScopeSPSController ||
		!reflect.DeepEqual(event.EntityIDs, []uuid.UUID{controllerID}) {
		t.Fatalf("unexpected outbox event: %#v", decoded)
	}
	if len(dispatcher.commands) != 1 {
		t.Fatalf("compatibility commands: got %d, want 1", len(dispatcher.commands))
	}
}

func TestDeleteSystemTypeRollsBackDeleteHistoryAndDatabaseCascadeFailures(t *testing.T) {
	loadErr := errors.New("load failed")
	deleteErr := errors.New("delete failed")
	historyErr := errors.New("history failed")
	commitErr := errors.New("commit failed")
	for _, test := range []struct {
		name                string
		getErr              error
		deleteErr           error
		deleteErrAfterWrite bool
		commitErr           error
		want                error
	}{
		{name: "load", getErr: loadErr, want: loadErr},
		{name: "delete", deleteErr: deleteErr, want: deleteErr},
		{name: "history", deleteErr: historyErr, deleteErrAfterWrite: true, want: historyErr},
		{name: "commit", commitErr: commitErr, want: commitErr},
	} {
		t.Run(test.name, func(t *testing.T) {
			systemTypeID := spsTestUUID(461)
			fieldDeviceID := spsTestUUID(462)
			harness := &globalSystemTypeDeleteHarness{
				committed: globalSystemTypeDeleteState{
					systemType: &domainFacility.SPSControllerSystemType{
						Base:            domain.Base{ID: systemTypeID},
						SPSControllerID: spsTestUUID(463),
						SystemTypeID:    spsTestUUID(464),
					},
					fieldDevices: []uuid.UUID{fieldDeviceID},
				},
				getErr:              test.getErr,
				deleteErr:           test.deleteErr,
				deleteErrAfterWrite: test.deleteErrAfterWrite,
				commitErr:           test.commitErr,
			}
			handler := NewDeleteSystemTypeHandler(DeleteSystemTypeDependencies{
				TransactionRunner:   harness.runner,
				TransactionWorkflow: harness.factory,
			})

			_, err := handler.Execute(context.Background(), DeleteSystemTypeCommand{
				SPSControllerSystemTypeID: systemTypeID,
			})
			if !errors.Is(err, test.want) {
				t.Fatalf("error: got %v, want %v", err, test.want)
			}
			if harness.committed.systemType == nil || harness.committed.systemType.ID != systemTypeID ||
				!reflect.DeepEqual(harness.committed.fieldDevices, []uuid.UUID{fieldDeviceID}) ||
				len(harness.committed.history) != 0 {
				t.Fatalf("failed transaction escaped: %+v", harness.committed)
			}
		})
	}
}

func TestDeleteSystemTypeMissingRowPreservesIdempotentSuccess(t *testing.T) {
	operationID := spsTestUUID(471)
	harness := &globalSystemTypeDeleteHarness{}
	handler := NewDeleteSystemTypeHandler(DeleteSystemTypeDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, globalSystemTypeDeleteBatchKey{}, batchID)
		},
		NewID: func() uuid.UUID { return operationID },
	})

	outcome, err := handler.Execute(context.Background(), DeleteSystemTypeCommand{
		SPSControllerSystemTypeID: spsTestUUID(472),
	})
	if err != nil {
		t.Fatalf("missing delete: %v", err)
	}
	if outcome.Existed || len(outcome.Mutation.Changes) != 0 || harness.deleteCalls != 1 ||
		harness.historyBatchID == nil || *harness.historyBatchID != operationID ||
		len(harness.committed.history) != 0 {
		t.Fatalf("missing-row outcome: outcome=%+v harness=%+v", outcome, harness)
	}
}

func TestDeleteSystemTypeValidatesConfigurationAndID(t *testing.T) {
	unconfigured := NewDeleteSystemTypeHandler(DeleteSystemTypeDependencies{})
	_, err := unconfigured.Execute(context.Background(), DeleteSystemTypeCommand{
		SPSControllerSystemTypeID: uuid.New(),
	})
	if !errors.Is(err, ErrDeleteSystemTypeTransactionNotConfigured) {
		t.Fatalf("unconfigured error: %v", err)
	}

	harness := &globalSystemTypeDeleteHarness{}
	configured := NewDeleteSystemTypeHandler(DeleteSystemTypeDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
	})
	_, err = configured.Execute(context.Background(), DeleteSystemTypeCommand{})
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("invalid ID error: %v", err)
	}
	if harness.runnerCalls != 0 {
		t.Fatalf("invalid ID started %d transactions", harness.runnerCalls)
	}
}
