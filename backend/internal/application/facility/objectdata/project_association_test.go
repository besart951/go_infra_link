package objectdata

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
	"github.com/google/uuid"
)

type associationHarness struct {
	projectID       uuid.UUID
	base            *domainFacility.ObjectData
	staged          *domainFacility.ObjectData
	inTransaction   bool
	committed       bool
	requireCalls    int
	getCalls        int
	updateCalls     int
	batchIDOnUpdate uuid.UUID
	updateErr       error
	commitErr       error
}

type associationWorkflowStub struct {
	harness *associationHarness
}

func (s *associationWorkflowStub) RequireProject(
	_ context.Context,
	projectID uuid.UUID,
) error {
	s.harness.requireCalls++
	if !s.harness.inTransaction {
		return errors.New("project checked outside transaction")
	}
	if projectID != s.harness.projectID {
		return domain.ErrNotFound
	}
	return nil
}

func (s *associationWorkflowStub) GetObjectData(
	_ context.Context,
	objectDataID uuid.UUID,
) (*domainFacility.ObjectData, error) {
	s.harness.getCalls++
	if !s.harness.inTransaction {
		return nil, errors.New("ObjectData loaded outside transaction")
	}
	if s.harness.staged == nil || s.harness.staged.ID != objectDataID {
		return nil, domain.ErrNotFound
	}
	return cloneObjectData(s.harness.staged), nil
}

func (s *associationWorkflowStub) UpdateObjectData(
	ctx context.Context,
	objectData *domainFacility.ObjectData,
) error {
	s.harness.updateCalls++
	if !s.harness.inTransaction {
		return errors.New("ObjectData updated outside transaction")
	}
	if batchID, ok := ctx.Value(associationBatchKey{}).(uuid.UUID); ok {
		s.harness.batchIDOnUpdate = batchID
	}
	if s.harness.updateErr != nil {
		return s.harness.updateErr
	}
	s.harness.staged = cloneObjectData(objectData)
	return nil
}

func (h *associationHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.inTransaction = true
	h.committed = false
	h.staged = cloneObjectData(h.base)
	err := run(ctx, h)
	h.inTransaction = false
	if err != nil {
		h.staged = nil
		return err
	}
	if h.commitErr != nil {
		h.staged = nil
		return h.commitErr
	}
	h.base = cloneObjectData(h.staged)
	h.staged = nil
	h.committed = true
	return nil
}

type associationDispatcherSpy struct {
	harness  *associationHarness
	commands []appcollaboration.Command
	err      error
}

func (s *associationDispatcherSpy) Dispatch(
	_ context.Context,
	command appcollaboration.Command,
) error {
	if !s.harness.committed {
		return errors.New("collaboration dispatched before commit")
	}
	s.commands = append(s.commands, command)
	return s.err
}

type associationBatchKey struct{}

func associationHistoryBatch(ctx context.Context, batchID uuid.UUID) context.Context {
	return context.WithValue(ctx, associationBatchKey{}, batchID)
}

func associationTestUUID(last byte) uuid.UUID {
	var id uuid.UUID
	id[15] = last
	return id
}

func newAssociationHandler(
	harness *associationHarness,
	dispatcher appcollaboration.CommandDispatcher,
	actorID uuid.UUID,
	operationID uuid.UUID,
	eventID uuid.UUID,
	occurredAt time.Time,
) *ProjectAssociationHandler {
	// The production handler allocates distinct durable-v2 and v1 compatibility
	// event IDs. This helper does not install an outbox, so reusing the expected
	// compatibility ID keeps the existing dispatch assertions focused.
	ids := []uuid.UUID{operationID, eventID, eventID}
	nextID := 0
	return NewProjectAssociationHandler(ProjectAssociationDependencies{
		TransactionRunner: harness.runner,
		TransactionWorkflow: func(unit apptransaction.UnitOfWork) (ProjectAssociationWorkflow, error) {
			if unit != harness {
				return nil, errors.New("unexpected transaction unit")
			}
			return &associationWorkflowStub{harness: harness}, nil
		},
		HistoryBatch: associationHistoryBatch,
		Dispatcher:   dispatcher,
		Actor: func(context.Context) *uuid.UUID {
			return &actorID
		},
		NewID: func() uuid.UUID {
			id := ids[nextID]
			nextID++
			return id
		},
		Now: func() time.Time { return occurredAt },
	})
}

func TestAttachToProjectCommitsHistoryCorrelationBeforeRefresh(t *testing.T) {
	projectID := associationTestUUID(1)
	objectDataID := associationTestUUID(2)
	actorID := associationTestUUID(3)
	operationID := associationTestUUID(4)
	eventID := associationTestUUID(5)
	occurredAt := time.Date(2026, 7, 22, 8, 30, 0, 0, time.FixedZone("test", 3600))
	harness := &associationHarness{
		projectID: projectID,
		base: &domainFacility.ObjectData{
			Base:        domain.Base{ID: objectDataID},
			Description: "AHU",
			Version:     "1",
		},
	}
	dispatcher := &associationDispatcherSpy{harness: harness}
	handler := newAssociationHandler(
		harness,
		dispatcher,
		actorID,
		operationID,
		eventID,
		occurredAt,
	)

	outcome, err := handler.ExecuteAttach(context.Background(), AttachToProjectCommand{
		ProjectID:    projectID,
		ObjectDataID: objectDataID,
	})
	if err != nil {
		t.Fatalf("attach ObjectData: %v", err)
	}
	if !harness.committed || harness.base.ProjectID == nil ||
		*harness.base.ProjectID != projectID || !harness.base.IsActive {
		t.Fatalf("association did not commit: %+v", harness.base)
	}
	if harness.requireCalls != 1 || harness.getCalls != 2 || harness.updateCalls != 1 {
		t.Fatalf(
			"unexpected workflow calls: require=%d get=%d update=%d",
			harness.requireCalls,
			harness.getCalls,
			harness.updateCalls,
		)
	}
	if harness.batchIDOnUpdate != operationID {
		t.Fatalf("history batch: got %s, want %s", harness.batchIDOnUpdate, operationID)
	}
	if outcome.Mutation.OperationID != operationID || outcome.Mutation.BatchID == nil ||
		*outcome.Mutation.BatchID != operationID || outcome.Mutation.ActorID == nil ||
		*outcome.Mutation.ActorID != actorID ||
		!outcome.Mutation.OccurredAt.Equal(occurredAt.UTC()) ||
		!reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectID}) {
		t.Fatalf("unexpected mutation envelope: %+v", outcome.Mutation)
	}
	if len(outcome.Mutation.Changes) != 1 {
		t.Fatalf("changes: got %d, want 1", len(outcome.Mutation.Changes))
	}
	change := outcome.Mutation.Changes[0]
	if change.EntityType != "object_data" || change.EntityID != objectDataID ||
		change.ParentID == nil || *change.ParentID != projectID ||
		!reflect.DeepEqual(change.ChangedFields, []mutation.FieldName{
			mutation.FieldNameProject,
			mutation.FieldNameIsActive,
		}) {
		t.Fatalf("unexpected association change: %+v", change)
	}
	var before, after objectDataSnapshot
	if err := json.Unmarshal(change.Before, &before); err != nil {
		t.Fatalf("decode before: %v", err)
	}
	if err := json.Unmarshal(change.After, &after); err != nil {
		t.Fatalf("decode after: %v", err)
	}
	if before.ProjectID != nil || before.IsActive || after.ProjectID == nil ||
		*after.ProjectID != projectID || !after.IsActive {
		t.Fatalf("unexpected snapshots: before=%+v after=%+v", before, after)
	}
	if len(dispatcher.commands) != 1 {
		t.Fatalf("commands: got %d, want 1", len(dispatcher.commands))
	}
	refresh, ok := dispatcher.commands[0].(appcollaboration.FacilityHierarchyRefreshRequired)
	if !ok || refresh.ProjectID != projectID || refresh.ActorID == nil ||
		*refresh.ActorID != actorID || refresh.OperationID != operationID ||
		refresh.CorrelationID != operationID || refresh.EventID != eventID ||
		refresh.Scope != appcollaboration.FacilityScopeProject || !refresh.FullRefresh ||
		len(refresh.EntityIDs) != 0 {
		t.Fatalf("unexpected refresh command: %+v", dispatcher.commands[0])
	}
}

func TestDeactivateForProjectRetainsOwnerAndDispatchesAfterCommit(t *testing.T) {
	projectID := associationTestUUID(11)
	objectDataID := associationTestUUID(12)
	harness := &associationHarness{
		projectID: projectID,
		base: &domainFacility.ObjectData{
			Base:      domain.Base{ID: objectDataID},
			ProjectID: &projectID,
			IsActive:  true,
		},
	}
	dispatcher := &associationDispatcherSpy{harness: harness}
	handler := newAssociationHandler(
		harness,
		dispatcher,
		associationTestUUID(13),
		associationTestUUID(14),
		associationTestUUID(15),
		time.Now(),
	)

	outcome, err := handler.ExecuteDeactivate(
		context.Background(),
		DeactivateForProjectCommand{ProjectID: projectID, ObjectDataID: objectDataID},
	)
	if err != nil {
		t.Fatalf("deactivate ObjectData: %v", err)
	}
	if outcome.ObjectData.ProjectID == nil || *outcome.ObjectData.ProjectID != projectID ||
		outcome.ObjectData.IsActive || harness.base.ProjectID == nil ||
		*harness.base.ProjectID != projectID || harness.base.IsActive {
		t.Fatalf("unexpected deactivated state: outcome=%+v base=%+v", outcome.ObjectData, harness.base)
	}
	if got := outcome.Mutation.Changes[0].ChangedFields; len(got) != 1 || got[0] != "is_active" {
		t.Fatalf("changed fields: got %v, want [is_active]", got)
	}
	if len(dispatcher.commands) != 1 {
		t.Fatalf("commands: got %d, want 1", len(dispatcher.commands))
	}
}

func TestProjectAssociationFailureRollsBackAndDoesNotDispatch(t *testing.T) {
	projectID := associationTestUUID(21)
	objectDataID := associationTestUUID(22)
	persistenceErr := errors.New("update failed")
	commitErr := errors.New("commit failed")

	for _, testCase := range []struct {
		name      string
		updateErr error
		commitErr error
		wantErr   error
	}{
		{name: "write", updateErr: persistenceErr, wantErr: persistenceErr},
		{name: "commit", commitErr: commitErr, wantErr: commitErr},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			harness := &associationHarness{
				projectID: projectID,
				base: &domainFacility.ObjectData{
					Base: domain.Base{ID: objectDataID},
				},
				updateErr: testCase.updateErr,
				commitErr: testCase.commitErr,
			}
			dispatcher := &associationDispatcherSpy{harness: harness}
			handler := newAssociationHandler(
				harness,
				dispatcher,
				associationTestUUID(23),
				associationTestUUID(24),
				associationTestUUID(25),
				time.Now(),
			)

			_, err := handler.ExecuteAttach(context.Background(), AttachToProjectCommand{
				ProjectID:    projectID,
				ObjectDataID: objectDataID,
			})

			if !errors.Is(err, testCase.wantErr) {
				t.Fatalf("expected %v, got %v", testCase.wantErr, err)
			}
			if harness.committed || harness.base.ProjectID != nil || harness.base.IsActive {
				t.Fatalf("failed association escaped transaction: %+v", harness.base)
			}
			if len(dispatcher.commands) != 0 {
				t.Fatalf("commands after rollback: %v", dispatcher.commands)
			}
		})
	}
}

func TestAttachToProjectRejectsAnotherOwnerWithoutWriteOrDispatch(t *testing.T) {
	projectID := associationTestUUID(31)
	ownerID := associationTestUUID(32)
	objectDataID := associationTestUUID(33)
	harness := &associationHarness{
		projectID: projectID,
		base: &domainFacility.ObjectData{
			Base:      domain.Base{ID: objectDataID},
			ProjectID: &ownerID,
			IsActive:  true,
		},
	}
	dispatcher := &associationDispatcherSpy{harness: harness}
	handler := newAssociationHandler(
		harness,
		dispatcher,
		associationTestUUID(34),
		associationTestUUID(35),
		associationTestUUID(36),
		time.Now(),
	)

	_, err := handler.ExecuteAttach(context.Background(), AttachToProjectCommand{
		ProjectID:    projectID,
		ObjectDataID: objectDataID,
	})

	if !errors.Is(err, domain.ErrConflict) {
		t.Fatalf("expected conflict, got %v", err)
	}
	if harness.updateCalls != 0 || harness.committed || len(dispatcher.commands) != 0 {
		t.Fatalf(
			"conflicting association had side effects: updates=%d committed=%v commands=%d",
			harness.updateCalls,
			harness.committed,
			len(dispatcher.commands),
		)
	}
}

func TestProjectAssociationReportsDispatchFailureWithoutFailingMutation(t *testing.T) {
	projectID := associationTestUUID(41)
	objectDataID := associationTestUUID(42)
	dispatchErr := errors.New("realtime unavailable")
	harness := &associationHarness{
		projectID: projectID,
		base: &domainFacility.ObjectData{
			Base: domain.Base{ID: objectDataID},
		},
	}
	dispatcher := &associationDispatcherSpy{harness: harness, err: dispatchErr}
	reported := make([]error, 0, 1)
	handler := newAssociationHandler(
		harness,
		dispatcher,
		associationTestUUID(43),
		associationTestUUID(44),
		associationTestUUID(45),
		time.Now(),
	)
	handler.reportError = func(err error) { reported = append(reported, err) }

	result, err := handler.AttachToProject(context.Background(), AttachToProjectCommand{
		ProjectID:    projectID,
		ObjectDataID: objectDataID,
	})

	if err != nil || result == nil || !harness.committed {
		t.Fatalf("committed mutation failed because dispatch failed: result=%+v err=%v", result, err)
	}
	if len(reported) != 1 || !errors.Is(reported[0], dispatchErr) {
		t.Fatalf("reported errors: %v", reported)
	}
}
