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
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type projectSPSReassignmentHistoryBatchKey struct{}

type projectSPSReassignmentState struct {
	link            *domainProject.ProjectSPSController
	descendantLinks []uuid.UUID
	history         []string
}

func (s projectSPSReassignmentState) clone() projectSPSReassignmentState {
	return projectSPSReassignmentState{
		link:            cloneProjectSPSControllerLink(s.link),
		descendantLinks: append([]uuid.UUID(nil), s.descendantLinks...),
		history:         append([]string(nil), s.history...),
	}
}

type projectSPSReassignmentUnit struct {
	state *projectSPSReassignmentState
}

type projectSPSReassignmentHarness struct {
	committed       projectSPSReassignmentState
	workflowErr     error
	commitErr       error
	updatedAt       time.Time
	runnerCalls     int
	readCalls       int
	updateCalls     int
	projectID       uuid.UUID
	linkID          uuid.UUID
	controllerID    uuid.UUID
	historyBatchIDs []uuid.UUID
}

func (h *projectSPSReassignmentHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	staged := h.committed.clone()
	if err := run(ctx, projectSPSReassignmentUnit{state: &staged}); err != nil {
		return err
	}
	if h.commitErr != nil {
		return h.commitErr
	}
	h.committed = staged
	return nil
}

func (h *projectSPSReassignmentHarness) factory(
	unit apptransaction.UnitOfWork,
) (ReassignProjectLinkWorkflow, error) {
	typed, ok := unit.(projectSPSReassignmentUnit)
	if !ok || typed.state == nil {
		return nil, errors.New("unexpected project SPS reassignment transaction unit")
	}
	return &projectSPSReassignmentWorkflowStub{harness: h, state: typed.state}, nil
}

type projectSPSReassignmentWorkflowStub struct {
	harness *projectSPSReassignmentHarness
	state   *projectSPSReassignmentState
}

func (s *projectSPSReassignmentWorkflowStub) GetByIds(
	_ context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectSPSController, error) {
	s.harness.readCalls++
	if s.state.link == nil || len(ids) != 1 || ids[0] != s.state.link.ID {
		return nil, nil
	}
	return []*domainProject.ProjectSPSController{
		cloneProjectSPSControllerLink(s.state.link),
	}, nil
}

func (s *projectSPSReassignmentWorkflowStub) UpdateSPSController(
	ctx context.Context,
	linkID uuid.UUID,
	projectID uuid.UUID,
	spsControllerID uuid.UUID,
) (*domainProject.ProjectSPSController, error) {
	s.harness.updateCalls++
	s.harness.linkID = linkID
	s.harness.projectID = projectID
	s.harness.controllerID = spsControllerID
	if batchID, ok := ctx.Value(projectSPSReassignmentHistoryBatchKey{}).(uuid.UUID); ok {
		s.harness.historyBatchIDs = append(s.harness.historyBatchIDs, batchID)
	}
	if s.state.link != nil {
		s.state.link.SPSControllerID = spsControllerID
		s.state.link.UpdatedAt = s.harness.updatedAt
		s.state.descendantLinks = append(s.state.descendantLinks, spsControllerID)
		s.state.history = append(
			s.state.history,
			"project_sps_controller:update",
			"project_field_device:create",
		)
	}
	if s.harness.workflowErr != nil {
		return nil, s.harness.workflowErr
	}
	return cloneProjectSPSControllerLink(s.state.link), nil
}

type projectSPSReassignmentDispatcherStub struct {
	harness           *projectSPSReassignmentHarness
	wantControllerID  uuid.UUID
	commands          []appcollaboration.Command
	err               error
	calledAfterCommit bool
}

func (s *projectSPSReassignmentDispatcherStub) Dispatch(
	_ context.Context,
	command appcollaboration.Command,
) error {
	s.commands = append(s.commands, command)
	s.calledAfterCommit = s.harness.committed.link != nil &&
		s.harness.committed.link.SPSControllerID == s.wantControllerID &&
		len(s.harness.committed.descendantLinks) == 2 &&
		len(s.harness.committed.history) == 2
	return s.err
}

func TestReassignProjectSPSControllerLinkCommitsDescendantsAndHistoryBeforeRefresh(t *testing.T) {
	projectID := uuid.New()
	linkID := uuid.New()
	previousSPSControllerID := uuid.New()
	newSPSControllerID := uuid.New()
	oldDescendantID := uuid.New()
	actorID := uuid.New()
	operationID := uuid.New()
	eventID := uuid.New()
	createdAt := time.Date(2026, time.July, 22, 2, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	occurredAt := updatedAt.Add(time.Minute)
	harness := &projectSPSReassignmentHarness{
		committed: projectSPSReassignmentState{
			link: &domainProject.ProjectSPSController{
				Base: domain.Base{
					ID:        linkID,
					CreatedAt: createdAt,
					UpdatedAt: createdAt,
				},
				ProjectID:       projectID,
				SPSControllerID: previousSPSControllerID,
			},
			descendantLinks: []uuid.UUID{oldDescendantID},
		},
		updatedAt: updatedAt,
	}
	dispatcher := &projectSPSReassignmentDispatcherStub{
		harness:          harness,
		wantControllerID: newSPSControllerID,
	}
	generatedIDs := []uuid.UUID{operationID, eventID}
	handler := NewReassignProjectLinkHandler(ReassignProjectLinkDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, projectSPSReassignmentHistoryBatchKey{}, batchID)
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

	outcome, err := handler.Execute(context.Background(), ReassignProjectLinkCommand{
		ProjectID:       projectID,
		LinkID:          linkID,
		SPSControllerID: newSPSControllerID,
	})

	if err != nil {
		t.Fatalf("reassign SPSController project link: %v", err)
	}
	if harness.runnerCalls != 1 || harness.readCalls != 1 || harness.updateCalls != 1 ||
		harness.projectID != projectID || harness.linkID != linkID ||
		harness.controllerID != newSPSControllerID ||
		harness.committed.link == nil ||
		harness.committed.link.SPSControllerID != newSPSControllerID ||
		!reflect.DeepEqual(harness.committed.descendantLinks, []uuid.UUID{
			oldDescendantID,
			newSPSControllerID,
		}) || !reflect.DeepEqual(harness.committed.history, []string{
		"project_sps_controller:update",
		"project_field_device:create",
	}) || !reflect.DeepEqual(harness.historyBatchIDs, []uuid.UUID{operationID}) {
		t.Fatalf("transaction state: %+v", harness)
	}
	if outcome.Link == nil || outcome.Link.ID != linkID ||
		outcome.Link.SPSControllerID != newSPSControllerID ||
		outcome.Mutation.OperationID != operationID || outcome.Mutation.BatchID == nil ||
		*outcome.Mutation.BatchID != operationID || outcome.Mutation.ActorID == nil ||
		*outcome.Mutation.ActorID != actorID ||
		!outcome.Mutation.OccurredAt.Equal(occurredAt) ||
		!reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectID}) ||
		len(outcome.Mutation.Changes) != 1 {
		t.Fatalf("reassignment outcome: %+v", outcome)
	}
	change := outcome.Mutation.Changes[0]
	if change.EntityType != mutation.EntityTypeProjectSPSController ||
		change.EntityID != linkID || change.ParentID == nil || *change.ParentID != projectID ||
		change.Action != domainHistory.ActionUpdate ||
		!reflect.DeepEqual(change.ChangedFields, []mutation.FieldName{mutation.FieldNameSPSController}) {
		t.Fatalf("canonical change: %+v", change)
	}
	var before projectSPSControllerLinkSnapshot
	if err := json.Unmarshal(change.Before, &before); err != nil {
		t.Fatalf("decode before snapshot: %v", err)
	}
	var after projectSPSControllerLinkSnapshot
	if err := json.Unmarshal(change.After, &after); err != nil {
		t.Fatalf("decode after snapshot: %v", err)
	}
	if before.SPSControllerID != previousSPSControllerID ||
		after.SPSControllerID != newSPSControllerID ||
		!before.UpdatedAt.Equal(createdAt) || !after.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("canonical snapshots: before=%+v after=%+v", before, after)
	}
	if len(dispatcher.commands) != 1 || !dispatcher.calledAfterCommit {
		t.Fatalf("dispatch timing: %+v", dispatcher)
	}
	dispatched, ok := dispatcher.commands[0].(appcollaboration.FacilityHierarchyRefreshRequired)
	if !ok || dispatched.ProjectID != projectID || dispatched.EventID != eventID ||
		dispatched.OperationID != operationID || dispatched.CorrelationID != operationID ||
		dispatched.ActorID == nil || *dispatched.ActorID != actorID ||
		dispatched.Scope != appcollaboration.FacilityScopeSPSController ||
		dispatched.FullRefresh ||
		!reflect.DeepEqual(dispatched.EntityIDs, []uuid.UUID{newSPSControllerID}) {
		t.Fatalf("typed collaboration command: %#v", dispatcher.commands[0])
	}
}

func TestReassignProjectSPSControllerLinkRollsBackAndDoesNotDispatch(t *testing.T) {
	projectID := uuid.New()
	linkID := uuid.New()
	previousSPSControllerID := uuid.New()
	newSPSControllerID := uuid.New()
	workflowErr := errors.New("link descendant failed")
	commitErr := errors.New("commit failed")

	for _, testCase := range []struct {
		name        string
		workflowErr error
		commitErr   error
	}{
		{name: "descendant link or history", workflowErr: workflowErr},
		{name: "commit", commitErr: commitErr},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			harness := &projectSPSReassignmentHarness{
				committed: projectSPSReassignmentState{
					link: &domainProject.ProjectSPSController{
						Base:            domain.Base{ID: linkID},
						ProjectID:       projectID,
						SPSControllerID: previousSPSControllerID,
					},
				},
				workflowErr: testCase.workflowErr,
				commitErr:   testCase.commitErr,
			}
			dispatcher := &projectSPSReassignmentDispatcherStub{
				harness:          harness,
				wantControllerID: newSPSControllerID,
			}
			handler := NewReassignProjectLinkHandler(ReassignProjectLinkDependencies{
				TransactionRunner:   harness.runner,
				TransactionWorkflow: harness.factory,
				Dispatcher:          dispatcher,
			})

			outcome, err := handler.Execute(context.Background(), ReassignProjectLinkCommand{
				ProjectID:       projectID,
				LinkID:          linkID,
				SPSControllerID: newSPSControllerID,
			})

			wantErr := testCase.workflowErr
			if wantErr == nil {
				wantErr = testCase.commitErr
			}
			if !errors.Is(err, wantErr) || outcome.Link != nil ||
				harness.committed.link == nil ||
				harness.committed.link.SPSControllerID != previousSPSControllerID ||
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

func TestReassignProjectSPSControllerLinkRejectsForeignProject(t *testing.T) {
	storedProjectID := uuid.New()
	harness := &projectSPSReassignmentHarness{
		committed: projectSPSReassignmentState{
			link: &domainProject.ProjectSPSController{
				Base:            domain.Base{ID: uuid.New()},
				ProjectID:       storedProjectID,
				SPSControllerID: uuid.New(),
			},
		},
	}
	dispatcher := &projectSPSReassignmentDispatcherStub{harness: harness}
	handler := NewReassignProjectLinkHandler(ReassignProjectLinkDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Dispatcher:          dispatcher,
	})

	outcome, err := handler.Execute(context.Background(), ReassignProjectLinkCommand{
		ProjectID:       uuid.New(),
		LinkID:          harness.committed.link.ID,
		SPSControllerID: uuid.New(),
	})

	if !errors.Is(err, domain.ErrNotFound) || outcome.Link != nil ||
		harness.readCalls != 1 || harness.updateCalls != 0 || len(dispatcher.commands) != 0 {
		t.Fatalf("foreign project result: outcome=%+v err=%v harness=%+v", outcome, err, harness)
	}
}

func TestReassignProjectSPSControllerLinkReportsDispatchFailureAfterCommit(t *testing.T) {
	projectID := uuid.New()
	linkID := uuid.New()
	newSPSControllerID := uuid.New()
	dispatchErr := errors.New("realtime unavailable")
	harness := &projectSPSReassignmentHarness{
		committed: projectSPSReassignmentState{
			link: &domainProject.ProjectSPSController{
				Base:            domain.Base{ID: linkID},
				ProjectID:       projectID,
				SPSControllerID: uuid.New(),
			},
			descendantLinks: []uuid.UUID{uuid.New()},
		},
	}
	dispatcher := &projectSPSReassignmentDispatcherStub{
		harness:          harness,
		wantControllerID: newSPSControllerID,
		err:              dispatchErr,
	}
	var reported []error
	handler := NewReassignProjectLinkHandler(ReassignProjectLinkDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Dispatcher:          dispatcher,
		ReportError:         func(err error) { reported = append(reported, err) },
	})

	updated, err := handler.ReassignProjectLink(context.Background(), ReassignProjectLinkCommand{
		ProjectID:       projectID,
		LinkID:          linkID,
		SPSControllerID: newSPSControllerID,
	})

	if err != nil || updated == nil || updated.SPSControllerID != newSPSControllerID ||
		harness.committed.link.SPSControllerID != newSPSControllerID ||
		len(reported) != 1 || !errors.Is(reported[0], dispatchErr) {
		t.Fatalf("best-effort dispatch: updated=%+v err=%v reported=%v", updated, err, reported)
	}
}

func TestReassignProjectSPSControllerLinkMissingConfigurationFailsWithoutWrites(t *testing.T) {
	handler := NewReassignProjectLinkHandler(ReassignProjectLinkDependencies{})

	updated, err := handler.ReassignProjectLink(context.Background(), ReassignProjectLinkCommand{
		ProjectID:       uuid.New(),
		LinkID:          uuid.New(),
		SPSControllerID: uuid.New(),
	})

	if updated != nil || !errors.Is(err, ErrReassignProjectLinkTransactionNotConfigured) {
		t.Fatalf("missing configuration: updated=%+v err=%v", updated, err)
	}
}
