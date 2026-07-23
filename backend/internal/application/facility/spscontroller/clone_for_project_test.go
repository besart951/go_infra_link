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
	"github.com/google/uuid"
)

type projectCloneHistoryBatchKey struct{}

type projectCloneTransactionState struct {
	controllers  map[uuid.UUID]*domainFacility.SPSController
	projectLinks map[uuid.UUID][]uuid.UUID
	history      []string
}

func (s projectCloneTransactionState) clone() projectCloneTransactionState {
	controllers := make(map[uuid.UUID]*domainFacility.SPSController, len(s.controllers))
	for id, controller := range s.controllers {
		controllers[id] = cloneSPSController(controller)
	}
	projectLinks := make(map[uuid.UUID][]uuid.UUID, len(s.projectLinks))
	for projectID, controllerIDs := range s.projectLinks {
		projectLinks[projectID] = append([]uuid.UUID(nil), controllerIDs...)
	}
	return projectCloneTransactionState{
		controllers:  controllers,
		projectLinks: projectLinks,
		history:      append([]string(nil), s.history...),
	}
}

type projectCloneTransactionUnit struct {
	state *projectCloneTransactionState
}

type projectCloneTransactionHarness struct {
	committed       projectCloneTransactionState
	copyEntity      *domainFacility.SPSController
	copyErr         error
	commitErr       error
	runnerCalls     int
	copyCalls       int
	projectID       uuid.UUID
	sourceID        uuid.UUID
	historyBatchIDs []uuid.UUID
}

func (h *projectCloneTransactionHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	staged := h.committed.clone()
	if err := run(ctx, projectCloneTransactionUnit{state: &staged}); err != nil {
		return err
	}
	if h.commitErr != nil {
		return h.commitErr
	}
	h.committed = staged
	return nil
}

func (h *projectCloneTransactionHarness) factory(
	unit apptransaction.UnitOfWork,
) (CloneForProjectWorkflow, error) {
	typed, ok := unit.(projectCloneTransactionUnit)
	if !ok || typed.state == nil {
		return nil, errors.New("unexpected project SPS clone transaction unit")
	}
	return &projectCloneWorkflowStub{harness: h, state: typed.state}, nil
}

type projectCloneWorkflowStub struct {
	harness *projectCloneTransactionHarness
	state   *projectCloneTransactionState
}

func (s *projectCloneWorkflowStub) CopySPSController(
	ctx context.Context,
	projectID uuid.UUID,
	sourceID uuid.UUID,
) (*domainFacility.SPSController, error) {
	s.harness.copyCalls++
	s.harness.projectID = projectID
	s.harness.sourceID = sourceID
	if _, ok := s.state.controllers[sourceID]; !ok {
		return nil, domain.ErrNotFound
	}
	copyEntity := cloneSPSController(s.harness.copyEntity)
	if copyEntity != nil && copyEntity.ID != uuid.Nil {
		s.state.controllers[copyEntity.ID] = cloneSPSController(copyEntity)
		s.state.projectLinks[projectID] = append(
			s.state.projectLinks[projectID],
			copyEntity.ID,
		)
		s.state.history = append(s.state.history,
			"sps_controller:create",
			"field_device:create",
			"project_sps_controller:create",
			"project_field_device:create",
		)
	}
	if batchID, ok := ctx.Value(projectCloneHistoryBatchKey{}).(uuid.UUID); ok {
		for range 4 {
			s.harness.historyBatchIDs = append(s.harness.historyBatchIDs, batchID)
		}
	}
	return copyEntity, s.harness.copyErr
}

type projectCloneDispatcherStub struct {
	harness           *projectCloneTransactionHarness
	projectID         uuid.UUID
	copyID            uuid.UUID
	commands          []appcollaboration.Command
	err               error
	calledAfterCommit bool
}

func (s *projectCloneDispatcherStub) Dispatch(
	_ context.Context,
	command appcollaboration.Command,
) error {
	s.commands = append(s.commands, command)
	copyEntity := s.harness.committed.controllers[s.copyID]
	links := s.harness.committed.projectLinks[s.projectID]
	s.calledAfterCommit = copyEntity != nil && reflect.DeepEqual(links, []uuid.UUID{s.copyID})
	return s.err
}

func TestCloneForProjectCommitsHierarchyAndLinksBeforeTypedDispatch(t *testing.T) {
	projectID := spsTestUUID(201)
	unrelatedProjectID := spsTestUUID(208)
	sourceID := spsTestUUID(202)
	copyID := spsTestUUID(203)
	cabinetID := spsTestUUID(204)
	actorID := spsTestUUID(205)
	operationID := spsTestUUID(206)
	eventID := spsTestUUID(207)
	createdAt := time.Date(2026, time.July, 20, 16, 0, 0, 0, time.UTC)
	occurredAt := createdAt.Add(time.Second)
	gaDevice := "AAC"
	harness := &projectCloneTransactionHarness{
		committed: projectCloneTransactionState{
			controllers: map[uuid.UUID]*domainFacility.SPSController{
				sourceID: {
					Base:             domain.Base{ID: sourceID},
					ControlCabinetID: cabinetID,
				},
			},
			projectLinks: map[uuid.UUID][]uuid.UUID{
				unrelatedProjectID: {sourceID},
			},
		},
		copyEntity: &domainFacility.SPSController{
			Base:             domain.Base{ID: copyID, CreatedAt: createdAt, UpdatedAt: createdAt},
			ControlCabinetID: cabinetID,
			GADevice:         &gaDevice,
			DeviceName:       "BLD_AK01_AAC",
		},
	}
	dispatcher := &projectCloneDispatcherStub{
		harness:   harness,
		projectID: projectID,
		copyID:    copyID,
	}
	generatedIDs := []uuid.UUID{operationID, eventID}
	handler := NewCloneForProjectHandler(CloneForProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, projectCloneHistoryBatchKey{}, batchID)
		},
		Dispatcher: dispatcher,
		Actor:      func(context.Context) *uuid.UUID { return &actorID },
		NewID: func() uuid.UUID {
			id := generatedIDs[0]
			generatedIDs = generatedIDs[1:]
			return id
		},
		Now: func() time.Time { return occurredAt },
	})

	outcome, err := handler.Execute(context.Background(), CloneForProjectCommand{
		ProjectID:             projectID,
		SourceSPSControllerID: sourceID,
	})
	if err != nil {
		t.Fatalf("execute project SPS clone: %v", err)
	}
	if harness.runnerCalls != 1 || harness.copyCalls != 1 ||
		harness.projectID != projectID || harness.sourceID != sourceID {
		t.Fatalf("workflow calls: runner=%d copy=%d project=%s source=%s",
			harness.runnerCalls,
			harness.copyCalls,
			harness.projectID,
			harness.sourceID,
		)
	}
	if harness.committed.controllers[sourceID] == nil || harness.committed.controllers[copyID] == nil ||
		!reflect.DeepEqual(harness.committed.projectLinks[projectID], []uuid.UUID{copyID}) ||
		!reflect.DeepEqual(harness.committed.projectLinks[unrelatedProjectID], []uuid.UUID{sourceID}) {
		t.Fatalf("committed project clone: %+v", harness.committed)
	}
	if !reflect.DeepEqual(harness.committed.history, []string{
		"sps_controller:create",
		"field_device:create",
		"project_sps_controller:create",
		"project_field_device:create",
	}) || !reflect.DeepEqual(harness.historyBatchIDs, []uuid.UUID{
		operationID,
		operationID,
		operationID,
		operationID,
	}) {
		t.Fatalf("history=%v batches=%v", harness.committed.history, harness.historyBatchIDs)
	}
	if outcome.SPSController == nil || outcome.SPSController.ID != copyID ||
		outcome.Mutation.OperationID != operationID ||
		outcome.Mutation.BatchID == nil || *outcome.Mutation.BatchID != operationID ||
		outcome.Mutation.ActorID == nil || *outcome.Mutation.ActorID != actorID ||
		!reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectID}) {
		t.Fatalf("project clone outcome: %+v", outcome)
	}
	if len(outcome.Mutation.Changes) != 1 {
		t.Fatalf("changes: got %d, want 1", len(outcome.Mutation.Changes))
	}
	change := outcome.Mutation.Changes[0]
	if change.EntityID != copyID || change.ParentID == nil || *change.ParentID != cabinetID ||
		change.Action != domainHistory.ActionCreate || len(change.Before) != 0 {
		t.Fatalf("project clone change: %+v", change)
	}
	var snapshot spsControllerSnapshot
	if err := json.Unmarshal(change.After, &snapshot); err != nil {
		t.Fatalf("decode project clone snapshot: %v", err)
	}
	if snapshot.ID != copyID || snapshot.GADevice == nil || *snapshot.GADevice != gaDevice {
		t.Fatalf("project clone snapshot: %+v", snapshot)
	}
	if !dispatcher.calledAfterCommit || len(dispatcher.commands) != 1 {
		t.Fatalf("dispatch timing: after=%t commands=%v", dispatcher.calledAfterCommit, dispatcher.commands)
	}
	command, ok := dispatcher.commands[0].(appcollaboration.SPSControllerCloned)
	if !ok || command.ProjectID != projectID ||
		command.SourceSPSControllerID != sourceID || command.SPSController.ID != copyID ||
		command.OperationID != operationID || command.EventID != eventID ||
		command.CorrelationID != operationID {
		t.Fatalf("project clone command: %+v", dispatcher.commands[0])
	}
}

func TestCloneForProjectWriteOrCommitFailureRollsBackLinksAndDoesNotDispatch(t *testing.T) {
	copyErr := errors.New("link copy failed")
	commitErr := errors.New("commit failed")
	for _, test := range []struct {
		name      string
		copyErr   error
		commitErr error
		wantErr   error
	}{
		{name: "copy or link", copyErr: copyErr, wantErr: copyErr},
		{name: "commit", commitErr: commitErr, wantErr: commitErr},
	} {
		t.Run(test.name, func(t *testing.T) {
			projectID := spsTestUUID(211)
			sourceID := spsTestUUID(212)
			copyID := spsTestUUID(213)
			harness := &projectCloneTransactionHarness{
				committed: projectCloneTransactionState{
					controllers: map[uuid.UUID]*domainFacility.SPSController{
						sourceID: {Base: domain.Base{ID: sourceID}},
					},
					projectLinks: map[uuid.UUID][]uuid.UUID{},
				},
				copyEntity: &domainFacility.SPSController{
					Base:             domain.Base{ID: copyID},
					ControlCabinetID: spsTestUUID(214),
				},
				copyErr:   test.copyErr,
				commitErr: test.commitErr,
			}
			dispatcher := &projectCloneDispatcherStub{
				harness:   harness,
				projectID: projectID,
				copyID:    copyID,
			}
			handler := NewCloneForProjectHandler(CloneForProjectDependencies{
				TransactionRunner:   harness.runner,
				TransactionWorkflow: harness.factory,
				Dispatcher:          dispatcher,
			})

			_, err := handler.Execute(context.Background(), CloneForProjectCommand{
				ProjectID:             projectID,
				SourceSPSControllerID: sourceID,
			})
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("project clone error: got %v, want %v", err, test.wantErr)
			}
			if len(harness.committed.controllers) != 1 ||
				len(harness.committed.projectLinks) != 0 ||
				len(harness.committed.history) != 0 {
				t.Fatalf("failed project clone escaped transaction: %+v", harness.committed)
			}
			if len(dispatcher.commands) != 0 {
				t.Fatalf("dispatched after rollback: %+v", dispatcher.commands)
			}
		})
	}
}

func TestCloneForProjectDispatchFailureIsBestEffortAndReported(t *testing.T) {
	projectID := spsTestUUID(221)
	sourceID := spsTestUUID(222)
	copyID := spsTestUUID(223)
	dispatchErr := errors.New("transport unavailable")
	harness := &projectCloneTransactionHarness{
		committed: projectCloneTransactionState{
			controllers: map[uuid.UUID]*domainFacility.SPSController{
				sourceID: {Base: domain.Base{ID: sourceID}},
			},
			projectLinks: map[uuid.UUID][]uuid.UUID{},
		},
		copyEntity: &domainFacility.SPSController{
			Base:             domain.Base{ID: copyID},
			ControlCabinetID: spsTestUUID(224),
		},
	}
	dispatcher := &projectCloneDispatcherStub{
		harness:   harness,
		projectID: projectID,
		copyID:    copyID,
		err:       dispatchErr,
	}
	var reported []error
	handler := NewCloneForProjectHandler(CloneForProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Dispatcher:          dispatcher,
		ReportError: func(err error) {
			reported = append(reported, err)
		},
	})

	copyEntity, err := handler.CloneForProject(context.Background(), CloneForProjectCommand{
		ProjectID:             projectID,
		SourceSPSControllerID: sourceID,
	})
	if err != nil || copyEntity == nil || copyEntity.ID != copyID {
		t.Fatalf("best-effort project clone: entity=%+v err=%v", copyEntity, err)
	}
	if harness.committed.controllers[copyID] == nil ||
		len(reported) != 1 || !errors.Is(reported[0], dispatchErr) {
		t.Fatalf("commit/report state: committed=%+v reported=%v", harness.committed, reported)
	}
}

func TestCloneForProjectRejectsMissingScopeOrSourceBeforeTransaction(t *testing.T) {
	harness := &projectCloneTransactionHarness{}
	handler := NewCloneForProjectHandler(CloneForProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
	})

	for _, command := range []CloneForProjectCommand{
		{SourceSPSControllerID: spsTestUUID(231)},
		{ProjectID: spsTestUUID(232)},
	} {
		_, err := handler.Execute(context.Background(), command)
		if !errors.Is(err, domain.ErrInvalidArgument) {
			t.Fatalf("validation error: got %v, want %v", err, domain.ErrInvalidArgument)
		}
	}
	if harness.runnerCalls != 0 {
		t.Fatalf("invalid commands opened %d transactions", harness.runnerCalls)
	}
}
