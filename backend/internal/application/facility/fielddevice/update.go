package fielddevice

import (
	"context"
	"errors"
	"fmt"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	"github.com/besart951/go_infra_link/backend/internal/application/facility/mutation"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	"github.com/google/uuid"
)

var ErrUpdateTransactionNotConfigured = errors.New("field device update transaction is not configured")

// UpdateWorkflow is the narrow Interface the application Module needs inside a
// transaction. The existing facility implementation is its first Adapter.
type UpdateWorkflow interface {
	GetByID(context.Context, uuid.UUID) (*domainFacility.FieldDevice, error)
	ListBacnetObjects(context.Context, uuid.UUID) ([]domainFacility.BacnetObject, error)
	UpdateWithBacnetObjects(
		context.Context,
		*domainFacility.FieldDevice,
		*uuid.UUID,
		*[]domainFacility.BacnetObject,
	) error
}

type HistoryBatchContext func(context.Context, uuid.UUID) context.Context

type UpdateCommand struct {
	FieldDeviceID             uuid.UUID
	BMK                       *string
	Description               *string
	TextIndividuell           *string
	ApparatNr                 *int
	SPSControllerSystemTypeID *uuid.UUID
	SystemPartID              *uuid.UUID
	ApparatID                 *uuid.UUID
	ObjectDataID              *uuid.UUID
	BacnetObjects             *[]domainFacility.BacnetObject
}

func (c UpdateCommand) validate() error {
	if c.FieldDeviceID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	if c.ObjectDataID != nil && c.BacnetObjects != nil {
		return domain.ErrInvalidArgument
	}
	return nil
}

func (c UpdateCommand) replacesBacnetObjects() bool {
	return c.ObjectDataID != nil || c.BacnetObjects != nil
}

func (c UpdateCommand) applyContentTo(fieldDevice *domainFacility.FieldDevice) {
	if fieldDevice == nil {
		return
	}
	if c.BMK != nil {
		fieldDevice.BMK = clonePointer(c.BMK)
	}
	if c.Description != nil {
		fieldDevice.Description = clonePointer(c.Description)
	}
	if c.TextIndividuell != nil {
		fieldDevice.TextIndividuell = clonePointer(c.TextIndividuell)
	}
}

type LoadError struct {
	Err error
}

func (e *LoadError) Error() string {
	if e == nil || e.Err == nil {
		return "load field device"
	}
	return "load field device: " + e.Err.Error()
}

func (e *LoadError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

type UpdateDependencies struct {
	TransactionRunner   apptransaction.Runner
	TransactionWorkflow apptransaction.Factory[UpdateWorkflow]
	HistoryBatch        HistoryBatchContext
	ProjectLinks        ProjectLinkReader
	Dispatcher          appcollaboration.CommandDispatcher
	Actor               ActorProvider
	NewID               IDGenerator
	Now                 Clock
	ReportError         ErrorReporter
}

type UpdateHandler struct {
	operation             apptransaction.Operation[UpdateWorkflow, UpdateWorkflow]
	transactionConfigured bool
	historyBatch          HistoryBatchContext
	projectLinks          ProjectLinkReader
	dispatcher            appcollaboration.CommandDispatcher
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
	reportError           ErrorReporter
}

type UpdateOutcome struct {
	FieldDevice    *domainFacility.FieldDevice
	Mutation       mutation.Result
	DispatchErrors []error
}

type committedUpdate struct {
	fieldDevice *domainFacility.FieldDevice
	changes     []mutation.EntityChange
	move        *MoveCommand
	batched     bool
}

func NewUpdateHandler(deps UpdateDependencies) *UpdateHandler {
	newID := deps.NewID
	if newID == nil {
		newID = uuid.New
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	boundary := apptransaction.NewBoundary[UpdateWorkflow](
		deps.TransactionRunner,
		deps.TransactionWorkflow,
	)
	var noNonTransactionalWorkflow UpdateWorkflow
	operation := apptransaction.Bind(
		boundary,
		noNonTransactionalWorkflow,
		func(workflow UpdateWorkflow) UpdateWorkflow { return workflow },
	)

	return &UpdateHandler{
		operation:             operation,
		transactionConfigured: deps.TransactionRunner != nil && deps.TransactionWorkflow != nil,
		historyBatch:          deps.HistoryBatch,
		projectLinks:          deps.ProjectLinks,
		dispatcher:            deps.Dispatcher,
		actor:                 deps.Actor,
		newID:                 newID,
		now:                   now,
		reportError:           deps.ReportError,
	}
}

// Update preserves the existing HTTP-facing result. Collaboration delivery is
// best effort after commit and cannot turn a committed mutation into an HTTP
// failure.
func (h *UpdateHandler) Update(
	ctx context.Context,
	command UpdateCommand,
) (*domainFacility.FieldDevice, error) {
	outcome, err := h.Execute(ctx, command)
	if err != nil {
		return nil, err
	}
	for _, dispatchErr := range outcome.DispatchErrors {
		if h.reportError != nil {
			h.reportError(dispatchErr)
		}
	}
	return outcome.FieldDevice, nil
}

func (h *UpdateHandler) Execute(
	ctx context.Context,
	command UpdateCommand,
) (UpdateOutcome, error) {
	if h == nil || !h.transactionConfigured {
		return UpdateOutcome{}, ErrUpdateTransactionNotConfigured
	}
	if err := command.validate(); err != nil {
		return UpdateOutcome{}, err
	}

	operationID := h.newID()
	actorID := actorFromContext(h.actor, ctx)
	committed, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(txCtx context.Context, workflow UpdateWorkflow) (committedUpdate, error) {
			return executeUpdateTransaction(
				txCtx,
				workflow,
				command,
				operationID,
				h.historyBatch,
			)
		},
	)
	if err != nil {
		return UpdateOutcome{}, err
	}

	occurredAt := h.now().UTC()
	result := mutation.Result{
		OperationID: operationID,
		ActorID:     actorID,
		OccurredAt:  occurredAt,
		Changes:     committed.changes,
	}
	if committed.batched {
		batchID := operationID
		result.BatchID = &batchID
	}
	outcome := UpdateOutcome{
		FieldDevice: committed.fieldDevice,
		Mutation:    result,
	}

	if h.projectLinks == nil || h.dispatcher == nil {
		return outcome, nil
	}

	dispatchCtx := context.WithoutCancel(ctx)
	links, err := h.projectLinks.GetByFieldDeviceIDs(
		dispatchCtx,
		[]uuid.UUID{command.FieldDeviceID},
	)
	if err != nil {
		outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
			"resolve FieldDevice collaboration projects: %w",
			err,
		))
		return outcome, nil
	}

	grouped := groupLinkedFieldDevices(links, []uuid.UUID{command.FieldDeviceID})
	projectIDs := sortedProjectIDs(grouped)
	outcome.Mutation.ProjectIDs = append([]uuid.UUID(nil), projectIDs...)
	for _, projectID := range projectIDs {
		envelope := appcollaboration.Envelope{
			SchemaVersion: appcollaboration.SchemaVersionV1,
			EventID:       h.newID(),
			OperationID:   operationID,
			CorrelationID: operationID,
			ProjectID:     projectID,
			ActorID:       actorID,
			OccurredAt:    occurredAt,
		}
		var collaborationCommand appcollaboration.Command
		if committed.move != nil && committed.move.movesParent() {
			collaborationCommand = appcollaboration.FieldDeviceMoved{
				Envelope:                      envelope,
				FieldDeviceID:                 command.FieldDeviceID,
				FromSPSControllerSystemTypeID: committed.move.From.SPSControllerSystemTypeID,
				ToSPSControllerSystemTypeID:   committed.move.To.SPSControllerSystemTypeID,
			}
		} else {
			collaborationCommand = appcollaboration.FieldDeviceUpdated{
				Envelope:      envelope,
				FieldDeviceID: command.FieldDeviceID,
			}
		}
		if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, collaborationCommand); dispatchErr != nil {
			outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
				"dispatch FieldDevice mutation for project %s: %w",
				projectID,
				dispatchErr,
			))
		}
	}

	return outcome, nil
}

func executeUpdateTransaction(
	ctx context.Context,
	workflow UpdateWorkflow,
	command UpdateCommand,
	operationID uuid.UUID,
	historyBatch HistoryBatchContext,
) (committedUpdate, error) {
	if workflow == nil {
		return committedUpdate{}, ErrUpdateTransactionNotConfigured
	}

	before, err := workflow.GetByID(ctx, command.FieldDeviceID)
	if err != nil {
		return committedUpdate{}, &LoadError{Err: err}
	}
	if before == nil {
		return committedUpdate{}, &LoadError{Err: domain.ErrNotFound}
	}

	var beforeBacnet []domainFacility.BacnetObject
	if command.replacesBacnetObjects() {
		beforeBacnet, err = workflow.ListBacnetObjects(ctx, command.FieldDeviceID)
		if err != nil {
			return committedUpdate{}, err
		}
	}

	updated := cloneFieldDevice(before)
	command.applyContentTo(updated)
	move, err := newMoveCommand(before, command)
	if err != nil {
		return committedUpdate{}, err
	}
	if move != nil {
		if err := move.applyTo(updated); err != nil {
			return committedUpdate{}, err
		}
	}

	batched := command.replacesBacnetObjects() || move != nil
	writeCtx := ctx
	if batched && historyBatch != nil {
		writeCtx = historyBatch(ctx, operationID)
	}
	objectDataID := clonePointer(command.ObjectDataID)
	bacnetObjects := cloneBacnetObjectSelection(command.BacnetObjects)
	if err := workflow.UpdateWithBacnetObjects(writeCtx, updated, objectDataID, bacnetObjects); err != nil {
		return committedUpdate{}, err
	}

	after, err := workflow.GetByID(writeCtx, command.FieldDeviceID)
	if err != nil {
		return committedUpdate{}, err
	}
	if after == nil {
		return committedUpdate{}, domain.ErrNotFound
	}

	var afterBacnet []domainFacility.BacnetObject
	if command.replacesBacnetObjects() {
		afterBacnet, err = workflow.ListBacnetObjects(writeCtx, command.FieldDeviceID)
		if err != nil {
			return committedUpdate{}, err
		}
	}

	changes, err := buildUpdateChanges(
		before,
		after,
		beforeBacnet,
		afterBacnet,
		command.replacesBacnetObjects(),
	)
	if err != nil {
		return committedUpdate{}, err
	}

	return committedUpdate{
		fieldDevice: cloneFieldDevice(after),
		changes:     changes,
		move:        move,
		batched:     batched,
	}, nil
}
