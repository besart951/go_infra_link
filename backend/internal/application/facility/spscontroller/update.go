package spscontroller

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

var ErrUpdateTransactionNotConfigured = errors.New("SPS controller update transaction is not configured")

// UpdateWorkflow is the transaction-scoped Interface consumed by this Module.
// The legacy facility service is the first Implementation.
type UpdateWorkflow interface {
	GetByID(context.Context, uuid.UUID) (*domainFacility.SPSController, error)
	Update(context.Context, *domainFacility.SPSController) error
	UpdateWithSystemTypes(
		context.Context,
		*domainFacility.SPSController,
		[]domainFacility.SPSControllerSystemType,
	) error
}

type ProjectLinkReader interface {
	GetBySPSControllerIDs(context.Context, []uuid.UUID) ([]*domainProject.ProjectSPSController, error)
}

type HistoryBatchContext func(context.Context, uuid.UUID) context.Context
type ActorProvider func(context.Context) *uuid.UUID
type IDGenerator func() uuid.UUID
type Clock func() time.Time
type ErrorReporter func(error)

type UpdateCommand struct {
	SPSControllerID   uuid.UUID
	ControlCabinetID  *uuid.UUID
	GADevice          *string
	DeviceName        *string
	DeviceDescription *string
	DeviceLocation    *string
	IPAddress         *string
	Subnet            *string
	Gateway           *string
	VLAN              *string
	SystemTypes       *[]domainFacility.SPSControllerSystemType
}

func (c UpdateCommand) validate() error {
	if c.SPSControllerID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	return nil
}

func (c UpdateCommand) applyContentTo(controller *domainFacility.SPSController) {
	if controller == nil {
		return
	}
	if c.GADevice != nil {
		controller.GADevice = clonePointer(c.GADevice)
	}
	if c.DeviceName != nil {
		controller.DeviceName = *c.DeviceName
	}
	if c.DeviceDescription != nil {
		controller.DeviceDescription = clonePointer(c.DeviceDescription)
	}
	if c.DeviceLocation != nil {
		controller.DeviceLocation = clonePointer(c.DeviceLocation)
	}
	if c.IPAddress != nil {
		controller.IPAddress = clonePointer(c.IPAddress)
	}
	if c.Subnet != nil {
		controller.Subnet = clonePointer(c.Subnet)
	}
	if c.Gateway != nil {
		controller.Gateway = clonePointer(c.Gateway)
	}
	if c.VLAN != nil {
		controller.Vlan = clonePointer(c.VLAN)
	}
}

type LoadError struct {
	Err error
}

func (e *LoadError) Error() string {
	if e == nil || e.Err == nil {
		return "load SPS controller"
	}
	return "load SPS controller: " + e.Err.Error()
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
	SPSController  *domainFacility.SPSController
	Mutation       mutation.Result
	DispatchErrors []error
}

type committedUpdate struct {
	controller *domainFacility.SPSController
	change     mutation.EntityChange
	move       *MoveCommand
	batched    bool
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

// Update preserves the existing HTTP result. Collaboration delivery remains
// best effort after commit and cannot turn a committed mutation into an HTTP
// failure.
func (h *UpdateHandler) Update(
	ctx context.Context,
	command UpdateCommand,
) (*domainFacility.SPSController, error) {
	outcome, err := h.Execute(ctx, command)
	if err != nil {
		return nil, err
	}
	for _, dispatchErr := range outcome.DispatchErrors {
		if h.reportError != nil {
			h.reportError(dispatchErr)
		}
	}
	return outcome.SPSController, nil
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
		Changes:     []mutation.EntityChange{committed.change},
	}
	if committed.batched {
		batchID := operationID
		result.BatchID = &batchID
	}
	outcome := UpdateOutcome{
		SPSController: committed.controller,
		Mutation:      result,
	}

	if h.projectLinks == nil || h.dispatcher == nil {
		return outcome, nil
	}

	dispatchCtx := context.WithoutCancel(ctx)
	links, err := h.projectLinks.GetBySPSControllerIDs(
		dispatchCtx,
		[]uuid.UUID{command.SPSControllerID},
	)
	if err != nil {
		outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
			"resolve SPSController collaboration projects: %w",
			err,
		))
		return outcome, nil
	}

	projectIDs := linkedProjectIDs(links, command.SPSControllerID)
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
		if committed.move != nil {
			collaborationCommand = appcollaboration.SPSControllerMoved{
				Envelope:             envelope,
				SPSControllerID:      command.SPSControllerID,
				FromControlCabinetID: committed.move.FromControlCabinetID,
				ToControlCabinetID:   committed.move.ToControlCabinetID,
			}
		} else {
			collaborationCommand = appcollaboration.SPSControllerUpdated{
				Envelope:        envelope,
				SPSControllerID: command.SPSControllerID,
			}
		}
		if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, collaborationCommand); dispatchErr != nil {
			outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
				"dispatch SPSController mutation for project %s: %w",
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

	before, err := workflow.GetByID(ctx, command.SPSControllerID)
	if err != nil {
		return committedUpdate{}, &LoadError{Err: err}
	}
	if before == nil {
		return committedUpdate{}, &LoadError{Err: domain.ErrNotFound}
	}

	updated := cloneSPSController(before)
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

	writeCtx := ctx
	batched := historyBatch != nil
	if batched {
		writeCtx = historyBatch(ctx, operationID)
	}
	if command.SystemTypes != nil {
		if err := workflow.UpdateWithSystemTypes(
			writeCtx,
			updated,
			cloneSystemTypes(*command.SystemTypes),
		); err != nil {
			return committedUpdate{}, err
		}
	} else if err := workflow.Update(writeCtx, updated); err != nil {
		return committedUpdate{}, err
	}

	after, err := workflow.GetByID(writeCtx, command.SPSControllerID)
	if err != nil {
		return committedUpdate{}, err
	}
	if after == nil {
		return committedUpdate{}, domain.ErrNotFound
	}

	change, err := buildUpdateChange(before, after, command.SystemTypes != nil)
	if err != nil {
		return committedUpdate{}, err
	}
	return committedUpdate{
		controller: cloneSPSController(after),
		change:     change,
		move:       move,
		batched:    batched,
	}, nil
}

func linkedProjectIDs(
	links []*domainProject.ProjectSPSController,
	spsControllerID uuid.UUID,
) []uuid.UUID {
	set := make(map[uuid.UUID]struct{}, len(links))
	for _, link := range links {
		if link == nil || link.ProjectID == uuid.Nil || link.SPSControllerID != spsControllerID {
			continue
		}
		set[link.ProjectID] = struct{}{}
	}
	projectIDs := make([]uuid.UUID, 0, len(set))
	for projectID := range set {
		projectIDs = append(projectIDs, projectID)
	}
	sort.Slice(projectIDs, func(i, j int) bool {
		return projectIDs[i].String() < projectIDs[j].String()
	})
	return projectIDs
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

func clonePointer[T any](value *T) *T {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func cloneSPSController(controller *domainFacility.SPSController) *domainFacility.SPSController {
	if controller == nil {
		return nil
	}
	clone := *controller
	clone.GADevice = clonePointer(controller.GADevice)
	clone.DeviceDescription = clonePointer(controller.DeviceDescription)
	clone.DeviceLocation = clonePointer(controller.DeviceLocation)
	clone.IPAddress = clonePointer(controller.IPAddress)
	clone.Subnet = clonePointer(controller.Subnet)
	clone.Gateway = clonePointer(controller.Gateway)
	clone.Vlan = clonePointer(controller.Vlan)
	clone.SPSControllerSystemTypes = cloneSystemTypes(controller.SPSControllerSystemTypes)
	return &clone
}

func cloneSystemTypes(
	systemTypes []domainFacility.SPSControllerSystemType,
) []domainFacility.SPSControllerSystemType {
	clones := make([]domainFacility.SPSControllerSystemType, len(systemTypes))
	for i := range systemTypes {
		clones[i] = systemTypes[i]
		clones[i].Number = clonePointer(systemTypes[i].Number)
		clones[i].DocumentName = clonePointer(systemTypes[i].DocumentName)
	}
	return clones
}
