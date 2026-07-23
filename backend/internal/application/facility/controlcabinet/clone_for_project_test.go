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
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

type projectCabinetCloneHistoryBatchKey struct{}

type projectCabinetCloneTransactionState struct {
	cabinets    map[uuid.UUID]*domainFacility.ControlCabinet
	projectRows map[uuid.UUID][]string
	history     []string
}

func (s projectCabinetCloneTransactionState) clone() projectCabinetCloneTransactionState {
	cabinets := make(map[uuid.UUID]*domainFacility.ControlCabinet, len(s.cabinets))
	for id, cabinet := range s.cabinets {
		cabinets[id] = cloneControlCabinet(cabinet)
	}
	projectRows := make(map[uuid.UUID][]string, len(s.projectRows))
	for projectID, rows := range s.projectRows {
		projectRows[projectID] = append([]string(nil), rows...)
	}
	return projectCabinetCloneTransactionState{
		cabinets:    cabinets,
		projectRows: projectRows,
		history:     append([]string(nil), s.history...),
	}
}

type projectCabinetCloneTransactionUnit struct {
	state *projectCabinetCloneTransactionState
}

type projectCabinetCloneTransactionHarness struct {
	committed       projectCabinetCloneTransactionState
	copyEntity      *domainFacility.ControlCabinet
	copyErr         error
	commitErr       error
	runnerCalls     int
	copyCalls       int
	projectID       uuid.UUID
	sourceID        uuid.UUID
	historyBatchIDs []uuid.UUID
}

func (h *projectCabinetCloneTransactionHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	staged := h.committed.clone()
	if err := run(ctx, projectCabinetCloneTransactionUnit{state: &staged}); err != nil {
		return err
	}
	if h.commitErr != nil {
		return h.commitErr
	}
	h.committed = staged
	return nil
}

func (h *projectCabinetCloneTransactionHarness) factory(
	unit apptransaction.UnitOfWork,
) (CloneForProjectWorkflow, error) {
	typed, ok := unit.(projectCabinetCloneTransactionUnit)
	if !ok || typed.state == nil {
		return nil, errors.New("unexpected project cabinet clone transaction unit")
	}
	return &projectCabinetCloneWorkflowStub{harness: h, state: typed.state}, nil
}

type projectCabinetCloneWorkflowStub struct {
	harness *projectCabinetCloneTransactionHarness
	state   *projectCabinetCloneTransactionState
}

func (s *projectCabinetCloneWorkflowStub) CopyControlCabinet(
	ctx context.Context,
	projectID uuid.UUID,
	sourceID uuid.UUID,
) (*domainFacility.ControlCabinet, error) {
	s.harness.copyCalls++
	s.harness.projectID = projectID
	s.harness.sourceID = sourceID
	if _, ok := s.state.cabinets[sourceID]; !ok {
		return nil, domain.ErrNotFound
	}
	copyEntity := cloneControlCabinet(s.harness.copyEntity)
	if copyEntity != nil && copyEntity.ID != uuid.Nil {
		s.state.cabinets[copyEntity.ID] = cloneControlCabinet(copyEntity)
		s.state.projectRows[projectID] = append(
			s.state.projectRows[projectID],
			"control_cabinet:"+copyEntity.ID.String(),
			"sps_controller:copied",
			"field_device:copied",
		)
		s.state.history = append(s.state.history,
			"control_cabinet:create",
			"sps_controller:create",
			"sps_controller_system_type:create",
			"field_device:create",
			"specification:create",
			"bacnet_object:create",
			"project_control_cabinet:create",
			"project_sps_controller:create",
			"project_field_device:create",
		)
	}
	if batchID, ok := ctx.Value(projectCabinetCloneHistoryBatchKey{}).(uuid.UUID); ok {
		for range 9 {
			s.harness.historyBatchIDs = append(s.harness.historyBatchIDs, batchID)
		}
	}
	return copyEntity, s.harness.copyErr
}

type projectCabinetCloneDispatcherStub struct {
	harness           *projectCabinetCloneTransactionHarness
	projectID         uuid.UUID
	copyID            uuid.UUID
	commands          []appcollaboration.Command
	err               error
	calledAfterCommit bool
}

func (s *projectCabinetCloneDispatcherStub) Dispatch(
	_ context.Context,
	command appcollaboration.Command,
) error {
	s.commands = append(s.commands, command)
	copyEntity := s.harness.committed.cabinets[s.copyID]
	rows := s.harness.committed.projectRows[s.projectID]
	s.calledAfterCommit = copyEntity != nil && len(rows) == 3 &&
		rows[0] == "control_cabinet:"+s.copyID.String()
	return s.err
}

func TestCloneForProjectCommitsHierarchyAndLinksBeforeTypedDispatch(t *testing.T) {
	projectID := cabinetTestUUID(501)
	unrelatedProjectID := cabinetTestUUID(502)
	sourceID := cabinetTestUUID(503)
	copyID := cabinetTestUUID(504)
	buildingID := cabinetTestUUID(505)
	actorID := cabinetTestUUID(506)
	operationID := cabinetTestUUID(507)
	eventID := cabinetTestUUID(508)
	createdAt := time.Date(2026, time.July, 20, 17, 0, 0, 0, time.UTC)
	occurredAt := createdAt.Add(time.Second)
	number := "AK03"
	harness := &projectCabinetCloneTransactionHarness{
		committed: projectCabinetCloneTransactionState{
			cabinets: map[uuid.UUID]*domainFacility.ControlCabinet{
				sourceID: {
					Base:       domain.Base{ID: sourceID},
					BuildingID: buildingID,
				},
			},
			projectRows: map[uuid.UUID][]string{
				unrelatedProjectID: {"control_cabinet:" + sourceID.String()},
			},
		},
		copyEntity: &domainFacility.ControlCabinet{
			Base:             domain.Base{ID: copyID, CreatedAt: createdAt, UpdatedAt: createdAt},
			BuildingID:       buildingID,
			ControlCabinetNr: &number,
		},
	}
	dispatcher := &projectCabinetCloneDispatcherStub{
		harness:   harness,
		projectID: projectID,
		copyID:    copyID,
	}
	generatedIDs := []uuid.UUID{operationID, eventID}
	handler := NewCloneForProjectHandler(CloneForProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, projectCabinetCloneHistoryBatchKey{}, batchID)
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
		ProjectID:              projectID,
		SourceControlCabinetID: sourceID,
	})
	if err != nil {
		t.Fatalf("execute project cabinet clone: %v", err)
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
	if harness.committed.cabinets[sourceID] == nil || harness.committed.cabinets[copyID] == nil ||
		!reflect.DeepEqual(harness.committed.projectRows[projectID], []string{
			"control_cabinet:" + copyID.String(),
			"sps_controller:copied",
			"field_device:copied",
		}) || !reflect.DeepEqual(harness.committed.projectRows[unrelatedProjectID], []string{
		"control_cabinet:" + sourceID.String(),
	}) {
		t.Fatalf("committed project clone: %+v", harness.committed)
	}
	if len(harness.committed.history) != 9 || len(harness.historyBatchIDs) != 9 {
		t.Fatalf("history=%v batches=%v", harness.committed.history, harness.historyBatchIDs)
	}
	for _, batchID := range harness.historyBatchIDs {
		if batchID != operationID {
			t.Fatalf("history batch: got %s, want %s", batchID, operationID)
		}
	}
	if outcome.ControlCabinet == nil || outcome.ControlCabinet.ID != copyID ||
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
	if change.EntityID != copyID || change.ParentID == nil || *change.ParentID != buildingID ||
		change.Action != domainHistory.ActionCreate || len(change.Before) != 0 {
		t.Fatalf("project clone change: %+v", change)
	}
	var snapshot controlCabinetSnapshot
	if err := json.Unmarshal(change.After, &snapshot); err != nil {
		t.Fatalf("decode project clone snapshot: %v", err)
	}
	if snapshot.ID != copyID || snapshot.ControlCabinetNr == nil ||
		*snapshot.ControlCabinetNr != number {
		t.Fatalf("project clone snapshot: %+v", snapshot)
	}
	if !dispatcher.calledAfterCommit || len(dispatcher.commands) != 1 {
		t.Fatalf("dispatch timing: after=%t commands=%v", dispatcher.calledAfterCommit, dispatcher.commands)
	}
	command, ok := dispatcher.commands[0].(appcollaboration.ControlCabinetCloned)
	if !ok || command.ProjectID != projectID ||
		command.SourceControlCabinetID != sourceID || command.ControlCabinet.ID != copyID ||
		command.OperationID != operationID || command.EventID != eventID ||
		command.CorrelationID != operationID ||
		command.SchemaVersion != appcollaboration.SchemaVersionV1 {
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
			projectID := cabinetTestUUID(511)
			sourceID := cabinetTestUUID(512)
			copyID := cabinetTestUUID(513)
			harness := &projectCabinetCloneTransactionHarness{
				committed: projectCabinetCloneTransactionState{
					cabinets: map[uuid.UUID]*domainFacility.ControlCabinet{
						sourceID: {Base: domain.Base{ID: sourceID}},
					},
					projectRows: map[uuid.UUID][]string{},
				},
				copyEntity: &domainFacility.ControlCabinet{
					Base:       domain.Base{ID: copyID},
					BuildingID: cabinetTestUUID(514),
				},
				copyErr:   test.copyErr,
				commitErr: test.commitErr,
			}
			dispatcher := &projectCabinetCloneDispatcherStub{
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
				ProjectID:              projectID,
				SourceControlCabinetID: sourceID,
			})
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("project clone error: got %v, want %v", err, test.wantErr)
			}
			if len(harness.committed.cabinets) != 1 ||
				len(harness.committed.projectRows) != 0 || len(harness.committed.history) != 0 {
				t.Fatalf("failed project clone escaped transaction: %+v", harness.committed)
			}
			if len(dispatcher.commands) != 0 {
				t.Fatalf("dispatched after rollback: %+v", dispatcher.commands)
			}
		})
	}
}

func TestCloneForProjectDispatchFailureIsBestEffortAndReported(t *testing.T) {
	projectID := cabinetTestUUID(521)
	sourceID := cabinetTestUUID(522)
	copyID := cabinetTestUUID(523)
	dispatchErr := errors.New("transport unavailable")
	harness := &projectCabinetCloneTransactionHarness{
		committed: projectCabinetCloneTransactionState{
			cabinets: map[uuid.UUID]*domainFacility.ControlCabinet{
				sourceID: {Base: domain.Base{ID: sourceID}},
			},
			projectRows: map[uuid.UUID][]string{},
		},
		copyEntity: &domainFacility.ControlCabinet{
			Base:       domain.Base{ID: copyID},
			BuildingID: cabinetTestUUID(524),
		},
	}
	dispatcher := &projectCabinetCloneDispatcherStub{
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
		ProjectID:              projectID,
		SourceControlCabinetID: sourceID,
	})
	if err != nil || copyEntity == nil || copyEntity.ID != copyID {
		t.Fatalf("best-effort project clone: entity=%+v err=%v", copyEntity, err)
	}
	if harness.committed.cabinets[copyID] == nil ||
		len(reported) != 1 || !errors.Is(reported[0], dispatchErr) {
		t.Fatalf("commit/report state: committed=%+v reported=%v", harness.committed, reported)
	}
}

func TestCloneForProjectRejectsMissingScopeOrSourceBeforeTransaction(t *testing.T) {
	harness := &projectCabinetCloneTransactionHarness{}
	handler := NewCloneForProjectHandler(CloneForProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
	})

	for _, command := range []CloneForProjectCommand{
		{SourceControlCabinetID: cabinetTestUUID(531)},
		{ProjectID: cabinetTestUUID(532)},
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
