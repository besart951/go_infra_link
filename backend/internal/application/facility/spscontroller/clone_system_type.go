package spscontroller

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
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
// transaction-scoped application port. Project recipients derive from the
// owning SPSController because SPSControllerSystemType has no direct
// project-link table.
type CloneSystemTypeWorkflow interface {
	CopyByID(context.Context, uuid.UUID) (*domainFacility.SPSControllerSystemType, error)
	GetByID(context.Context, uuid.UUID) (*domainFacility.SPSControllerSystemType, error)
	GetOwningProjectIDs(context.Context, uuid.UUID) ([]uuid.UUID, error)
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
	Dispatcher          appcollaboration.CommandDispatcher
	Actor               ActorProvider
	NewID               IDGenerator
	Now                 Clock
	ReportError         ErrorReporter
}

type CloneSystemTypeHandler struct {
	operation apptransaction.Operation[
		CloneSystemTypeWorkflow,
		CloneSystemTypeWorkflow,
	]
	transactionConfigured bool
	historyBatch          HistoryBatchContext
	dispatcher            appcollaboration.CommandDispatcher
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
	reportError           ErrorReporter
}

type CloneSystemTypeOutcome struct {
	SPSControllerSystemType *domainFacility.SPSControllerSystemType
	Mutation                mutation.Result
	DispatchErrors          []error
}

type committedSystemTypeClone struct {
	systemType *domainFacility.SPSControllerSystemType
	projectIDs []uuid.UUID
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
		dispatcher:            deps.Dispatcher,
		actor:                 deps.Actor,
		newID:                 newID,
		now:                   now,
		reportError:           deps.ReportError,
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
	for _, dispatchErr := range outcome.DispatchErrors {
		if h.reportError != nil {
			h.reportError(dispatchErr)
		}
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
	occurredAt := h.now().UTC()
	var collaborationCommands []appcollaboration.Command
	committed, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(
			txCtx context.Context,
			workflow CloneSystemTypeWorkflow,
		) (committedSystemTypeClone, error) {
			result, err := executeCloneSystemTypeTransaction(
				txCtx,
				workflow,
				command,
				operationID,
				h.historyBatch,
			)
			if err != nil {
				return committedSystemTypeClone{}, err
			}
			state := result.systemType
			for _, projectID := range result.projectIDs {
				collaborationCommand := appcollaboration.SPSControllerSystemTypeCloned{
					Envelope: appcollaboration.Envelope{
						SchemaVersion: appcollaboration.SchemaVersionV2,
						EventID:       h.newID(),
						OperationID:   operationID,
						CorrelationID: operationID,
						ProjectID:     projectID,
						ActorID:       actorID,
						OccurredAt:    occurredAt,
					},
					SourceSPSControllerSystemTypeID: command.SourceSPSControllerSystemTypeID,
					SPSControllerSystemTypeID:       state.ID,
					SPSControllerID:                 state.SPSControllerID,
				}
				if _, err := appcollaboration.EnqueueCommand(
					txCtx,
					collaborationCommand,
				); err != nil {
					return committedSystemTypeClone{}, fmt.Errorf(
						"enqueue global SPSControllerSystemType clone for project %s: %w",
						projectID,
						err,
					)
				}
				collaborationCommands = append(
					collaborationCommands,
					collaborationCommand,
				)
			}
			return result, nil
		},
	)
	if err != nil {
		return CloneSystemTypeOutcome{}, err
	}

	result := mutation.Result{
		OperationID: operationID,
		ActorID:     actorID,
		OccurredAt:  occurredAt,
		ProjectIDs:  append([]uuid.UUID(nil), committed.projectIDs...),
		Changes:     []mutation.EntityChange{committed.change},
	}
	if committed.batched {
		batchID := operationID
		result.BatchID = &batchID
	}
	outcome := CloneSystemTypeOutcome{
		SPSControllerSystemType: committed.systemType,
		Mutation:                result,
	}
	if h.dispatcher == nil {
		return outcome, nil
	}
	dispatchCtx := context.WithoutCancel(ctx)
	for _, collaborationCommand := range collaborationCommands {
		if dispatchErr := h.dispatcher.Dispatch(
			dispatchCtx,
			collaborationCommand,
		); dispatchErr != nil {
			outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
				"dispatch global SPSControllerSystemType clone: %w",
				dispatchErr,
			))
		}
	}
	return outcome, nil
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
	projectIDs, err := workflow.GetOwningProjectIDs(
		writeCtx,
		after.SPSControllerID,
	)
	if err != nil {
		return committedSystemTypeClone{}, err
	}
	return committedSystemTypeClone{
		systemType: cloneSPSControllerSystemType(after),
		projectIDs: normalizeProjectIDs(projectIDs),
		change:     change,
		batched:    batched,
	}, nil
}

func normalizeProjectIDs(projectIDs []uuid.UUID) []uuid.UUID {
	set := make(map[uuid.UUID]struct{}, len(projectIDs))
	for _, projectID := range projectIDs {
		if projectID != uuid.Nil {
			set[projectID] = struct{}{}
		}
	}
	normalized := make([]uuid.UUID, 0, len(set))
	for projectID := range set {
		normalized = append(normalized, projectID)
	}
	sort.Slice(normalized, func(i, j int) bool {
		return normalized[i].String() < normalized[j].String()
	})
	return normalized
}
