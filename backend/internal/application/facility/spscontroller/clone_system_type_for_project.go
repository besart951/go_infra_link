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

var ErrCloneSystemTypeForProjectTransactionNotConfigured = errors.New(
	"SPS controller system-type project clone transaction is not configured",
)

// CloneSystemTypeForProjectWorkflow is implemented by a transaction-scoped
// ProjectFacilityLinkService. It retains number-allocation, hierarchy-copy,
// and descendant project-link policy while the application handler owns the
// outer transaction and after-commit publication gate.
type CloneSystemTypeForProjectWorkflow interface {
	RequireSourceAccess(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) error
	CopySPSControllerSystemType(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) (*domainFacility.SPSControllerSystemType, error)
}

type CloneSystemTypeForProjectCommand struct {
	ProjectID                       uuid.UUID
	SourceSPSControllerSystemTypeID uuid.UUID
}

func (c CloneSystemTypeForProjectCommand) validate() error {
	if c.ProjectID == uuid.Nil || c.SourceSPSControllerSystemTypeID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	return nil
}

type CloneSystemTypeForProjectDependencies struct {
	TransactionRunner   apptransaction.Runner
	TransactionWorkflow apptransaction.Factory[CloneSystemTypeForProjectWorkflow]
	HistoryBatch        HistoryBatchContext
	Dispatcher          appcollaboration.CommandDispatcher
	Actor               ActorProvider
	NewID               IDGenerator
	Now                 Clock
	ReportError         ErrorReporter
}

type CloneSystemTypeForProjectHandler struct {
	operation apptransaction.Operation[
		CloneSystemTypeForProjectWorkflow,
		CloneSystemTypeForProjectWorkflow,
	]
	transactionConfigured bool
	historyBatch          HistoryBatchContext
	dispatcher            appcollaboration.CommandDispatcher
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
	reportError           ErrorReporter
}

type CloneSystemTypeForProjectOutcome struct {
	SPSControllerSystemType *domainFacility.SPSControllerSystemType
	Mutation                mutation.Result
	DispatchErrors          []error
}

type committedProjectSystemTypeClone struct {
	systemType *domainFacility.SPSControllerSystemType
	change     mutation.EntityChange
	batched    bool
}

type spsControllerSystemTypeSnapshot struct {
	ID              uuid.UUID `json:"id"`
	CreatedAt       time.Time `json:"created_at"`
	UpdatedAt       time.Time `json:"updated_at"`
	Number          *int      `json:"number"`
	DocumentName    *string   `json:"document_name"`
	SPSControllerID uuid.UUID `json:"sps_controller_id"`
	SystemTypeID    uuid.UUID `json:"system_type_id"`
}

func NewCloneSystemTypeForProjectHandler(
	deps CloneSystemTypeForProjectDependencies,
) *CloneSystemTypeForProjectHandler {
	newID := deps.NewID
	if newID == nil {
		newID = uuid.New
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	boundary := apptransaction.NewBoundary[CloneSystemTypeForProjectWorkflow](
		deps.TransactionRunner,
		deps.TransactionWorkflow,
	)
	var noNonTransactionalWorkflow CloneSystemTypeForProjectWorkflow
	operation := apptransaction.Bind(
		boundary,
		noNonTransactionalWorkflow,
		func(workflow CloneSystemTypeForProjectWorkflow) CloneSystemTypeForProjectWorkflow {
			return workflow
		},
	)
	return &CloneSystemTypeForProjectHandler{
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

func (h *CloneSystemTypeForProjectHandler) CloneSystemTypeForProject(
	ctx context.Context,
	command CloneSystemTypeForProjectCommand,
) (*domainFacility.SPSControllerSystemType, error) {
	outcome, err := h.Execute(ctx, command)
	if err != nil {
		return nil, err
	}
	for _, dispatchErr := range outcome.DispatchErrors {
		if h.reportError != nil {
			h.reportError(dispatchErr)
		}
	}
	return outcome.SPSControllerSystemType, nil
}

func (h *CloneSystemTypeForProjectHandler) Execute(
	ctx context.Context,
	command CloneSystemTypeForProjectCommand,
) (CloneSystemTypeForProjectOutcome, error) {
	if h == nil || !h.transactionConfigured {
		return CloneSystemTypeForProjectOutcome{}, ErrCloneSystemTypeForProjectTransactionNotConfigured
	}
	if err := command.validate(); err != nil {
		return CloneSystemTypeForProjectOutcome{}, err
	}

	operationID := h.newID()
	eventID := h.newID()
	actorID := actorFromContext(h.actor, ctx)
	occurredAt := h.now().UTC()
	var collaborationCommand appcollaboration.Command
	committed, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(
			txCtx context.Context,
			workflow CloneSystemTypeForProjectWorkflow,
		) (committedProjectSystemTypeClone, error) {
			result, err := executeCloneSystemTypeForProjectTransaction(
				txCtx,
				workflow,
				command,
				operationID,
				h.historyBatch,
			)
			if err != nil {
				return committedProjectSystemTypeClone{}, err
			}
			collaborationCommand = appcollaboration.SPSControllerSystemTypeCloned{
				Envelope: appcollaboration.Envelope{
					SchemaVersion: appcollaboration.SchemaVersionV2,
					EventID:       eventID, OperationID: operationID, CorrelationID: operationID,
					ProjectID: command.ProjectID, ActorID: actorID, OccurredAt: occurredAt,
				},
				SourceSPSControllerSystemTypeID: command.SourceSPSControllerSystemTypeID,
				SPSControllerSystemTypeID:       result.systemType.ID,
				SPSControllerID:                 result.systemType.SPSControllerID,
			}
			if _, err := appcollaboration.EnqueueCommand(txCtx, collaborationCommand); err != nil {
				return committedProjectSystemTypeClone{}, fmt.Errorf("enqueue project-scoped SPSControllerSystemType clone: %w", err)
			}
			return result, nil
		},
	)
	if err != nil {
		return CloneSystemTypeForProjectOutcome{}, err
	}

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
	outcome := CloneSystemTypeForProjectOutcome{
		SPSControllerSystemType: committed.systemType,
		Mutation:                result,
	}
	if h.dispatcher == nil {
		return outcome, nil
	}

	dispatchCtx := context.WithoutCancel(ctx)
	if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, collaborationCommand); dispatchErr != nil {
		outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
			"dispatch project-scoped cloned SPSControllerSystemType for project %s: %w",
			command.ProjectID,
			dispatchErr,
		))
	}
	return outcome, nil
}

func executeCloneSystemTypeForProjectTransaction(
	ctx context.Context,
	workflow CloneSystemTypeForProjectWorkflow,
	command CloneSystemTypeForProjectCommand,
	operationID uuid.UUID,
	historyBatch HistoryBatchContext,
) (committedProjectSystemTypeClone, error) {
	if workflow == nil {
		return committedProjectSystemTypeClone{}, ErrCloneSystemTypeForProjectTransactionNotConfigured
	}
	if err := workflow.RequireSourceAccess(
		ctx,
		command.ProjectID,
		command.SourceSPSControllerSystemTypeID,
	); err != nil {
		return committedProjectSystemTypeClone{}, err
	}

	writeCtx := ctx
	batched := historyBatch != nil
	if batched {
		writeCtx = historyBatch(ctx, operationID)
	}
	copyEntity, err := workflow.CopySPSControllerSystemType(
		writeCtx,
		command.ProjectID,
		command.SourceSPSControllerSystemTypeID,
	)
	if err != nil {
		return committedProjectSystemTypeClone{}, err
	}
	if copyEntity == nil || copyEntity.ID == uuid.Nil ||
		copyEntity.ID == command.SourceSPSControllerSystemTypeID ||
		copyEntity.SPSControllerID == uuid.Nil || copyEntity.SystemTypeID == uuid.Nil {
		return committedProjectSystemTypeClone{}, domain.ErrInvalidArgument
	}
	change, err := buildSystemTypeCreateChange(copyEntity)
	if err != nil {
		return committedProjectSystemTypeClone{}, err
	}
	return committedProjectSystemTypeClone{
		systemType: cloneSPSControllerSystemType(copyEntity),
		change:     change,
		batched:    batched,
	}, nil
}

func buildSystemTypeCreateChange(
	after *domainFacility.SPSControllerSystemType,
) (mutation.EntityChange, error) {
	afterJSON, err := json.Marshal(toSystemTypeSnapshot(after))
	if err != nil {
		return mutation.EntityChange{}, err
	}
	parentID := after.SPSControllerID
	return mutation.EntityChange{
		EntityType: mutation.EntityTypeSPSControllerSystemType,
		EntityID:   after.ID,
		ParentID:   &parentID,
		Action:     domainHistory.ActionCreate,
		After:      json.RawMessage(afterJSON),
	}, nil
}

func toSystemTypeSnapshot(
	entity *domainFacility.SPSControllerSystemType,
) spsControllerSystemTypeSnapshot {
	if entity == nil {
		return spsControllerSystemTypeSnapshot{}
	}
	return spsControllerSystemTypeSnapshot{
		ID:              entity.ID,
		CreatedAt:       entity.CreatedAt,
		UpdatedAt:       entity.UpdatedAt,
		Number:          clonePointer(entity.Number),
		DocumentName:    clonePointer(entity.DocumentName),
		SPSControllerID: entity.SPSControllerID,
		SystemTypeID:    entity.SystemTypeID,
	}
}

func cloneSPSControllerSystemType(
	entity *domainFacility.SPSControllerSystemType,
) *domainFacility.SPSControllerSystemType {
	if entity == nil {
		return nil
	}
	clone := *entity
	clone.Number = clonePointer(entity.Number)
	clone.DocumentName = clonePointer(entity.DocumentName)
	clone.FieldDevices = append([]domainFacility.FieldDevice(nil), entity.FieldDevices...)
	return &clone
}
