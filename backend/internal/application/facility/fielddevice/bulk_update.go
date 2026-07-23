package fielddevice

import (
	"context"
	"fmt"
	"sort"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	"github.com/besart951/go_infra_link/backend/internal/application/facility/mutation"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainFieldDevice "github.com/besart951/go_infra_link/backend/internal/domain/facility/fielddevice"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

const defaultMaxTargetedRefreshIDs = 100

type BulkUpdateExecutor interface {
	ExecuteBulkUpdate(
		context.Context,
		[]domainFacility.BulkFieldDeviceUpdate,
	) domainFieldDevice.BulkUpdateExecution
}

type ProjectLinkReader interface {
	GetByFieldDeviceIDs(context.Context, []uuid.UUID) ([]*domainProject.ProjectFieldDevice, error)
}

type ActorProvider func(context.Context) *uuid.UUID
type IDGenerator func() uuid.UUID
type Clock func() time.Time
type ErrorReporter func(error)

type BulkUpdateDependencies struct {
	Executor              BulkUpdateExecutor
	HistoryBatch          HistoryBatchContext
	ProjectLinks          ProjectLinkReader
	Dispatcher            appcollaboration.CommandDispatcher
	Actor                 ActorProvider
	NewID                 IDGenerator
	Now                   Clock
	ReportError           ErrorReporter
	MaxTargetedRefreshIDs int
}

type BulkUpdateHandler struct {
	executor              BulkUpdateExecutor
	historyBatch          HistoryBatchContext
	projectLinks          ProjectLinkReader
	dispatcher            appcollaboration.CommandDispatcher
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
	reportError           ErrorReporter
	maxTargetedRefreshIDs int
}

type BulkUpdateOutcome struct {
	Result            *domainFacility.BulkOperationResult
	Mutation          mutation.Result
	ReconciliationIDs []uuid.UUID
	DispatchErrors    []error
}

func NewBulkUpdateHandler(deps BulkUpdateDependencies) *BulkUpdateHandler {
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

	return &BulkUpdateHandler{
		executor:              deps.Executor,
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

// BulkUpdate preserves the existing HTTP-facing result while the richer
// application outcome remains available to tests and later response mappers.
func (h *BulkUpdateHandler) BulkUpdate(
	ctx context.Context,
	updates []domainFacility.BulkFieldDeviceUpdate,
) *domainFacility.BulkOperationResult {
	outcome := h.Execute(ctx, updates)
	for _, err := range outcome.DispatchErrors {
		if h.reportError != nil {
			h.reportError(err)
		}
	}
	return outcome.Result
}

func (h *BulkUpdateHandler) Execute(
	ctx context.Context,
	updates []domainFacility.BulkFieldDeviceUpdate,
) BulkUpdateOutcome {
	if h == nil || h.executor == nil {
		return BulkUpdateOutcome{
			Result: failedBulkResult(updates, "field device bulk updater is not configured"),
		}
	}

	operationID := h.newID()
	actorID := actorFromContext(h.actor, ctx)
	mutationCtx := ctx
	if h.historyBatch != nil {
		mutationCtx = h.historyBatch(ctx, operationID)
	}
	execution := h.executor.ExecuteBulkUpdate(mutationCtx, updates)
	result := execution.Result
	occurredAt := h.now().UTC()
	batchID := operationID
	outcome := BulkUpdateOutcome{
		Result: result,
		Mutation: mutation.Result{
			OperationID: operationID,
			BatchID:     &batchID,
			ActorID:     actorID,
			OccurredAt:  occurredAt,
			Changes:     successfulChanges(updates, execution.Items),
		},
		ReconciliationIDs: reconciliationIDs(updates),
	}

	if len(outcome.ReconciliationIDs) == 0 || h.projectLinks == nil || h.dispatcher == nil {
		return outcome
	}

	dispatchCtx := context.WithoutCancel(ctx)
	links, err := h.projectLinks.GetByFieldDeviceIDs(dispatchCtx, outcome.ReconciliationIDs)
	if err != nil {
		outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf("resolve FieldDevice collaboration projects: %w", err))
		return outcome
	}

	byProject := groupLinkedFieldDevices(links, outcome.ReconciliationIDs)
	projectIDs := sortedProjectIDs(byProject)
	outcome.Mutation.ProjectIDs = append([]uuid.UUID(nil), projectIDs...)

	for _, projectID := range projectIDs {
		entityIDs := byProject[projectID]
		fullRefresh := len(entityIDs) > h.maxTargetedRefreshIDs
		if fullRefresh {
			entityIDs = nil
		}

		command := appcollaboration.FacilityHierarchyRefreshRequired{
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
		if err := h.dispatcher.Dispatch(dispatchCtx, command); err != nil {
			outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
				"dispatch FieldDevice collaboration refresh for project %s: %w",
				projectID,
				err,
			))
		}
	}

	return outcome
}

func failedBulkResult(
	updates []domainFacility.BulkFieldDeviceUpdate,
	message string,
) *domainFacility.BulkOperationResult {
	result := &domainFacility.BulkOperationResult{
		Results:      make([]domainFacility.BulkOperationResultItem, len(updates)),
		TotalCount:   len(updates),
		FailureCount: len(updates),
	}
	for i, update := range updates {
		result.Results[i] = domainFacility.BulkOperationResultItem{
			ID:    update.ID,
			Error: message,
		}
	}
	return result
}

func actorFromContext(provider ActorProvider, ctx context.Context) *uuid.UUID {
	if provider == nil {
		return nil
	}
	actorID := provider(ctx)
	if actorID == nil || *actorID == uuid.Nil {
		return nil
	}
	clone := *actorID
	return &clone
}

func reconciliationIDs(updates []domainFacility.BulkFieldDeviceUpdate) []uuid.UUID {
	set := make(map[uuid.UUID]struct{}, len(updates))
	for _, update := range updates {
		if update.ID == uuid.Nil || len(requestedFields(update)) == 0 {
			continue
		}
		set[update.ID] = struct{}{}
	}
	return sortedUUIDSet(set)
}

func successfulChanges(
	updates []domainFacility.BulkFieldDeviceUpdate,
	executions []domainFieldDevice.BulkUpdateItemExecution,
) []mutation.EntityChange {
	if len(executions) == 0 {
		return nil
	}

	fieldsByID := make(map[uuid.UUID]map[mutation.FieldName]struct{})
	for _, execution := range executions {
		if execution.ID == uuid.Nil || execution.Index < 0 || execution.Index >= len(updates) {
			continue
		}
		update := updates[execution.Index]
		if update.ID != execution.ID {
			continue
		}
		successfulFields := successfulRequestedFields(update, execution.Phases)
		if len(successfulFields) == 0 {
			continue
		}
		if fieldsByID[execution.ID] == nil {
			fieldsByID[execution.ID] = map[mutation.FieldName]struct{}{}
		}
		for _, field := range successfulFields {
			fieldsByID[execution.ID][field] = struct{}{}
		}
	}

	ids := make(map[uuid.UUID]struct{}, len(fieldsByID))
	for id := range fieldsByID {
		ids[id] = struct{}{}
	}
	sortedIDs := sortedUUIDSet(ids)
	changes := make([]mutation.EntityChange, 0, len(sortedIDs))
	for _, id := range sortedIDs {
		fields := make([]mutation.FieldName, 0, len(fieldsByID[id]))
		for field := range fieldsByID[id] {
			fields = append(fields, field)
		}
		sort.Slice(fields, func(i, j int) bool { return fields[i] < fields[j] })
		changes = append(changes, mutation.EntityChange{
			EntityType:    mutation.EntityTypeFieldDevice,
			EntityID:      id,
			Action:        domainHistory.ActionUpdate,
			ChangedFields: fields,
		})
	}
	return changes
}

func successfulRequestedFields(
	update domainFacility.BulkFieldDeviceUpdate,
	phases []domainFieldDevice.BulkUpdatePhaseResult,
) []mutation.FieldName {
	fields := make([]mutation.FieldName, 0, 8)
	for _, phase := range phases {
		if phase.Status != domainFieldDevice.BulkUpdatePhaseSucceeded {
			continue
		}
		fields = append(fields, requestedFieldsForPhase(update, phase.Phase)...)
	}
	return fields
}

func requestedFields(update domainFacility.BulkFieldDeviceUpdate) []mutation.FieldName {
	fields := make([]mutation.FieldName, 0, 8)
	fields = append(fields, requestedFieldsForPhase(
		update,
		domainFieldDevice.BulkUpdatePhaseFieldDevice,
	)...)
	fields = append(fields, requestedFieldsForPhase(
		update,
		domainFieldDevice.BulkUpdatePhaseSpecification,
	)...)
	fields = append(fields, requestedFieldsForPhase(
		update,
		domainFieldDevice.BulkUpdatePhaseBacnetObjects,
	)...)
	return fields
}

func requestedFieldsForPhase(
	update domainFacility.BulkFieldDeviceUpdate,
	phase domainFieldDevice.BulkUpdatePhase,
) []mutation.FieldName {
	fields := make([]mutation.FieldName, 0, 6)
	if phase != domainFieldDevice.BulkUpdatePhaseFieldDevice {
		switch phase {
		case domainFieldDevice.BulkUpdatePhaseSpecification:
			if update.Specification != nil && update.Specification.HasChanges() {
				return []mutation.FieldName{mutation.FieldNameSpecification}
			}
		case domainFieldDevice.BulkUpdatePhaseBacnetObjects:
			if update.BacnetObjects != nil {
				return []mutation.FieldName{mutation.FieldNameBacnetObjects}
			}
		}
		return fields
	}

	if update.HasBMKUpdate() {
		fields = append(fields, mutation.FieldNameBMK)
	}
	if update.HasDescriptionUpdate() {
		fields = append(fields, mutation.FieldNameDescription)
	}
	if update.HasTextIndividuellUpdate() {
		fields = append(fields, mutation.FieldNameTextFix)
	}
	if update.ApparatNr != nil {
		fields = append(fields, mutation.FieldNameApparatNumber)
	}
	if update.ApparatID != nil {
		fields = append(fields, mutation.FieldNameApparat)
	}
	if update.SystemPartID != nil {
		fields = append(fields, mutation.FieldNameSystemPart)
	}
	return fields
}

func groupLinkedFieldDevices(
	links []*domainProject.ProjectFieldDevice,
	requestedIDs []uuid.UUID,
) map[uuid.UUID][]uuid.UUID {
	requested := make(map[uuid.UUID]struct{}, len(requestedIDs))
	for _, id := range requestedIDs {
		requested[id] = struct{}{}
	}

	sets := make(map[uuid.UUID]map[uuid.UUID]struct{})
	for _, link := range links {
		if link == nil || link.ProjectID == uuid.Nil || link.FieldDeviceID == uuid.Nil {
			continue
		}
		if _, ok := requested[link.FieldDeviceID]; !ok {
			continue
		}
		if sets[link.ProjectID] == nil {
			sets[link.ProjectID] = map[uuid.UUID]struct{}{}
		}
		sets[link.ProjectID][link.FieldDeviceID] = struct{}{}
	}

	grouped := make(map[uuid.UUID][]uuid.UUID, len(sets))
	for projectID, ids := range sets {
		grouped[projectID] = sortedUUIDSet(ids)
	}
	return grouped
}

func sortedProjectIDs(byProject map[uuid.UUID][]uuid.UUID) []uuid.UUID {
	set := make(map[uuid.UUID]struct{}, len(byProject))
	for projectID := range byProject {
		set[projectID] = struct{}{}
	}
	return sortedUUIDSet(set)
}

func sortedUUIDSet(set map[uuid.UUID]struct{}) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(set))
	for id := range set {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool {
		return ids[i].String() < ids[j].String()
	})
	return ids
}
