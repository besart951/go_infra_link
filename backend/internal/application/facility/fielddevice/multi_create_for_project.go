package fielddevice

import (
	"context"
	"errors"
	"fmt"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	"github.com/besart951/go_infra_link/backend/internal/application/facility/mutation"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

var ErrMultiCreateForProjectTransactionNotConfigured = errors.New(
	"project field device multi-create transaction is not configured",
)

// MultiCreateForProjectWorkflow is implemented by the transaction-scoped
// ProjectFacilityLinkService. That compatibility service retains the current
// create-and-link policy while this handler owns the outer commit gate. Every
// failed item must roll back its own root, children, and history before the
// partial result is returned; successful items remain staged for the outer
// project-link transaction.
type MultiCreateForProjectWorkflow interface {
	MultiCreateAndAssignFieldDevices(
		context.Context,
		uuid.UUID,
		[]domainFacility.FieldDeviceCreateItem,
	) (*domainFacility.FieldDeviceMultiCreateResult, error)
}

type MultiCreateForProjectCommand struct {
	ProjectID uuid.UUID
	Items     []domainFacility.FieldDeviceCreateItem
}

type MultiCreateForProjectDependencies struct {
	TransactionRunner   apptransaction.Runner
	TransactionWorkflow apptransaction.Factory[MultiCreateForProjectWorkflow]
	HistoryBatch        HistoryBatchContext
	Dispatcher          appcollaboration.CommandDispatcher
	Actor               ActorProvider
	NewID               IDGenerator
	Now                 Clock
	ReportError         ErrorReporter
}

type MultiCreateForProjectHandler struct {
	operation             apptransaction.Operation[MultiCreateForProjectWorkflow, MultiCreateForProjectWorkflow]
	transactionConfigured bool
	historyBatch          HistoryBatchContext
	dispatcher            appcollaboration.CommandDispatcher
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
	reportError           ErrorReporter
}

type MultiCreateForProjectOutcome struct {
	Result         *domainFacility.FieldDeviceMultiCreateResult
	Mutation       mutation.Result
	DispatchErrors []error
}

type committedProjectMultiCreate struct {
	result       *domainFacility.FieldDeviceMultiCreateResult
	changes      []mutation.EntityChange
	fieldDevices []appcollaboration.FieldDeviceState
	batched      bool
}

func NewMultiCreateForProjectHandler(
	deps MultiCreateForProjectDependencies,
) *MultiCreateForProjectHandler {
	newID := deps.NewID
	if newID == nil {
		newID = uuid.New
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	boundary := apptransaction.NewBoundary[MultiCreateForProjectWorkflow](
		deps.TransactionRunner,
		deps.TransactionWorkflow,
	)
	var noNonTransactionalWorkflow MultiCreateForProjectWorkflow
	operation := apptransaction.Bind(
		boundary,
		noNonTransactionalWorkflow,
		func(workflow MultiCreateForProjectWorkflow) MultiCreateForProjectWorkflow {
			return workflow
		},
	)
	return &MultiCreateForProjectHandler{
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

// MultiCreateForProject preserves the endpoint's existing partial-result DTO.
// Collaboration delivery is best effort after the outer transaction commits.
func (h *MultiCreateForProjectHandler) MultiCreateForProject(
	ctx context.Context,
	command MultiCreateForProjectCommand,
) (*domainFacility.FieldDeviceMultiCreateResult, error) {
	outcome, err := h.Execute(ctx, command)
	if err != nil {
		return nil, err
	}
	for _, dispatchErr := range outcome.DispatchErrors {
		if h.reportError != nil {
			h.reportError(dispatchErr)
		}
	}
	return outcome.Result, nil
}

func (h *MultiCreateForProjectHandler) Execute(
	ctx context.Context,
	command MultiCreateForProjectCommand,
) (MultiCreateForProjectOutcome, error) {
	if h == nil || !h.transactionConfigured {
		return MultiCreateForProjectOutcome{}, ErrMultiCreateForProjectTransactionNotConfigured
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
			workflow MultiCreateForProjectWorkflow,
		) (committedProjectMultiCreate, error) {
			result, err := executeMultiCreateForProjectTransaction(
				txCtx,
				workflow,
				command,
				operationID,
				h.historyBatch,
			)
			if err != nil {
				return committedProjectMultiCreate{}, err
			}
			if len(result.fieldDevices) > 0 {
				collaborationCommand = appcollaboration.FieldDevicesCreated{
					Envelope: appcollaboration.Envelope{
						SchemaVersion: appcollaboration.SchemaVersionV2,
						EventID:       eventID, OperationID: operationID, CorrelationID: operationID,
						ProjectID: command.ProjectID, ActorID: actorID, OccurredAt: occurredAt,
					},
					FieldDevices: result.fieldDevices,
				}
				if _, err := appcollaboration.EnqueueCommand(txCtx, collaborationCommand); err != nil {
					return committedProjectMultiCreate{}, fmt.Errorf("enqueue project FieldDevice multi-create: %w", err)
				}
			}
			return result, nil
		},
	)
	if err != nil {
		return MultiCreateForProjectOutcome{}, err
	}

	batchID := operationID
	result := mutation.Result{
		OperationID: operationID,
		ActorID:     actorID,
		OccurredAt:  occurredAt,
		ProjectIDs:  []uuid.UUID{command.ProjectID},
		Changes:     committed.changes,
	}
	if committed.batched {
		result.BatchID = &batchID
	}
	outcome := MultiCreateForProjectOutcome{
		Result:   committed.result,
		Mutation: result,
	}
	if h.dispatcher == nil || len(committed.fieldDevices) == 0 {
		return outcome, nil
	}

	dispatchCtx := context.WithoutCancel(ctx)
	if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, collaborationCommand); dispatchErr != nil {
		outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
			"dispatch project FieldDevice multi-create for project %s: %w",
			command.ProjectID,
			dispatchErr,
		))
	}
	return outcome, nil
}

func executeMultiCreateForProjectTransaction(
	ctx context.Context,
	workflow MultiCreateForProjectWorkflow,
	command MultiCreateForProjectCommand,
	operationID uuid.UUID,
	historyBatch HistoryBatchContext,
) (committedProjectMultiCreate, error) {
	if workflow == nil {
		return committedProjectMultiCreate{}, ErrMultiCreateForProjectTransactionNotConfigured
	}

	writeCtx := ctx
	batched := historyBatch != nil
	if batched {
		writeCtx = historyBatch(ctx, operationID)
	}
	result, err := workflow.MultiCreateAndAssignFieldDevices(
		writeCtx,
		command.ProjectID,
		command.Items,
	)
	if err != nil {
		return committedProjectMultiCreate{}, err
	}
	normalizeMultiCreateResult(result, command.Items)
	return committedProjectMultiCreate{
		result:       result,
		changes:      successfulCreateChanges(result),
		fieldDevices: successfulFieldDeviceStates(result),
		batched:      batched,
	}, nil
}

func successfulFieldDeviceStates(
	result *domainFacility.FieldDeviceMultiCreateResult,
) []appcollaboration.FieldDeviceState {
	if result == nil || result.SuccessCount == 0 {
		return nil
	}

	states := make([]appcollaboration.FieldDeviceState, 0, result.SuccessCount)
	for _, item := range result.Results {
		if !item.Success || item.FieldDevice == nil || item.FieldDevice.ID == uuid.Nil {
			continue
		}
		fieldDevice := item.FieldDevice
		states = append(states, appcollaboration.FieldDeviceState{
			ID:                        fieldDevice.ID,
			Revision:                  fieldDevice.Revision,
			BMK:                       clonePointer(fieldDevice.BMK),
			Description:               clonePointer(fieldDevice.Description),
			TextFix:                   clonePointer(fieldDevice.TextIndividuell),
			ApparatNumber:             fieldDevice.ApparatNr,
			SPSControllerSystemTypeID: fieldDevice.SPSControllerSystemTypeID,
			SystemPartID:              fieldDevice.SystemPartID,
			SpecificationID:           clonePointer(fieldDevice.SpecificationID),
			ApparatID:                 fieldDevice.ApparatID,
			CreatedAt:                 fieldDevice.CreatedAt,
			UpdatedAt:                 fieldDevice.UpdatedAt,
		})
	}
	return states
}
