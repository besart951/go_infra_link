package controlcabinet

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

var ErrCloneTransactionNotConfigured = errors.New(
	"control cabinet clone transaction is not configured",
)

// CloneWorkflow keeps the existing HierarchyCopier behind a narrow,
// transaction-scoped port. CopyByID owns the deep-copy policy; GetByID returns
// the authoritative copied root used by history, the HTTP response, and
// collaboration.
type CloneWorkflow interface {
	CopyByID(context.Context, uuid.UUID) (*domainFacility.ControlCabinet, error)
	GetByID(context.Context, uuid.UUID) (*domainFacility.ControlCabinet, error)
	GetSourceProjectIDs(context.Context, uuid.UUID) ([]uuid.UUID, error)
	AssignCopyToProject(context.Context, uuid.UUID, uuid.UUID) error
}

type CloneCommand struct {
	SourceControlCabinetID uuid.UUID
}

func (c CloneCommand) validate() error {
	if c.SourceControlCabinetID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	return nil
}

type CloneDependencies struct {
	TransactionRunner   apptransaction.Runner
	TransactionWorkflow apptransaction.Factory[CloneWorkflow]
	HistoryBatch        HistoryBatchContext
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
	dispatcher            appcollaboration.CommandDispatcher
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
	reportError           ErrorReporter
}

type CloneOutcome struct {
	ControlCabinet *domainFacility.ControlCabinet
	Mutation       mutation.Result
	DispatchErrors []error
}

type committedClone struct {
	cabinet    *domainFacility.ControlCabinet
	projectIDs []uuid.UUID
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
	occurredAt := h.now().UTC()
	var collaborationCommands []appcollaboration.Command
	committed, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(txCtx context.Context, workflow CloneWorkflow) (committedClone, error) {
			result, err := executeCloneTransaction(
				txCtx,
				workflow,
				command,
				operationID,
				h.historyBatch,
			)
			if err != nil {
				return committedClone{}, err
			}
			state := toCollaborationState(result.cabinet)
			for _, projectID := range result.projectIDs {
				collaborationCommand := appcollaboration.ControlCabinetCloned{
					Envelope: appcollaboration.Envelope{
						SchemaVersion: appcollaboration.SchemaVersionV2,
						EventID:       h.newID(),
						OperationID:   operationID,
						CorrelationID: operationID,
						ProjectID:     projectID,
						ActorID:       actorID,
						OccurredAt:    occurredAt,
					},
					SourceControlCabinetID: command.SourceControlCabinetID,
					ControlCabinet:         state,
				}
				if _, err := appcollaboration.EnqueueCommand(
					txCtx,
					collaborationCommand,
				); err != nil {
					return committedClone{}, fmt.Errorf(
						"enqueue inherited ControlCabinet clone for project %s: %w",
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
		return CloneOutcome{}, err
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
	outcome := CloneOutcome{
		ControlCabinet: committed.cabinet,
		Mutation:       result,
	}
	if h.dispatcher == nil {
		return outcome, nil
	}

	dispatchCtx := context.WithoutCancel(ctx)
	for _, collaborationCommand := range collaborationCommands {
		if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, collaborationCommand); dispatchErr != nil {
			outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
				"dispatch inherited ControlCabinet clone: %w",
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
	projectIDs, err := workflow.GetSourceProjectIDs(
		writeCtx,
		command.SourceControlCabinetID,
	)
	if err != nil {
		return committedClone{}, err
	}
	projectIDs = normalizeProjectIDs(projectIDs)
	copyEntity, err := workflow.CopyByID(writeCtx, command.SourceControlCabinetID)
	if err != nil {
		return committedClone{}, err
	}
	if copyEntity == nil || copyEntity.ID == uuid.Nil ||
		copyEntity.ID == command.SourceControlCabinetID {
		return committedClone{}, domain.ErrInvalidArgument
	}
	after, err := workflow.GetByID(writeCtx, copyEntity.ID)
	if err != nil {
		return committedClone{}, err
	}
	if after == nil || after.ID != copyEntity.ID {
		return committedClone{}, domain.ErrNotFound
	}
	for _, projectID := range projectIDs {
		if err := workflow.AssignCopyToProject(
			writeCtx,
			projectID,
			after.ID,
		); err != nil {
			return committedClone{}, err
		}
	}
	change, err := buildCreateChange(after)
	if err != nil {
		return committedClone{}, err
	}
	return committedClone{
		cabinet:    cloneControlCabinet(after),
		projectIDs: projectIDs,
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
