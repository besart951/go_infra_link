package projectlink

import (
	"context"
	"errors"
	"fmt"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

var ErrUnlinkTransactionNotConfigured = errors.New(
	"project facility unlink transaction is not configured",
)

type Kind string

const (
	KindControlCabinet Kind = "control_cabinet"
	KindSPSController  Kind = "sps_controller"
	KindFieldDevice    Kind = "field_device"
)

func (kind Kind) valid() bool {
	switch kind {
	case KindControlCabinet, KindSPSController, KindFieldDevice:
		return true
	default:
		return false
	}
}

func (kind Kind) collaborationScope() appcollaboration.FacilityScope {
	switch kind {
	case KindControlCabinet:
		return appcollaboration.FacilityScopeControlCabinet
	case KindSPSController:
		return appcollaboration.FacilityScopeSPSController
	case KindFieldDevice:
		return appcollaboration.FacilityScopeFieldDevice
	default:
		return ""
	}
}

// Link is the normalized application view of one project association. The
// global facility entity is deliberately not exposed as a deletion capability.
type Link struct {
	ID        uuid.UUID
	ProjectID uuid.UUID
	EntityID  uuid.UUID
}

// Workflow is implemented by a transaction-scoped adapter over the three
// project-link repositories. Delete removes only the selected association.
type Workflow interface {
	GetProjectFacilityLink(context.Context, Kind, uuid.UUID) (*Link, error)
	DeleteProjectFacilityLink(context.Context, Kind, uuid.UUID) error
}

type HistoryBatchContext func(context.Context, uuid.UUID) context.Context
type ActorProvider func(context.Context) *uuid.UUID
type IDGenerator func() uuid.UUID
type Clock func() time.Time

type Command struct {
	Kind      Kind
	ProjectID uuid.UUID
	LinkID    uuid.UUID
}

func (command Command) validate() error {
	if !command.Kind.valid() || command.ProjectID == uuid.Nil || command.LinkID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	return nil
}

type Dependencies struct {
	TransactionRunner   apptransaction.Runner
	TransactionWorkflow apptransaction.Factory[Workflow]
	HistoryBatch        HistoryBatchContext
	Actor               ActorProvider
	NewID               IDGenerator
	Now                 Clock
}

type Handler struct {
	operation             apptransaction.Operation[Workflow, Workflow]
	transactionConfigured bool
	historyBatch          HistoryBatchContext
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
}

type Outcome struct {
	Link        Link
	OperationID uuid.UUID
	EventID     uuid.UUID
	OccurredAt  time.Time
}

func NewHandler(deps Dependencies) *Handler {
	newID := deps.NewID
	if newID == nil {
		newID = uuid.New
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	boundary := apptransaction.NewBoundary[Workflow](
		deps.TransactionRunner,
		deps.TransactionWorkflow,
	)
	var noNonTransactionalWorkflow Workflow
	return &Handler{
		operation: apptransaction.Bind(
			boundary,
			noNonTransactionalWorkflow,
			func(workflow Workflow) Workflow { return workflow },
		),
		transactionConfigured: deps.TransactionRunner != nil &&
			deps.TransactionWorkflow != nil,
		historyBatch: deps.HistoryBatch,
		actor:        deps.Actor,
		newID:        newID,
		now:          now,
	}
}

func (h *Handler) Unlink(ctx context.Context, command Command) error {
	_, err := h.Execute(ctx, command)
	return err
}

func (h *Handler) Execute(ctx context.Context, command Command) (Outcome, error) {
	if h == nil || !h.transactionConfigured {
		return Outcome{}, ErrUnlinkTransactionNotConfigured
	}
	if err := command.validate(); err != nil {
		return Outcome{}, err
	}

	operationID := h.newID()
	eventID := h.newID()
	occurredAt := h.now().UTC()
	actorID := actorFromContext(h.actor, ctx)

	link, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(txCtx context.Context, workflow Workflow) (Link, error) {
			if h.historyBatch != nil {
				txCtx = h.historyBatch(txCtx, operationID)
			}
			stored, err := workflow.GetProjectFacilityLink(txCtx, command.Kind, command.LinkID)
			if err != nil {
				return Link{}, err
			}
			if stored == nil || stored.ID != command.LinkID ||
				stored.ProjectID != command.ProjectID ||
				stored.EntityID == uuid.Nil {
				return Link{}, domain.ErrNotFound
			}
			if err := workflow.DeleteProjectFacilityLink(txCtx, command.Kind, command.LinkID); err != nil {
				return Link{}, err
			}

			durableCommand := appcollaboration.FacilityHierarchyRefreshRequired{
				Envelope: appcollaboration.Envelope{
					SchemaVersion: appcollaboration.SchemaVersionV2,
					EventID:       eventID,
					OperationID:   operationID,
					CorrelationID: operationID,
					ProjectID:     command.ProjectID,
					ActorID:       actorID,
					OccurredAt:    occurredAt,
				},
				Scope:       command.Kind.collaborationScope(),
				EntityIDs:   []uuid.UUID{stored.EntityID},
				FullRefresh: true,
			}
			if _, err := appcollaboration.EnqueueCommand(txCtx, durableCommand); err != nil {
				return Link{}, fmt.Errorf("enqueue project %s unlink: %w", command.Kind, err)
			}
			return *stored, nil
		},
	)
	if err != nil {
		return Outcome{}, err
	}
	return Outcome{
		Link:        link,
		OperationID: operationID,
		EventID:     eventID,
		OccurredAt:  occurredAt,
	}, nil
}

func actorFromContext(provider ActorProvider, ctx context.Context) *uuid.UUID {
	if provider == nil {
		return nil
	}
	actorID := provider(ctx)
	if actorID == nil || *actorID == uuid.Nil {
		return nil
	}
	copyID := *actorID
	return &copyID
}
