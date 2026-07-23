package spscontroller

import (
	"context"
	"encoding/json"
	"errors"
	"time"

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
	Actor               ActorProvider
	NewID               IDGenerator
	Now                 Clock
}

type DeleteSystemTypeHandler struct {
	operation apptransaction.Operation[
		DeleteSystemTypeWorkflow,
		DeleteSystemTypeWorkflow,
	]
	transactionConfigured bool
	historyBatch          HistoryBatchContext
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
}

type DeleteSystemTypeOutcome struct {
	Mutation mutation.Result
	Existed  bool
}

type committedSystemTypeDelete struct {
	change  *mutation.EntityChange
	batched bool
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
		actor:                 deps.Actor,
		newID:                 newID,
		now:                   now,
	}
}

// DeleteSystemType preserves the existing idempotent 204 transport contract.
// The global endpoint remains silent on the collaboration stream.
func (h *DeleteSystemTypeHandler) DeleteSystemType(
	ctx context.Context,
	command DeleteSystemTypeCommand,
) error {
	_, err := h.Execute(ctx, command)
	return err
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
	committed, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(
			txCtx context.Context,
			workflow DeleteSystemTypeWorkflow,
		) (committedSystemTypeDelete, error) {
			return executeDeleteSystemTypeTransaction(
				txCtx,
				workflow,
				command,
				operationID,
				h.historyBatch,
			)
		},
	)
	if err != nil {
		return DeleteSystemTypeOutcome{}, err
	}

	result := mutation.Result{
		OperationID: operationID,
		ActorID:     actorID,
		OccurredAt:  h.now().UTC(),
	}
	if committed.batched {
		batchID := operationID
		result.BatchID = &batchID
	}
	if committed.change != nil {
		result.Changes = []mutation.EntityChange{*committed.change}
	}
	return DeleteSystemTypeOutcome{
		Mutation: result,
		Existed:  committed.change != nil,
	}, nil
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
