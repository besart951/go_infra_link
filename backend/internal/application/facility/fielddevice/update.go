package fielddevice

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
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
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

type transactionalUpdateOutbox interface {
	UpdateWorkflow
	GetByFieldDeviceIDs(context.Context, []uuid.UUID) ([]*domainProject.ProjectFieldDevice, error)
}

type projectAssignmentMoveReconciler interface {
	ReconcileFieldDeviceMove(
		context.Context,
		uuid.UUID,
		uuid.UUID,
		uuid.UUID,
	) ([]uuid.UUID, error)
}

type HistoryBatchContext func(context.Context, uuid.UUID) context.Context

type UpdateCommand struct {
	FieldDeviceID             uuid.UUID
	ExpectedVersion           uint64
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
	projectIDs  []uuid.UUID
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
	occurredAt := h.now().UTC()
	committed, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(txCtx context.Context, workflow UpdateWorkflow) (committedUpdate, error) {
			return executeUpdateTransaction(
				txCtx,
				workflow,
				command,
				operationID,
				actorID,
				occurredAt,
				h.newID,
				h.historyBatch,
			)
		},
	)
	if err != nil {
		return UpdateOutcome{}, err
	}

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

	if h.dispatcher == nil {
		return outcome, nil
	}

	dispatchCtx := context.WithoutCancel(ctx)
	projectIDs := append([]uuid.UUID(nil), committed.projectIDs...)
	if len(projectIDs) == 0 && h.projectLinks != nil {
		links, err := h.projectLinks.GetByFieldDeviceIDs(dispatchCtx, []uuid.UUID{command.FieldDeviceID})
		if err != nil {
			outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf("resolve FieldDevice collaboration projects: %w", err))
			return outcome, nil
		}
		projectIDs = sortedProjectIDs(groupLinkedFieldDevices(links, []uuid.UUID{command.FieldDeviceID}))
	}
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
	actorID *uuid.UUID,
	occurredAt time.Time,
	newID IDGenerator,
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
	if command.ExpectedVersion != 0 && before.Revision != command.ExpectedVersion {
		return committedUpdate{}, &domain.RevisionConflict{
			EntityID: before.ID,
			Expected: command.ExpectedVersion,
			Current:  before.Revision,
		}
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

	var reconciliationProjectIDs []uuid.UUID
	if move != nil && move.movesParent() {
		if reconciler, ok := workflow.(projectAssignmentMoveReconciler); ok {
			reconciliationProjectIDs, err = reconciler.ReconcileFieldDeviceMove(
				writeCtx,
				command.FieldDeviceID,
				move.From.SPSControllerSystemTypeID,
				move.To.SPSControllerSystemTypeID,
			)
			if err != nil {
				return committedUpdate{}, fmt.Errorf(
					"reconcile FieldDevice project assignments: %w",
					err,
				)
			}
		}
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
	projectIDs, err := enqueueTransactionalUpdateCommands(
		writeCtx, workflow, command.FieldDeviceID, after.Revision, move,
		reconciliationProjectIDs, operationID, actorID, occurredAt, newID,
	)
	if err != nil {
		return committedUpdate{}, err
	}

	return committedUpdate{
		fieldDevice: cloneFieldDevice(after),
		changes:     changes,
		move:        move,
		batched:     batched,
		projectIDs:  projectIDs,
	}, nil
}

func enqueueTransactionalUpdateCommands(
	ctx context.Context,
	workflow UpdateWorkflow,
	fieldDeviceID uuid.UUID,
	revision uint64,
	move *MoveCommand,
	reconciliationProjectIDs []uuid.UUID,
	operationID uuid.UUID,
	actorID *uuid.UUID,
	occurredAt time.Time,
	newID IDGenerator,
) ([]uuid.UUID, error) {
	outbox, ok := workflow.(transactionalUpdateOutbox)
	if !ok {
		return nil, nil
	}
	links, err := outbox.GetByFieldDeviceIDs(ctx, []uuid.UUID{fieldDeviceID})
	if err != nil {
		return nil, fmt.Errorf("resolve FieldDevice collaboration projects for outbox: %w", err)
	}
	projectIDs := mergeProjectIDs(
		sortedProjectIDs(groupLinkedFieldDevices(links, []uuid.UUID{fieldDeviceID})),
		reconciliationProjectIDs,
	)
	for _, projectID := range projectIDs {
		envelope := appcollaboration.Envelope{
			SchemaVersion:   appcollaboration.SchemaVersionV2,
			EventID:         newID(),
			OperationID:     operationID,
			CorrelationID:   operationID,
			ProjectID:       projectID,
			ActorID:         actorID,
			OccurredAt:      occurredAt,
			EntityRevisions: map[string]uint64{fieldDeviceID.String(): revision},
		}
		var event appcollaboration.Command = appcollaboration.FieldDeviceUpdated{Envelope: envelope, FieldDeviceID: fieldDeviceID}
		if move != nil && move.movesParent() {
			event = appcollaboration.FieldDeviceMoved{
				Envelope: envelope, FieldDeviceID: fieldDeviceID,
				FromSPSControllerSystemTypeID: move.From.SPSControllerSystemTypeID,
				ToSPSControllerSystemTypeID:   move.To.SPSControllerSystemTypeID,
			}
		}
		configured, err := appcollaboration.EnqueueCommand(ctx, event)
		if err != nil {
			return nil, fmt.Errorf("enqueue FieldDevice collaboration event for project %s: %w", projectID, err)
		}
		if !configured {
			return nil, nil
		}
	}
	return projectIDs, nil
}

func mergeProjectIDs(groups ...[]uuid.UUID) []uuid.UUID {
	set := make(map[uuid.UUID]struct{})
	for _, group := range groups {
		for _, projectID := range group {
			if projectID != uuid.Nil {
				set[projectID] = struct{}{}
			}
		}
	}
	out := make([]uuid.UUID, 0, len(set))
	for projectID := range set {
		out = append(out, projectID)
	}
	sort.Slice(out, func(i, j int) bool {
		return out[i].String() < out[j].String()
	})
	return out
}
