package controlcabinet

import (
	"context"
	"errors"
	"testing"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainCollaboration "github.com/besart951/go_infra_link/backend/internal/domain/collaboration"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type projectRestoreScopeStub struct {
	err              error
	actorID          uuid.UUID
	projectID        uuid.UUID
	controlCabinetID uuid.UUID
	calls            int
	order            *[]string
}

func (s *projectRestoreScopeStub) RequireControlCabinetRestoreScope(
	_ context.Context,
	actorID uuid.UUID,
	projectID uuid.UUID,
	controlCabinetID uuid.UUID,
) error {
	s.calls++
	s.actorID = actorID
	s.projectID = projectID
	s.controlCabinetID = controlCabinetID
	if s.order != nil {
		*s.order = append(*s.order, "scope")
	}
	return s.err
}

type projectHistoryRestorerStub struct {
	result           *domainHistory.RestoreResult
	err              error
	controlCabinetID uuid.UUID
	request          domainHistory.RestoreControlCabinetRequest
	calls            int
	order            *[]string
}

func (s *projectHistoryRestorerStub) RestoreControlCabinet(
	_ context.Context,
	controlCabinetID uuid.UUID,
	request domainHistory.RestoreControlCabinetRequest,
) (*domainHistory.RestoreResult, error) {
	s.calls++
	s.controlCabinetID = controlCabinetID
	s.request = request
	if s.order != nil {
		*s.order = append(*s.order, "restore_committed")
	}
	return s.result, s.err
}

type projectRestoreLinkReaderStub struct {
	links []*domainProject.ProjectControlCabinet
	err   error
	calls int
}

func (s *projectRestoreLinkReaderStub) GetByControlCabinetIDs(
	_ context.Context,
	_ []uuid.UUID,
) ([]*domainProject.ProjectControlCabinet, error) {
	s.calls++
	return s.links, s.err
}

type projectRestoreDispatcherStub struct {
	commands []appcollaboration.Command
	err      error
	order    *[]string
}

type transactionalProjectRestoreWorkflowStub struct {
	*projectHistoryRestorerStub
	*projectRestoreLinkReaderStub
}

func TestTransactionalProjectRestoreWritesVersionTwoOutboxBeforeReturn(t *testing.T) {
	projectID := cabinetTestUUID(951)
	otherProjectID := cabinetTestUUID(952)
	controlCabinetID := cabinetTestUUID(953)
	batchID := cabinetTestUUID(954)
	eventIDs := []uuid.UUID{cabinetTestUUID(955), cabinetTestUUID(956)}
	workflow := &transactionalProjectRestoreWorkflowStub{
		projectHistoryRestorerStub: &projectHistoryRestorerStub{
			result: &domainHistory.RestoreResult{RestoredCount: 4, BatchID: batchID},
		},
		projectRestoreLinkReaderStub: &projectRestoreLinkReaderStub{links: []*domainProject.ProjectControlCabinet{
			{ProjectID: projectID, ControlCabinetID: controlCabinetID},
			{ProjectID: otherProjectID, ControlCabinetID: controlCabinetID},
		}},
	}
	store := &updateOutboxStoreStub{}
	ctx := domainCollaboration.WithOutboxStore(context.Background(), store)

	committed, err := executeTransactionalProjectRestore(
		ctx,
		workflow,
		RestoreForProjectCommand{
			ProjectID: projectID, ControlCabinetID: controlCabinetID,
		},
		nil,
		time.Date(2026, time.July, 23, 17, 0, 0, 0, time.UTC),
		func() uuid.UUID {
			id := eventIDs[0]
			eventIDs = eventIDs[1:]
			return id
		},
	)
	if err != nil {
		t.Fatalf("transactional restore: %v", err)
	}
	if committed.restore != workflow.result || committed.operationID != batchID {
		t.Fatalf("committed restore: %+v", committed)
	}
	assertProjectIDSet(t, committed.projectIDs, projectID, otherProjectID)
	if len(store.events) != 2 {
		t.Fatalf("outbox events: got %d, want 2", len(store.events))
	}
	for _, event := range store.events {
		decoded, err := appcollaboration.DecodeCommand(appcollaboration.EncodedCommand{
			Type: event.EventType, Payload: event.Payload,
		})
		if err != nil {
			t.Fatalf("decode outbox event: %v", err)
		}
		refresh, ok := decoded.(appcollaboration.FacilityHierarchyRefreshRequired)
		if !ok || refresh.SchemaVersion != appcollaboration.SchemaVersionV2 ||
			refresh.Scope != appcollaboration.FacilityScopeProject || !refresh.FullRefresh {
			t.Fatalf("unexpected durable refresh: %#v", decoded)
		}
	}
}

func (s *projectRestoreDispatcherStub) Dispatch(
	_ context.Context,
	command appcollaboration.Command,
) error {
	if s.order != nil {
		*s.order = append(*s.order, "dispatch")
	}
	s.commands = append(s.commands, command)
	return s.err
}

func TestRestoreForProjectValidatesScopeThenDispatchesAllAffectedProjectsAfterCommit(
	t *testing.T,
) {
	actorID := uuid.New()
	projectID := uuid.New()
	otherProjectID := uuid.New()
	controlCabinetID := uuid.New()
	batchID := uuid.New()
	eventID := uuid.New()
	asOf := time.Date(2026, time.July, 22, 11, 30, 0, 0, time.UTC)
	occurredAt := asOf.Add(time.Minute)
	order := []string{}
	scope := &projectRestoreScopeStub{order: &order}
	restorer := &projectHistoryRestorerStub{
		result: &domainHistory.RestoreResult{
			RestoredCount: 4,
			DeletedCount:  1,
			SkippedCount:  2,
			BatchID:       batchID,
		},
		order: &order,
	}
	links := &projectRestoreLinkReaderStub{links: []*domainProject.ProjectControlCabinet{
		{ProjectID: otherProjectID, ControlCabinetID: controlCabinetID},
		{ProjectID: projectID, ControlCabinetID: controlCabinetID},
		{ProjectID: otherProjectID, ControlCabinetID: controlCabinetID},
	}}
	dispatcher := &projectRestoreDispatcherStub{order: &order}
	eventIDs := []uuid.UUID{uuid.New(), uuid.New()}
	nextID := 0
	handler := NewRestoreForProjectHandler(RestoreForProjectDependencies{
		Scope:        scope,
		Restorer:     restorer,
		ProjectLinks: links,
		Dispatcher:   dispatcher,
		Actor: func(context.Context) *uuid.UUID {
			return &actorID
		},
		NewID: func() uuid.UUID {
			id := eventIDs[nextID]
			nextID++
			return id
		},
		Now: func() time.Time { return occurredAt },
	})

	outcome, err := handler.Execute(context.Background(), RestoreForProjectCommand{
		ProjectID:        projectID,
		ControlCabinetID: controlCabinetID,
		AsOf:             &asOf,
		EventID:          &eventID,
	})
	if err != nil {
		t.Fatalf("Execute returned error: %v", err)
	}
	if scope.calls != 1 || scope.actorID != actorID || scope.projectID != projectID ||
		scope.controlCabinetID != controlCabinetID {
		t.Fatalf("scope call: %+v", scope)
	}
	if restorer.calls != 1 || restorer.controlCabinetID != controlCabinetID {
		t.Fatalf("restore call: %+v", restorer)
	}
	if restorer.request.ProjectID == nil || *restorer.request.ProjectID != projectID ||
		restorer.request.AsOf != &asOf || restorer.request.EventID != &eventID {
		t.Fatalf("restore request: %+v", restorer.request)
	}
	if len(order) != 4 || order[0] != "scope" || order[1] != "restore_committed" ||
		order[2] != "dispatch" || order[3] != "dispatch" {
		t.Fatalf("execution order: got %v", order)
	}
	if outcome.Restore != restorer.result || outcome.Mutation.OperationID != batchID ||
		outcome.Mutation.BatchID == nil || *outcome.Mutation.BatchID != batchID ||
		outcome.Mutation.ActorID == nil || *outcome.Mutation.ActorID != actorID ||
		!outcome.Mutation.OccurredAt.Equal(occurredAt) {
		t.Fatalf("mutation result: %+v", outcome.Mutation)
	}
	assertProjectIDSet(t, outcome.Mutation.ProjectIDs, projectID, otherProjectID)
	if len(outcome.Mutation.Changes) != 1 ||
		outcome.Mutation.Changes[0].EntityID != controlCabinetID ||
		outcome.Mutation.Changes[0].Action != domainHistory.ActionRestore {
		t.Fatalf("canonical restore marker: %+v", outcome.Mutation.Changes)
	}
	if len(dispatcher.commands) != 2 {
		t.Fatalf("commands: got %d, want 2", len(dispatcher.commands))
	}
	commandProjects := make([]uuid.UUID, 0, len(dispatcher.commands))
	for index, raw := range dispatcher.commands {
		refresh, ok := raw.(appcollaboration.FacilityHierarchyRefreshRequired)
		if !ok {
			t.Fatalf("command %d: got %T", index, raw)
		}
		if refresh.Scope != appcollaboration.FacilityScopeProject || !refresh.FullRefresh ||
			refresh.OperationID != batchID || refresh.CorrelationID != batchID ||
			refresh.ActorID == nil || *refresh.ActorID != actorID ||
			!refresh.OccurredAt.Equal(occurredAt) {
			t.Fatalf("refresh %d: %+v", index, refresh)
		}
		commandProjects = append(commandProjects, refresh.ProjectID)
	}
	assertProjectIDSet(t, commandProjects, projectID, otherProjectID)
}

func TestRestoreForProjectRejectsMissingActorOrScopeBeforeRestore(t *testing.T) {
	command := RestoreForProjectCommand{
		ProjectID:        uuid.New(),
		ControlCabinetID: uuid.New(),
	}

	for _, test := range []struct {
		name  string
		actor ActorProvider
		scope *projectRestoreScopeStub
		want  error
	}{
		{
			name:  "missing actor",
			scope: &projectRestoreScopeStub{},
			want:  ErrProjectRestoreAccessDenied,
		},
		{
			name: "scope denied",
			actor: func(context.Context) *uuid.UUID {
				actorID := uuid.New()
				return &actorID
			},
			scope: &projectRestoreScopeStub{err: ErrProjectRestoreAccessDenied},
			want:  ErrProjectRestoreAccessDenied,
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			restorer := &projectHistoryRestorerStub{}
			dispatcher := &projectRestoreDispatcherStub{}
			handler := NewRestoreForProjectHandler(RestoreForProjectDependencies{
				Scope:      test.scope,
				Restorer:   restorer,
				Dispatcher: dispatcher,
				Actor:      test.actor,
			})

			_, err := handler.Execute(context.Background(), command)
			if !errors.Is(err, test.want) {
				t.Fatalf("error: got %v, want %v", err, test.want)
			}
			if restorer.calls != 0 || len(dispatcher.commands) != 0 {
				t.Fatalf("restore or dispatch occurred after denial")
			}
		})
	}
}

func TestRestoreForProjectDoesNotDispatchAfterRestoreFailureOrNoChange(t *testing.T) {
	actorID := uuid.New()
	command := RestoreForProjectCommand{
		ProjectID:        uuid.New(),
		ControlCabinetID: uuid.New(),
	}
	restoreErr := errors.New("restore rolled back")

	for _, test := range []struct {
		name       string
		restorer   *projectHistoryRestorerStub
		wantErr    error
		wantChange int
	}{
		{
			name:     "restore failed",
			restorer: &projectHistoryRestorerStub{err: restoreErr},
			wantErr:  restoreErr,
		},
		{
			name: "all targets skipped",
			restorer: &projectHistoryRestorerStub{result: &domainHistory.RestoreResult{
				SkippedCount: 3,
				BatchID:      uuid.New(),
			}},
		},
	} {
		t.Run(test.name, func(t *testing.T) {
			dispatcher := &projectRestoreDispatcherStub{}
			handler := NewRestoreForProjectHandler(RestoreForProjectDependencies{
				Scope:      &projectRestoreScopeStub{},
				Restorer:   test.restorer,
				Dispatcher: dispatcher,
				Actor:      func(context.Context) *uuid.UUID { return &actorID },
			})

			outcome, err := handler.Execute(context.Background(), command)
			if !errors.Is(err, test.wantErr) {
				t.Fatalf("error: got %v, want %v", err, test.wantErr)
			}
			if len(dispatcher.commands) != 0 {
				t.Fatalf("commands after failed/no-op restore: %+v", dispatcher.commands)
			}
			if err == nil && len(outcome.Mutation.Changes) != test.wantChange {
				t.Fatalf("changes: got %+v", outcome.Mutation.Changes)
			}
		})
	}
}

func TestRestoreForProjectKeepsCommittedResultWhenScopeOrDispatchDeliveryFails(
	t *testing.T,
) {
	actorID := uuid.New()
	projectID := uuid.New()
	batchID := uuid.New()
	linkErr := errors.New("link lookup failed")
	dispatchErr := errors.New("transport unavailable")
	dispatcher := &projectRestoreDispatcherStub{err: dispatchErr}
	reported := []error{}
	handler := NewRestoreForProjectHandler(RestoreForProjectDependencies{
		Scope: &projectRestoreScopeStub{},
		Restorer: &projectHistoryRestorerStub{result: &domainHistory.RestoreResult{
			RestoredCount: 1,
			BatchID:       batchID,
		}},
		ProjectLinks: &projectRestoreLinkReaderStub{err: linkErr},
		Dispatcher:   dispatcher,
		Actor:        func(context.Context) *uuid.UUID { return &actorID },
		ReportError: func(err error) {
			reported = append(reported, err)
		},
	})

	result, err := handler.RestoreForProject(context.Background(), RestoreForProjectCommand{
		ProjectID:        projectID,
		ControlCabinetID: uuid.New(),
	})
	if err != nil {
		t.Fatalf("RestoreForProject returned transport error: %v", err)
	}
	if result == nil || result.BatchID != batchID {
		t.Fatalf("committed result: %+v", result)
	}
	if len(dispatcher.commands) != 1 {
		t.Fatalf("target project must still receive one attempted refresh")
	}
	if len(reported) != 2 || !errors.Is(reported[0], linkErr) ||
		!errors.Is(reported[1], dispatchErr) {
		t.Fatalf("reported errors: %+v", reported)
	}
}

func TestRestoreForProjectValidatesCommandAndConfiguration(t *testing.T) {
	handler := NewRestoreForProjectHandler(RestoreForProjectDependencies{})
	if _, err := handler.Execute(context.Background(), RestoreForProjectCommand{}); !errors.Is(
		err,
		ErrProjectRestoreNotConfigured,
	) {
		t.Fatalf("configuration error: got %v", err)
	}

	actorID := uuid.New()
	handler = NewRestoreForProjectHandler(RestoreForProjectDependencies{
		Scope:    &projectRestoreScopeStub{},
		Restorer: &projectHistoryRestorerStub{},
		Actor:    func(context.Context) *uuid.UUID { return &actorID },
	})
	if _, err := handler.Execute(context.Background(), RestoreForProjectCommand{}); !errors.Is(
		err,
		domain.ErrInvalidArgument,
	) {
		t.Fatalf("validation error: got %v", err)
	}
}

func assertProjectIDSet(t *testing.T, got []uuid.UUID, want ...uuid.UUID) {
	t.Helper()
	if len(got) != len(want) {
		t.Fatalf("project IDs: got %v, want %v", got, want)
	}
	wantSet := make(map[uuid.UUID]struct{}, len(want))
	for _, id := range want {
		wantSet[id] = struct{}{}
	}
	for _, id := range got {
		if _, ok := wantSet[id]; !ok {
			t.Fatalf("unexpected project ID %s in %v", id, got)
		}
	}
}
