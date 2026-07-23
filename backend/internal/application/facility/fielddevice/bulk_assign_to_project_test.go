package fielddevice

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type bulkAssignTransactionState struct {
	links   map[uuid.UUID]*domainProject.ProjectFieldDevice
	history []uuid.UUID
}

func (s bulkAssignTransactionState) clone() bulkAssignTransactionState {
	links := make(map[uuid.UUID]*domainProject.ProjectFieldDevice, len(s.links))
	for fieldDeviceID, link := range s.links {
		links[fieldDeviceID] = cloneProjectFieldDeviceLink(link)
	}
	return bulkAssignTransactionState{
		links:   links,
		history: append([]uuid.UUID(nil), s.history...),
	}
}

type bulkAssignTransactionUnit struct {
	state *bulkAssignTransactionState
}

type bulkAssignTransactionHarness struct {
	committed       bulkAssignTransactionState
	candidates      map[uuid.UUID]*domainProject.ProjectFieldDevice
	workflowErrs    map[uuid.UUID]error
	commitErrs      map[uuid.UUID]error
	runnerCalls     int
	workflowCalls   []uuid.UUID
	historyBatchIDs []uuid.UUID
	lastAttempt     uuid.UUID
}

func (h *bulkAssignTransactionHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	staged := h.committed.clone()
	h.lastAttempt = uuid.Nil
	if err := run(ctx, bulkAssignTransactionUnit{state: &staged}); err != nil {
		return err
	}
	if err := h.commitErrs[h.lastAttempt]; err != nil {
		return err
	}
	h.committed = staged
	return nil
}

func (h *bulkAssignTransactionHarness) factory(
	unit apptransaction.UnitOfWork,
) (AssignToProjectWorkflow, error) {
	typed, ok := unit.(bulkAssignTransactionUnit)
	if !ok || typed.state == nil {
		return nil, errors.New("unexpected bulk-assignment transaction unit")
	}
	return &bulkAssignWorkflowStub{harness: h, state: typed.state}, nil
}

type bulkAssignWorkflowStub struct {
	harness *bulkAssignTransactionHarness
	state   *bulkAssignTransactionState
}

func (s *bulkAssignWorkflowStub) CreateFieldDevice(
	ctx context.Context,
	projectID uuid.UUID,
	fieldDeviceID uuid.UUID,
) (*domainProject.ProjectFieldDevice, error) {
	s.harness.workflowCalls = append(s.harness.workflowCalls, fieldDeviceID)
	s.harness.lastAttempt = fieldDeviceID
	if batchID, ok := ctx.Value(assignProjectHistoryBatchKey{}).(uuid.UUID); ok {
		s.harness.historyBatchIDs = append(s.harness.historyBatchIDs, batchID)
	}
	if s.state.links[fieldDeviceID] != nil {
		return nil, domain.ErrConflict
	}
	link := cloneProjectFieldDeviceLink(s.harness.candidates[fieldDeviceID])
	if link != nil {
		link.ProjectID = projectID
		s.state.links[fieldDeviceID] = link
		s.state.history = append(s.state.history, link.ID)
	}
	if err := s.harness.workflowErrs[fieldDeviceID]; err != nil {
		return nil, err
	}
	return link, nil
}

type bulkAssignProjectReaderStub struct {
	harness                 *bulkAssignTransactionHarness
	project                 *domainProject.Project
	err                     error
	calls                   int
	ids                     []uuid.UUID
	calledBeforeTransaction bool
}

func (s *bulkAssignProjectReaderStub) GetByIds(
	_ context.Context,
	ids []uuid.UUID,
) ([]*domainProject.Project, error) {
	s.calls++
	s.ids = append([]uuid.UUID(nil), ids...)
	s.calledBeforeTransaction = s.harness.runnerCalls == 0
	if s.err != nil || s.project == nil {
		return nil, s.err
	}
	return []*domainProject.Project{s.project}, nil
}

type bulkAssignDispatcherStub struct {
	harness           *bulkAssignTransactionHarness
	wantCommitted     []uuid.UUID
	wantAbsent        []uuid.UUID
	commands          []appcollaboration.Command
	err               error
	calledAfterCommit bool
}

func (s *bulkAssignDispatcherStub) Dispatch(
	_ context.Context,
	command appcollaboration.Command,
) error {
	s.commands = append(s.commands, command)
	committed := true
	for _, id := range s.wantCommitted {
		committed = committed && s.harness.committed.links[id] != nil
	}
	for _, id := range s.wantAbsent {
		committed = committed && s.harness.committed.links[id] == nil
	}
	s.calledAfterCommit = committed
	return s.err
}

func TestBulkAssignToProjectPreservesPartialResultAndDispatchesAfterAllItemCommits(t *testing.T) {
	projectID := testUUID(801)
	fieldDeviceA := testUUID(802)
	fieldDeviceB := testUUID(803)
	fieldDeviceC := testUUID(804)
	linkA := testUUID(805)
	linkB := testUUID(806)
	linkC := testUUID(807)
	actorID := testUUID(808)
	operationID := testUUID(809)
	eventID := testUUID(810)
	workflowErr := errors.New("history failed")
	commitErr := errors.New("commit failed")
	occurredAt := time.Date(2026, time.July, 21, 20, 0, 0, 0, time.UTC)
	harness := &bulkAssignTransactionHarness{
		committed: bulkAssignTransactionState{links: map[uuid.UUID]*domainProject.ProjectFieldDevice{}},
		candidates: map[uuid.UUID]*domainProject.ProjectFieldDevice{
			fieldDeviceA: projectFieldDeviceLink(linkA, projectID, fieldDeviceA),
			fieldDeviceB: projectFieldDeviceLink(linkB, projectID, fieldDeviceB),
			fieldDeviceC: projectFieldDeviceLink(linkC, projectID, fieldDeviceC),
		},
		workflowErrs: map[uuid.UUID]error{fieldDeviceB: workflowErr},
		commitErrs:   map[uuid.UUID]error{fieldDeviceC: commitErr},
	}
	project := &domainProject.Project{}
	project.ID = projectID
	projects := &bulkAssignProjectReaderStub{harness: harness, project: project}
	dispatcher := &bulkAssignDispatcherStub{
		harness:       harness,
		wantCommitted: []uuid.UUID{fieldDeviceA},
		wantAbsent:    []uuid.UUID{fieldDeviceB, fieldDeviceC},
	}
	generatedIDs := []uuid.UUID{operationID, eventID}
	handler := NewBulkAssignToProjectHandler(BulkAssignToProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Projects:            projects,
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
	requestIDs := []uuid.UUID{fieldDeviceA, fieldDeviceB, fieldDeviceC, fieldDeviceA}

	outcome := handler.Execute(context.Background(), BulkAssignToProjectCommand{
		ProjectID:      projectID,
		FieldDeviceIDs: requestIDs,
	})

	if projects.calls != 1 || !projects.calledBeforeTransaction ||
		!reflect.DeepEqual(projects.ids, []uuid.UUID{projectID}) ||
		harness.runnerCalls != len(requestIDs) ||
		!reflect.DeepEqual(harness.workflowCalls, requestIDs) {
		t.Fatalf("precheck/item transactions: projects=%+v harness=%+v", projects, harness)
	}
	if !reflect.DeepEqual(outcome.Result.SuccessFieldDeviceIDs, []uuid.UUID{fieldDeviceA}) ||
		!reflect.DeepEqual(outcome.Result.AssociationErrors, []string{
			workflowErr.Error(),
			commitErr.Error(),
			domain.ErrConflict.Error(),
		}) {
		t.Fatalf("partial compatibility result: %+v", outcome.Result)
	}
	if len(harness.committed.links) != 1 || harness.committed.links[fieldDeviceA] == nil ||
		!reflect.DeepEqual(harness.committed.history, []uuid.UUID{linkA}) {
		t.Fatalf("committed partial state: %+v", harness.committed)
	}
	if len(harness.historyBatchIDs) != len(requestIDs) {
		t.Fatalf("history BatchIDs: %v", harness.historyBatchIDs)
	}
	for _, batchID := range harness.historyBatchIDs {
		if batchID != operationID {
			t.Fatalf("history BatchID: got %s want %s", batchID, operationID)
		}
	}
	if outcome.Mutation.OperationID != operationID || outcome.Mutation.BatchID == nil ||
		*outcome.Mutation.BatchID != operationID || outcome.Mutation.ActorID == nil ||
		*outcome.Mutation.ActorID != actorID ||
		!reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectID}) ||
		len(outcome.Mutation.Changes) != 1 || outcome.Mutation.Changes[0].EntityID != linkA ||
		!reflect.DeepEqual(outcome.ReconciliationIDs, []uuid.UUID{fieldDeviceA}) {
		t.Fatalf("canonical outcome: %+v", outcome)
	}
	if len(dispatcher.commands) != 1 || !dispatcher.calledAfterCommit {
		t.Fatalf("dispatch timing: %+v", dispatcher)
	}
	command, ok := dispatcher.commands[0].(appcollaboration.FacilityHierarchyRefreshRequired)
	if !ok || command.ProjectID != projectID || command.EventID != eventID ||
		command.OperationID != operationID || command.Scope != appcollaboration.FacilityScopeFieldDevice ||
		command.FullRefresh || !reflect.DeepEqual(command.EntityIDs, []uuid.UUID{fieldDeviceA}) {
		t.Fatalf("collaboration command: %#v", dispatcher.commands[0])
	}
}

func TestBulkAssignToProjectMissingProjectPreservesSingleLegacyAssociationError(t *testing.T) {
	projectID := testUUID(821)
	harness := &bulkAssignTransactionHarness{
		committed: bulkAssignTransactionState{links: map[uuid.UUID]*domainProject.ProjectFieldDevice{}},
	}
	projects := &bulkAssignProjectReaderStub{harness: harness}
	dispatcher := &bulkAssignDispatcherStub{harness: harness}
	handler := NewBulkAssignToProjectHandler(BulkAssignToProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Projects:            projects,
		Dispatcher:          dispatcher,
	})

	outcome := handler.Execute(context.Background(), BulkAssignToProjectCommand{
		ProjectID:      projectID,
		FieldDeviceIDs: []uuid.UUID{testUUID(822), testUUID(823)},
	})

	if len(outcome.Result.SuccessFieldDeviceIDs) != 0 ||
		!reflect.DeepEqual(outcome.Result.AssociationErrors, []string{projectNotFoundAssociationError}) ||
		harness.runnerCalls != 0 || len(dispatcher.commands) != 0 {
		t.Fatalf("missing-project result: outcome=%+v harness=%+v", outcome, harness)
	}
}

func TestBulkAssignToProjectLargeSuccessSetFallsBackToOneFullRefresh(t *testing.T) {
	projectID := testUUID(831)
	ids := make([]uuid.UUID, defaultMaxTargetedRefreshIDs+1)
	candidates := make(map[uuid.UUID]*domainProject.ProjectFieldDevice, len(ids))
	for index := range ids {
		ids[index] = testUUID(900 + index)
		candidates[ids[index]] = projectFieldDeviceLink(
			testUUID(1100+index),
			projectID,
			ids[index],
		)
	}
	harness := &bulkAssignTransactionHarness{
		committed:  bulkAssignTransactionState{links: map[uuid.UUID]*domainProject.ProjectFieldDevice{}},
		candidates: candidates,
	}
	project := &domainProject.Project{}
	project.ID = projectID
	projects := &bulkAssignProjectReaderStub{harness: harness, project: project}
	dispatcher := &bulkAssignDispatcherStub{harness: harness, wantCommitted: ids}
	handler := NewBulkAssignToProjectHandler(BulkAssignToProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Projects:            projects,
		Dispatcher:          dispatcher,
	})

	outcome := handler.Execute(context.Background(), BulkAssignToProjectCommand{
		ProjectID:      projectID,
		FieldDeviceIDs: ids,
	})

	if len(outcome.Result.SuccessFieldDeviceIDs) != len(ids) || len(dispatcher.commands) != 1 {
		t.Fatalf("large assignment result: successes=%d commands=%d",
			len(outcome.Result.SuccessFieldDeviceIDs),
			len(dispatcher.commands),
		)
	}
	command, ok := dispatcher.commands[0].(appcollaboration.FacilityHierarchyRefreshRequired)
	if !ok || !command.FullRefresh || len(command.EntityIDs) != 0 ||
		command.ProjectID != projectID || command.Scope != appcollaboration.FacilityScopeFieldDevice {
		t.Fatalf("full refresh fallback: %#v", dispatcher.commands[0])
	}
}

func TestBulkAssignToProjectReportsDispatchFailureWithoutChangingPartialResult(t *testing.T) {
	projectID := testUUID(835)
	fieldDeviceID := testUUID(836)
	harness := &bulkAssignTransactionHarness{
		committed: bulkAssignTransactionState{links: map[uuid.UUID]*domainProject.ProjectFieldDevice{}},
		candidates: map[uuid.UUID]*domainProject.ProjectFieldDevice{
			fieldDeviceID: projectFieldDeviceLink(testUUID(837), projectID, fieldDeviceID),
		},
	}
	project := &domainProject.Project{}
	project.ID = projectID
	dispatchErr := errors.New("realtime unavailable")
	dispatcher := &bulkAssignDispatcherStub{harness: harness, err: dispatchErr}
	var reported []error
	handler := NewBulkAssignToProjectHandler(BulkAssignToProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Projects:            &bulkAssignProjectReaderStub{harness: harness, project: project},
		Dispatcher:          dispatcher,
		ReportError:         func(err error) { reported = append(reported, err) },
	})

	result := handler.BulkAssignToProject(context.Background(), BulkAssignToProjectCommand{
		ProjectID:      projectID,
		FieldDeviceIDs: []uuid.UUID{fieldDeviceID},
	})

	if !reflect.DeepEqual(result.SuccessFieldDeviceIDs, []uuid.UUID{fieldDeviceID}) ||
		len(result.AssociationErrors) != 0 || harness.committed.links[fieldDeviceID] == nil ||
		len(reported) != 1 || !errors.Is(reported[0], dispatchErr) {
		t.Fatalf("best-effort dispatch: result=%+v state=%+v reported=%v",
			result,
			harness.committed,
			reported,
		)
	}
}

func TestBulkAssignToProjectMissingConfigurationReturnsCompatibilityError(t *testing.T) {
	handler := NewBulkAssignToProjectHandler(BulkAssignToProjectDependencies{})

	result := handler.BulkAssignToProject(context.Background(), BulkAssignToProjectCommand{
		ProjectID:      testUUID(841),
		FieldDeviceIDs: []uuid.UUID{testUUID(842)},
	})

	if len(result.SuccessFieldDeviceIDs) != 0 || len(result.AssociationErrors) != 1 {
		t.Fatalf("missing configuration result: %+v", result)
	}
}

func projectFieldDeviceLink(
	linkID uuid.UUID,
	projectID uuid.UUID,
	fieldDeviceID uuid.UUID,
) *domainProject.ProjectFieldDevice {
	link := &domainProject.ProjectFieldDevice{
		ProjectID:     projectID,
		FieldDeviceID: fieldDeviceID,
	}
	link.ID = linkID
	return link
}
