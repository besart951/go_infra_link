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

const projectNotFoundAssociationError = "project not found"

// BulkAssignProjectReader preserves the existing request-wide project
// existence check before any independently committed link assignment.
type BulkAssignProjectReader interface {
	GetByIds(context.Context, []uuid.UUID) ([]*domainProject.Project, error)
}

type BulkAssignToProjectCommand struct {
	ProjectID      uuid.UUID
	FieldDeviceIDs []uuid.UUID
}

type BulkAssignToProjectResult struct {
	SuccessFieldDeviceIDs []uuid.UUID
	AssociationErrors     []string
	Results               []domainFacility.BulkOperationResultItem
}

type BulkAssignToProjectDependencies struct {
	TransactionRunner     apptransaction.Runner
	TransactionWorkflow   apptransaction.Factory[AssignToProjectWorkflow]
	Projects              BulkAssignProjectReader
	HistoryBatch          HistoryBatchContext
	Dispatcher            appcollaboration.CommandDispatcher
	Actor                 ActorProvider
	NewID                 IDGenerator
	Now                   Clock
	ReportError           ErrorReporter
	MaxTargetedRefreshIDs int
}

type BulkAssignToProjectHandler struct {
	operation             apptransaction.Operation[AssignToProjectWorkflow, AssignToProjectWorkflow]
	transactionConfigured bool
	projects              BulkAssignProjectReader
	historyBatch          HistoryBatchContext
	dispatcher            appcollaboration.CommandDispatcher
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
	reportError           ErrorReporter
	maxTargetedRefreshIDs int
}

type BulkAssignToProjectOutcome struct {
	Result            BulkAssignToProjectResult
	Mutation          mutation.Result
	ReconciliationIDs []uuid.UUID
	DispatchErrors    []error
}

func NewBulkAssignToProjectHandler(
	deps BulkAssignToProjectDependencies,
) *BulkAssignToProjectHandler {
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
	boundary := apptransaction.NewBoundary[AssignToProjectWorkflow](
		deps.TransactionRunner,
		deps.TransactionWorkflow,
	)
	var noNonTransactionalWorkflow AssignToProjectWorkflow
	operation := apptransaction.Bind(
		boundary,
		noNonTransactionalWorkflow,
		func(workflow AssignToProjectWorkflow) AssignToProjectWorkflow { return workflow },
	)
	return &BulkAssignToProjectHandler{
		operation:             operation,
		transactionConfigured: deps.TransactionRunner != nil && deps.TransactionWorkflow != nil,
		projects:              deps.Projects,
		historyBatch:          deps.HistoryBatch,
		dispatcher:            deps.Dispatcher,
		actor:                 deps.Actor,
		newID:                 newID,
		now:                   now,
		reportError:           deps.ReportError,
		maxTargetedRefreshIDs: maxTargetedRefreshIDs,
	}
}

// BulkAssignToProject preserves the endpoint's two-list partial-success
// response. Collaboration failures remain best effort after all item commits.
func (h *BulkAssignToProjectHandler) BulkAssignToProject(
	ctx context.Context,
	command BulkAssignToProjectCommand,
) BulkAssignToProjectResult {
	outcome := h.Execute(ctx, command)
	for _, dispatchErr := range outcome.DispatchErrors {
		if h != nil && h.reportError != nil {
			h.reportError(dispatchErr)
		}
	}
	return outcome.Result
}

func (h *BulkAssignToProjectHandler) Execute(
	ctx context.Context,
	command BulkAssignToProjectCommand,
) BulkAssignToProjectOutcome {
	if h == nil || !h.transactionConfigured || h.projects == nil {
		result := BulkAssignToProjectResult{
			AssociationErrors: []string{"project FieldDevice bulk assignment is not configured"},
			Results:           make([]domainFacility.BulkOperationResultItem, len(command.FieldDeviceIDs)),
		}
		for index, fieldDeviceID := range command.FieldDeviceIDs {
			result.Results[index] = domainFacility.BulkOperationResultItem{
				ID:         fieldDeviceID,
				Error:      "project FieldDevice bulk assignment is not configured",
				ErrorCode:  itemErrorCodeNotConfigured,
				ErrorField: "fielddevice",
				Reason:     "project FieldDevice bulk assignment is not configured",
			}
		}
		return BulkAssignToProjectOutcome{Result: result}
	}

	operationID := h.newID()
	actorID := actorFromContext(h.actor, ctx)
	occurredAt := h.now().UTC()
	mutationResult := mutation.Result{
		OperationID: operationID,
		ActorID:     actorID,
		OccurredAt:  occurredAt,
		ProjectIDs:  []uuid.UUID{command.ProjectID},
	}
	if h.historyBatch != nil {
		batchID := operationID
		mutationResult.BatchID = &batchID
	}
	outcome := BulkAssignToProjectOutcome{
		Mutation: mutationResult,
		Result: BulkAssignToProjectResult{
			Results: make(
				[]domainFacility.BulkOperationResultItem,
				len(command.FieldDeviceIDs),
			),
		},
	}

	projects, err := h.projects.GetByIds(ctx, []uuid.UUID{command.ProjectID})
	if err != nil || !containsProject(projects, command.ProjectID) {
		outcome.Result.AssociationErrors = []string{projectNotFoundAssociationError}
		for index, fieldDeviceID := range command.FieldDeviceIDs {
			outcome.Result.Results[index] = domainFacility.BulkOperationResultItem{
				ID:         fieldDeviceID,
				Error:      projectNotFoundAssociationError,
				ErrorCode:  itemErrorCodeNotFound,
				ErrorField: "project_id",
				Reason:     projectNotFoundAssociationError,
			}
		}
		return outcome
	}

	changes := make([]mutation.EntityChange, 0, len(command.FieldDeviceIDs))
	for index, fieldDeviceID := range command.FieldDeviceIDs {
		item := &outcome.Result.Results[index]
		item.ID = fieldDeviceID
		committed, assignErr := apptransaction.RunResult(
			ctx,
			h.operation,
			func(
				txCtx context.Context,
				workflow AssignToProjectWorkflow,
			) (committedProjectAssignment, error) {
				result, err := executeAssignToProjectTransaction(
					txCtx,
					workflow,
					AssignToProjectCommand{
						ProjectID:     command.ProjectID,
						FieldDeviceID: fieldDeviceID,
					},
					operationID,
					h.historyBatch,
				)
				if err != nil {
					return committedProjectAssignment{}, err
				}
				if appcollaboration.OutboxConfigured(txCtx) {
					durableCommand := appcollaboration.FacilityHierarchyRefreshRequired{
						Envelope: appcollaboration.Envelope{
							SchemaVersion: appcollaboration.SchemaVersionV2,
							EventID:       h.newID(), OperationID: operationID, CorrelationID: operationID,
							ProjectID: command.ProjectID, ActorID: actorID, OccurredAt: occurredAt,
						},
						Scope:     appcollaboration.FacilityScopeFieldDevice,
						EntityIDs: []uuid.UUID{result.link.FieldDeviceID},
					}
					if _, err := appcollaboration.EnqueueCommand(txCtx, durableCommand); err != nil {
						return committedProjectAssignment{}, fmt.Errorf("enqueue bulk ProjectFieldDevice assignment: %w", err)
					}
				}
				return result, nil
			},
		)
		if assignErr != nil {
			item.Error = assignErr.Error()
			item.ErrorField = "fielddevice_id"
			item.ErrorCode = classifyItemError(item.Error, item.ErrorField)
			item.Reason = item.Error
			outcome.Result.AssociationErrors = append(
				outcome.Result.AssociationErrors,
				assignErr.Error(),
			)
			continue
		}
		item.Success = true
		outcome.Result.SuccessFieldDeviceIDs = append(
			outcome.Result.SuccessFieldDeviceIDs,
			committed.link.FieldDeviceID,
		)
		changes = append(changes, committed.change)
	}

	outcome.Mutation.Changes = changes
	outcome.ReconciliationIDs = uniqueSortedAssignmentIDs(
		outcome.Result.SuccessFieldDeviceIDs,
	)
	if len(outcome.ReconciliationIDs) == 0 || h.dispatcher == nil {
		return outcome
	}

	entityIDs := outcome.ReconciliationIDs
	fullRefresh := len(entityIDs) > h.maxTargetedRefreshIDs
	if fullRefresh {
		entityIDs = nil
	}
	dispatchCtx := context.WithoutCancel(ctx)
	commandToDispatch := appcollaboration.FacilityHierarchyRefreshRequired{
		Envelope: appcollaboration.Envelope{
			SchemaVersion: appcollaboration.SchemaVersionV1,
			EventID:       h.newID(),
			OperationID:   operationID,
			CorrelationID: operationID,
			ProjectID:     command.ProjectID,
			ActorID:       actorID,
			OccurredAt:    occurredAt,
		},
		Scope:       appcollaboration.FacilityScopeFieldDevice,
		EntityIDs:   append([]uuid.UUID(nil), entityIDs...),
		FullRefresh: fullRefresh,
	}
	if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, commandToDispatch); dispatchErr != nil {
		outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
			"dispatch bulk ProjectFieldDevice assignments for project %s: %w",
			command.ProjectID,
			dispatchErr,
		))
	}
	return outcome
}

func containsProject(projects []*domainProject.Project, projectID uuid.UUID) bool {
	for _, project := range projects {
		if project != nil && project.ID == projectID {
			return true
		}
	}
	return false
}

func uniqueSortedAssignmentIDs(ids []uuid.UUID) []uuid.UUID {
	set := make(map[uuid.UUID]struct{}, len(ids))
	for _, id := range ids {
		if id != uuid.Nil {
			set[id] = struct{}{}
		}
	}
	return sortedUUIDSet(set)
}
