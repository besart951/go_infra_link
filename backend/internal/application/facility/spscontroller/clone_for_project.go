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

var ErrCloneForProjectTransactionNotConfigured = errors.New(
	"SPS controller project clone transaction is not configured",
)

// CloneForProjectWorkflow is implemented first by the transaction-scoped
// ProjectFacilityLinkService. That service retains deep-copy and link policy;
// the application command owns the outer transaction and commit gate.
type CloneForProjectWorkflow interface {
	CopySPSController(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) (*domainFacility.SPSController, error)
}

type CloneForProjectCommand struct {
	ProjectID             uuid.UUID
	SourceSPSControllerID uuid.UUID
}

func (c CloneForProjectCommand) validate() error {
	if c.ProjectID == uuid.Nil || c.SourceSPSControllerID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	return nil
}

type CloneForProjectDependencies struct {
	TransactionRunner   apptransaction.Runner
	TransactionWorkflow apptransaction.Factory[CloneForProjectWorkflow]
	HistoryBatch        HistoryBatchContext
	Dispatcher          appcollaboration.CommandDispatcher
	Actor               ActorProvider
	NewID               IDGenerator
	Now                 Clock
	ReportError         ErrorReporter
}

type CloneForProjectHandler struct {
	operation             apptransaction.Operation[CloneForProjectWorkflow, CloneForProjectWorkflow]
	transactionConfigured bool
	historyBatch          HistoryBatchContext
	dispatcher            appcollaboration.CommandDispatcher
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
	reportError           ErrorReporter
}

type CloneForProjectOutcome struct {
	SPSController  *domainFacility.SPSController
	Mutation       mutation.Result
	DispatchErrors []error
}

type committedProjectClone struct {
	controller *domainFacility.SPSController
	change     mutation.EntityChange
	batched    bool
}

func NewCloneForProjectHandler(
	deps CloneForProjectDependencies,
) *CloneForProjectHandler {
	newID := deps.NewID
	if newID == nil {
		newID = uuid.New
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	boundary := apptransaction.NewBoundary[CloneForProjectWorkflow](
		deps.TransactionRunner,
		deps.TransactionWorkflow,
	)
	var noNonTransactionalWorkflow CloneForProjectWorkflow
	operation := apptransaction.Bind(
		boundary,
		noNonTransactionalWorkflow,
		func(workflow CloneForProjectWorkflow) CloneForProjectWorkflow { return workflow },
	)
	return &CloneForProjectHandler{
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

func (h *CloneForProjectHandler) CloneForProject(
	ctx context.Context,
	command CloneForProjectCommand,
) (*domainFacility.SPSController, error) {
	outcome, err := h.Execute(ctx, command)
	if err != nil {
		return nil, err
	}
	for _, dispatchErr := range outcome.DispatchErrors {
		if h.reportError != nil {
			h.reportError(dispatchErr)
		}
	}
	return outcome.SPSController, nil
}

func (h *CloneForProjectHandler) Execute(
	ctx context.Context,
	command CloneForProjectCommand,
) (CloneForProjectOutcome, error) {
	if h == nil || !h.transactionConfigured {
		return CloneForProjectOutcome{}, ErrCloneForProjectTransactionNotConfigured
	}
	if err := command.validate(); err != nil {
		return CloneForProjectOutcome{}, err
	}

	operationID := h.newID()
	actorID := actorFromContext(h.actor, ctx)
	committed, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(
			txCtx context.Context,
			workflow CloneForProjectWorkflow,
		) (committedProjectClone, error) {
			return executeCloneForProjectTransaction(
				txCtx,
				workflow,
				command,
				operationID,
				h.historyBatch,
			)
		},
	)
	if err != nil {
		return CloneForProjectOutcome{}, err
	}

	occurredAt := h.now().UTC()
	batchID := operationID
	result := mutation.Result{
		OperationID: operationID,
		ProjectIDs:  []uuid.UUID{command.ProjectID},
		ActorID:     actorID,
		OccurredAt:  occurredAt,
		Changes:     []mutation.EntityChange{committed.change},
	}
	if committed.batched {
		result.BatchID = &batchID
	}
	outcome := CloneForProjectOutcome{
		SPSController: committed.controller,
		Mutation:      result,
	}
	if h.dispatcher == nil {
		return outcome, nil
	}

	dispatchCtx := context.WithoutCancel(ctx)
	collaborationCommand := appcollaboration.SPSControllerCloned{
		Envelope: appcollaboration.Envelope{
			SchemaVersion: appcollaboration.SchemaVersionV1,
			EventID:       h.newID(),
			OperationID:   operationID,
			CorrelationID: operationID,
			ProjectID:     command.ProjectID,
			ActorID:       actorID,
			OccurredAt:    occurredAt,
		},
		SourceSPSControllerID: command.SourceSPSControllerID,
		SPSController:         toCollaborationState(committed.controller),
	}
	if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, collaborationCommand); dispatchErr != nil {
		outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
			"dispatch project-scoped cloned SPSController for project %s: %w",
			command.ProjectID,
			dispatchErr,
		))
	}
	return outcome, nil
}

func executeCloneForProjectTransaction(
	ctx context.Context,
	workflow CloneForProjectWorkflow,
	command CloneForProjectCommand,
	operationID uuid.UUID,
	historyBatch HistoryBatchContext,
) (committedProjectClone, error) {
	if workflow == nil {
		return committedProjectClone{}, ErrCloneForProjectTransactionNotConfigured
	}

	writeCtx := ctx
	batched := historyBatch != nil
	if batched {
		writeCtx = historyBatch(ctx, operationID)
	}
	copyEntity, err := workflow.CopySPSController(
		writeCtx,
		command.ProjectID,
		command.SourceSPSControllerID,
	)
	if err != nil {
		return committedProjectClone{}, err
	}
	if copyEntity == nil || copyEntity.ID == uuid.Nil ||
		copyEntity.ID == command.SourceSPSControllerID {
		return committedProjectClone{}, domain.ErrInvalidArgument
	}
	change, err := buildCreateChange(copyEntity)
	if err != nil {
		return committedProjectClone{}, err
	}
	return committedProjectClone{
		controller: cloneSPSController(copyEntity),
		change:     change,
		batched:    batched,
	}, nil
}
