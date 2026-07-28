package controlcabinet

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	"github.com/besart951/go_infra_link/backend/internal/application/facility/mutation"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

var ErrAssignToProjectTransactionNotConfigured = errors.New(
	"project control cabinet assignment transaction is not configured",
)

// AssignToProjectWorkflow is implemented by a transaction-scoped
// ProjectFacilityLinkService. That compatibility service retains descendant
// SPSController and FieldDevice link materialization while this handler owns
// the outer transaction and commit gate.
type AssignToProjectWorkflow interface {
	CreateControlCabinet(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) (*domainProject.ProjectControlCabinet, error)
}

type AssignToProjectCommand struct {
	ProjectID        uuid.UUID
	ControlCabinetID uuid.UUID
}

type AssignToProjectDependencies struct {
	TransactionRunner   apptransaction.Runner
	TransactionWorkflow apptransaction.Factory[AssignToProjectWorkflow]
	HistoryBatch        HistoryBatchContext
	Dispatcher          appcollaboration.CommandDispatcher
	Actor               ActorProvider
	NewID               IDGenerator
	Now                 Clock
	ReportError         ErrorReporter
}

type AssignToProjectHandler struct {
	operation             apptransaction.Operation[AssignToProjectWorkflow, AssignToProjectWorkflow]
	transactionConfigured bool
	historyBatch          HistoryBatchContext
	dispatcher            appcollaboration.CommandDispatcher
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
	reportError           ErrorReporter
}

type AssignToProjectOutcome struct {
	Link           *domainProject.ProjectControlCabinet
	Mutation       mutation.Result
	DispatchErrors []error
}

type committedProjectAssignment struct {
	link    *domainProject.ProjectControlCabinet
	change  mutation.EntityChange
	batched bool
}

type projectControlCabinetLinkSnapshot struct {
	ID               uuid.UUID `json:"id"`
	ProjectID        uuid.UUID `json:"project_id"`
	ControlCabinetID uuid.UUID `json:"control_cabinet_id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
}

func NewAssignToProjectHandler(
	deps AssignToProjectDependencies,
) *AssignToProjectHandler {
	newID := deps.NewID
	if newID == nil {
		newID = uuid.New
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
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
	return &AssignToProjectHandler{
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

func (h *AssignToProjectHandler) AssignToProject(
	ctx context.Context,
	command AssignToProjectCommand,
) (*domainProject.ProjectControlCabinet, error) {
	outcome, err := h.Execute(ctx, command)
	if err != nil {
		return nil, err
	}
	for _, dispatchErr := range outcome.DispatchErrors {
		if h.reportError != nil {
			h.reportError(dispatchErr)
		}
	}
	return outcome.Link, nil
}

func (h *AssignToProjectHandler) Execute(
	ctx context.Context,
	command AssignToProjectCommand,
) (AssignToProjectOutcome, error) {
	if h == nil || !h.transactionConfigured {
		return AssignToProjectOutcome{}, ErrAssignToProjectTransactionNotConfigured
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
			workflow AssignToProjectWorkflow,
		) (committedProjectAssignment, error) {
			result, err := executeAssignToProjectTransaction(
				txCtx,
				workflow,
				command,
				operationID,
				h.historyBatch,
			)
			if err != nil {
				return committedProjectAssignment{}, err
			}
			collaborationCommand = appcollaboration.FacilityHierarchyRefreshRequired{
				Envelope: appcollaboration.Envelope{
					SchemaVersion: appcollaboration.SchemaVersionV2,
					EventID:       eventID, OperationID: operationID, CorrelationID: operationID,
					ProjectID: command.ProjectID, ActorID: actorID, OccurredAt: occurredAt,
				},
				Scope:     appcollaboration.FacilityScopeControlCabinet,
				EntityIDs: []uuid.UUID{result.link.ControlCabinetID},
			}
			if _, err := appcollaboration.EnqueueCommand(txCtx, collaborationCommand); err != nil {
				return committedProjectAssignment{}, fmt.Errorf("enqueue project control cabinet assignment: %w", err)
			}
			return result, nil
		},
	)
	if err != nil {
		return AssignToProjectOutcome{}, err
	}

	result := mutation.Result{
		OperationID: operationID,
		ActorID:     actorID,
		OccurredAt:  occurredAt,
		ProjectIDs:  []uuid.UUID{command.ProjectID},
		Changes:     []mutation.EntityChange{committed.change},
	}
	if committed.batched {
		batchID := operationID
		result.BatchID = &batchID
	}
	outcome := AssignToProjectOutcome{
		Link:     committed.link,
		Mutation: result,
	}
	if h.dispatcher == nil {
		return outcome, nil
	}

	dispatchCtx := context.WithoutCancel(ctx)
	if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, collaborationCommand); dispatchErr != nil {
		outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
			"dispatch project control cabinet assignment for project %s: %w",
			command.ProjectID,
			dispatchErr,
		))
	}
	return outcome, nil
}

func executeAssignToProjectTransaction(
	ctx context.Context,
	workflow AssignToProjectWorkflow,
	command AssignToProjectCommand,
	operationID uuid.UUID,
	historyBatch HistoryBatchContext,
) (committedProjectAssignment, error) {
	if workflow == nil {
		return committedProjectAssignment{}, ErrAssignToProjectTransactionNotConfigured
	}

	writeCtx := ctx
	batched := historyBatch != nil
	if batched {
		writeCtx = historyBatch(ctx, operationID)
	}
	link, err := workflow.CreateControlCabinet(
		writeCtx,
		command.ProjectID,
		command.ControlCabinetID,
	)
	if err != nil {
		return committedProjectAssignment{}, err
	}
	if link == nil || link.ID == uuid.Nil || link.ProjectID != command.ProjectID ||
		link.ControlCabinetID != command.ControlCabinetID {
		return committedProjectAssignment{}, errors.New("invalid project control cabinet assignment result")
	}
	change, err := buildProjectControlCabinetCreateChange(link)
	if err != nil {
		return committedProjectAssignment{}, err
	}
	return committedProjectAssignment{
		link:    cloneProjectControlCabinetLink(link),
		change:  change,
		batched: batched,
	}, nil
}

func buildProjectControlCabinetCreateChange(
	link *domainProject.ProjectControlCabinet,
) (mutation.EntityChange, error) {
	after, err := json.Marshal(projectControlCabinetSnapshot(link))
	if err != nil {
		return mutation.EntityChange{}, fmt.Errorf("marshal project control cabinet create snapshot: %w", err)
	}
	projectID := link.ProjectID
	return mutation.EntityChange{
		EntityType:    mutation.EntityTypeProjectControlCabinet,
		EntityID:      link.ID,
		ParentID:      &projectID,
		Action:        domainHistory.ActionCreate,
		After:         after,
		ChangedFields: []mutation.FieldName{mutation.FieldNameControlCabinet},
	}, nil
}

func projectControlCabinetSnapshot(
	link *domainProject.ProjectControlCabinet,
) projectControlCabinetLinkSnapshot {
	if link == nil {
		return projectControlCabinetLinkSnapshot{}
	}
	return projectControlCabinetLinkSnapshot{
		ID:               link.ID,
		ProjectID:        link.ProjectID,
		ControlCabinetID: link.ControlCabinetID,
		CreatedAt:        link.CreatedAt,
		UpdatedAt:        link.UpdatedAt,
	}
}

func cloneProjectControlCabinetLink(
	link *domainProject.ProjectControlCabinet,
) *domainProject.ProjectControlCabinet {
	if link == nil {
		return nil
	}
	clone := *link
	return &clone
}
