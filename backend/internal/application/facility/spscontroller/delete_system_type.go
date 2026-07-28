package spscontroller

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	"github.com/besart951/go_infra_link/backend/internal/application/facility/mutation"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

var ErrDeleteSystemTypeTransactionNotConfigured = errors.New(
	"SPS controller system-type delete transaction is not configured",
)

// DeleteSystemTypeWorkflow intentionally loads and deletes only the assignment
// root. Existing database cascade behavior remains the compatibility policy;
// descendant history and project-link reconciliation require a separate,
// bounded hierarchy-delete use case.
type DeleteSystemTypeWorkflow interface {
	GetByID(context.Context, uuid.UUID) (*domainFacility.SPSControllerSystemType, error)
	DeleteByID(context.Context, uuid.UUID) error
}

type transactionalDeleteSystemTypeScope interface {
	DeleteSystemTypeWorkflow
	GetDeleteProjectScope(
		context.Context,
		uuid.UUID,
	) (uuid.UUID, []uuid.UUID, error)
}

type DeleteSystemTypeCommand struct {
	SPSControllerSystemTypeID uuid.UUID
}

func (c DeleteSystemTypeCommand) validate() error {
	if c.SPSControllerSystemTypeID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	return nil
}

type DeleteSystemTypeDependencies struct {
	TransactionRunner   apptransaction.Runner
	TransactionWorkflow apptransaction.Factory[DeleteSystemTypeWorkflow]
	HistoryBatch        HistoryBatchContext
	Dispatcher          appcollaboration.CommandDispatcher
	Actor               ActorProvider
	NewID               IDGenerator
	Now                 Clock
	ReportError         ErrorReporter
}

type DeleteSystemTypeHandler struct {
	operation apptransaction.Operation[
		DeleteSystemTypeWorkflow,
		DeleteSystemTypeWorkflow,
	]
	transactionConfigured bool
	historyBatch          HistoryBatchContext
	dispatcher            appcollaboration.CommandDispatcher
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
	reportError           ErrorReporter
}

type DeleteSystemTypeOutcome struct {
	Mutation       mutation.Result
	Existed        bool
	DispatchErrors []error
}

type committedSystemTypeDelete struct {
	change     *mutation.EntityChange
	batched    bool
	projectIDs []uuid.UUID
	commands   []appcollaboration.Command
}

func NewDeleteSystemTypeHandler(
	deps DeleteSystemTypeDependencies,
) *DeleteSystemTypeHandler {
	newID := deps.NewID
	if newID == nil {
		newID = uuid.New
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	boundary := apptransaction.NewBoundary[DeleteSystemTypeWorkflow](
		deps.TransactionRunner,
		deps.TransactionWorkflow,
	)
	var noNonTransactionalWorkflow DeleteSystemTypeWorkflow
	operation := apptransaction.Bind(
		boundary,
		noNonTransactionalWorkflow,
		func(workflow DeleteSystemTypeWorkflow) DeleteSystemTypeWorkflow {
			return workflow
		},
	)
	return &DeleteSystemTypeHandler{
		operation:             operation,
		transactionConfigured: deps.TransactionRunner != nil && deps.TransactionWorkflow != nil,
		historyBatch:          deps.HistoryBatch,
		dispatcher:            deps.Dispatcher,
		actor:                 deps.Actor,
		newID:                 newID,
		now:                   now,
		reportError:           deps.ReportError,
	}
}

// DeleteSystemType preserves the existing idempotent 204 transport contract.
func (h *DeleteSystemTypeHandler) DeleteSystemType(
	ctx context.Context,
	command DeleteSystemTypeCommand,
) error {
	outcome, err := h.Execute(ctx, command)
	if err != nil {
		return err
	}
	for _, dispatchErr := range outcome.DispatchErrors {
		if h.reportError != nil {
			h.reportError(dispatchErr)
		}
	}
	return nil
}

func (h *DeleteSystemTypeHandler) Execute(
	ctx context.Context,
	command DeleteSystemTypeCommand,
) (DeleteSystemTypeOutcome, error) {
	if h == nil || !h.transactionConfigured {
		return DeleteSystemTypeOutcome{}, ErrDeleteSystemTypeTransactionNotConfigured
	}
	if err := command.validate(); err != nil {
		return DeleteSystemTypeOutcome{}, err
	}

	operationID := h.newID()
	actorID := actorFromContext(h.actor, ctx)
	occurredAt := h.now().UTC()
	committed, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(
			txCtx context.Context,
			workflow DeleteSystemTypeWorkflow,
		) (committedSystemTypeDelete, error) {
			var (
				ownerID    uuid.UUID
				projectIDs []uuid.UUID
			)
			if scope, ok := workflow.(transactionalDeleteSystemTypeScope); ok {
				var scopeErr error
				ownerID, projectIDs, scopeErr = scope.GetDeleteProjectScope(
					txCtx,
					command.SPSControllerSystemTypeID,
				)
				if scopeErr != nil {
					return committedSystemTypeDelete{}, fmt.Errorf(
						"resolve SPSControllerSystemType delete scope: %w",
						scopeErr,
					)
				}
			}
			result, err := executeDeleteSystemTypeTransaction(
				txCtx,
				workflow,
				command,
				operationID,
				h.historyBatch,
			)
			if err != nil || result.change == nil {
				return result, err
			}
			if ownerID == uuid.Nil {
				return result, nil
			}
			result.projectIDs = append([]uuid.UUID(nil), projectIDs...)
			for _, projectID := range projectIDs {
				event := appcollaboration.FacilityHierarchyRefreshRequired{
					Envelope: appcollaboration.Envelope{
						SchemaVersion: appcollaboration.SchemaVersionV2,
						EventID:       h.newID(),
						OperationID:   operationID,
						CorrelationID: operationID,
						ProjectID:     projectID,
						ActorID:       actorID,
						OccurredAt:    occurredAt,
					},
					Scope:     appcollaboration.FacilityScopeSPSController,
					EntityIDs: []uuid.UUID{ownerID},
				}
				configured, err := appcollaboration.EnqueueCommand(txCtx, event)
				if err != nil {
					return committedSystemTypeDelete{}, fmt.Errorf(
						"enqueue SPSControllerSystemType delete for project %s: %w",
						projectID,
						err,
					)
				}
				if configured {
					result.commands = append(result.commands, event)
				}
			}
			return result, nil
		},
	)
	if err != nil {
		return DeleteSystemTypeOutcome{}, err
	}

	result := mutation.Result{
		OperationID: operationID,
		ActorID:     actorID,
		OccurredAt:  occurredAt,
		ProjectIDs:  append([]uuid.UUID(nil), committed.projectIDs...),
	}
	if committed.batched {
		batchID := operationID
		result.BatchID = &batchID
	}
	if committed.change != nil {
		result.Changes = []mutation.EntityChange{*committed.change}
	}
	outcome := DeleteSystemTypeOutcome{
		Mutation: result,
		Existed:  committed.change != nil,
	}
	if committed.change == nil || h.dispatcher == nil {
		return outcome, nil
	}
	dispatchCtx := context.WithoutCancel(ctx)
	for _, event := range committed.commands {
		envelope, _ := appcollaboration.CommandEnvelope(event)
		if err := h.dispatcher.Dispatch(dispatchCtx, event); err != nil {
			outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
				"dispatch SPSControllerSystemType delete for project %s: %w",
				envelope.ProjectID,
				err,
			))
		}
	}
	return outcome, nil
}

func executeDeleteSystemTypeTransaction(
	ctx context.Context,
	workflow DeleteSystemTypeWorkflow,
	command DeleteSystemTypeCommand,
	operationID uuid.UUID,
	historyBatch HistoryBatchContext,
) (committedSystemTypeDelete, error) {
	if workflow == nil {
		return committedSystemTypeDelete{}, ErrDeleteSystemTypeTransactionNotConfigured
	}

	before, err := workflow.GetByID(ctx, command.SPSControllerSystemTypeID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return committedSystemTypeDelete{}, err
	}
	var change *mutation.EntityChange
	if before != nil && err == nil {
		built, buildErr := buildSystemTypeDeleteChange(before)
		if buildErr != nil {
			return committedSystemTypeDelete{}, buildErr
		}
		change = &built
	}

	writeCtx := ctx
	batched := historyBatch != nil
	if batched {
		writeCtx = historyBatch(ctx, operationID)
	}
	if err := workflow.DeleteByID(writeCtx, command.SPSControllerSystemTypeID); err != nil {
		return committedSystemTypeDelete{}, err
	}
	return committedSystemTypeDelete{
		change:  change,
		batched: batched,
	}, nil
}

func buildSystemTypeDeleteChange(
	before *domainFacility.SPSControllerSystemType,
) (mutation.EntityChange, error) {
	beforeJSON, err := json.Marshal(toSystemTypeSnapshot(before))
	if err != nil {
		return mutation.EntityChange{}, err
	}
	parentID := before.SPSControllerID
	return mutation.EntityChange{
		EntityType: mutation.EntityTypeSPSControllerSystemType,
		EntityID:   before.ID,
		ParentID:   &parentID,
		Action:     domainHistory.ActionDelete,
		Before:     json.RawMessage(beforeJSON),
	}, nil
}
