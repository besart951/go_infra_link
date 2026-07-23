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
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type reassignProjectTransactionState struct {
	link    *domainProject.ProjectFieldDevice
	history []uuid.UUID
}

func (s reassignProjectTransactionState) clone() reassignProjectTransactionState {
	return reassignProjectTransactionState{
		link:    cloneProjectFieldDeviceLink(s.link),
		history: append([]uuid.UUID(nil), s.history...),
	}
}

type reassignProjectTransactionUnit struct {
	state *reassignProjectTransactionState
}

type reassignProjectTransactionHarness struct {
	committed       reassignProjectTransactionState
	updateErr       error
	commitErr       error
	updatedAt       time.Time
	runnerCalls     int
	getCalls        int
	updateCalls     int
	requestedIDs    []uuid.UUID
	historyBatchIDs []uuid.UUID
}

func (h *reassignProjectTransactionHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	staged := h.committed.clone()
	if err := run(ctx, reassignProjectTransactionUnit{state: &staged}); err != nil {
		return err
	}
	if h.commitErr != nil {
		return h.commitErr
	}
	h.committed = staged
	return nil
}

func (h *reassignProjectTransactionHarness) factory(
	unit apptransaction.UnitOfWork,
) (ReassignProjectLinkWorkflow, error) {
	typed, ok := unit.(reassignProjectTransactionUnit)
	if !ok || typed.state == nil {
		return nil, errors.New("unexpected ProjectFieldDevice reassignment transaction unit")
	}
	return &reassignProjectWorkflowStub{harness: h, state: typed.state}, nil
}

type reassignProjectWorkflowStub struct {
	harness *reassignProjectTransactionHarness
	state   *reassignProjectTransactionState
}

func (s *reassignProjectWorkflowStub) GetByIds(
	_ context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectFieldDevice, error) {
	s.harness.getCalls++
	s.harness.requestedIDs = append([]uuid.UUID(nil), ids...)
	if s.state.link == nil || len(ids) != 1 || s.state.link.ID != ids[0] {
		return nil, nil
	}
	return []*domainProject.ProjectFieldDevice{cloneProjectFieldDeviceLink(s.state.link)}, nil
}

func (s *reassignProjectWorkflowStub) Update(
	ctx context.Context,
	link *domainProject.ProjectFieldDevice,
) error {
	s.harness.updateCalls++
	if batchID, ok := ctx.Value(assignProjectHistoryBatchKey{}).(uuid.UUID); ok {
		s.harness.historyBatchIDs = append(s.harness.historyBatchIDs, batchID)
	}
	updated := cloneProjectFieldDeviceLink(link)
	if !s.harness.updatedAt.IsZero() {
		updated.UpdatedAt = s.harness.updatedAt
		link.UpdatedAt = s.harness.updatedAt
	}
	s.state.link = updated
	s.state.history = append(s.state.history, link.ID)
	return s.harness.updateErr
}

type reassignProjectDispatcherStub struct {
	harness           *reassignProjectTransactionHarness
	wantFieldDeviceID uuid.UUID
	commands          []appcollaboration.Command
	err               error
	calledAfterCommit bool
}

func (s *reassignProjectDispatcherStub) Dispatch(
	_ context.Context,
	command appcollaboration.Command,
) error {
	s.commands = append(s.commands, command)
	s.calledAfterCommit = s.harness.committed.link != nil &&
		s.harness.committed.link.FieldDeviceID == s.wantFieldDeviceID &&
		len(s.harness.committed.history) == 1
	return s.err
}

func TestReassignProjectLinkCommitsHistoryBeforeTypedRefresh(t *testing.T) {
	projectID := testUUID(1201)
	linkID := testUUID(1202)
	oldFieldDeviceID := testUUID(1203)
	newFieldDeviceID := testUUID(1204)
	actorID := testUUID(1205)
	operationID := testUUID(1206)
	eventID := testUUID(1207)
	createdAt := time.Date(2026, time.July, 21, 21, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Hour)
	occurredAt := updatedAt.Add(time.Minute)
	link := projectFieldDeviceLink(linkID, projectID, oldFieldDeviceID)
	link.CreatedAt = createdAt
	link.UpdatedAt = createdAt
	harness := &reassignProjectTransactionHarness{
		committed: reassignProjectTransactionState{link: link},
		updatedAt: updatedAt,
	}
	dispatcher := &reassignProjectDispatcherStub{
		harness:           harness,
		wantFieldDeviceID: newFieldDeviceID,
	}
	generatedIDs := []uuid.UUID{operationID, eventID}
	handler := NewReassignProjectLinkHandler(ReassignProjectLinkDependencies{
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

	outcome, err := handler.Execute(context.Background(), ReassignProjectLinkCommand{
		ProjectID:     projectID,
		LinkID:        linkID,
		FieldDeviceID: newFieldDeviceID,
	})

	if err != nil {
		t.Fatalf("reassign ProjectFieldDevice: %v", err)
	}
	if harness.runnerCalls != 1 || harness.getCalls != 1 || harness.updateCalls != 1 ||
		!reflect.DeepEqual(harness.requestedIDs, []uuid.UUID{linkID}) ||
		harness.committed.link == nil ||
		harness.committed.link.FieldDeviceID != newFieldDeviceID ||
		!reflect.DeepEqual(harness.committed.history, []uuid.UUID{linkID}) ||
		!reflect.DeepEqual(harness.historyBatchIDs, []uuid.UUID{operationID}) {
		t.Fatalf("transaction state: %+v", harness)
	}
	if outcome.Link == nil || outcome.Link.ID != linkID ||
		outcome.Link.FieldDeviceID != newFieldDeviceID ||
		!outcome.Link.UpdatedAt.Equal(updatedAt) ||
		outcome.Mutation.OperationID != operationID || outcome.Mutation.BatchID == nil ||
		*outcome.Mutation.BatchID != operationID || outcome.Mutation.ActorID == nil ||
		*outcome.Mutation.ActorID != actorID ||
		!outcome.Mutation.OccurredAt.Equal(occurredAt) ||
		!reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectID}) ||
		len(outcome.Mutation.Changes) != 1 {
		t.Fatalf("reassignment outcome: %+v", outcome)
	}
	change := outcome.Mutation.Changes[0]
	if change.EntityType != mutation.EntityTypeProjectFieldDevice ||
		change.EntityID != linkID || change.ParentID == nil || *change.ParentID != projectID ||
		change.Action != domainHistory.ActionUpdate ||
		!reflect.DeepEqual(change.ChangedFields, []mutation.FieldName{mutation.FieldNameFieldDevice}) {
		t.Fatalf("canonical change: %+v", change)
	}
	var before projectFieldDeviceLinkSnapshot
	if err := json.Unmarshal(change.Before, &before); err != nil {
		t.Fatalf("decode before snapshot: %v", err)
	}
	var after projectFieldDeviceLinkSnapshot
	if err := json.Unmarshal(change.After, &after); err != nil {
		t.Fatalf("decode after snapshot: %v", err)
	}
	if before.FieldDeviceID != oldFieldDeviceID || !before.UpdatedAt.Equal(createdAt) ||
		after.FieldDeviceID != newFieldDeviceID || !after.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("before/after snapshots: before=%+v after=%+v", before, after)
	}
	if len(dispatcher.commands) != 1 || !dispatcher.calledAfterCommit {
		t.Fatalf("dispatch timing: %+v", dispatcher)
	}
	command, ok := dispatcher.commands[0].(appcollaboration.FacilityHierarchyRefreshRequired)
	if !ok || command.ProjectID != projectID || command.EventID != eventID ||
		command.OperationID != operationID || command.CorrelationID != operationID ||
		command.ActorID == nil || *command.ActorID != actorID ||
		command.Scope != appcollaboration.FacilityScopeFieldDevice || command.FullRefresh ||
		!reflect.DeepEqual(command.EntityIDs, []uuid.UUID{newFieldDeviceID}) {
		t.Fatalf("typed collaboration command: %#v", dispatcher.commands[0])
	}
}

func TestReassignProjectLinkRejectsDifferentProjectWithoutWriting(t *testing.T) {
	linkedProjectID := testUUID(1211)
	requestedProjectID := testUUID(1212)
	linkID := testUUID(1213)
	oldFieldDeviceID := testUUID(1214)
	link := projectFieldDeviceLink(linkID, linkedProjectID, oldFieldDeviceID)
	harness := &reassignProjectTransactionHarness{
		committed: reassignProjectTransactionState{link: link},
	}
	dispatcher := &reassignProjectDispatcherStub{harness: harness}
	handler := NewReassignProjectLinkHandler(ReassignProjectLinkDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Dispatcher:          dispatcher,
	})

	outcome, err := handler.Execute(context.Background(), ReassignProjectLinkCommand{
		ProjectID:     requestedProjectID,
		LinkID:        linkID,
		FieldDeviceID: testUUID(1215),
	})

	if !errors.Is(err, domain.ErrNotFound) || outcome.Link != nil ||
		harness.updateCalls != 0 ||
		harness.committed.link.FieldDeviceID != oldFieldDeviceID ||
		len(harness.committed.history) != 0 || len(dispatcher.commands) != 0 {
		t.Fatalf("cross-project reassignment advanced: outcome=%+v err=%v harness=%+v commands=%v",
			outcome,
			err,
			harness,
			dispatcher.commands,
		)
	}
}

func TestReassignProjectLinkRollsBackAndDoesNotDispatchOnUpdateOrCommitFailure(t *testing.T) {
	projectID := testUUID(1221)
	linkID := testUUID(1222)
	oldFieldDeviceID := testUUID(1223)
	newFieldDeviceID := testUUID(1224)
	link := projectFieldDeviceLink(linkID, projectID, oldFieldDeviceID)

	for _, testCase := range []struct {
		name      string
		updateErr error
		commitErr error
	}{
		{name: "update or history", updateErr: errors.New("history write failed")},
		{name: "commit", commitErr: errors.New("commit failed")},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			harness := &reassignProjectTransactionHarness{
				committed: reassignProjectTransactionState{link: link},
				updateErr: testCase.updateErr,
				commitErr: testCase.commitErr,
			}
			dispatcher := &reassignProjectDispatcherStub{
				harness:           harness,
				wantFieldDeviceID: newFieldDeviceID,
			}
			handler := NewReassignProjectLinkHandler(ReassignProjectLinkDependencies{
				TransactionRunner:   harness.runner,
				TransactionWorkflow: harness.factory,
				Dispatcher:          dispatcher,
			})

			outcome, err := handler.Execute(context.Background(), ReassignProjectLinkCommand{
				ProjectID:     projectID,
				LinkID:        linkID,
				FieldDeviceID: newFieldDeviceID,
			})

			wantErr := testCase.updateErr
			if wantErr == nil {
				wantErr = testCase.commitErr
			}
			if !errors.Is(err, wantErr) || outcome.Link != nil ||
				harness.committed.link.FieldDeviceID != oldFieldDeviceID ||
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

func TestReassignProjectLinkReportsDispatchFailureWithoutChangingCommittedResult(t *testing.T) {
	projectID := testUUID(1231)
	linkID := testUUID(1232)
	oldFieldDeviceID := testUUID(1233)
	newFieldDeviceID := testUUID(1234)
	dispatchErr := errors.New("realtime unavailable")
	harness := &reassignProjectTransactionHarness{
		committed: reassignProjectTransactionState{
			link: projectFieldDeviceLink(linkID, projectID, oldFieldDeviceID),
		},
	}
	dispatcher := &reassignProjectDispatcherStub{
		harness:           harness,
		wantFieldDeviceID: newFieldDeviceID,
		err:               dispatchErr,
	}
	var reported []error
	handler := NewReassignProjectLinkHandler(ReassignProjectLinkDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Dispatcher:          dispatcher,
		ReportError:         func(err error) { reported = append(reported, err) },
	})

	updated, err := handler.ReassignProjectLink(context.Background(), ReassignProjectLinkCommand{
		ProjectID:     projectID,
		LinkID:        linkID,
		FieldDeviceID: newFieldDeviceID,
	})

	if err != nil || updated == nil || updated.FieldDeviceID != newFieldDeviceID ||
		harness.committed.link.FieldDeviceID != newFieldDeviceID ||
		len(reported) != 1 || !errors.Is(reported[0], dispatchErr) {
		t.Fatalf("best-effort dispatch: updated=%+v err=%v reported=%v", updated, err, reported)
	}
}

func TestReassignProjectLinkMissingConfigurationFailsWithoutWrites(t *testing.T) {
	handler := NewReassignProjectLinkHandler(ReassignProjectLinkDependencies{})

	updated, err := handler.ReassignProjectLink(context.Background(), ReassignProjectLinkCommand{
		ProjectID:     testUUID(1241),
		LinkID:        testUUID(1242),
		FieldDeviceID: testUUID(1243),
	})

	if updated != nil || !errors.Is(err, ErrReassignProjectLinkTransactionNotConfigured) {
		t.Fatalf("missing configuration: updated=%+v err=%v", updated, err)
	}
}
