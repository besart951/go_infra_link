package controlcabinet

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

var ErrDeleteTransactionNotConfigured = errors.New(
	"control cabinet delete transaction is not configured",
)

// DeleteWorkflow intentionally loads only the cabinet root. Existing database
// cascade behavior remains the compatibility policy; descendant audit
// completeness requires a separate bounded, set-based delete design.
type DeleteWorkflow interface {
	GetByID(context.Context, uuid.UUID) (*domainFacility.ControlCabinet, error)
	DeleteByID(context.Context, uuid.UUID) error
}

type transactionalDeleteProjectLinks interface {
	DeleteWorkflow
	GetByControlCabinetIDs(context.Context, []uuid.UUID) ([]*domainProject.ProjectControlCabinet, error)
}

type DeleteCommand struct {
	ControlCabinetID uuid.UUID
}

func (c DeleteCommand) validate() error {
	if c.ControlCabinetID == uuid.Nil {
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
	buildingID     uuid.UUID
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
	occurredAt := h.now().UTC()
	var (
		projectIDs []uuid.UUID
		scopeErr   error
	)
	if h.projectLinks != nil && h.dispatcher != nil {
		links, err := h.projectLinks.GetByControlCabinetIDs(
			ctx,
			[]uuid.UUID{command.ControlCabinetID},
		)
		if err != nil {
			scopeErr = err
		} else {
			projectIDs = linkedProjectIDs(links, command.ControlCabinetID)
		}
	}

	committed, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(txCtx context.Context, workflow DeleteWorkflow) (committedDelete, error) {
			if links, ok := workflow.(transactionalDeleteProjectLinks); ok {
				resolved, err := links.GetByControlCabinetIDs(txCtx, []uuid.UUID{command.ControlCabinetID})
				if err != nil {
					return committedDelete{}, fmt.Errorf("resolve deleted ControlCabinet collaboration projects: %w", err)
				}
				projectIDs = linkedProjectIDs(resolved, command.ControlCabinetID)
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
				event := appcollaboration.ControlCabinetDeleted{
					Envelope: appcollaboration.Envelope{
						SchemaVersion: appcollaboration.SchemaVersionV2,
						EventID:       h.newID(), OperationID: operationID, CorrelationID: operationID,
						ProjectID: projectID, ActorID: actorID, OccurredAt: occurredAt,
					},
					ControlCabinetID: command.ControlCabinetID, BuildingID: result.buildingID,
				}
				if _, err := appcollaboration.EnqueueCommand(txCtx, event); err != nil {
					return committedDelete{}, fmt.Errorf("enqueue deleted ControlCabinet for project %s: %w", projectID, err)
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
			"resolve deleted ControlCabinet collaboration projects: %w",
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
				"dispatch deleted ControlCabinet for project %s: %w",
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

	before, err := workflow.GetByID(ctx, command.ControlCabinetID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return committedDelete{}, err
	}
	var (
		change     *mutation.EntityChange
		buildingID uuid.UUID
	)
	if before != nil && err == nil {
		built, buildErr := buildDeleteChange(before)
		if buildErr != nil {
			return committedDelete{}, buildErr
		}
		change = &built
		buildingID = before.BuildingID
	}

	writeCtx := ctx
	historyBatched := historyBatch != nil
	if historyBatched {
		writeCtx = historyBatch(ctx, operationID)
	}
	if err := workflow.DeleteByID(writeCtx, command.ControlCabinetID); err != nil {
		return committedDelete{}, err
	}
	return committedDelete{
		change:         change,
		buildingID:     buildingID,
		historyBatched: historyBatched,
	}, nil
}
