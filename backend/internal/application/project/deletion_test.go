package project

import (
	"context"
	"errors"
	"reflect"
	"sort"
	"testing"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainCollaboration "github.com/besart951/go_infra_link/backend/internal/domain/collaboration"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type deletionBatchKey struct{}

type deletionState struct {
	projects          map[uuid.UUID]Snapshot
	roles             map[uuid.UUID]domainUser.Role
	hierarchyLinks    map[uuid.UUID]bool
	objectData        map[uuid.UUID]uuid.UUID
	memberships       map[uuid.UUID][]uuid.UUID
	objectDeleteCalls [][]uuid.UUID
	projectDeletes    []uuid.UUID
	batchIDs          []uuid.UUID
	events            []domainCollaboration.OutboxEvent
}

func (state deletionState) clone() deletionState {
	cloned := deletionState{
		projects:          make(map[uuid.UUID]Snapshot, len(state.projects)),
		roles:             make(map[uuid.UUID]domainUser.Role, len(state.roles)),
		hierarchyLinks:    make(map[uuid.UUID]bool, len(state.hierarchyLinks)),
		objectData:        make(map[uuid.UUID]uuid.UUID, len(state.objectData)),
		memberships:       make(map[uuid.UUID][]uuid.UUID, len(state.memberships)),
		objectDeleteCalls: make([][]uuid.UUID, len(state.objectDeleteCalls)),
		projectDeletes:    append([]uuid.UUID(nil), state.projectDeletes...),
		batchIDs:          append([]uuid.UUID(nil), state.batchIDs...),
		events:            append([]domainCollaboration.OutboxEvent(nil), state.events...),
	}
	for id, project := range state.projects {
		cloned.projects[id] = project
	}
	for id, role := range state.roles {
		cloned.roles[id] = role
	}
	for id, linked := range state.hierarchyLinks {
		cloned.hierarchyLinks[id] = linked
	}
	for id, projectID := range state.objectData {
		cloned.objectData[id] = projectID
	}
	for projectID, members := range state.memberships {
		cloned.memberships[projectID] = append([]uuid.UUID(nil), members...)
	}
	for i, ids := range state.objectDeleteCalls {
		cloned.objectDeleteCalls[i] = append([]uuid.UUID(nil), ids...)
	}
	return cloned
}

type deletionHarness struct {
	state     deletionState
	outboxErr error
	commitErr error
}

func (h *deletionHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	staged := h.state.clone()
	runCtx := domainCollaboration.WithOutboxStore(
		ctx,
		&deletionOutboxStore{events: &staged.events, err: h.outboxErr},
	)
	if err := run(runCtx, &staged); err != nil {
		return err
	}
	if h.commitErr != nil {
		return h.commitErr
	}
	h.state = staged
	return nil
}

func (h *deletionHarness) factory(unit apptransaction.UnitOfWork) (DeletionWorkflow, error) {
	state, ok := unit.(*deletionState)
	if !ok {
		return nil, errors.New("unexpected transaction unit")
	}
	return (*deletionWorkflow)(state), nil
}

type deletionWorkflow deletionState

func (workflow *deletionWorkflow) recordBatch(ctx context.Context) {
	if batchID, ok := ctx.Value(deletionBatchKey{}).(uuid.UUID); ok {
		workflow.batchIDs = append(workflow.batchIDs, batchID)
	}
}

func (workflow *deletionWorkflow) GetProjectForDeletion(
	ctx context.Context,
	id uuid.UUID,
) (*Snapshot, error) {
	workflow.recordBatch(ctx)
	project, ok := workflow.projects[id]
	if !ok {
		return nil, domain.ErrNotFound
	}
	copyProject := project
	return &copyProject, nil
}

func (workflow *deletionWorkflow) GetActiveUserRole(
	_ context.Context,
	id uuid.UUID,
) (domainUser.Role, error) {
	role, ok := workflow.roles[id]
	if !ok {
		return "", domain.ErrNotFound
	}
	return role, nil
}

func (workflow *deletionWorkflow) HasHierarchyLinks(
	_ context.Context,
	projectID uuid.UUID,
) (bool, error) {
	return workflow.hierarchyLinks[projectID], nil
}

func (workflow *deletionWorkflow) ListProjectObjectDataIDs(
	_ context.Context,
	projectID, after uuid.UUID,
	limit int,
) ([]uuid.UUID, error) {
	ids := make([]uuid.UUID, 0)
	for id, ownerID := range workflow.objectData {
		if ownerID == projectID && (after == uuid.Nil || id.String() > after.String()) {
			ids = append(ids, id)
		}
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].String() < ids[j].String()
	})
	if len(ids) > limit {
		ids = ids[:limit]
	}
	return ids, nil
}

func (workflow *deletionWorkflow) DeleteObjectData(
	ctx context.Context,
	ids []uuid.UUID,
) error {
	workflow.recordBatch(ctx)
	workflow.objectDeleteCalls = append(
		workflow.objectDeleteCalls,
		append([]uuid.UUID(nil), ids...),
	)
	for _, id := range ids {
		delete(workflow.objectData, id)
	}
	return nil
}

func (workflow *deletionWorkflow) DeleteProjectMemberships(
	_ context.Context,
	projectID uuid.UUID,
) error {
	delete(workflow.memberships, projectID)
	return nil
}

func (workflow *deletionWorkflow) DeleteProject(
	ctx context.Context,
	projectID uuid.UUID,
) error {
	workflow.recordBatch(ctx)
	delete(workflow.projects, projectID)
	workflow.projectDeletes = append(workflow.projectDeletes, projectID)
	return nil
}

type deletionOutboxStore struct {
	events *[]domainCollaboration.OutboxEvent
	err    error
}

func (store *deletionOutboxStore) Enqueue(
	_ context.Context,
	event *domainCollaboration.OutboxEvent,
) error {
	if store.err != nil {
		return store.err
	}
	*store.events = append(*store.events, *event)
	return nil
}

func (*deletionOutboxStore) ClaimDue(context.Context, time.Time, int) ([]domainCollaboration.OutboxEvent, error) {
	return nil, nil
}
func (*deletionOutboxStore) WasProcessed(context.Context, string, uuid.UUID) (bool, error) {
	return false, nil
}
func (*deletionOutboxStore) MarkDelivered(context.Context, string, domainCollaboration.OutboxEvent, time.Time) error {
	return nil
}
func (*deletionOutboxStore) MarkFailed(context.Context, domainCollaboration.OutboxEvent, string, time.Time, time.Time) error {
	return nil
}

func TestDeleteProjectEnforcesAdminPolicyCleansOwnedDataAndPersistsV2Event(t *testing.T) {
	statuses := []domainProject.ProjectStatus{
		domainProject.StatusCompleted,
		domainProject.StatusOngoing,
	}
	roles := []domainUser.Role{
		domainUser.RoleSuperAdmin,
		domainUser.RoleAdminFZAG,
	}

	for _, status := range statuses {
		for _, role := range roles {
			t.Run(string(status)+"/"+string(role), func(t *testing.T) {
				projectID := uuid.New()
				actorID := uuid.New()
				otherProjectID := uuid.New()
				operationID := uuid.New()
				eventID := uuid.New()
				occurredAt := time.Date(2026, 7, 23, 18, 0, 0, 0, time.UTC)
				objectData := make(map[uuid.UUID]uuid.UUID, 207)
				for range 205 {
					objectData[uuid.New()] = projectID
				}
				otherObjectDataID := uuid.New()
				objectData[otherObjectDataID] = otherProjectID

				harness := &deletionHarness{state: deletionState{
					projects: map[uuid.UUID]Snapshot{
						projectID:      {ID: projectID, Status: status},
						otherProjectID: {ID: otherProjectID, Status: domainProject.StatusPlanned},
					},
					roles:          map[uuid.UUID]domainUser.Role{actorID: role},
					hierarchyLinks: map[uuid.UUID]bool{},
					objectData:     objectData,
					memberships: map[uuid.UUID][]uuid.UUID{
						projectID:      {uuid.New(), uuid.New()},
						otherProjectID: {uuid.New()},
					},
				}}
				ids := []uuid.UUID{operationID, eventID}
				handler := NewDeleteHandler(DeleteDependencies{
					TransactionRunner:   harness.runner,
					TransactionWorkflow: harness.factory,
					HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
						return context.WithValue(ctx, deletionBatchKey{}, batchID)
					},
					Actor: func(context.Context) *uuid.UUID { return &actorID },
					NewID: func() uuid.UUID {
						id := ids[0]
						ids = ids[1:]
						return id
					},
					Now: func() time.Time { return occurredAt },
				})

				outcome, err := handler.Execute(
					context.Background(),
					DeleteCommand{ProjectID: projectID},
				)
				if err != nil {
					t.Fatalf("delete project: %v", err)
				}
				if outcome.Project.ID != projectID || outcome.Project.Status != status ||
					outcome.DeletedObjectData != 205 ||
					outcome.OperationID != operationID || outcome.EventID != eventID ||
					outcome.OccurredAt != occurredAt {
					t.Fatalf("unexpected outcome: %+v", outcome)
				}
				if _, exists := harness.state.projects[projectID]; exists {
					t.Fatal("project still exists")
				}
				if _, exists := harness.state.projects[otherProjectID]; !exists {
					t.Fatal("another project was deleted")
				}
				if _, exists := harness.state.objectData[otherObjectDataID]; !exists {
					t.Fatal("another project's ObjectData was deleted")
				}
				for _, ownerID := range harness.state.objectData {
					if ownerID == projectID {
						t.Fatal("project-owned ObjectData remains")
					}
				}
				if _, exists := harness.state.memberships[projectID]; exists {
					t.Fatal("project memberships remain")
				}
				if _, exists := harness.state.memberships[otherProjectID]; !exists {
					t.Fatal("another project's memberships were deleted")
				}
				if len(harness.state.objectDeleteCalls) != 3 {
					t.Fatalf("ObjectData delete batches: %d", len(harness.state.objectDeleteCalls))
				}
				for _, batch := range harness.state.objectDeleteCalls {
					if len(batch) == 0 || len(batch) > projectOwnedDataDeleteBatchSize {
						t.Fatalf("invalid ObjectData delete batch size: %d", len(batch))
					}
				}
				if !reflect.DeepEqual(harness.state.projectDeletes, []uuid.UUID{projectID}) {
					t.Fatalf("project deletes: %v", harness.state.projectDeletes)
				}
				if len(harness.state.batchIDs) != 5 {
					t.Fatalf("history batch uses: %v", harness.state.batchIDs)
				}
				for _, batchID := range harness.state.batchIDs {
					if batchID != operationID {
						t.Fatalf("history batch ID: got %s, want %s", batchID, operationID)
					}
				}
				if len(harness.state.events) != 1 {
					t.Fatalf("outbox events: %d", len(harness.state.events))
				}
				event := harness.state.events[0]
				decoded, err := appcollaboration.DecodeCommand(
					appcollaboration.EncodedCommand{
						Type: event.EventType, Payload: event.Payload,
					},
				)
				if err != nil {
					t.Fatalf("decode outbox command: %v", err)
				}
				refresh, ok := decoded.(appcollaboration.FacilityHierarchyRefreshRequired)
				if !ok || refresh.SchemaVersion != appcollaboration.SchemaVersionV2 ||
					refresh.ProjectID != projectID ||
					refresh.OperationID != operationID ||
					refresh.CorrelationID != operationID ||
					refresh.EventID != eventID ||
					refresh.ActorID == nil || *refresh.ActorID != actorID ||
					refresh.OccurredAt != occurredAt ||
					refresh.Scope != appcollaboration.FacilityScopeProject ||
					!refresh.FullRefresh ||
					!reflect.DeepEqual(refresh.EntityIDs, []uuid.UUID{projectID}) {
					t.Fatalf("unexpected durable refresh: %#v", decoded)
				}
			})
		}
	}
}

func TestDeleteProjectRejectsHierarchyLinksForEveryStatus(t *testing.T) {
	for _, status := range []domainProject.ProjectStatus{
		domainProject.StatusPlanned,
		domainProject.StatusOngoing,
		domainProject.StatusCompleted,
	} {
		t.Run(string(status), func(t *testing.T) {
			projectID := uuid.New()
			actorID := uuid.New()
			harness := &deletionHarness{state: deletionState{
				projects:       map[uuid.UUID]Snapshot{projectID: {ID: projectID, Status: status}},
				roles:          map[uuid.UUID]domainUser.Role{actorID: domainUser.RoleSuperAdmin},
				hierarchyLinks: map[uuid.UUID]bool{projectID: true},
				objectData:     map[uuid.UUID]uuid.UUID{uuid.New(): projectID},
				memberships:    map[uuid.UUID][]uuid.UUID{projectID: {actorID}},
			}}
			handler := NewDeleteHandler(DeleteDependencies{
				TransactionRunner:   harness.runner,
				TransactionWorkflow: harness.factory,
				Actor:               func(context.Context) *uuid.UUID { return &actorID },
			})

			err := handler.Delete(context.Background(), DeleteCommand{ProjectID: projectID})
			if !errors.Is(err, ErrHierarchyLinksRemain) {
				t.Fatalf("expected hierarchy-link conflict, got %v", err)
			}
			assertDeletionStateUnchanged(t, harness.state, projectID)
		})
	}
}

func TestDeleteProjectRejectsEveryNonFZAGAdminRoleAndMissingActor(t *testing.T) {
	roles := []domainUser.Role{
		domainUser.RoleFZAG,
		domainUser.RoleAdminPlaner,
		domainUser.RolePlaner,
		domainUser.RoleAdminEnterpreneur,
		domainUser.RoleEnterpreneur,
	}
	for _, role := range roles {
		t.Run(string(role), func(t *testing.T) {
			projectID := uuid.New()
			actorID := uuid.New()
			harness := deletionHarnessForAuthorization(projectID, actorID, role)
			handler := NewDeleteHandler(DeleteDependencies{
				TransactionRunner:   harness.runner,
				TransactionWorkflow: harness.factory,
				Actor:               func(context.Context) *uuid.UUID { return &actorID },
			})

			err := handler.Delete(context.Background(), DeleteCommand{ProjectID: projectID})
			if !errors.Is(err, ErrDeletionForbidden) {
				t.Fatalf("expected forbidden, got %v", err)
			}
			assertDeletionStateUnchanged(t, harness.state, projectID)
		})
	}

	t.Run("missing actor", func(t *testing.T) {
		projectID := uuid.New()
		harness := deletionHarnessForAuthorization(
			projectID,
			uuid.New(),
			domainUser.RoleSuperAdmin,
		)
		handler := NewDeleteHandler(DeleteDependencies{
			TransactionRunner:   harness.runner,
			TransactionWorkflow: harness.factory,
		})
		err := handler.Delete(context.Background(), DeleteCommand{ProjectID: projectID})
		if !errors.Is(err, ErrDeletionForbidden) {
			t.Fatalf("expected forbidden, got %v", err)
		}
		assertDeletionStateUnchanged(t, harness.state, projectID)
	})
}

func TestDeleteProjectRollsBackAllWritesWhenOutboxFails(t *testing.T) {
	projectID := uuid.New()
	actorID := uuid.New()
	outboxErr := errors.New("outbox unavailable")
	harness := deletionHarnessForAuthorization(
		projectID,
		actorID,
		domainUser.RoleAdminFZAG,
	)
	harness.outboxErr = outboxErr
	handler := NewDeleteHandler(DeleteDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Actor:               func(context.Context) *uuid.UUID { return &actorID },
	})

	err := handler.Delete(context.Background(), DeleteCommand{ProjectID: projectID})
	if !errors.Is(err, outboxErr) {
		t.Fatalf("expected outbox failure, got %v", err)
	}
	assertDeletionStateUnchanged(t, harness.state, projectID)
}

func deletionHarnessForAuthorization(
	projectID, actorID uuid.UUID,
	role domainUser.Role,
) *deletionHarness {
	return &deletionHarness{state: deletionState{
		projects: map[uuid.UUID]Snapshot{
			projectID: {ID: projectID, Status: domainProject.StatusPlanned},
		},
		roles:          map[uuid.UUID]domainUser.Role{actorID: role},
		hierarchyLinks: map[uuid.UUID]bool{},
		objectData:     map[uuid.UUID]uuid.UUID{uuid.New(): projectID},
		memberships:    map[uuid.UUID][]uuid.UUID{projectID: {actorID}},
	}}
}

func assertDeletionStateUnchanged(
	t *testing.T,
	state deletionState,
	projectID uuid.UUID,
) {
	t.Helper()
	if _, exists := state.projects[projectID]; !exists {
		t.Fatal("project mutation escaped rollback/rejection")
	}
	if len(state.memberships[projectID]) == 0 {
		t.Fatal("project membership mutation escaped rollback/rejection")
	}
	hasObjectData := false
	for _, ownerID := range state.objectData {
		if ownerID == projectID {
			hasObjectData = true
			break
		}
	}
	if !hasObjectData {
		t.Fatal("ObjectData mutation escaped rollback/rejection")
	}
	if len(state.objectDeleteCalls) != 0 ||
		len(state.projectDeletes) != 0 ||
		len(state.events) != 0 {
		t.Fatalf(
			"unexpected writes: object=%d project=%d event=%d",
			len(state.objectDeleteCalls),
			len(state.projectDeletes),
			len(state.events),
		)
	}
}
