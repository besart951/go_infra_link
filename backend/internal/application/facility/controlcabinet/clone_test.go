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
	domainCollaboration "github.com/besart951/go_infra_link/backend/internal/domain/collaboration"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

type cloneHistoryBatchKey struct{}

type cloneTransactionState struct {
	cabinets     map[uuid.UUID]*domainFacility.ControlCabinet
	projectLinks map[uuid.UUID][]uuid.UUID
	descendants  []string
	history      []string
}

func (s cloneTransactionState) clone() cloneTransactionState {
	cabinets := make(map[uuid.UUID]*domainFacility.ControlCabinet, len(s.cabinets))
	for id, cabinet := range s.cabinets {
		cabinets[id] = cloneControlCabinet(cabinet)
	}
	return cloneTransactionState{
		cabinets:     cabinets,
		projectLinks: cloneCabinetProjectLinks(s.projectLinks),
		descendants:  append([]string(nil), s.descendants...),
		history:      append([]string(nil), s.history...),
	}
}

type cloneTransactionUnit struct {
	state *cloneTransactionState
}

type cloneTransactionHarness struct {
	committed       cloneTransactionState
	copyEntity      *domainFacility.ControlCabinet
	projectIDs      []uuid.UUID
	projectIDsErr   error
	assignErr       error
	outbox          domainCollaboration.OutboxStore
	cloneErr        error
	reloadErr       error
	commitErr       error
	runnerCalls     int
	cloneCalls      int
	reloadCalls     int
	projectIDCalls  int
	assignCalls     int
	historyBatchIDs []uuid.UUID
}

func (h *cloneTransactionHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	staged := h.committed.clone()
	runCtx := ctx
	if h.outbox != nil {
		runCtx = domainCollaboration.WithOutboxStore(ctx, h.outbox)
	}
	if err := run(runCtx, cloneTransactionUnit{state: &staged}); err != nil {
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
		return nil, errors.New("unexpected cabinet clone transaction unit")
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
) (*domainFacility.ControlCabinet, error) {
	s.harness.cloneCalls++
	if _, ok := s.state.cabinets[sourceID]; !ok {
		return nil, domain.ErrNotFound
	}
	copyEntity := cloneControlCabinet(s.harness.copyEntity)
	if copyEntity != nil && copyEntity.ID != uuid.Nil {
		s.state.cabinets[copyEntity.ID] = cloneControlCabinet(copyEntity)
		s.state.descendants = append(s.state.descendants,
			"sps_controller",
			"sps_controller_system_type",
			"field_device",
			"specification",
			"bacnet_object",
		)
		s.state.history = append(s.state.history,
			"control_cabinet:create",
			"sps_controller:create",
			"sps_controller_system_type:create",
			"field_device:create",
			"specification:create",
			"bacnet_object:create",
		)
	}
	if batchID, ok := ctx.Value(cloneHistoryBatchKey{}).(uuid.UUID); ok {
		for range s.state.history {
			s.harness.historyBatchIDs = append(s.harness.historyBatchIDs, batchID)
		}
	}
	return copyEntity, s.harness.cloneErr
}

func (s *cloneWorkflowStub) GetByID(
	_ context.Context,
	id uuid.UUID,
) (*domainFacility.ControlCabinet, error) {
	s.harness.reloadCalls++
	if s.harness.reloadErr != nil {
		return nil, s.harness.reloadErr
	}
	cabinet, ok := s.state.cabinets[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	return cloneControlCabinet(cabinet), nil
}

func (s *cloneWorkflowStub) GetSourceProjectIDs(
	_ context.Context,
	_ uuid.UUID,
) ([]uuid.UUID, error) {
	s.harness.projectIDCalls++
	return append([]uuid.UUID(nil), s.harness.projectIDs...), s.harness.projectIDsErr
}

func (s *cloneWorkflowStub) AssignCopyToProject(
	ctx context.Context,
	projectID, controlCabinetID uuid.UUID,
) error {
	s.harness.assignCalls++
	if s.harness.assignErr != nil {
		return s.harness.assignErr
	}
	if s.state.projectLinks == nil {
		s.state.projectLinks = make(map[uuid.UUID][]uuid.UUID)
	}
	s.state.projectLinks[projectID] = append(
		s.state.projectLinks[projectID],
		controlCabinetID,
	)
	s.state.history = append(s.state.history, "project_control_cabinet:create")
	if batchID, ok := ctx.Value(cloneHistoryBatchKey{}).(uuid.UUID); ok {
		s.harness.historyBatchIDs = append(s.harness.historyBatchIDs, batchID)
	}
	return nil
}

func cloneCabinetProjectLinks(
	source map[uuid.UUID][]uuid.UUID,
) map[uuid.UUID][]uuid.UUID {
	cloned := make(map[uuid.UUID][]uuid.UUID, len(source))
	for projectID, cabinetIDs := range source {
		cloned[projectID] = append([]uuid.UUID(nil), cabinetIDs...)
	}
	return cloned
}

func TestCloneCommitsDeepCopyHistoryBeforeResolvingProjectsAndDispatching(t *testing.T) {
	sourceID := cabinetTestUUID(401)
	copyID := cabinetTestUUID(402)
	buildingID := cabinetTestUUID(403)
	projectOne := cabinetTestUUID(404)
	projectTwo := cabinetTestUUID(405)
	actorID := cabinetTestUUID(406)
	operationID := cabinetTestUUID(407)
	eventOne := cabinetTestUUID(408)
	eventTwo := cabinetTestUUID(409)
	createdAt := time.Date(2026, time.July, 20, 16, 0, 0, 0, time.UTC)
	occurredAt := createdAt.Add(time.Second)
	number := "AK02"
	harness := &cloneTransactionHarness{
		committed: cloneTransactionState{
			cabinets: map[uuid.UUID]*domainFacility.ControlCabinet{
				sourceID: {Base: domain.Base{ID: sourceID}, BuildingID: buildingID},
			},
			projectLinks: map[uuid.UUID][]uuid.UUID{},
		},
		copyEntity: &domainFacility.ControlCabinet{
			Base:             domain.Base{ID: copyID, CreatedAt: createdAt, UpdatedAt: createdAt},
			BuildingID:       buildingID,
			ControlCabinetNr: &number,
		},
		projectIDs: []uuid.UUID{projectTwo, projectOne, projectOne, uuid.Nil},
		outbox:     &updateOutboxStoreStub{},
	}
	dispatcher := &updateCommandDispatcherStub{}
	generatedIDs := []uuid.UUID{operationID, eventOne, eventTwo}
	handler := NewCloneHandler(CloneDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, cloneHistoryBatchKey{}, batchID)
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

	outcome, err := handler.Execute(context.Background(), CloneCommand{
		SourceControlCabinetID: sourceID,
	})
	if err != nil {
		t.Fatalf("execute cabinet clone: %v", err)
	}
	if harness.runnerCalls != 1 || harness.cloneCalls != 1 || harness.reloadCalls != 1 ||
		harness.projectIDCalls != 1 || harness.assignCalls != 2 {
		t.Fatalf("transaction calls: runner=%d clone=%d reload=%d project_ids=%d assignments=%d",
			harness.runnerCalls,
			harness.cloneCalls,
			harness.reloadCalls,
			harness.projectIDCalls,
			harness.assignCalls,
		)
	}
	if !reflect.DeepEqual(harness.committed.descendants, []string{
		"sps_controller",
		"sps_controller_system_type",
		"field_device",
		"specification",
		"bacnet_object",
	}) || len(harness.committed.history) != 8 ||
		!reflect.DeepEqual(harness.committed.projectLinks, map[uuid.UUID][]uuid.UUID{
			projectOne: {copyID},
			projectTwo: {copyID},
		}) {
		t.Fatalf("deep copy did not commit atomically: %+v", harness.committed)
	}
	wantBatches := []uuid.UUID{
		operationID, operationID, operationID,
		operationID, operationID, operationID,
		operationID, operationID,
	}
	if !reflect.DeepEqual(harness.historyBatchIDs, wantBatches) ||
		outcome.Mutation.BatchID == nil || *outcome.Mutation.BatchID != operationID {
		t.Fatalf("history batches=%v result=%v", harness.historyBatchIDs, outcome.Mutation.BatchID)
	}
	if outcome.ControlCabinet == nil || outcome.ControlCabinet.ID != copyID ||
		outcome.Mutation.OperationID != operationID ||
		outcome.Mutation.ActorID == nil || *outcome.Mutation.ActorID != actorID ||
		!outcome.Mutation.OccurredAt.Equal(occurredAt) ||
		!reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectOne, projectTwo}) {
		t.Fatalf("clone outcome: %+v", outcome)
	}
	change := outcome.Mutation.Changes[0]
	if change.EntityID != copyID || change.ParentID == nil || *change.ParentID != buildingID ||
		change.Action != domainHistory.ActionCreate || len(change.Before) != 0 {
		t.Fatalf("clone root change: %+v", change)
	}
	var snapshot controlCabinetSnapshot
	if err := json.Unmarshal(change.After, &snapshot); err != nil {
		t.Fatalf("decode clone snapshot: %v", err)
	}
	if snapshot.ID != copyID || snapshot.BuildingID != buildingID ||
		snapshot.ControlCabinetNr == nil || *snapshot.ControlCabinetNr != number {
		t.Fatalf("clone snapshot: %+v", snapshot)
	}
	if len(dispatcher.commands) != 2 {
		t.Fatalf("commands: got %d, want 2", len(dispatcher.commands))
	}
	for index, raw := range dispatcher.commands {
		command, ok := raw.(appcollaboration.ControlCabinetCloned)
		if !ok {
			t.Fatalf("command: got %T, want ControlCabinetCloned", raw)
		}
		wantProjectID := []uuid.UUID{projectOne, projectTwo}[index]
		if command.ProjectID != wantProjectID ||
			command.SourceControlCabinetID != sourceID ||
			command.ControlCabinet.ID != copyID ||
			command.OperationID != operationID || command.CorrelationID != operationID ||
			command.SchemaVersion != appcollaboration.SchemaVersionV2 {
			t.Fatalf("clone command: %+v", command)
		}
	}
	if len(harness.outbox.(*updateOutboxStoreStub).events) != 2 {
		t.Fatalf("durable clone events: %+v", harness.outbox)
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
			sourceID := cabinetTestUUID(421)
			copyID := cabinetTestUUID(422)
			harness := &cloneTransactionHarness{
				committed: cloneTransactionState{cabinets: map[uuid.UUID]*domainFacility.ControlCabinet{
					sourceID: {Base: domain.Base{ID: sourceID}},
				}},
				copyEntity: &domainFacility.ControlCabinet{
					Base:       domain.Base{ID: copyID},
					BuildingID: cabinetTestUUID(423),
				},
				cloneErr:  test.cloneErr,
				reloadErr: test.reloadErr,
				commitErr: test.commitErr,
			}
			dispatcher := &updateCommandDispatcherStub{}
			handler := NewCloneHandler(CloneDependencies{
				TransactionRunner:   harness.runner,
				TransactionWorkflow: harness.factory,
				Dispatcher:          dispatcher,
			})

			_, err := handler.Execute(context.Background(), CloneCommand{
				SourceControlCabinetID: sourceID,
			})
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("clone error: got %v, want %v", err, test.wantErr)
			}
			if len(harness.committed.cabinets) != 1 ||
				harness.committed.cabinets[sourceID] == nil ||
				len(harness.committed.descendants) != 0 || len(harness.committed.history) != 0 {
				t.Fatalf("failed clone escaped transaction: %+v", harness.committed)
			}
			if len(dispatcher.commands) != 0 {
				t.Fatalf("post-commit work ran after rollback: commands=%v", dispatcher.commands)
			}
		})
	}
}

func TestCloneSourceProjectFailureRollsBackBeforeCopy(t *testing.T) {
	sourceID := cabinetTestUUID(431)
	copyID := cabinetTestUUID(432)
	scopeErr := errors.New("scope lookup failed")
	harness := &cloneTransactionHarness{
		committed: cloneTransactionState{cabinets: map[uuid.UUID]*domainFacility.ControlCabinet{
			sourceID: {Base: domain.Base{ID: sourceID}},
		}},
		copyEntity: &domainFacility.ControlCabinet{
			Base:       domain.Base{ID: copyID},
			BuildingID: cabinetTestUUID(433),
		},
	}
	harness.projectIDsErr = scopeErr
	dispatcher := &updateCommandDispatcherStub{}
	handler := NewCloneHandler(CloneDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Dispatcher:          dispatcher,
	})

	outcome, err := handler.Execute(context.Background(), CloneCommand{
		SourceControlCabinetID: sourceID,
	})
	if !errors.Is(err, scopeErr) {
		t.Fatalf("clone source-project failure: got %v, want %v", err, scopeErr)
	}
	if harness.committed.cabinets[copyID] != nil || outcome.ControlCabinet != nil ||
		harness.cloneCalls != 0 {
		t.Fatalf("scope failure changed state: outcome=%+v state=%+v", outcome, harness.committed)
	}
	if len(dispatcher.commands) != 0 {
		t.Fatalf("unexpected command after scope failure: %+v", dispatcher.commands)
	}
}

func TestCloneReportsDispatchFailureWithoutChangingCommittedResult(t *testing.T) {
	sourceID := cabinetTestUUID(441)
	copyID := cabinetTestUUID(442)
	projectID := cabinetTestUUID(443)
	dispatchErr := errors.New("realtime unavailable")
	harness := &cloneTransactionHarness{
		committed: cloneTransactionState{cabinets: map[uuid.UUID]*domainFacility.ControlCabinet{
			sourceID: {Base: domain.Base{ID: sourceID}},
		}},
		copyEntity: &domainFacility.ControlCabinet{
			Base:       domain.Base{ID: copyID},
			BuildingID: cabinetTestUUID(444),
		},
		projectIDs: []uuid.UUID{projectID},
	}
	dispatcher := &updateCommandDispatcherStub{err: dispatchErr}
	var reported []error
	handler := NewCloneHandler(CloneDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Dispatcher:          dispatcher,
		ReportError: func(err error) {
			reported = append(reported, err)
		},
	})

	copyEntity, err := handler.Clone(context.Background(), CloneCommand{
		SourceControlCabinetID: sourceID,
	})
	if err != nil || copyEntity == nil || copyEntity.ID != copyID {
		t.Fatalf("best-effort clone result: entity=%+v err=%v", copyEntity, err)
	}
	if len(reported) != 1 || !errors.Is(reported[0], dispatchErr) {
		t.Fatalf("reported errors: got %v, want wrapped %v", reported, dispatchErr)
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
