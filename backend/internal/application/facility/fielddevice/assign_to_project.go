package fielddevice

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
	"project FieldDevice assignment transaction is not configured",
)

// AssignToProjectWorkflow is the transaction-scoped capability required to
// create one explicit ProjectFieldDevice link. The existing project service is
// the compatibility Adapter while the application owns commit and dispatch.
type AssignToProjectWorkflow interface {
	CreateFieldDevice(
		context.Context,
		uuid.UUID,
		uuid.UUID,
	) (*domainProject.ProjectFieldDevice, error)
}

type AssignToProjectCommand struct {
	ProjectID     uuid.UUID
	FieldDeviceID uuid.UUID
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
	Link           *domainProject.ProjectFieldDevice
	Mutation       mutation.Result
	DispatchErrors []error
}

type committedProjectAssignment struct {
	link    *domainProject.ProjectFieldDevice
	change  mutation.EntityChange
	batched bool
}

type projectFieldDeviceLinkSnapshot struct {
	ID            uuid.UUID `json:"id"`
	ProjectID     uuid.UUID `json:"project_id"`
	FieldDeviceID uuid.UUID `json:"field_device_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
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

// AssignToProject returns the established ProjectFieldDevice link. Realtime
// delivery is best effort after commit and cannot change that HTTP result.
func (h *AssignToProjectHandler) AssignToProject(
	ctx context.Context,
	command AssignToProjectCommand,
) (*domainProject.ProjectFieldDevice, error) {
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
	actorID := actorFromContext(h.actor, ctx)
	committed, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(
			txCtx context.Context,
			workflow AssignToProjectWorkflow,
		) (committedProjectAssignment, error) {
			return executeAssignToProjectTransaction(
				txCtx,
				workflow,
				command,
				operationID,
				h.historyBatch,
			)
		},
	)
	if err != nil {
		return AssignToProjectOutcome{}, err
	}

	occurredAt := h.now().UTC()
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
		Scope:     appcollaboration.FacilityScopeFieldDevice,
		EntityIDs: []uuid.UUID{committed.link.FieldDeviceID},
	}
	if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, commandToDispatch); dispatchErr != nil {
		outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
			"dispatch ProjectFieldDevice assignment for project %s: %w",
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
	link, err := workflow.CreateFieldDevice(
		writeCtx,
		command.ProjectID,
		command.FieldDeviceID,
	)
	if err != nil {
		return committedProjectAssignment{}, err
	}
	if link == nil || link.ID == uuid.Nil || link.ProjectID != command.ProjectID ||
		link.FieldDeviceID != command.FieldDeviceID {
		return committedProjectAssignment{}, errors.New("invalid ProjectFieldDevice assignment result")
	}
	change, err := buildProjectFieldDeviceCreateChange(link)
	if err != nil {
		return committedProjectAssignment{}, err
	}
	return committedProjectAssignment{
		link:    cloneProjectFieldDeviceLink(link),
		change:  change,
		batched: batched,
	}, nil
}

func buildProjectFieldDeviceCreateChange(
	link *domainProject.ProjectFieldDevice,
) (mutation.EntityChange, error) {
	snapshot := projectFieldDeviceSnapshot(link)
	after, err := json.Marshal(snapshot)
	if err != nil {
		return mutation.EntityChange{}, fmt.Errorf("marshal ProjectFieldDevice create snapshot: %w", err)
	}
	projectID := link.ProjectID
	return mutation.EntityChange{
		EntityType:    mutation.EntityTypeProjectFieldDevice,
		EntityID:      link.ID,
		ParentID:      &projectID,
		Action:        domainHistory.ActionCreate,
		After:         after,
		ChangedFields: []mutation.FieldName{mutation.FieldNameFieldDevice},
	}, nil
}

func cloneProjectFieldDeviceLink(
	link *domainProject.ProjectFieldDevice,
) *domainProject.ProjectFieldDevice {
	if link == nil {
		return nil
	}
	clone := *link
	return &clone
}
