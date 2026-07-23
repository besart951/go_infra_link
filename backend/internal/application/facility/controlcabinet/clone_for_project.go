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
	"github.com/google/uuid"
)

var ErrCloneForProjectTransactionNotConfigured = errors.New(
	"control cabinet project clone transaction is not configured",
)

// CloneForProjectWorkflow is implemented by a transaction-scoped
// ProjectFacilityLinkService. That service retains the deep-copy and descendant
// link policy while this handler owns the outer transaction and commit gate.
type CloneForProjectWorkflow interface {
	CopyControlCabinet(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) (*domainFacility.ControlCabinet, error)
}

type CloneForProjectCommand struct {
	ProjectID              uuid.UUID
	SourceControlCabinetID uuid.UUID
}

func (c CloneForProjectCommand) validate() error {
	if c.ProjectID == uuid.Nil || c.SourceControlCabinetID == uuid.Nil {
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
	ControlCabinet *domainFacility.ControlCabinet
	Mutation       mutation.Result
	DispatchErrors []error
}

type committedProjectClone struct {
	cabinet *domainFacility.ControlCabinet
	change  mutation.EntityChange
	batched bool
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
) (*domainFacility.ControlCabinet, error) {
	outcome, err := h.Execute(ctx, command)
	if err != nil {
		return nil, err
	}
	for _, dispatchErr := range outcome.DispatchErrors {
		if h.reportError != nil {
			h.reportError(dispatchErr)
		}
	}
	return outcome.ControlCabinet, nil
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
	result := mutation.Result{
		OperationID: operationID,
		ProjectIDs:  []uuid.UUID{command.ProjectID},
		ActorID:     actorID,
		OccurredAt:  occurredAt,
		Changes:     []mutation.EntityChange{committed.change},
	}
	if committed.batched {
		batchID := operationID
		result.BatchID = &batchID
	}
	outcome := CloneForProjectOutcome{
		ControlCabinet: committed.cabinet,
		Mutation:       result,
	}
	if h.dispatcher == nil {
		return outcome, nil
	}

	dispatchCtx := context.WithoutCancel(ctx)
	collaborationCommand := appcollaboration.ControlCabinetCloned{
		Envelope: appcollaboration.Envelope{
			SchemaVersion: appcollaboration.SchemaVersionV1,
			EventID:       h.newID(),
			OperationID:   operationID,
			CorrelationID: operationID,
			ProjectID:     command.ProjectID,
			ActorID:       actorID,
			OccurredAt:    occurredAt,
		},
		SourceControlCabinetID: command.SourceControlCabinetID,
		ControlCabinet:         toCollaborationState(committed.cabinet),
	}
	if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, collaborationCommand); dispatchErr != nil {
		outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
			"dispatch project-scoped cloned ControlCabinet for project %s: %w",
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
	copyEntity, err := workflow.CopyControlCabinet(
		writeCtx,
		command.ProjectID,
		command.SourceControlCabinetID,
	)
	if err != nil {
		return committedProjectClone{}, err
	}
	if copyEntity == nil || copyEntity.ID == uuid.Nil ||
		copyEntity.ID == command.SourceControlCabinetID {
		return committedProjectClone{}, domain.ErrInvalidArgument
	}
	change, err := buildCreateChange(copyEntity)
	if err != nil {
		return committedProjectClone{}, err
	}
	return committedProjectClone{
		cabinet: cloneControlCabinet(copyEntity),
		change:  change,
		batched: batched,
	}, nil
}
