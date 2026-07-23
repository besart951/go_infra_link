package fielddevice

import (
	"context"
	"fmt"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	"github.com/besart951/go_infra_link/backend/internal/application/facility/mutation"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

// BulkDeleteSnapshotReader loads all candidate roots in one query so the
// application does not add a per-item read merely to build canonical changes.
type BulkDeleteSnapshotReader interface {
	GetByIds(context.Context, []uuid.UUID) ([]*domainFacility.FieldDevice, error)
}

// BulkDeleteWorkflow is the minimal transaction-scoped delete capability.
// The decorated repository behind the first Adapter records history before
// deleting and therefore participates in each item's transaction.
type BulkDeleteWorkflow interface {
	DeleteByID(context.Context, uuid.UUID) error
}

type BulkDeleteCommand struct {
	FieldDeviceIDs []uuid.UUID
}

type BulkDeleteDependencies struct {
	TransactionRunner     apptransaction.Runner
	TransactionWorkflow   apptransaction.Factory[BulkDeleteWorkflow]
	Snapshots             BulkDeleteSnapshotReader
	HistoryBatch          HistoryBatchContext
	ProjectLinks          ProjectLinkReader
	Dispatcher            appcollaboration.CommandDispatcher
	Actor                 ActorProvider
	NewID                 IDGenerator
	Now                   Clock
	ReportError           ErrorReporter
	MaxTargetedRefreshIDs int
}

type BulkDeleteHandler struct {
	operation             apptransaction.Operation[BulkDeleteWorkflow, BulkDeleteWorkflow]
	transactionConfigured bool
	snapshots             BulkDeleteSnapshotReader
	historyBatch          HistoryBatchContext
	projectLinks          ProjectLinkReader
	dispatcher            appcollaboration.CommandDispatcher
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
	reportError           ErrorReporter
	maxTargetedRefreshIDs int
}

type BulkDeleteOutcome struct {
	Result            *domainFacility.BulkOperationResult
	Mutation          mutation.Result
	ReconciliationIDs []uuid.UUID
	DispatchErrors    []error
}

func NewBulkDeleteHandler(deps BulkDeleteDependencies) *BulkDeleteHandler {
	newID := deps.NewID
	if newID == nil {
		newID = uuid.New
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	maxTargetedRefreshIDs := deps.MaxTargetedRefreshIDs
	if maxTargetedRefreshIDs <= 0 {
		maxTargetedRefreshIDs = defaultMaxTargetedRefreshIDs
	}
	boundary := apptransaction.NewBoundary[BulkDeleteWorkflow](
		deps.TransactionRunner,
		deps.TransactionWorkflow,
	)
	var noNonTransactionalWorkflow BulkDeleteWorkflow
	operation := apptransaction.Bind(
		boundary,
		noNonTransactionalWorkflow,
		func(workflow BulkDeleteWorkflow) BulkDeleteWorkflow { return workflow },
	)
	return &BulkDeleteHandler{
		operation:             operation,
		transactionConfigured: deps.TransactionRunner != nil && deps.TransactionWorkflow != nil,
		snapshots:             deps.Snapshots,
		historyBatch:          deps.HistoryBatch,
		projectLinks:          deps.ProjectLinks,
		dispatcher:            deps.Dispatcher,
		actor:                 deps.Actor,
		newID:                 newID,
		now:                   now,
		reportError:           deps.ReportError,
		maxTargetedRefreshIDs: maxTargetedRefreshIDs,
	}
}

// BulkDelete preserves the existing index-aligned, partial-success HTTP result.
// Scope and collaboration failures remain best effort after committed deletes.
func (h *BulkDeleteHandler) BulkDelete(
	ctx context.Context,
	command BulkDeleteCommand,
) *domainFacility.BulkOperationResult {
	outcome := h.Execute(ctx, command)
	for _, dispatchErr := range outcome.DispatchErrors {
		if h.reportError != nil {
			h.reportError(dispatchErr)
		}
	}
	return outcome.Result
}

func (h *BulkDeleteHandler) Execute(
	ctx context.Context,
	command BulkDeleteCommand,
) BulkDeleteOutcome {
	if h == nil || !h.transactionConfigured || h.snapshots == nil {
		return BulkDeleteOutcome{
			Result: failedBulkDeleteResult(
				command.FieldDeviceIDs,
				"field device bulk delete is not configured",
			),
		}
	}

	operationID := h.newID()
	actorID := actorFromContext(h.actor, ctx)
	candidateIDs := uniqueBulkDeleteIDs(command.FieldDeviceIDs)
	var snapshots []*domainFacility.FieldDevice
	if len(candidateIDs) > 0 {
		var err error
		snapshots, err = h.snapshots.GetByIds(ctx, candidateIDs)
		if err != nil {
			return BulkDeleteOutcome{
				Result:   failedBulkDeleteResult(command.FieldDeviceIDs, err.Error()),
				Mutation: newBulkDeleteMutation(operationID, actorID, h.now, h.historyBatch != nil),
			}
		}
	}
	snapshotByID := make(map[uuid.UUID]*domainFacility.FieldDevice, len(snapshots))
	for _, fieldDevice := range snapshots {
		if fieldDevice == nil || fieldDevice.ID == uuid.Nil {
			continue
		}
		snapshotByID[fieldDevice.ID] = cloneFieldDevice(fieldDevice)
	}

	var (
		links    []*domainProject.ProjectFieldDevice
		scopeErr error
	)
	if len(candidateIDs) > 0 && h.projectLinks != nil && h.dispatcher != nil {
		resolved, resolveErr := h.projectLinks.GetByFieldDeviceIDs(ctx, candidateIDs)
		if resolveErr != nil {
			scopeErr = resolveErr
		} else {
			links = resolved
		}
	}

	result := &domainFacility.BulkOperationResult{
		Results:    make([]domainFacility.BulkOperationResultItem, len(command.FieldDeviceIDs)),
		TotalCount: len(command.FieldDeviceIDs),
	}
	changes := make([]mutation.EntityChange, 0, len(snapshotByID))
	changedIDs := make(map[uuid.UUID]struct{}, len(snapshotByID))
	for index, fieldDeviceID := range command.FieldDeviceIDs {
		item := &result.Results[index]
		item.ID = fieldDeviceID

		committed, deleteErr := apptransaction.RunResult(
			ctx,
			h.operation,
			func(txCtx context.Context, workflow BulkDeleteWorkflow) (bool, error) {
				return executeBulkDeleteItem(
					txCtx,
					workflow,
					fieldDeviceID,
					operationID,
					h.historyBatch,
				)
			},
		)
		if deleteErr != nil || !committed {
			if deleteErr != nil {
				item.Error = deleteErr.Error()
			} else {
				item.Error = "field device delete did not commit"
			}
			result.FailureCount++
			continue
		}

		item.Success = true
		result.SuccessCount++
		before := snapshotByID[fieldDeviceID]
		if before == nil {
			continue
		}
		if _, duplicate := changedIDs[fieldDeviceID]; duplicate {
			continue
		}
		change, buildErr := buildDeleteChange(before)
		if buildErr != nil {
			// FieldDevice snapshots contain only JSON-safe values. If that invariant
			// changes, retain the committed compatibility result and omit the
			// derived application projection rather than report a false rollback.
			continue
		}
		changedIDs[fieldDeviceID] = struct{}{}
		changes = append(changes, change)
	}

	occurredAt := h.now().UTC()
	reconciliationIDs := sortedUUIDSet(changedIDs)
	mutationResult := mutation.Result{
		OperationID: operationID,
		ActorID:     actorID,
		OccurredAt:  occurredAt,
		Changes:     changes,
	}
	if h.historyBatch != nil {
		batchID := operationID
		mutationResult.BatchID = &batchID
	}
	outcome := BulkDeleteOutcome{
		Result:            result,
		Mutation:          mutationResult,
		ReconciliationIDs: reconciliationIDs,
	}
	if len(reconciliationIDs) == 0 || h.dispatcher == nil {
		return outcome
	}
	if scopeErr != nil {
		outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
			"resolve bulk-deleted FieldDevice collaboration projects: %w",
			scopeErr,
		))
		return outcome
	}

	grouped := groupLinkedFieldDevices(links, reconciliationIDs)
	projectIDs := sortedProjectIDs(grouped)
	outcome.Mutation.ProjectIDs = append([]uuid.UUID(nil), projectIDs...)
	dispatchCtx := context.WithoutCancel(ctx)
	for _, projectID := range projectIDs {
		entityIDs := grouped[projectID]
		fullRefresh := len(entityIDs) > h.maxTargetedRefreshIDs
		if fullRefresh {
			entityIDs = nil
		}
		collaborationCommand := appcollaboration.FacilityHierarchyRefreshRequired{
			Envelope: appcollaboration.Envelope{
				SchemaVersion: appcollaboration.SchemaVersionV1,
				EventID:       h.newID(),
				OperationID:   operationID,
				CorrelationID: operationID,
				ProjectID:     projectID,
				ActorID:       actorID,
				OccurredAt:    occurredAt,
			},
			Scope:       appcollaboration.FacilityScopeFieldDevice,
			EntityIDs:   append([]uuid.UUID(nil), entityIDs...),
			FullRefresh: fullRefresh,
		}
		if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, collaborationCommand); dispatchErr != nil {
			outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
				"dispatch bulk-deleted FieldDevice refresh for project %s: %w",
				projectID,
				dispatchErr,
			))
		}
	}
	return outcome
}

func executeBulkDeleteItem(
	ctx context.Context,
	workflow BulkDeleteWorkflow,
	fieldDeviceID uuid.UUID,
	operationID uuid.UUID,
	historyBatch HistoryBatchContext,
) (bool, error) {
	if workflow == nil {
		return false, ErrDeleteTransactionNotConfigured
	}
	writeCtx := ctx
	if historyBatch != nil {
		writeCtx = historyBatch(ctx, operationID)
	}
	if err := workflow.DeleteByID(writeCtx, fieldDeviceID); err != nil {
		return false, err
	}
	return true, nil
}

func failedBulkDeleteResult(
	ids []uuid.UUID,
	message string,
) *domainFacility.BulkOperationResult {
	result := &domainFacility.BulkOperationResult{
		Results:      make([]domainFacility.BulkOperationResultItem, len(ids)),
		TotalCount:   len(ids),
		FailureCount: len(ids),
	}
	for index, id := range ids {
		result.Results[index] = domainFacility.BulkOperationResultItem{
			ID:    id,
			Error: message,
		}
	}
	return result
}

func newBulkDeleteMutation(
	operationID uuid.UUID,
	actorID *uuid.UUID,
	now Clock,
	batched bool,
) mutation.Result {
	result := mutation.Result{
		OperationID: operationID,
		ActorID:     actorID,
		OccurredAt:  now().UTC(),
	}
	if batched {
		batchID := operationID
		result.BatchID = &batchID
	}
	return result
}

func uniqueBulkDeleteIDs(ids []uuid.UUID) []uuid.UUID {
	set := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		if id != uuid.Nil {
			set[id] = struct{}{}
		}
	}
	return sortedUUIDSet(set)
}
