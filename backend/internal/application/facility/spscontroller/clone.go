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

var ErrCloneTransactionNotConfigured = errors.New(
	"SPS controller clone transaction is not configured",
)

// CloneWorkflow keeps the mature HierarchyCopier behind a narrow,
// transaction-scoped application port. CopyByID owns the deep-copy policy;
// GetByID supplies the authoritative committed root projection.
type CloneWorkflow interface {
	CopyByID(context.Context, uuid.UUID) (*domainFacility.SPSController, error)
	GetByID(context.Context, uuid.UUID) (*domainFacility.SPSController, error)
}

type CloneCommand struct {
	SourceSPSControllerID uuid.UUID
}

func (c CloneCommand) validate() error {
	if c.SourceSPSControllerID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	return nil
}

type CloneDependencies struct {
	TransactionRunner   apptransaction.Runner
	TransactionWorkflow apptransaction.Factory[CloneWorkflow]
	HistoryBatch        HistoryBatchContext
	ProjectLinks        ProjectLinkReader
	Dispatcher          appcollaboration.CommandDispatcher
	Actor               ActorProvider
	NewID               IDGenerator
	Now                 Clock
	ReportError         ErrorReporter
}

type CloneHandler struct {
	operation             apptransaction.Operation[CloneWorkflow, CloneWorkflow]
	transactionConfigured bool
	historyBatch          HistoryBatchContext
	projectLinks          ProjectLinkReader
	dispatcher            appcollaboration.CommandDispatcher
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
	reportError           ErrorReporter
}

type CloneOutcome struct {
	SPSController  *domainFacility.SPSController
	Mutation       mutation.Result
	DispatchErrors []error
}

type committedClone struct {
	controller *domainFacility.SPSController
	change     mutation.EntityChange
	batched    bool
}

func NewCloneHandler(deps CloneDependencies) *CloneHandler {
	newID := deps.NewID
	if newID == nil {
		newID = uuid.New
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	boundary := apptransaction.NewBoundary[CloneWorkflow](
		deps.TransactionRunner,
		deps.TransactionWorkflow,
	)
	var noNonTransactionalWorkflow CloneWorkflow
	operation := apptransaction.Bind(
		boundary,
		noNonTransactionalWorkflow,
		func(workflow CloneWorkflow) CloneWorkflow { return workflow },
	)
	return &CloneHandler{
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

func (h *CloneHandler) Clone(
	ctx context.Context,
	command CloneCommand,
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

func (h *CloneHandler) Execute(
	ctx context.Context,
	command CloneCommand,
) (CloneOutcome, error) {
	if h == nil || !h.transactionConfigured {
		return CloneOutcome{}, ErrCloneTransactionNotConfigured
	}
	if err := command.validate(); err != nil {
		return CloneOutcome{}, err
	}

	operationID := h.newID()
	actorID := actorFromContext(h.actor, ctx)
	committed, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(txCtx context.Context, workflow CloneWorkflow) (committedClone, error) {
			return executeCloneTransaction(
				txCtx,
				workflow,
				command,
				operationID,
				h.historyBatch,
			)
		},
	)
	if err != nil {
		return CloneOutcome{}, err
	}

	occurredAt := h.now().UTC()
	result := mutation.Result{
		OperationID: operationID,
		ActorID:     actorID,
		OccurredAt:  occurredAt,
		Changes:     []mutation.EntityChange{committed.change},
	}
	if committed.batched {
		batchID := operationID
		result.BatchID = &batchID
	}
	outcome := CloneOutcome{
		SPSController: committed.controller,
		Mutation:      result,
	}
	if h.projectLinks == nil || h.dispatcher == nil {
		return outcome, nil
	}

	dispatchCtx := context.WithoutCancel(ctx)
	links, err := h.projectLinks.GetBySPSControllerIDs(
		dispatchCtx,
		[]uuid.UUID{committed.controller.ID},
	)
	if err != nil {
		outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
			"resolve cloned SPSController collaboration projects: %w",
			err,
		))
		return outcome, nil
	}

	projectIDs := linkedProjectIDs(links, committed.controller.ID)
	outcome.Mutation.ProjectIDs = append([]uuid.UUID(nil), projectIDs...)
	state := toCollaborationState(committed.controller)
	for _, projectID := range projectIDs {
		collaborationCommand := appcollaboration.SPSControllerCloned{
			Envelope: appcollaboration.Envelope{
				SchemaVersion: appcollaboration.SchemaVersionV1,
				EventID:       h.newID(),
				OperationID:   operationID,
				CorrelationID: operationID,
				ProjectID:     projectID,
				ActorID:       actorID,
				OccurredAt:    occurredAt,
			},
			SourceSPSControllerID: command.SourceSPSControllerID,
			SPSController:         state,
		}
		if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, collaborationCommand); dispatchErr != nil {
			outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
				"dispatch cloned SPSController for project %s: %w",
				projectID,
				dispatchErr,
			))
		}
	}
	return outcome, nil
}

func executeCloneTransaction(
	ctx context.Context,
	workflow CloneWorkflow,
	command CloneCommand,
	operationID uuid.UUID,
	historyBatch HistoryBatchContext,
) (committedClone, error) {
	if workflow == nil {
		return committedClone{}, ErrCloneTransactionNotConfigured
	}

	writeCtx := ctx
	batched := historyBatch != nil
	if batched {
		writeCtx = historyBatch(ctx, operationID)
	}
	copyEntity, err := workflow.CopyByID(writeCtx, command.SourceSPSControllerID)
	if err != nil {
		return committedClone{}, err
	}
	if copyEntity == nil || copyEntity.ID == uuid.Nil ||
		copyEntity.ID == command.SourceSPSControllerID {
		return committedClone{}, domain.ErrInvalidArgument
	}
	after, err := workflow.GetByID(writeCtx, copyEntity.ID)
	if err != nil {
		return committedClone{}, err
	}
	if after == nil || after.ID != copyEntity.ID {
		return committedClone{}, domain.ErrNotFound
	}
	change, err := buildCreateChange(after)
	if err != nil {
		return committedClone{}, err
	}
	return committedClone{
		controller: cloneSPSController(after),
		change:     change,
		batched:    batched,
	}, nil
}
