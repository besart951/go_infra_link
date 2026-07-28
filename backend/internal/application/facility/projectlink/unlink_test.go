package projectlink

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainCollaboration "github.com/besart951/go_infra_link/backend/internal/domain/collaboration"
	"github.com/google/uuid"
)

type unlinkBatchKey struct{}

type unlinkState struct {
	links       map[Kind]map[uuid.UUID]Link
	events      []domainCollaboration.OutboxEvent
	deleteCalls []Command
	batchIDs    []uuid.UUID
}

func (state unlinkState) clone() unlinkState {
	cloned := unlinkState{
		links:       make(map[Kind]map[uuid.UUID]Link, len(state.links)),
		events:      append([]domainCollaboration.OutboxEvent(nil), state.events...),
		deleteCalls: append([]Command(nil), state.deleteCalls...),
		batchIDs:    append([]uuid.UUID(nil), state.batchIDs...),
	}
	for kind, links := range state.links {
		cloned.links[kind] = make(map[uuid.UUID]Link, len(links))
		for id, link := range links {
			cloned.links[kind][id] = link
		}
	}
	return cloned
}

type unlinkHarness struct {
	state      unlinkState
	outboxErr  error
	commitErr  error
	runnerRuns int
}

func (h *unlinkHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerRuns++
	staged := h.state.clone()
	store := &unlinkOutboxStore{events: &staged.events, err: h.outboxErr}
	runCtx := domainCollaboration.WithOutboxStore(ctx, store)
	if err := run(runCtx, &staged); err != nil {
		return err
	}
	if h.commitErr != nil {
		return h.commitErr
	}
	h.state = staged
	return nil
}

func (h *unlinkHarness) factory(unit apptransaction.UnitOfWork) (Workflow, error) {
	state, ok := unit.(*unlinkState)
	if !ok {
		return nil, errors.New("unexpected transaction unit")
	}
	return (*unlinkWorkflow)(state), nil
}

type unlinkWorkflow unlinkState

func (workflow *unlinkWorkflow) GetProjectFacilityLink(
	ctx context.Context,
	kind Kind,
	linkID uuid.UUID,
) (*Link, error) {
	if batchID, ok := ctx.Value(unlinkBatchKey{}).(uuid.UUID); ok {
		workflow.batchIDs = append(workflow.batchIDs, batchID)
	}
	link, ok := workflow.links[kind][linkID]
	if !ok {
		return nil, nil
	}
	copyLink := link
	return &copyLink, nil
}

func (workflow *unlinkWorkflow) DeleteProjectFacilityLink(
	_ context.Context,
	kind Kind,
	linkID uuid.UUID,
) error {
	link, ok := workflow.links[kind][linkID]
	if !ok {
		return nil
	}
	workflow.deleteCalls = append(workflow.deleteCalls, Command{
		Kind:      kind,
		ProjectID: link.ProjectID,
		LinkID:    linkID,
	})
	delete(workflow.links[kind], linkID)
	return nil
}

type unlinkOutboxStore struct {
	events *[]domainCollaboration.OutboxEvent
	err    error
}

func (store *unlinkOutboxStore) Enqueue(
	_ context.Context,
	event *domainCollaboration.OutboxEvent,
) error {
	if store.err != nil {
		return store.err
	}
	*store.events = append(*store.events, *event)
	return nil
}

func (*unlinkOutboxStore) ClaimDue(context.Context, time.Time, int) ([]domainCollaboration.OutboxEvent, error) {
	return nil, nil
}
func (*unlinkOutboxStore) WasProcessed(context.Context, string, uuid.UUID) (bool, error) {
	return false, nil
}
func (*unlinkOutboxStore) MarkDelivered(context.Context, string, domainCollaboration.OutboxEvent, time.Time) error {
	return nil
}
func (*unlinkOutboxStore) MarkFailed(context.Context, domainCollaboration.OutboxEvent, string, time.Time, time.Time) error {
	return nil
}

func TestUnlinkDeletesOnlyTheSelectedProjectAssociationAndPersistsV2Scope(t *testing.T) {
	projectID := uuid.New()
	otherProjectID := uuid.New()
	actorID := uuid.New()
	occurredAt := time.Date(2026, 7, 23, 15, 0, 0, 0, time.UTC)

	tests := []struct {
		kind  Kind
		scope appcollaboration.FacilityScope
	}{
		{kind: KindControlCabinet, scope: appcollaboration.FacilityScopeControlCabinet},
		{kind: KindSPSController, scope: appcollaboration.FacilityScopeSPSController},
		{kind: KindFieldDevice, scope: appcollaboration.FacilityScopeFieldDevice},
	}

	for _, test := range tests {
		t.Run(string(test.kind), func(t *testing.T) {
			linkID := uuid.New()
			entityID := uuid.New()
			otherLinkID := uuid.New()
			operationID := uuid.New()
			eventID := uuid.New()
			harness := &unlinkHarness{state: unlinkState{links: map[Kind]map[uuid.UUID]Link{
				test.kind: {
					linkID: {
						ID: linkID, ProjectID: projectID, EntityID: entityID,
					},
					otherLinkID: {
						ID: otherLinkID, ProjectID: otherProjectID, EntityID: entityID,
					},
				},
			}}}
			ids := []uuid.UUID{operationID, eventID}
			handler := NewHandler(Dependencies{
				TransactionRunner:   harness.runner,
				TransactionWorkflow: harness.factory,
				HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
					return context.WithValue(ctx, unlinkBatchKey{}, batchID)
				},
				Actor: func(context.Context) *uuid.UUID { return &actorID },
				NewID: func() uuid.UUID {
					id := ids[0]
					ids = ids[1:]
					return id
				},
				Now: func() time.Time { return occurredAt },
			})

			outcome, err := handler.Execute(context.Background(), Command{
				Kind: test.kind, ProjectID: projectID, LinkID: linkID,
			})
			if err != nil {
				t.Fatalf("unlink: %v", err)
			}
			if outcome.Link.ID != linkID || outcome.Link.EntityID != entityID ||
				outcome.OperationID != operationID || outcome.EventID != eventID {
				t.Fatalf("unexpected outcome: %+v", outcome)
			}
			if _, exists := harness.state.links[test.kind][linkID]; exists {
				t.Fatal("selected project association still exists")
			}
			if _, exists := harness.state.links[test.kind][otherLinkID]; !exists {
				t.Fatal("another project's association was removed")
			}
			if len(harness.state.deleteCalls) != 1 ||
				harness.state.deleteCalls[0].LinkID != linkID {
				t.Fatalf("unexpected deletes: %+v", harness.state.deleteCalls)
			}
			if !reflect.DeepEqual(harness.state.batchIDs, []uuid.UUID{operationID}) {
				t.Fatalf("history batch IDs: %v", harness.state.batchIDs)
			}
			if len(harness.state.events) != 1 {
				t.Fatalf("outbox events: %d", len(harness.state.events))
			}
			event := harness.state.events[0]
			decoded, err := appcollaboration.DecodeCommand(appcollaboration.EncodedCommand{
				Type: event.EventType, Payload: event.Payload,
			})
			if err != nil {
				t.Fatalf("decode outbox command: %v", err)
			}
			refresh, ok := decoded.(appcollaboration.FacilityHierarchyRefreshRequired)
			if !ok || refresh.SchemaVersion != appcollaboration.SchemaVersionV2 ||
				refresh.ProjectID != projectID || refresh.OperationID != operationID ||
				refresh.CorrelationID != operationID || refresh.EventID != eventID ||
				refresh.ActorID == nil || *refresh.ActorID != actorID ||
				refresh.OccurredAt != occurredAt || refresh.Scope != test.scope ||
				!refresh.FullRefresh ||
				!reflect.DeepEqual(refresh.EntityIDs, []uuid.UUID{entityID}) {
				t.Fatalf("unexpected durable refresh: %#v", decoded)
			}
		})
	}
}

func TestUnlinkRejectsCrossProjectLinkWithoutDeletingOrPublishing(t *testing.T) {
	linkID := uuid.New()
	storedProjectID := uuid.New()
	harness := &unlinkHarness{state: unlinkState{links: map[Kind]map[uuid.UUID]Link{
		KindFieldDevice: {
			linkID: {ID: linkID, ProjectID: storedProjectID, EntityID: uuid.New()},
		},
	}}}
	handler := NewHandler(Dependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
	})

	err := handler.Unlink(context.Background(), Command{
		Kind: KindFieldDevice, ProjectID: uuid.New(), LinkID: linkID,
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("expected not found, got %v", err)
	}
	if _, exists := harness.state.links[KindFieldDevice][linkID]; !exists {
		t.Fatal("cross-project link was deleted")
	}
	if len(harness.state.events) != 0 || len(harness.state.deleteCalls) != 0 {
		t.Fatalf("unexpected mutation: events=%d deletes=%d", len(harness.state.events), len(harness.state.deleteCalls))
	}
}

func TestUnlinkRollsBackLinkAndHistoryWhenOutboxEnqueueFails(t *testing.T) {
	linkID := uuid.New()
	projectID := uuid.New()
	outboxErr := errors.New("outbox unavailable")
	harness := &unlinkHarness{
		outboxErr: outboxErr,
		state: unlinkState{links: map[Kind]map[uuid.UUID]Link{
			KindControlCabinet: {
				linkID: {ID: linkID, ProjectID: projectID, EntityID: uuid.New()},
			},
		}},
	}
	handler := NewHandler(Dependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
	})

	err := handler.Unlink(context.Background(), Command{
		Kind: KindControlCabinet, ProjectID: projectID, LinkID: linkID,
	})
	if !errors.Is(err, outboxErr) {
		t.Fatalf("expected outbox failure, got %v", err)
	}
	if _, exists := harness.state.links[KindControlCabinet][linkID]; !exists {
		t.Fatal("link deletion survived outbox rollback")
	}
	if len(harness.state.events) != 0 || len(harness.state.deleteCalls) != 0 {
		t.Fatalf("rollback leaked state: events=%d deletes=%d", len(harness.state.events), len(harness.state.deleteCalls))
	}
}
