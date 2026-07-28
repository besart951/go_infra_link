package fielddevice

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
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

var ErrDeleteTransactionNotConfigured = errors.New("field device delete transaction is not configured")

// DeleteWorkflow is the transaction-scoped Interface for a FieldDevice delete.
type DeleteWorkflow interface {
	GetByID(context.Context, uuid.UUID) (*domainFacility.FieldDevice, error)
	DeleteByID(context.Context, uuid.UUID) error
}

type transactionalDeleteProjectLinks interface {
	DeleteWorkflow
	GetByFieldDeviceIDs(context.Context, []uuid.UUID) ([]*domainProject.ProjectFieldDevice, error)
}

type DeleteCommand struct {
	FieldDeviceID uuid.UUID
}

func (c DeleteCommand) validate() error {
	if c.FieldDeviceID == uuid.Nil {
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
	parentID       uuid.UUID
	historyBatched bool
	projectIDs     []uuid.UUID
	commands       []appcollaboration.Command
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

// Delete preserves the existing HTTP contract: deleting a missing row remains
// successful. Collaboration delivery is best effort after commit.
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
	occurredAt := h.now().UTC()
	var (
		projectIDs []uuid.UUID
		scopeErr   error
	)
	if h.projectLinks != nil && h.dispatcher != nil {
		links, err := h.projectLinks.GetByFieldDeviceIDs(
			ctx,
			[]uuid.UUID{command.FieldDeviceID},
		)
		if err != nil {
			scopeErr = err
		} else {
			grouped := groupLinkedFieldDevices(links, []uuid.UUID{command.FieldDeviceID})
			projectIDs = sortedProjectIDs(grouped)
		}
	}
	committed, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(txCtx context.Context, workflow DeleteWorkflow) (committedDelete, error) {
			if links, ok := workflow.(transactionalDeleteProjectLinks); ok {
				resolved, err := links.GetByFieldDeviceIDs(txCtx, []uuid.UUID{command.FieldDeviceID})
				if err != nil {
					return committedDelete{}, fmt.Errorf("resolve deleted FieldDevice collaboration projects: %w", err)
				}
				projectIDs = sortedProjectIDs(groupLinkedFieldDevices(resolved, []uuid.UUID{command.FieldDeviceID}))
				scopeErr = nil
			}
			result, err := executeDeleteTransaction(
				txCtx,
				workflow,
				command,
				operationID,
				h.historyBatch,
			)
			if err != nil || result.change == nil {
				return result, err
			}
			result.projectIDs = append([]uuid.UUID(nil), projectIDs...)
			for _, projectID := range projectIDs {
				event := appcollaboration.FieldDeviceDeleted{
					Envelope: appcollaboration.Envelope{
						SchemaVersion: appcollaboration.SchemaVersionV2,
						EventID:       h.newID(), OperationID: operationID, CorrelationID: operationID,
						ProjectID: projectID, ActorID: actorID, OccurredAt: occurredAt,
					},
					FieldDeviceID: command.FieldDeviceID, SPSControllerSystemTypeID: result.parentID,
				}
				if _, err := appcollaboration.EnqueueCommand(txCtx, event); err != nil {
					return committedDelete{}, fmt.Errorf("enqueue deleted FieldDevice for project %s: %w", projectID, err)
				}
				result.commands = append(result.commands, event)
			}
			return result, nil
		},
	)
	if err != nil {
		return DeleteOutcome{}, err
	}

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
		result.ProjectIDs = append([]uuid.UUID(nil), committed.projectIDs...)
	}
	outcome := DeleteOutcome{
		Mutation: result,
		Existed:  committed.change != nil,
	}
	if committed.change != nil && scopeErr != nil {
		outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
			"resolve deleted FieldDevice collaboration projects: %w",
			scopeErr,
		))
		return outcome, nil
	}
	if committed.change == nil || h.dispatcher == nil {
		return outcome, nil
	}

	dispatchCtx := context.WithoutCancel(ctx)
	for _, collaborationCommand := range committed.commands {
		envelope, _ := appcollaboration.CommandEnvelope(collaborationCommand)
		if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, collaborationCommand); dispatchErr != nil {
			outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
				"dispatch deleted FieldDevice for project %s: %w",
				envelope.ProjectID,
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

	before, err := workflow.GetByID(ctx, command.FieldDeviceID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return committedDelete{}, err
	}

	var (
		change   *mutation.EntityChange
		parentID uuid.UUID
	)
	if before != nil && err == nil {
		built, buildErr := buildDeleteChange(before)
		if buildErr != nil {
			return committedDelete{}, buildErr
		}
		change = &built
		parentID = before.SPSControllerSystemTypeID
	}

	writeCtx := ctx
	historyBatched := historyBatch != nil
	if historyBatched {
		writeCtx = historyBatch(ctx, operationID)
	}
	if err := workflow.DeleteByID(writeCtx, command.FieldDeviceID); err != nil {
		return committedDelete{}, err
	}

	return committedDelete{
		change:         change,
		parentID:       parentID,
		historyBatched: historyBatched,
	}, nil
}
