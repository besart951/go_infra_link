package fielddevice

import (
	"context"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/application/facility/mutation"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

// MultiCreateExecutor preserves the existing per-item transaction and
// partial-success policy behind a narrow application port.
type MultiCreateExecutor interface {
	MultiCreate(
		context.Context,
		[]domainFacility.FieldDeviceCreateItem,
	) *domainFacility.FieldDeviceMultiCreateResult
}

type MultiCreateCommand struct {
	Items []domainFacility.FieldDeviceCreateItem
}

type MultiCreateDependencies struct {
	Executor     MultiCreateExecutor
	HistoryBatch HistoryBatchContext
	Actor        ActorProvider
	NewID        IDGenerator
	Now          Clock
}

type MultiCreateHandler struct {
	executor     MultiCreateExecutor
	historyBatch HistoryBatchContext
	actor        ActorProvider
	newID        IDGenerator
	now          Clock
}

type MultiCreateOutcome struct {
	Result   *domainFacility.FieldDeviceMultiCreateResult
	Mutation mutation.Result
}

func NewMultiCreateHandler(deps MultiCreateDependencies) *MultiCreateHandler {
	newID := deps.NewID
	if newID == nil {
		newID = uuid.New
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &MultiCreateHandler{
		executor:     deps.Executor,
		historyBatch: deps.HistoryBatch,
		actor:        deps.Actor,
		newID:        newID,
		now:          now,
	}
}

// MultiCreate preserves the existing HTTP result and per-item partial-success
// semantics. The richer mutation outcome remains available to tests and later
// project-scoped orchestration.
func (h *MultiCreateHandler) MultiCreate(
	ctx context.Context,
	command MultiCreateCommand,
) *domainFacility.FieldDeviceMultiCreateResult {
	return h.Execute(ctx, command).Result
}

func (h *MultiCreateHandler) Execute(
	ctx context.Context,
	command MultiCreateCommand,
) MultiCreateOutcome {
	if h == nil || h.executor == nil {
		return MultiCreateOutcome{
			Result: failedMultiCreateResult(
				command.Items,
				"field device multi-create is not configured",
			),
		}
	}

	operationID := h.newID()
	mutationCtx := ctx
	batched := h.historyBatch != nil
	if batched {
		mutationCtx = h.historyBatch(ctx, operationID)
	}
	result := h.executor.MultiCreate(mutationCtx, command.Items)
	if result == nil {
		result = failedMultiCreateResult(
			command.Items,
			"field device multi-create returned no result",
		)
	}
	normalizeMultiCreateResult(result, command.Items)

	mutationResult := mutation.Result{
		OperationID: operationID,
		ActorID:     actorFromContext(h.actor, ctx),
		OccurredAt:  h.now().UTC(),
		Changes:     successfulCreateChanges(result),
	}
	if batched {
		batchID := operationID
		mutationResult.BatchID = &batchID
	}
	return MultiCreateOutcome{
		Result:   result,
		Mutation: mutationResult,
	}
}

func successfulCreateChanges(
	result *domainFacility.FieldDeviceMultiCreateResult,
) []mutation.EntityChange {
	if result == nil || result.SuccessCount == 0 {
		return nil
	}

	changes := make([]mutation.EntityChange, 0, result.SuccessCount)
	for _, item := range result.Results {
		if !item.Success || item.FieldDevice == nil || item.FieldDevice.ID == uuid.Nil {
			continue
		}
		after, err := marshalSnapshot(toFieldDeviceSnapshot(item.FieldDevice))
		if err != nil {
			// This snapshot contains only JSON-safe primitives. Keep the existing
			// partial-success response if that invariant changes later.
			continue
		}
		parentID := item.FieldDevice.SPSControllerSystemTypeID
		changes = append(changes, mutation.EntityChange{
			EntityType: mutation.EntityTypeFieldDevice,
			EntityID:   item.FieldDevice.ID,
			ParentID:   &parentID,
			Action:     domainHistory.ActionCreate,
			After:      after,
		})
	}
	return changes
}

func failedMultiCreateResult(
	items []domainFacility.FieldDeviceCreateItem,
	message string,
) *domainFacility.FieldDeviceMultiCreateResult {
	result := &domainFacility.FieldDeviceMultiCreateResult{
		Results:       make([]domainFacility.FieldDeviceCreateResult, len(items)),
		TotalRequests: len(items),
		FailureCount:  len(items),
	}
	for index := range items {
		result.Results[index] = domainFacility.FieldDeviceCreateResult{
			Index:      index,
			Error:      message,
			ErrorField: "fielddevice",
		}
	}
	return result
}
