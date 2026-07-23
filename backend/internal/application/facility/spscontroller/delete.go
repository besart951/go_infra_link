package spscontroller

import (
	"context"
	"errors"
	"fmt"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	"github.com/besart951/go_infra_link/backend/internal/application/facility/mutation"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

var ErrDeleteTransactionNotConfigured = errors.New(
	"SPS controller delete transaction is not configured",
)

// DeleteWorkflow is the narrow transaction-scoped Interface needed by the
// existing global SPSController DELETE endpoint.
type DeleteWorkflow interface {
	GetByID(context.Context, uuid.UUID) (*domainFacility.SPSController, error)
	DeleteByID(context.Context, uuid.UUID) error
}

type DeleteCommand struct {
	SPSControllerID uuid.UUID
}

func (c DeleteCommand) validate() error {
	if c.SPSControllerID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	return nil
}

type DeleteDependencies struct {
	TransactionRunner   apptransaction.Runner
	TransactionWorkflow apptransaction.Factory[DeleteWorkflow]
	HistoryBatch        HistoryBatchContext
	ProjectLinks        ProjectLinkReader
	Dispatcher          appcollaboration.CommandDispatcher
	Actor               ActorProvider
	NewID               IDGenerator
	Now                 Clock
	ReportError         ErrorReporter
}

type DeleteHandler struct {
	operation             apptransaction.Operation[DeleteWorkflow, DeleteWorkflow]
	transactionConfigured bool
	historyBatch          HistoryBatchContext
	projectLinks          ProjectLinkReader
	dispatcher            appcollaboration.CommandDispatcher
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
	reportError           ErrorReporter
}

type DeleteOutcome struct {
	Mutation       mutation.Result
	Existed        bool
	DispatchErrors []error
}

type committedDelete struct {
	change         *mutation.EntityChange
	controlCabinet uuid.UUID
	historyBatched bool
}

func NewDeleteHandler(deps DeleteDependencies) *DeleteHandler {
	newID := deps.NewID
	if newID == nil {
		newID = uuid.New
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	boundary := apptransaction.NewBoundary[DeleteWorkflow](
		deps.TransactionRunner,
		deps.TransactionWorkflow,
	)
	var noNonTransactionalWorkflow DeleteWorkflow
	operation := apptransaction.Bind(
		boundary,
		noNonTransactionalWorkflow,
		func(workflow DeleteWorkflow) DeleteWorkflow { return workflow },
	)
	return &DeleteHandler{
		operation:             operation,
		transactionConfigured: deps.TransactionRunner != nil && deps.TransactionWorkflow != nil,
		historyBatch:          deps.HistoryBatch,
		projectLinks:          deps.ProjectLinks,
		dispatcher:            deps.Dispatcher,
		actor:                 deps.Actor,
		newID:                 newID,
		now:                   now,
		reportError:           deps.ReportError,
	}
}

// Delete preserves the existing idempotent HTTP contract. Collaboration
// delivery remains best effort after the database transaction commits.
func (h *DeleteHandler) Delete(ctx context.Context, command DeleteCommand) error {
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

func (h *DeleteHandler) Execute(
	ctx context.Context,
	command DeleteCommand,
) (DeleteOutcome, error) {
	if h == nil || !h.transactionConfigured {
		return DeleteOutcome{}, ErrDeleteTransactionNotConfigured
	}
	if err := command.validate(); err != nil {
		return DeleteOutcome{}, err
	}

	operationID := h.newID()
	actorID := actorFromContext(h.actor, ctx)
	var (
		projectIDs []uuid.UUID
		scopeErr   error
	)
	if h.projectLinks != nil && h.dispatcher != nil {
		links, err := h.projectLinks.GetBySPSControllerIDs(
			ctx,
			[]uuid.UUID{command.SPSControllerID},
		)
		if err != nil {
			scopeErr = err
		} else {
			projectIDs = linkedProjectIDs(links, command.SPSControllerID)
		}
	}

	committed, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(txCtx context.Context, workflow DeleteWorkflow) (committedDelete, error) {
			return executeDeleteTransaction(
				txCtx,
				workflow,
				command,
				operationID,
				h.historyBatch,
			)
		},
	)
	if err != nil {
		return DeleteOutcome{}, err
	}

	occurredAt := h.now().UTC()
	result := mutation.Result{
		OperationID: operationID,
		ActorID:     actorID,
		OccurredAt:  occurredAt,
	}
	if committed.historyBatched {
		batchID := operationID
		result.BatchID = &batchID
	}
	if committed.change != nil {
		result.Changes = []mutation.EntityChange{*committed.change}
		result.ProjectIDs = append([]uuid.UUID(nil), projectIDs...)
	}
	outcome := DeleteOutcome{
		Mutation: result,
		Existed:  committed.change != nil,
	}
	if committed.change != nil && scopeErr != nil {
		outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
			"resolve deleted SPSController collaboration projects: %w",
			scopeErr,
		))
		return outcome, nil
	}
	if committed.change == nil || h.dispatcher == nil {
		return outcome, nil
	}

	dispatchCtx := context.WithoutCancel(ctx)
	for _, projectID := range projectIDs {
		collaborationCommand := appcollaboration.SPSControllerDeleted{
			Envelope: appcollaboration.Envelope{
				SchemaVersion: appcollaboration.SchemaVersionV1,
				EventID:       h.newID(),
				OperationID:   operationID,
				CorrelationID: operationID,
				ProjectID:     projectID,
				ActorID:       actorID,
				OccurredAt:    occurredAt,
			},
			SPSControllerID:  command.SPSControllerID,
			ControlCabinetID: committed.controlCabinet,
		}
		if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, collaborationCommand); dispatchErr != nil {
			outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
				"dispatch deleted SPSController for project %s: %w",
				projectID,
				dispatchErr,
			))
		}
	}
	return outcome, nil
}

func executeDeleteTransaction(
	ctx context.Context,
	workflow DeleteWorkflow,
	command DeleteCommand,
	operationID uuid.UUID,
	historyBatch HistoryBatchContext,
) (committedDelete, error) {
	if workflow == nil {
		return committedDelete{}, ErrDeleteTransactionNotConfigured
	}

	before, err := workflow.GetByID(ctx, command.SPSControllerID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return committedDelete{}, err
	}
	var (
		change         *mutation.EntityChange
		controlCabinet uuid.UUID
	)
	if before != nil && err == nil {
		built, buildErr := buildDeleteChange(before)
		if buildErr != nil {
			return committedDelete{}, buildErr
		}
		change = &built
		controlCabinet = before.ControlCabinetID
	}

	writeCtx := ctx
	historyBatched := historyBatch != nil
	if historyBatched {
		writeCtx = historyBatch(ctx, operationID)
	}
	if err := workflow.DeleteByID(writeCtx, command.SPSControllerID); err != nil {
		return committedDelete{}, err
	}
	return committedDelete{
		change:         change,
		controlCabinet: controlCabinet,
		historyBatched: historyBatched,
	}, nil
}
