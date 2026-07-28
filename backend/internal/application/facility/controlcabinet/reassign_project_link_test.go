package controlcabinet

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

type projectCabinetReassignmentHistoryBatchKey struct{}

type projectCabinetReassignmentState struct {
	link            *domainProject.ProjectControlCabinet
	descendantLinks []string
	history         []string
}

func (s projectCabinetReassignmentState) clone() projectCabinetReassignmentState {
	return projectCabinetReassignmentState{
		link:            cloneProjectControlCabinetLink(s.link),
		descendantLinks: append([]string(nil), s.descendantLinks...),
		history:         append([]string(nil), s.history...),
	}
}

type projectCabinetReassignmentUnit struct {
	state *projectCabinetReassignmentState
}

type projectCabinetReassignmentHarness struct {
	committed       projectCabinetReassignmentState
	workflowErr     error
	commitErr       error
	updatedAt       time.Time
	runnerCalls     int
	readCalls       int
	updateCalls     int
	projectID       uuid.UUID
	linkID          uuid.UUID
	cabinetID       uuid.UUID
	historyBatchIDs []uuid.UUID
}

func (h *projectCabinetReassignmentHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	staged := h.committed.clone()
	if err := run(ctx, projectCabinetReassignmentUnit{state: &staged}); err != nil {
		return err
	}
	if h.commitErr != nil {
		return h.commitErr
	}
	h.committed = staged
	return nil
}

func (h *projectCabinetReassignmentHarness) factory(
	unit apptransaction.UnitOfWork,
) (ReassignProjectLinkWorkflow, error) {
	typed, ok := unit.(projectCabinetReassignmentUnit)
	if !ok || typed.state == nil {
		return nil, errors.New("unexpected project cabinet reassignment transaction unit")
	}
	return &projectCabinetReassignmentWorkflowStub{harness: h, state: typed.state}, nil
}

type projectCabinetReassignmentWorkflowStub struct {
	harness *projectCabinetReassignmentHarness
	state   *projectCabinetReassignmentState
}

func (s *projectCabinetReassignmentWorkflowStub) GetByIds(
	_ context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectControlCabinet, error) {
	s.harness.readCalls++
	if s.state.link == nil || len(ids) != 1 || ids[0] != s.state.link.ID {
		return nil, nil
	}
	return []*domainProject.ProjectControlCabinet{
		cloneProjectControlCabinetLink(s.state.link),
	}, nil
}

func (s *projectCabinetReassignmentWorkflowStub) UpdateControlCabinet(
	ctx context.Context,
	linkID uuid.UUID,
	projectID uuid.UUID,
	controlCabinetID uuid.UUID,
) (*domainProject.ProjectControlCabinet, error) {
	s.harness.updateCalls++
	s.harness.linkID = linkID
	s.harness.projectID = projectID
	s.harness.cabinetID = controlCabinetID
	if batchID, ok := ctx.Value(projectCabinetReassignmentHistoryBatchKey{}).(uuid.UUID); ok {
		s.harness.historyBatchIDs = append(s.harness.historyBatchIDs, batchID)
	}
	if s.state.link != nil {
		s.state.link.ControlCabinetID = controlCabinetID
		s.state.link.UpdatedAt = s.harness.updatedAt
		s.state.descendantLinks = append(
			s.state.descendantLinks,
			"new:sps_controller",
			"new:field_device",
		)
		s.state.history = append(
			s.state.history,
			"project_control_cabinet:update",
			"project_sps_controller:create",
			"project_field_device:create",
		)
	}
	if s.harness.workflowErr != nil {
		return nil, s.harness.workflowErr
	}
	return cloneProjectControlCabinetLink(s.state.link), nil
}

type projectCabinetReassignmentDispatcherStub struct {
	harness         *projectCabinetReassignmentHarness
	wantCabinetID   uuid.UUID
	commands        []appcollaboration.Command
	err             error
	afterCommitSeen bool
}

func (s *projectCabinetReassignmentDispatcherStub) Dispatch(
	_ context.Context,
	command appcollaboration.Command,
) error {
	s.commands = append(s.commands, command)
	s.afterCommitSeen = s.harness.committed.link != nil &&
		s.harness.committed.link.ControlCabinetID == s.wantCabinetID &&
		len(s.harness.committed.descendantLinks) == 4 &&
		len(s.harness.committed.history) == 3
	return s.err
}

func TestReassignProjectControlCabinetLinkCommitsDescendantsAndHistoryBeforeRefresh(t *testing.T) {
	projectID := uuid.New()
	linkID := uuid.New()
	previousCabinetID := uuid.New()
	newCabinetID := uuid.New()
	actorID := uuid.New()
	operationID := uuid.New()
	eventID := uuid.New()
	createdAt := time.Date(2026, time.July, 22, 4, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	occurredAt := updatedAt.Add(time.Minute)
	harness := &projectCabinetReassignmentHarness{
		committed: projectCabinetReassignmentState{
			link: &domainProject.ProjectControlCabinet{
				Base: domain.Base{
					ID:        linkID,
					CreatedAt: createdAt,
					UpdatedAt: createdAt,
				},
				ProjectID:        projectID,
				ControlCabinetID: previousCabinetID,
			},
			descendantLinks: []string{"old:sps_controller", "old:field_device"},
		},
		updatedAt: updatedAt,
	}
	dispatcher := &projectCabinetReassignmentDispatcherStub{
		harness:       harness,
		wantCabinetID: newCabinetID,
	}
	generatedIDs := []uuid.UUID{operationID, eventID}
	handler := NewReassignProjectLinkHandler(ReassignProjectLinkDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, projectCabinetReassignmentHistoryBatchKey{}, batchID)
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
		ProjectID:        projectID,
		LinkID:           linkID,
		ControlCabinetID: newCabinetID,
	})

	if err != nil {
		t.Fatalf("reassign ControlCabinet project link: %v", err)
	}
	if harness.runnerCalls != 1 || harness.readCalls != 1 || harness.updateCalls != 1 ||
		harness.projectID != projectID || harness.linkID != linkID ||
		harness.cabinetID != newCabinetID || harness.committed.link == nil ||
		harness.committed.link.ControlCabinetID != newCabinetID ||
		!reflect.DeepEqual(harness.committed.descendantLinks, []string{
			"old:sps_controller",
			"old:field_device",
			"new:sps_controller",
			"new:field_device",
		}) || !reflect.DeepEqual(harness.committed.history, []string{
		"project_control_cabinet:update",
		"project_sps_controller:create",
		"project_field_device:create",
	}) || !reflect.DeepEqual(harness.historyBatchIDs, []uuid.UUID{operationID}) {
		t.Fatalf("transaction state: %+v", harness)
	}
	if outcome.Link == nil || outcome.Link.ID != linkID ||
		outcome.Link.ControlCabinetID != newCabinetID ||
		outcome.Mutation.OperationID != operationID || outcome.Mutation.BatchID == nil ||
		*outcome.Mutation.BatchID != operationID || outcome.Mutation.ActorID == nil ||
		*outcome.Mutation.ActorID != actorID ||
		!outcome.Mutation.OccurredAt.Equal(occurredAt) ||
		!reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectID}) ||
		len(outcome.Mutation.Changes) != 1 {
		t.Fatalf("reassignment outcome: %+v", outcome)
	}
	change := outcome.Mutation.Changes[0]
	if change.EntityType != mutation.EntityTypeProjectControlCabinet ||
		change.EntityID != linkID || change.ParentID == nil || *change.ParentID != projectID ||
		change.Action != domainHistory.ActionUpdate ||
		!reflect.DeepEqual(change.ChangedFields, []mutation.FieldName{mutation.FieldNameControlCabinet}) {
		t.Fatalf("canonical change: %+v", change)
	}
	var before projectControlCabinetLinkSnapshot
	if err := json.Unmarshal(change.Before, &before); err != nil {
		t.Fatalf("decode before snapshot: %v", err)
	}
	var after projectControlCabinetLinkSnapshot
	if err := json.Unmarshal(change.After, &after); err != nil {
		t.Fatalf("decode after snapshot: %v", err)
	}
	if before.ControlCabinetID != previousCabinetID ||
		after.ControlCabinetID != newCabinetID ||
		!before.UpdatedAt.Equal(createdAt) || !after.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("canonical snapshots: before=%+v after=%+v", before, after)
	}
	if len(dispatcher.commands) != 1 || !dispatcher.afterCommitSeen {
		t.Fatalf("dispatch timing: %+v", dispatcher)
	}
	dispatched, ok := dispatcher.commands[0].(appcollaboration.FacilityHierarchyRefreshRequired)
	if !ok || dispatched.ProjectID != projectID || dispatched.EventID != eventID ||
		dispatched.OperationID != operationID || dispatched.CorrelationID != operationID ||
		dispatched.ActorID == nil || *dispatched.ActorID != actorID ||
		dispatched.Scope != appcollaboration.FacilityScopeControlCabinet ||
		dispatched.FullRefresh ||
		!reflect.DeepEqual(
			dispatched.EntityIDs,
			[]uuid.UUID{previousCabinetID, newCabinetID},
		) {
		t.Fatalf("typed collaboration command: %#v", dispatcher.commands[0])
	}
}

func TestReassignProjectControlCabinetLinkRollsBackAndDoesNotDispatch(t *testing.T) {
	projectID := uuid.New()
	linkID := uuid.New()
	previousCabinetID := uuid.New()
	newCabinetID := uuid.New()

	for _, testCase := range []struct {
		name        string
		workflowErr error
		commitErr   error
	}{
		{name: "descendant link or history", workflowErr: errors.New("link descendant failed")},
		{name: "commit", commitErr: errors.New("commit failed")},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			harness := &projectCabinetReassignmentHarness{
				committed: projectCabinetReassignmentState{
					link: &domainProject.ProjectControlCabinet{
						Base:             domain.Base{ID: linkID},
						ProjectID:        projectID,
						ControlCabinetID: previousCabinetID,
					},
				},
				workflowErr: testCase.workflowErr,
				commitErr:   testCase.commitErr,
			}
			dispatcher := &projectCabinetReassignmentDispatcherStub{
				harness:       harness,
				wantCabinetID: newCabinetID,
			}
			handler := NewReassignProjectLinkHandler(ReassignProjectLinkDependencies{
				TransactionRunner:   harness.runner,
				TransactionWorkflow: harness.factory,
				Dispatcher:          dispatcher,
			})

			outcome, err := handler.Execute(context.Background(), ReassignProjectLinkCommand{
				ProjectID:        projectID,
				LinkID:           linkID,
				ControlCabinetID: newCabinetID,
			})

			wantErr := testCase.workflowErr
			if wantErr == nil {
				wantErr = testCase.commitErr
			}
			if !errors.Is(err, wantErr) || outcome.Link != nil ||
				harness.committed.link == nil ||
				harness.committed.link.ControlCabinetID != previousCabinetID ||
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

func TestReassignProjectControlCabinetLinkRejectsForeignProject(t *testing.T) {
	harness := &projectCabinetReassignmentHarness{
		committed: projectCabinetReassignmentState{
			link: &domainProject.ProjectControlCabinet{
				Base:             domain.Base{ID: uuid.New()},
				ProjectID:        uuid.New(),
				ControlCabinetID: uuid.New(),
			},
		},
	}
	dispatcher := &projectCabinetReassignmentDispatcherStub{harness: harness}
	handler := NewReassignProjectLinkHandler(ReassignProjectLinkDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Dispatcher:          dispatcher,
	})

	outcome, err := handler.Execute(context.Background(), ReassignProjectLinkCommand{
		ProjectID:        uuid.New(),
		LinkID:           harness.committed.link.ID,
		ControlCabinetID: uuid.New(),
	})

	if !errors.Is(err, domain.ErrNotFound) || outcome.Link != nil ||
		harness.readCalls != 1 || harness.updateCalls != 0 || len(dispatcher.commands) != 0 {
		t.Fatalf("foreign project result: outcome=%+v err=%v harness=%+v", outcome, err, harness)
	}
}

func TestReassignProjectControlCabinetLinkReportsDispatchFailureAfterCommit(t *testing.T) {
	projectID := uuid.New()
	linkID := uuid.New()
	newCabinetID := uuid.New()
	dispatchErr := errors.New("realtime unavailable")
	harness := &projectCabinetReassignmentHarness{
		committed: projectCabinetReassignmentState{
			link: &domainProject.ProjectControlCabinet{
				Base:             domain.Base{ID: linkID},
				ProjectID:        projectID,
				ControlCabinetID: uuid.New(),
			},
			descendantLinks: []string{"old:sps_controller", "old:field_device"},
		},
	}
	dispatcher := &projectCabinetReassignmentDispatcherStub{
		harness:       harness,
		wantCabinetID: newCabinetID,
		err:           dispatchErr,
	}
	var reported []error
	handler := NewReassignProjectLinkHandler(ReassignProjectLinkDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Dispatcher:          dispatcher,
		ReportError:         func(err error) { reported = append(reported, err) },
	})

	updated, err := handler.ReassignProjectLink(context.Background(), ReassignProjectLinkCommand{
		ProjectID:        projectID,
		LinkID:           linkID,
		ControlCabinetID: newCabinetID,
	})

	if err != nil || updated == nil || updated.ControlCabinetID != newCabinetID ||
		harness.committed.link.ControlCabinetID != newCabinetID ||
		len(reported) != 1 || !errors.Is(reported[0], dispatchErr) {
		t.Fatalf("best-effort dispatch: updated=%+v err=%v reported=%v", updated, err, reported)
	}
}

func TestReassignProjectControlCabinetLinkMissingConfigurationFailsWithoutWrites(t *testing.T) {
	handler := NewReassignProjectLinkHandler(ReassignProjectLinkDependencies{})

	updated, err := handler.ReassignProjectLink(context.Background(), ReassignProjectLinkCommand{
		ProjectID:        uuid.New(),
		LinkID:           uuid.New(),
		ControlCabinetID: uuid.New(),
	})

	if updated != nil || !errors.Is(err, ErrReassignProjectLinkTransactionNotConfigured) {
		t.Fatalf("missing configuration: updated=%+v err=%v", updated, err)
	}
}
