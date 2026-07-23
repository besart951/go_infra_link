package spscontroller

import (
	"context"
	"errors"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/application/facility/mutation"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

var ErrCloneSystemTypeTransactionNotConfigured = errors.New(
	"SPS controller system-type clone transaction is not configured",
)

// CloneSystemTypeWorkflow keeps the existing HierarchyCopier behind a narrow,
// transaction-scoped application port. The global route deliberately has no
// project-recipient port because its established collaboration behavior is
// silence and SPSControllerSystemType has no direct project-link table.
type CloneSystemTypeWorkflow interface {
	CopyByID(context.Context, uuid.UUID) (*domainFacility.SPSControllerSystemType, error)
	GetByID(context.Context, uuid.UUID) (*domainFacility.SPSControllerSystemType, error)
}

type CloneSystemTypeCommand struct {
	SourceSPSControllerSystemTypeID uuid.UUID
}

func (c CloneSystemTypeCommand) validate() error {
	if c.SourceSPSControllerSystemTypeID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	return nil
}

type CloneSystemTypeDependencies struct {
	TransactionRunner   apptransaction.Runner
	TransactionWorkflow apptransaction.Factory[CloneSystemTypeWorkflow]
	HistoryBatch        HistoryBatchContext
	Actor               ActorProvider
	NewID               IDGenerator
	Now                 Clock
}

type CloneSystemTypeHandler struct {
	operation apptransaction.Operation[
		CloneSystemTypeWorkflow,
		CloneSystemTypeWorkflow,
	]
	transactionConfigured bool
	historyBatch          HistoryBatchContext
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
}

type CloneSystemTypeOutcome struct {
	SPSControllerSystemType *domainFacility.SPSControllerSystemType
	Mutation                mutation.Result
}

type committedSystemTypeClone struct {
	systemType *domainFacility.SPSControllerSystemType
	change     mutation.EntityChange
	batched    bool
}

func NewCloneSystemTypeHandler(
	deps CloneSystemTypeDependencies,
) *CloneSystemTypeHandler {
	newID := deps.NewID
	if newID == nil {
		newID = uuid.New
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	boundary := apptransaction.NewBoundary[CloneSystemTypeWorkflow](
		deps.TransactionRunner,
		deps.TransactionWorkflow,
	)
	var noNonTransactionalWorkflow CloneSystemTypeWorkflow
	operation := apptransaction.Bind(
		boundary,
		noNonTransactionalWorkflow,
		func(workflow CloneSystemTypeWorkflow) CloneSystemTypeWorkflow {
			return workflow
		},
	)
	return &CloneSystemTypeHandler{
		operation:             operation,
		transactionConfigured: deps.TransactionRunner != nil && deps.TransactionWorkflow != nil,
		historyBatch:          deps.HistoryBatch,
		actor:                 deps.Actor,
		newID:                 newID,
		now:                   now,
	}
}

func (h *CloneSystemTypeHandler) CloneSystemType(
	ctx context.Context,
	command CloneSystemTypeCommand,
) (*domainFacility.SPSControllerSystemType, error) {
	outcome, err := h.Execute(ctx, command)
	if err != nil {
		return nil, err
	}
	return outcome.SPSControllerSystemType, nil
}

func (h *CloneSystemTypeHandler) Execute(
	ctx context.Context,
	command CloneSystemTypeCommand,
) (CloneSystemTypeOutcome, error) {
	if h == nil || !h.transactionConfigured {
		return CloneSystemTypeOutcome{}, ErrCloneSystemTypeTransactionNotConfigured
	}
	if err := command.validate(); err != nil {
		return CloneSystemTypeOutcome{}, err
	}

	operationID := h.newID()
	actorID := actorFromContext(h.actor, ctx)
	committed, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(
			txCtx context.Context,
			workflow CloneSystemTypeWorkflow,
		) (committedSystemTypeClone, error) {
			return executeCloneSystemTypeTransaction(
				txCtx,
				workflow,
				command,
				operationID,
				h.historyBatch,
			)
		},
	)
	if err != nil {
		return CloneSystemTypeOutcome{}, err
	}

	result := mutation.Result{
		OperationID: operationID,
		ActorID:     actorID,
		OccurredAt:  h.now().UTC(),
		Changes:     []mutation.EntityChange{committed.change},
	}
	if committed.batched {
		batchID := operationID
		result.BatchID = &batchID
	}
	return CloneSystemTypeOutcome{
		SPSControllerSystemType: committed.systemType,
		Mutation:                result,
	}, nil
}

func executeCloneSystemTypeTransaction(
	ctx context.Context,
	workflow CloneSystemTypeWorkflow,
	command CloneSystemTypeCommand,
	operationID uuid.UUID,
	historyBatch HistoryBatchContext,
) (committedSystemTypeClone, error) {
	if workflow == nil {
		return committedSystemTypeClone{}, ErrCloneSystemTypeTransactionNotConfigured
	}

	writeCtx := ctx
	batched := historyBatch != nil
	if batched {
		writeCtx = historyBatch(ctx, operationID)
	}
	copyEntity, err := workflow.CopyByID(
		writeCtx,
		command.SourceSPSControllerSystemTypeID,
	)
	if err != nil {
		return committedSystemTypeClone{}, err
	}
	if copyEntity == nil || copyEntity.ID == uuid.Nil ||
		copyEntity.ID == command.SourceSPSControllerSystemTypeID {
		return committedSystemTypeClone{}, domain.ErrInvalidArgument
	}
	after, err := workflow.GetByID(writeCtx, copyEntity.ID)
	if err != nil {
		return committedSystemTypeClone{}, err
	}
	if after == nil || after.ID != copyEntity.ID ||
		after.SPSControllerID == uuid.Nil || after.SystemTypeID == uuid.Nil {
		return committedSystemTypeClone{}, domain.ErrNotFound
	}
	change, err := buildSystemTypeCreateChange(after)
	if err != nil {
		return committedSystemTypeClone{}, err
	}
	return committedSystemTypeClone{
		systemType: cloneSPSControllerSystemType(after),
		change:     change,
		batched:    batched,
	}, nil
}
