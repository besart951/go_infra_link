package spscontroller

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

type projectSPSAssignmentHistoryBatchKey struct{}

type projectSPSAssignmentState struct {
	link            *domainProject.ProjectSPSController
	descendantLinks []string
	history         []string
}

func (s projectSPSAssignmentState) clone() projectSPSAssignmentState {
	return projectSPSAssignmentState{
		link:            cloneProjectSPSControllerLink(s.link),
		descendantLinks: append([]string(nil), s.descendantLinks...),
		history:         append([]string(nil), s.history...),
	}
}

type projectSPSAssignmentUnit struct {
	state *projectSPSAssignmentState
}

type projectSPSAssignmentHarness struct {
	committed       projectSPSAssignmentState
	link            *domainProject.ProjectSPSController
	workflowErr     error
	commitErr       error
	runnerCalls     int
	workflowCalls   int
	projectID       uuid.UUID
	controllerID    uuid.UUID
	historyBatchIDs []uuid.UUID
}

func (h *projectSPSAssignmentHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	staged := h.committed.clone()
	if err := run(ctx, projectSPSAssignmentUnit{state: &staged}); err != nil {
		return err
	}
	if h.commitErr != nil {
		return h.commitErr
	}
	h.committed = staged
	return nil
}

func (h *projectSPSAssignmentHarness) factory(
	unit apptransaction.UnitOfWork,
) (AssignToProjectWorkflow, error) {
	typed, ok := unit.(projectSPSAssignmentUnit)
	if !ok || typed.state == nil {
		return nil, errors.New("unexpected project SPS assignment transaction unit")
	}
	return &projectSPSAssignmentWorkflowStub{harness: h, state: typed.state}, nil
}

type projectSPSAssignmentWorkflowStub struct {
	harness *projectSPSAssignmentHarness
	state   *projectSPSAssignmentState
}

func (s *projectSPSAssignmentWorkflowStub) CreateSPSController(
	ctx context.Context,
	projectID uuid.UUID,
	spsControllerID uuid.UUID,
) (*domainProject.ProjectSPSController, error) {
	s.harness.workflowCalls++
	s.harness.projectID = projectID
	s.harness.controllerID = spsControllerID
	if batchID, ok := ctx.Value(projectSPSAssignmentHistoryBatchKey{}).(uuid.UUID); ok {
		s.harness.historyBatchIDs = append(s.harness.historyBatchIDs, batchID)
	}
	link := cloneProjectSPSControllerLink(s.harness.link)
	if link != nil {
		s.state.link = link
		s.state.descendantLinks = append(s.state.descendantLinks, "field_device:linked")
		s.state.history = append(
			s.state.history,
			"project_sps_controller:create",
			"project_field_device:create",
		)
	}
	if s.harness.workflowErr != nil {
		return nil, s.harness.workflowErr
	}
	return link, nil
}

type projectSPSAssignmentDispatcherStub struct {
	harness           *projectSPSAssignmentHarness
	commands          []appcollaboration.Command
	err               error
	calledAfterCommit bool
}

func (s *projectSPSAssignmentDispatcherStub) Dispatch(
	_ context.Context,
	command appcollaboration.Command,
) error {
	s.commands = append(s.commands, command)
	s.calledAfterCommit = s.harness.committed.link != nil &&
		len(s.harness.committed.descendantLinks) == 1 &&
		len(s.harness.committed.history) == 2
	return s.err
}

func TestAssignSPSControllerToProjectCommitsDescendantLinksAndHistoryBeforeRefresh(t *testing.T) {
	projectID := uuid.New()
	spsControllerID := uuid.New()
	linkID := uuid.New()
	actorID := uuid.New()
	operationID := uuid.New()
	eventID := uuid.New()
	createdAt := time.Date(2026, time.July, 22, 1, 0, 0, 0, time.UTC)
	occurredAt := createdAt.Add(time.Minute)
	link := &domainProject.ProjectSPSController{
		ProjectID:       projectID,
		SPSControllerID: spsControllerID,
	}
	link.ID = linkID
	link.CreatedAt = createdAt
	link.UpdatedAt = createdAt
	harness := &projectSPSAssignmentHarness{link: link}
	dispatcher := &projectSPSAssignmentDispatcherStub{harness: harness}
	generatedIDs := []uuid.UUID{operationID, eventID}
	handler := NewAssignToProjectHandler(AssignToProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, projectSPSAssignmentHistoryBatchKey{}, batchID)
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
		ProjectID:       projectID,
		SPSControllerID: spsControllerID,
	})

	if err != nil {
		t.Fatalf("assign SPSController to project: %v", err)
	}
	if harness.runnerCalls != 1 || harness.workflowCalls != 1 ||
		harness.projectID != projectID || harness.controllerID != spsControllerID ||
		harness.committed.link == nil || harness.committed.link.ID != linkID ||
		!reflect.DeepEqual(harness.committed.descendantLinks, []string{"field_device:linked"}) ||
		!reflect.DeepEqual(harness.committed.history, []string{
			"project_sps_controller:create",
			"project_field_device:create",
		}) || !reflect.DeepEqual(harness.historyBatchIDs, []uuid.UUID{operationID}) {
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
	if change.EntityType != mutation.EntityTypeProjectSPSController ||
		change.EntityID != linkID || change.ParentID == nil || *change.ParentID != projectID ||
		change.Action != domainHistory.ActionCreate || len(change.Before) != 0 ||
		!reflect.DeepEqual(change.ChangedFields, []mutation.FieldName{mutation.FieldNameSPSController}) {
		t.Fatalf("canonical change: %+v", change)
	}
	var snapshot projectSPSControllerLinkSnapshot
	if err := json.Unmarshal(change.After, &snapshot); err != nil {
		t.Fatalf("decode canonical link snapshot: %v", err)
	}
	if snapshot.ID != linkID || snapshot.ProjectID != projectID ||
		snapshot.SPSControllerID != spsControllerID ||
		!snapshot.CreatedAt.Equal(createdAt) {
		t.Fatalf("canonical snapshot: %+v", snapshot)
	}
	if len(dispatcher.commands) != 1 || !dispatcher.calledAfterCommit {
		t.Fatalf("dispatch timing: %+v", dispatcher)
	}
	command, ok := dispatcher.commands[0].(appcollaboration.FacilityHierarchyRefreshRequired)
	if !ok || command.ProjectID != projectID || command.EventID != eventID ||
		command.OperationID != operationID || command.CorrelationID != operationID ||
		command.ActorID == nil || *command.ActorID != actorID ||
		command.Scope != appcollaboration.FacilityScopeSPSController || command.FullRefresh ||
		!reflect.DeepEqual(command.EntityIDs, []uuid.UUID{spsControllerID}) {
		t.Fatalf("typed collaboration command: %#v", dispatcher.commands[0])
	}
}

func TestAssignSPSControllerToProjectRollsBackCascadeAndDoesNotDispatch(t *testing.T) {
	projectID := uuid.New()
	spsControllerID := uuid.New()
	link := &domainProject.ProjectSPSController{
		ProjectID:       projectID,
		SPSControllerID: spsControllerID,
	}
	link.ID = uuid.New()

	for _, testCase := range []struct {
		name        string
		workflowErr error
		commitErr   error
	}{
		{name: "descendant link or history", workflowErr: errors.New("link descendant failed")},
		{name: "commit", commitErr: errors.New("commit failed")},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			harness := &projectSPSAssignmentHarness{
				link:        link,
				workflowErr: testCase.workflowErr,
				commitErr:   testCase.commitErr,
			}
			dispatcher := &projectSPSAssignmentDispatcherStub{harness: harness}
			handler := NewAssignToProjectHandler(AssignToProjectDependencies{
				TransactionRunner:   harness.runner,
				TransactionWorkflow: harness.factory,
				Dispatcher:          dispatcher,
			})

			outcome, err := handler.Execute(context.Background(), AssignToProjectCommand{
				ProjectID:       projectID,
				SPSControllerID: spsControllerID,
			})

			wantErr := testCase.workflowErr
			if wantErr == nil {
				wantErr = testCase.commitErr
			}
			if !errors.Is(err, wantErr) || outcome.Link != nil ||
				harness.committed.link != nil ||
				len(harness.committed.descendantLinks) != 0 ||
				len(harness.committed.history) != 0 || len(dispatcher.commands) != 0 {
				t.Fatalf("rollback outcome: outcome=%+v err=%v harness=%+v commands=%v",
					outcome,
					err,
					harness,
					dispatcher.commands,
				)
			}
		})
	}
}

func TestAssignSPSControllerToProjectReportsDispatchFailureAfterCommit(t *testing.T) {
	projectID := uuid.New()
	spsControllerID := uuid.New()
	link := &domainProject.ProjectSPSController{
		ProjectID:       projectID,
		SPSControllerID: spsControllerID,
	}
	link.ID = uuid.New()
	dispatchErr := errors.New("realtime unavailable")
	harness := &projectSPSAssignmentHarness{link: link}
	dispatcher := &projectSPSAssignmentDispatcherStub{
		harness: harness,
		err:     dispatchErr,
	}
	var reported []error
	handler := NewAssignToProjectHandler(AssignToProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Dispatcher:          dispatcher,
		ReportError:         func(err error) { reported = append(reported, err) },
	})

	created, err := handler.AssignToProject(context.Background(), AssignToProjectCommand{
		ProjectID:       projectID,
		SPSControllerID: spsControllerID,
	})

	if err != nil || created == nil || created.ID != link.ID ||
		harness.committed.link == nil || len(harness.committed.history) != 2 ||
		len(reported) != 1 || !errors.Is(reported[0], dispatchErr) {
		t.Fatalf("best-effort dispatch: created=%+v err=%v reported=%v", created, err, reported)
	}
}

func TestAssignSPSControllerToProjectMissingConfigurationFailsWithoutWrites(t *testing.T) {
	handler := NewAssignToProjectHandler(AssignToProjectDependencies{})

	created, err := handler.AssignToProject(context.Background(), AssignToProjectCommand{
		ProjectID:       uuid.New(),
		SPSControllerID: uuid.New(),
	})

	if created != nil || !errors.Is(err, ErrAssignToProjectTransactionNotConfigured) {
		t.Fatalf("missing configuration: created=%+v err=%v", created, err)
	}
}
