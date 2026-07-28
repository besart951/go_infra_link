package controlcabinet

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
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

var ErrUpdateTransactionNotConfigured = errors.New("control cabinet update transaction is not configured")

// UpdateWorkflow is the transaction-scoped Interface consumed by this Module.
// The legacy facility service remains the first Implementation so its mature
// destination, uniqueness, and descendant-name rules stay in one place.
type UpdateWorkflow interface {
	GetByID(context.Context, uuid.UUID) (*domainFacility.ControlCabinet, error)
	Update(context.Context, *domainFacility.ControlCabinet) error
}

type ProjectLinkReader interface {
	GetByControlCabinetIDs(context.Context, []uuid.UUID) ([]*domainProject.ProjectControlCabinet, error)
}

type transactionalUpdateOutbox interface {
	UpdateWorkflow
	GetByControlCabinetIDs(context.Context, []uuid.UUID) ([]*domainProject.ProjectControlCabinet, error)
}

type descendantChangeCounter interface {
	CountSPSControllers(context.Context, uuid.UUID) (int64, error)
}

type HistoryBatchContext func(context.Context, uuid.UUID) context.Context
type ActorProvider func(context.Context) *uuid.UUID
type IDGenerator func() uuid.UUID
type Clock func() time.Time
type ErrorReporter func(error)

type UpdateCommand struct {
	ControlCabinetID uuid.UUID
	ExpectedVersion  uint64
	BuildingID       *uuid.UUID
	ControlCabinetNr *string
}

func (c UpdateCommand) validate() error {
	if c.ControlCabinetID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	return nil
}

func (c UpdateCommand) applyContentTo(cabinet *domainFacility.ControlCabinet) {
	if cabinet == nil {
		return
	}
	if c.ControlCabinetNr != nil {
		cabinet.ControlCabinetNr = clonePointer(c.ControlCabinetNr)
	}
}

type LoadError struct {
	Err error
}

func (e *LoadError) Error() string {
	if e == nil || e.Err == nil {
		return "load control cabinet"
	}
	return "load control cabinet: " + e.Err.Error()
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
	ControlCabinet *domainFacility.ControlCabinet
	Mutation       mutation.Result
	DispatchErrors []error
}

type committedUpdate struct {
	cabinet    *domainFacility.ControlCabinet
	change     mutation.EntityChange
	move       *MoveCommand
	batched    bool
	projectIDs []uuid.UUID
	childCount int64
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

// Update preserves the existing HTTP result. Collaboration remains best effort
// after commit and cannot turn a committed mutation into an HTTP failure.
func (h *UpdateHandler) Update(
	ctx context.Context,
	command UpdateCommand,
) (*domainFacility.ControlCabinet, error) {
	outcome, err := h.Execute(ctx, command)
	if err != nil {
		return nil, err
	}
	for _, dispatchErr := range outcome.DispatchErrors {
		if h.reportError != nil {
			h.reportError(dispatchErr)
		}
	}
	return outcome.ControlCabinet, nil
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
		Changes:     []mutation.EntityChange{committed.change},
	}
	if committed.childCount > 0 {
		result.Aggregates = []mutation.AggregateChange{{
			EntityType:    mutation.EntityTypeSPSController,
			ParentID:      command.ControlCabinetID,
			Action:        domainHistory.ActionUpdate,
			ChangedFields: []mutation.FieldName{mutation.FieldNameDeviceName},
			Count:         committed.childCount,
		}}
	}
	if committed.batched {
		batchID := operationID
		result.BatchID = &batchID
	}
	outcome := UpdateOutcome{
		ControlCabinet: committed.cabinet,
		Mutation:       result,
	}

	if h.dispatcher == nil {
		return outcome, nil
	}

	dispatchCtx := context.WithoutCancel(ctx)
	projectIDs := append([]uuid.UUID(nil), committed.projectIDs...)
	if len(projectIDs) == 0 && h.projectLinks != nil {
		links, err := h.projectLinks.GetByControlCabinetIDs(
			dispatchCtx,
			[]uuid.UUID{command.ControlCabinetID},
		)
		if err != nil {
			outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
				"resolve ControlCabinet collaboration projects: %w",
				err,
			))
			return outcome, nil
		}
		projectIDs = linkedProjectIDs(links, command.ControlCabinetID)
	}

	outcome.Mutation.ProjectIDs = append([]uuid.UUID(nil), projectIDs...)
	state := toCollaborationState(committed.cabinet)
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
			collaborationCommand = appcollaboration.ControlCabinetMoved{
				Envelope:       envelope,
				ControlCabinet: state,
				FromBuildingID: committed.move.FromBuildingID,
				ToBuildingID:   committed.move.ToBuildingID,
			}
		} else {
			collaborationCommand = appcollaboration.ControlCabinetUpdated{
				Envelope:       envelope,
				ControlCabinet: state,
			}
		}
		if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, collaborationCommand); dispatchErr != nil {
			outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
				"dispatch ControlCabinet mutation for project %s: %w",
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

	before, err := workflow.GetByID(ctx, command.ControlCabinetID)
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

	updated := cloneControlCabinet(before)
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
	if err := workflow.Update(writeCtx, updated); err != nil {
		return committedUpdate{}, err
	}
	var childCount int64
	if command.BuildingID != nil || command.ControlCabinetNr != nil {
		if counter, ok := workflow.(descendantChangeCounter); ok {
			childCount, err = counter.CountSPSControllers(writeCtx, command.ControlCabinetID)
			if err != nil {
				return committedUpdate{}, fmt.Errorf("count regenerated SPSController names: %w", err)
			}
		}
	}

	after, err := workflow.GetByID(writeCtx, command.ControlCabinetID)
	if err != nil {
		return committedUpdate{}, err
	}
	if after == nil {
		return committedUpdate{}, domain.ErrNotFound
	}

	change, err := buildUpdateChange(before, after)
	if err != nil {
		return committedUpdate{}, err
	}
	projectIDs, err := enqueueTransactionalUpdateCommands(
		writeCtx,
		workflow,
		command.ControlCabinetID,
		after,
		move,
		operationID,
		actorID,
		occurredAt,
		newID,
	)
	if err != nil {
		return committedUpdate{}, err
	}
	return committedUpdate{
		cabinet:    cloneControlCabinet(after),
		change:     change,
		move:       move,
		batched:    batched,
		projectIDs: projectIDs,
		childCount: childCount,
	}, nil
}

func enqueueTransactionalUpdateCommands(
	ctx context.Context,
	workflow UpdateWorkflow,
	controlCabinetID uuid.UUID,
	after *domainFacility.ControlCabinet,
	move *MoveCommand,
	operationID uuid.UUID,
	actorID *uuid.UUID,
	occurredAt time.Time,
	newID IDGenerator,
) ([]uuid.UUID, error) {
	outbox, ok := workflow.(transactionalUpdateOutbox)
	if !ok {
		return nil, nil
	}
	links, err := outbox.GetByControlCabinetIDs(ctx, []uuid.UUID{controlCabinetID})
	if err != nil {
		return nil, fmt.Errorf("resolve ControlCabinet collaboration projects for outbox: %w", err)
	}
	projectIDs := linkedProjectIDs(links, controlCabinetID)
	state := toCollaborationState(after)
	for _, projectID := range projectIDs {
		envelope := appcollaboration.Envelope{
			SchemaVersion: appcollaboration.SchemaVersionV2,
			EventID:       newID(),
			OperationID:   operationID,
			CorrelationID: operationID,
			ProjectID:     projectID,
			ActorID:       actorID,
			OccurredAt:    occurredAt,
			EntityRevisions: map[string]uint64{
				controlCabinetID.String(): after.Revision,
			},
		}
		var event appcollaboration.Command = appcollaboration.ControlCabinetUpdated{
			Envelope:       envelope,
			ControlCabinet: state,
		}
		if move != nil {
			event = appcollaboration.ControlCabinetMoved{
				Envelope:       envelope,
				ControlCabinet: state,
				FromBuildingID: move.FromBuildingID,
				ToBuildingID:   move.ToBuildingID,
			}
		}
		configured, err := appcollaboration.EnqueueCommand(ctx, event)
		if err != nil {
			return nil, fmt.Errorf(
				"enqueue ControlCabinet collaboration event for project %s: %w",
				projectID,
				err,
			)
		}
		if !configured {
			return nil, nil
		}
	}
	return projectIDs, nil
}

func linkedProjectIDs(
	links []*domainProject.ProjectControlCabinet,
	controlCabinetID uuid.UUID,
) []uuid.UUID {
	set := make(map[uuid.UUID]struct{}, len(links))
	for _, link := range links {
		if link == nil || link.ProjectID == uuid.Nil || link.ControlCabinetID != controlCabinetID {
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

func cloneControlCabinet(cabinet *domainFacility.ControlCabinet) *domainFacility.ControlCabinet {
	if cabinet == nil {
		return nil
	}
	clone := *cabinet
	clone.ControlCabinetNr = clonePointer(cabinet.ControlCabinetNr)
	clone.SPSControllers = append([]domainFacility.SPSController(nil), cabinet.SPSControllers...)
	return &clone
}
