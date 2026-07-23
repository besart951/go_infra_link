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
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type assignProjectHistoryBatchKey struct{}

type assignProjectTransactionState struct {
	link    *domainProject.ProjectFieldDevice
	history []uuid.UUID
}

func (s assignProjectTransactionState) clone() assignProjectTransactionState {
	return assignProjectTransactionState{
		link:    cloneProjectFieldDeviceLink(s.link),
		history: append([]uuid.UUID(nil), s.history...),
	}
}

type assignProjectTransactionUnit struct {
	state *assignProjectTransactionState
}

type assignProjectTransactionHarness struct {
	committed       assignProjectTransactionState
	link            *domainProject.ProjectFieldDevice
	workflowErr     error
	commitErr       error
	runnerCalls     int
	workflowCalls   int
	historyBatchIDs []uuid.UUID
	callProjectID   uuid.UUID
	callDeviceID    uuid.UUID
}

func (h *assignProjectTransactionHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	staged := h.committed.clone()
	if err := run(ctx, assignProjectTransactionUnit{state: &staged}); err != nil {
		return err
	}
	if h.commitErr != nil {
		return h.commitErr
	}
	h.committed = staged
	return nil
}

func (h *assignProjectTransactionHarness) factory(
	unit apptransaction.UnitOfWork,
) (AssignToProjectWorkflow, error) {
	typed, ok := unit.(assignProjectTransactionUnit)
	if !ok || typed.state == nil {
		return nil, errors.New("unexpected project-assignment transaction unit")
	}
	return &assignProjectWorkflowStub{harness: h, state: typed.state}, nil
}

type assignProjectWorkflowStub struct {
	harness *assignProjectTransactionHarness
	state   *assignProjectTransactionState
}

func (s *assignProjectWorkflowStub) CreateFieldDevice(
	ctx context.Context,
	projectID uuid.UUID,
	fieldDeviceID uuid.UUID,
) (*domainProject.ProjectFieldDevice, error) {
	s.harness.workflowCalls++
	s.harness.callProjectID = projectID
	s.harness.callDeviceID = fieldDeviceID
	if batchID, ok := ctx.Value(assignProjectHistoryBatchKey{}).(uuid.UUID); ok {
		s.harness.historyBatchIDs = append(s.harness.historyBatchIDs, batchID)
	}
	if s.harness.link != nil {
		s.state.link = cloneProjectFieldDeviceLink(s.harness.link)
		s.state.history = append(s.state.history, s.harness.link.ID)
	}
	if s.harness.workflowErr != nil {
		return nil, s.harness.workflowErr
	}
	return cloneProjectFieldDeviceLink(s.harness.link), nil
}

type assignProjectDispatcherStub struct {
	harness           *assignProjectTransactionHarness
	commands          []appcollaboration.Command
	err               error
	calledAfterCommit bool
}

func (s *assignProjectDispatcherStub) Dispatch(
	_ context.Context,
	command appcollaboration.Command,
) error {
	s.commands = append(s.commands, command)
	s.calledAfterCommit = s.harness.committed.link != nil &&
		len(s.harness.committed.history) == 1
	return s.err
}

func TestAssignToProjectCommitsLinkAndHistoryBeforeTypedRefresh(t *testing.T) {
	projectID := testUUID(701)
	fieldDeviceID := testUUID(702)
	linkID := testUUID(703)
	actorID := testUUID(704)
	operationID := testUUID(705)
	eventID := testUUID(706)
	createdAt := time.Date(2026, time.July, 21, 18, 0, 0, 0, time.UTC)
	occurredAt := createdAt.Add(time.Minute)
	link := &domainProject.ProjectFieldDevice{
		ProjectID:     projectID,
		FieldDeviceID: fieldDeviceID,
	}
	link.ID = linkID
	link.CreatedAt = createdAt
	link.UpdatedAt = createdAt
	harness := &assignProjectTransactionHarness{link: link}
	dispatcher := &assignProjectDispatcherStub{harness: harness}
	generatedIDs := []uuid.UUID{operationID, eventID}
	handler := NewAssignToProjectHandler(AssignToProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, assignProjectHistoryBatchKey{}, batchID)
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

	outcome, err := handler.Execute(context.Background(), AssignToProjectCommand{
		ProjectID:     projectID,
		FieldDeviceID: fieldDeviceID,
	})

	if err != nil {
		t.Fatalf("assign existing FieldDevice: %v", err)
	}
	if harness.runnerCalls != 1 || harness.workflowCalls != 1 ||
		harness.callProjectID != projectID || harness.callDeviceID != fieldDeviceID ||
		harness.committed.link == nil || harness.committed.link.ID != linkID ||
		!reflect.DeepEqual(harness.committed.history, []uuid.UUID{linkID}) ||
		!reflect.DeepEqual(harness.historyBatchIDs, []uuid.UUID{operationID}) {
		t.Fatalf("transaction state: %+v", harness)
	}
	if outcome.Link == nil || outcome.Link.ID != linkID ||
		outcome.Mutation.OperationID != operationID || outcome.Mutation.BatchID == nil ||
		*outcome.Mutation.BatchID != operationID || outcome.Mutation.ActorID == nil ||
		*outcome.Mutation.ActorID != actorID ||
		!outcome.Mutation.OccurredAt.Equal(occurredAt) ||
		!reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectID}) ||
		len(outcome.Mutation.Changes) != 1 {
		t.Fatalf("assignment outcome: %+v", outcome)
	}
	change := outcome.Mutation.Changes[0]
	if change.EntityType != "project_field_device" || change.EntityID != linkID ||
		change.ParentID == nil || *change.ParentID != projectID ||
		change.Action != domainHistory.ActionCreate || len(change.Before) != 0 ||
		!reflect.DeepEqual(change.ChangedFields, []mutation.FieldName{mutation.FieldNameFieldDevice}) {
		t.Fatalf("canonical change: %+v", change)
	}
	var snapshot projectFieldDeviceLinkSnapshot
	if err := json.Unmarshal(change.After, &snapshot); err != nil {
		t.Fatalf("decode canonical link snapshot: %v", err)
	}
	if snapshot.ID != linkID || snapshot.ProjectID != projectID ||
		snapshot.FieldDeviceID != fieldDeviceID || !snapshot.CreatedAt.Equal(createdAt) {
		t.Fatalf("canonical snapshot: %+v", snapshot)
	}
	if len(dispatcher.commands) != 1 || !dispatcher.calledAfterCommit {
		t.Fatalf("dispatch timing: %+v", dispatcher)
	}
	command, ok := dispatcher.commands[0].(appcollaboration.FacilityHierarchyRefreshRequired)
	if !ok || command.ProjectID != projectID || command.EventID != eventID ||
		command.OperationID != operationID || command.CorrelationID != operationID ||
		command.ActorID == nil || *command.ActorID != actorID ||
		command.Scope != appcollaboration.FacilityScopeFieldDevice || command.FullRefresh ||
		!reflect.DeepEqual(command.EntityIDs, []uuid.UUID{fieldDeviceID}) {
		t.Fatalf("typed collaboration command: %#v", dispatcher.commands[0])
	}
}

func TestAssignToProjectRollsBackAndDoesNotDispatchOnWorkflowOrCommitFailure(t *testing.T) {
	projectID := testUUID(711)
	fieldDeviceID := testUUID(712)
	link := &domainProject.ProjectFieldDevice{ProjectID: projectID, FieldDeviceID: fieldDeviceID}
	link.ID = testUUID(713)

	for _, testCase := range []struct {
		name        string
		workflowErr error
		commitErr   error
	}{
		{name: "workflow", workflowErr: errors.New("history write failed")},
		{name: "commit", commitErr: errors.New("commit failed")},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			harness := &assignProjectTransactionHarness{
				link:        link,
				workflowErr: testCase.workflowErr,
				commitErr:   testCase.commitErr,
			}
			dispatcher := &assignProjectDispatcherStub{harness: harness}
			handler := NewAssignToProjectHandler(AssignToProjectDependencies{
				TransactionRunner:   harness.runner,
				TransactionWorkflow: harness.factory,
				Dispatcher:          dispatcher,
			})

			outcome, err := handler.Execute(context.Background(), AssignToProjectCommand{
				ProjectID:     projectID,
				FieldDeviceID: fieldDeviceID,
			})

			wantErr := testCase.workflowErr
			if wantErr == nil {
				wantErr = testCase.commitErr
			}
			if !errors.Is(err, wantErr) || outcome.Link != nil ||
				harness.committed.link != nil || len(harness.committed.history) != 0 ||
				len(dispatcher.commands) != 0 {
				t.Fatalf("rollback outcome: outcome=%+v err=%v state=%+v commands=%v",
					outcome,
					err,
					harness.committed,
					dispatcher.commands,
				)
			}
		})
	}
}

func TestAssignToProjectReportsDispatchFailureWithoutChangingCommittedResult(t *testing.T) {
	projectID := testUUID(721)
	fieldDeviceID := testUUID(722)
	link := &domainProject.ProjectFieldDevice{ProjectID: projectID, FieldDeviceID: fieldDeviceID}
	link.ID = testUUID(723)
	dispatchErr := errors.New("realtime unavailable")
	harness := &assignProjectTransactionHarness{link: link}
	dispatcher := &assignProjectDispatcherStub{harness: harness, err: dispatchErr}
	var reported []error
	handler := NewAssignToProjectHandler(AssignToProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Dispatcher:          dispatcher,
		ReportError:         func(err error) { reported = append(reported, err) },
	})

	created, err := handler.AssignToProject(context.Background(), AssignToProjectCommand{
		ProjectID:     projectID,
		FieldDeviceID: fieldDeviceID,
	})

	if err != nil || created == nil || created.ID != link.ID ||
		harness.committed.link == nil || len(reported) != 1 ||
		!errors.Is(reported[0], dispatchErr) {
		t.Fatalf("best-effort dispatch: created=%+v err=%v reported=%v", created, err, reported)
	}
}

func TestAssignToProjectMissingConfigurationFailsWithoutWrites(t *testing.T) {
	handler := NewAssignToProjectHandler(AssignToProjectDependencies{})

	created, err := handler.AssignToProject(context.Background(), AssignToProjectCommand{
		ProjectID:     testUUID(731),
		FieldDeviceID: testUUID(732),
	})

	if created != nil || !errors.Is(err, ErrAssignToProjectTransactionNotConfigured) {
		t.Fatalf("missing configuration: created=%+v err=%v", created, err)
	}
}
