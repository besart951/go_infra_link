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

type cloneHistoryBatchKey struct{}

type cloneTransactionState struct {
	controllers map[uuid.UUID]*domainFacility.SPSController
	history     []string
}

func (s cloneTransactionState) clone() cloneTransactionState {
	controllers := make(map[uuid.UUID]*domainFacility.SPSController, len(s.controllers))
	for id, controller := range s.controllers {
		controllers[id] = cloneSPSController(controller)
	}
	return cloneTransactionState{
		controllers: controllers,
		history:     append([]string(nil), s.history...),
	}
}

type cloneTransactionUnit struct {
	state *cloneTransactionState
}

type cloneTransactionHarness struct {
	committed       cloneTransactionState
	copyEntity      *domainFacility.SPSController
	cloneErr        error
	reloadErr       error
	commitErr       error
	runnerCalls     int
	cloneCalls      int
	reloadCalls     int
	historyBatchIDs []uuid.UUID
}

func (h *cloneTransactionHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	staged := h.committed.clone()
	if err := run(ctx, cloneTransactionUnit{state: &staged}); err != nil {
		return err
	}
	if h.commitErr != nil {
		return h.commitErr
	}
	h.committed = staged
	return nil
}

func (h *cloneTransactionHarness) factory(
	unit apptransaction.UnitOfWork,
) (CloneWorkflow, error) {
	typed, ok := unit.(cloneTransactionUnit)
	if !ok || typed.state == nil {
		return nil, errors.New("unexpected SPS clone transaction unit")
	}
	return &cloneWorkflowStub{harness: h, state: typed.state}, nil
}

type cloneWorkflowStub struct {
	harness *cloneTransactionHarness
	state   *cloneTransactionState
}

func (s *cloneWorkflowStub) CopyByID(
	ctx context.Context,
	sourceID uuid.UUID,
) (*domainFacility.SPSController, error) {
	s.harness.cloneCalls++
	if _, ok := s.state.controllers[sourceID]; !ok {
		return nil, domain.ErrNotFound
	}
	copyEntity := cloneSPSController(s.harness.copyEntity)
	if copyEntity != nil && copyEntity.ID != uuid.Nil {
		s.state.controllers[copyEntity.ID] = cloneSPSController(copyEntity)
		s.state.history = append(s.state.history,
			"sps_controller:create",
			"sps_controller_system_type:create",
			"field_device:create",
		)
	}
	if batchID, ok := ctx.Value(cloneHistoryBatchKey{}).(uuid.UUID); ok {
		s.harness.historyBatchIDs = append(
			s.harness.historyBatchIDs,
			batchID,
			batchID,
			batchID,
		)
	}
	return copyEntity, s.harness.cloneErr
}

func (s *cloneWorkflowStub) GetByID(
	_ context.Context,
	id uuid.UUID,
) (*domainFacility.SPSController, error) {
	s.harness.reloadCalls++
	if s.harness.reloadErr != nil {
		return nil, s.harness.reloadErr
	}
	controller, ok := s.state.controllers[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return cloneSPSController(controller), nil
}

type cloneProjectLinkReaderStub struct {
	harness           *cloneTransactionHarness
	copyID            uuid.UUID
	links             []*domainProject.ProjectSPSController
	err               error
	calls             int
	received          []uuid.UUID
	calledAfterCommit bool
}

func (s *cloneProjectLinkReaderStub) GetBySPSControllerIDs(
	_ context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectSPSController, error) {
	s.calls++
	s.received = append([]uuid.UUID(nil), ids...)
	_, s.calledAfterCommit = s.harness.committed.controllers[s.copyID]
	return s.links, s.err
}

func TestCloneCommitsDeepCopyHistoryBeforeResolvingProjectsAndDispatching(t *testing.T) {
	sourceID := spsTestUUID(171)
	copyID := spsTestUUID(172)
	cabinetID := spsTestUUID(173)
	projectOne := spsTestUUID(174)
	projectTwo := spsTestUUID(175)
	actorID := spsTestUUID(176)
	operationID := spsTestUUID(177)
	eventOne := spsTestUUID(178)
	eventTwo := spsTestUUID(179)
	createdAt := time.Date(2026, time.July, 20, 15, 0, 0, 0, time.UTC)
	occurredAt := createdAt.Add(time.Second)
	gaDevice := "AAB"
	description := "Cloned controller"
	source := &domainFacility.SPSController{
		Base:             domain.Base{ID: sourceID, CreatedAt: createdAt.Add(-time.Hour)},
		ControlCabinetID: cabinetID,
		DeviceName:       "BLD_AK01_AAA",
	}
	copyEntity := &domainFacility.SPSController{
		Base:              domain.Base{ID: copyID, CreatedAt: createdAt, UpdatedAt: createdAt},
		ControlCabinetID:  cabinetID,
		GADevice:          &gaDevice,
		DeviceName:        "BLD_AK01_AAB",
		DeviceDescription: &description,
	}
	harness := &cloneTransactionHarness{
		committed: cloneTransactionState{controllers: map[uuid.UUID]*domainFacility.SPSController{
			sourceID: source,
		}},
		copyEntity: copyEntity,
	}
	links := &cloneProjectLinkReaderStub{
		harness: harness,
		copyID:  copyID,
		links: []*domainProject.ProjectSPSController{
			{ProjectID: projectTwo, SPSControllerID: copyID},
			{ProjectID: projectOne, SPSControllerID: copyID},
			{ProjectID: projectOne, SPSControllerID: copyID},
			{ProjectID: spsTestUUID(180), SPSControllerID: sourceID},
		},
	}
	dispatcher := &updateCommandDispatcherStub{}
	generatedIDs := []uuid.UUID{operationID, eventOne, eventTwo}
	handler := NewCloneHandler(CloneDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, cloneHistoryBatchKey{}, batchID)
		},
		ProjectLinks: links,
		Dispatcher:   dispatcher,
		Actor:        func(context.Context) *uuid.UUID { return &actorID },
		NewID: func() uuid.UUID {
			id := generatedIDs[0]
			generatedIDs = generatedIDs[1:]
			return id
		},
		Now: func() time.Time { return occurredAt },
	})

	outcome, err := handler.Execute(context.Background(), CloneCommand{
		SourceSPSControllerID: sourceID,
	})
	if err != nil {
		t.Fatalf("execute SPS clone: %v", err)
	}
	if harness.runnerCalls != 1 || harness.cloneCalls != 1 || harness.reloadCalls != 1 {
		t.Fatalf("transaction calls: runner=%d clone=%d reload=%d",
			harness.runnerCalls,
			harness.cloneCalls,
			harness.reloadCalls,
		)
	}
	if _, ok := harness.committed.controllers[sourceID]; !ok {
		t.Fatal("source controller was not preserved")
	}
	if _, ok := harness.committed.controllers[copyID]; !ok ||
		!reflect.DeepEqual(harness.committed.history, []string{
			"sps_controller:create",
			"sps_controller_system_type:create",
			"field_device:create",
		}) {
		t.Fatalf("committed clone state: %+v", harness.committed)
	}
	if !links.calledAfterCommit || links.calls != 1 ||
		!reflect.DeepEqual(links.received, []uuid.UUID{copyID}) {
		t.Fatalf("post-commit scope: after=%t calls=%d ids=%v",
			links.calledAfterCommit,
			links.calls,
			links.received,
		)
	}
	if !reflect.DeepEqual(harness.historyBatchIDs, []uuid.UUID{
		operationID,
		operationID,
		operationID,
	}) || outcome.Mutation.BatchID == nil || *outcome.Mutation.BatchID != operationID {
		t.Fatalf("copy history batches=%v result=%v", harness.historyBatchIDs, outcome.Mutation.BatchID)
	}
	if outcome.SPSController == nil || outcome.SPSController.ID != copyID ||
		outcome.Mutation.OperationID != operationID ||
		outcome.Mutation.ActorID == nil || *outcome.Mutation.ActorID != actorID ||
		!outcome.Mutation.OccurredAt.Equal(occurredAt) ||
		!reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectOne, projectTwo}) {
		t.Fatalf("clone outcome: %+v", outcome)
	}
	if len(outcome.Mutation.Changes) != 1 {
		t.Fatalf("changes: got %d, want 1", len(outcome.Mutation.Changes))
	}
	change := outcome.Mutation.Changes[0]
	if change.EntityID != copyID || change.ParentID == nil || *change.ParentID != cabinetID ||
		change.Action != domainHistory.ActionCreate || len(change.Before) != 0 {
		t.Fatalf("clone root change: %+v", change)
	}
	var snapshot spsControllerSnapshot
	if err := json.Unmarshal(change.After, &snapshot); err != nil {
		t.Fatalf("decode copy snapshot: %v", err)
	}
	if snapshot.ID != copyID || snapshot.ControlCabinetID != cabinetID ||
		snapshot.GADevice == nil || *snapshot.GADevice != gaDevice ||
		snapshot.DeviceDescription == nil || *snapshot.DeviceDescription != description {
		t.Fatalf("copy snapshot: %+v", snapshot)
	}
	if len(dispatcher.commands) != 2 {
		t.Fatalf("commands: got %d, want 2", len(dispatcher.commands))
	}
	for index, raw := range dispatcher.commands {
		command, ok := raw.(appcollaboration.SPSControllerCloned)
		if !ok {
			t.Fatalf("command: got %T, want SPSControllerCloned", raw)
		}
		wantProjectID := []uuid.UUID{projectOne, projectTwo}[index]
		if command.ProjectID != wantProjectID ||
			command.SourceSPSControllerID != sourceID ||
			command.SPSController.ID != copyID ||
			command.OperationID != operationID || command.CorrelationID != operationID ||
			command.SchemaVersion != appcollaboration.SchemaVersionV1 {
			t.Fatalf("clone command: %+v", command)
		}
	}
}

func TestCloneFailureOrCommitFailureRollsBackAndDoesNotResolveOrDispatch(t *testing.T) {
	cloneErr := errors.New("copy children failed")
	reloadErr := errors.New("reload failed")
	commitErr := errors.New("commit failed")
	for _, test := range []struct {
		name      string
		cloneErr  error
		reloadErr error
		commitErr error
		wantErr   error
	}{
		{name: "deep copy", cloneErr: cloneErr, wantErr: cloneErr},
		{name: "authoritative reload", reloadErr: reloadErr, wantErr: reloadErr},
		{name: "commit", commitErr: commitErr, wantErr: commitErr},
	} {
		t.Run(test.name, func(t *testing.T) {
			sourceID := spsTestUUID(181)
			copyID := spsTestUUID(182)
			harness := &cloneTransactionHarness{
				committed: cloneTransactionState{controllers: map[uuid.UUID]*domainFacility.SPSController{
					sourceID: {Base: domain.Base{ID: sourceID}},
				}},
				copyEntity: &domainFacility.SPSController{
					Base:             domain.Base{ID: copyID},
					ControlCabinetID: spsTestUUID(183),
				},
				cloneErr:  test.cloneErr,
				reloadErr: test.reloadErr,
				commitErr: test.commitErr,
			}
			links := &cloneProjectLinkReaderStub{harness: harness, copyID: copyID}
			dispatcher := &updateCommandDispatcherStub{}
			handler := NewCloneHandler(CloneDependencies{
				TransactionRunner:   harness.runner,
				TransactionWorkflow: harness.factory,
				ProjectLinks:        links,
				Dispatcher:          dispatcher,
			})

			_, err := handler.Execute(context.Background(), CloneCommand{
				SourceSPSControllerID: sourceID,
			})
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("clone error: got %v, want %v", err, test.wantErr)
			}
			if len(harness.committed.controllers) != 1 ||
				harness.committed.controllers[sourceID] == nil ||
				len(harness.committed.history) != 0 {
				t.Fatalf("failed clone escaped transaction: %+v", harness.committed)
			}
			if links.calls != 0 || len(dispatcher.commands) != 0 {
				t.Fatalf("post-commit work ran after rollback: links=%d commands=%v",
					links.calls,
					dispatcher.commands,
				)
			}
		})
	}
}

func TestCloneScopeFailureIsBestEffortAfterCommit(t *testing.T) {
	sourceID := spsTestUUID(191)
	copyID := spsTestUUID(192)
	scopeErr := errors.New("scope lookup failed")
	harness := &cloneTransactionHarness{
		committed: cloneTransactionState{controllers: map[uuid.UUID]*domainFacility.SPSController{
			sourceID: {Base: domain.Base{ID: sourceID}},
		}},
		copyEntity: &domainFacility.SPSController{
			Base:             domain.Base{ID: copyID},
			ControlCabinetID: spsTestUUID(193),
		},
	}
	links := &cloneProjectLinkReaderStub{
		harness: harness,
		copyID:  copyID,
		err:     scopeErr,
	}
	dispatcher := &updateCommandDispatcherStub{}
	handler := NewCloneHandler(CloneDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		ProjectLinks:        links,
		Dispatcher:          dispatcher,
	})

	outcome, err := handler.Execute(context.Background(), CloneCommand{
		SourceSPSControllerID: sourceID,
	})
	if err != nil {
		t.Fatalf("clone with scope failure: %v", err)
	}
	if harness.committed.controllers[copyID] == nil || outcome.SPSController == nil {
		t.Fatalf("clone did not commit: outcome=%+v state=%+v", outcome, harness.committed)
	}
	if len(outcome.DispatchErrors) != 1 || !errors.Is(outcome.DispatchErrors[0], scopeErr) {
		t.Fatalf("dispatch errors: got %v, want wrapped %v", outcome.DispatchErrors, scopeErr)
	}
	if len(dispatcher.commands) != 0 {
		t.Fatalf("unexpected command after scope failure: %+v", dispatcher.commands)
	}
}

func TestCloneRejectsNilSourceBeforeOpeningTransaction(t *testing.T) {
	harness := &cloneTransactionHarness{}
	handler := NewCloneHandler(CloneDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
	})

	_, err := handler.Execute(context.Background(), CloneCommand{})
	if !errors.Is(err, domain.ErrInvalidArgument) {
		t.Fatalf("clone error: got %v, want %v", err, domain.ErrInvalidArgument)
	}
	if harness.runnerCalls != 0 {
		t.Fatalf("invalid clone opened %d transactions", harness.runnerCalls)
	}
}
