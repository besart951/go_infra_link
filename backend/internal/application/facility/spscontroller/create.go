package spscontroller

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

var ErrCreateTransactionNotConfigured = errors.New("SPS controller create transaction is not configured")

type CreateWorkflow interface {
	CreateWithSystemTypes(
		context.Context,
		*domainFacility.SPSController,
		[]domainFacility.SPSControllerSystemType,
	) error
	GetByID(context.Context, uuid.UUID) (*domainFacility.SPSController, error)
}

type CreateCommand struct {
	ControlCabinetID  uuid.UUID
	GADevice          *string
	DeviceName        string
	DeviceDescription *string
	DeviceLocation    *string
	IPAddress         *string
	Subnet            *string
	Gateway           *string
	VLAN              *string
	SystemTypes       []domainFacility.SPSControllerSystemType
}

func (c CreateCommand) toDomain() *domainFacility.SPSController {
	return &domainFacility.SPSController{
		ControlCabinetID:  c.ControlCabinetID,
		GADevice:          clonePointer(c.GADevice),
		DeviceName:        c.DeviceName,
		DeviceDescription: clonePointer(c.DeviceDescription),
		DeviceLocation:    clonePointer(c.DeviceLocation),
		IPAddress:         clonePointer(c.IPAddress),
		Subnet:            clonePointer(c.Subnet),
		Gateway:           clonePointer(c.Gateway),
		Vlan:              clonePointer(c.VLAN),
	}
}

type CreateDependencies struct {
	TransactionRunner   apptransaction.Runner
	TransactionWorkflow apptransaction.Factory[CreateWorkflow]
	HistoryBatch        HistoryBatchContext
	ProjectLinks        ProjectLinkReader
	Dispatcher          appcollaboration.CommandDispatcher
	Actor               ActorProvider
	NewID               IDGenerator
	Now                 Clock
	ReportError         ErrorReporter
}

type CreateHandler struct {
	operation             apptransaction.Operation[CreateWorkflow, CreateWorkflow]
	transactionConfigured bool
	historyBatch          HistoryBatchContext
	projectLinks          ProjectLinkReader
	dispatcher            appcollaboration.CommandDispatcher
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
	reportError           ErrorReporter
}

type CreateOutcome struct {
	SPSController  *domainFacility.SPSController
	Mutation       mutation.Result
	DispatchErrors []error
}

type committedCreate struct {
	controller *domainFacility.SPSController
	change     mutation.EntityChange
	batched    bool
}

func NewCreateHandler(deps CreateDependencies) *CreateHandler {
	newID := deps.NewID
	if newID == nil {
		newID = uuid.New
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	boundary := apptransaction.NewBoundary[CreateWorkflow](
		deps.TransactionRunner,
		deps.TransactionWorkflow,
	)
	var noNonTransactionalWorkflow CreateWorkflow
	operation := apptransaction.Bind(
		boundary,
		noNonTransactionalWorkflow,
		func(workflow CreateWorkflow) CreateWorkflow { return workflow },
	)
	return &CreateHandler{
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

func (h *CreateHandler) Create(
	ctx context.Context,
	command CreateCommand,
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

func (h *CreateHandler) Execute(
	ctx context.Context,
	command CreateCommand,
) (CreateOutcome, error) {
	if h == nil || !h.transactionConfigured {
		return CreateOutcome{}, ErrCreateTransactionNotConfigured
	}

	operationID := h.newID()
	actorID := actorFromContext(h.actor, ctx)
	committed, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(txCtx context.Context, workflow CreateWorkflow) (committedCreate, error) {
			return executeCreateTransaction(
				txCtx,
				workflow,
				command,
				operationID,
				h.historyBatch,
			)
		},
	)
	if err != nil {
		return CreateOutcome{}, err
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
	outcome := CreateOutcome{
		SPSController: committed.controller,
		Mutation:      result,
	}
	if h.projectLinks == nil || h.dispatcher == nil {
		return outcome, nil
	}

	dispatchCtx := context.WithoutCancel(ctx)
	links, err := h.projectLinks.GetBySPSControllerIDs(
		dispatchCtx,
		[]uuid.UUID{committed.controller.ID},
	)
	if err != nil {
		outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
			"resolve created SPSController collaboration projects: %w",
			err,
		))
		return outcome, nil
	}

	projectIDs := linkedProjectIDs(links, committed.controller.ID)
	outcome.Mutation.ProjectIDs = append([]uuid.UUID(nil), projectIDs...)
	state := toCollaborationState(committed.controller)
	for _, projectID := range projectIDs {
		collaborationCommand := appcollaboration.SPSControllerCreated{
			Envelope: appcollaboration.Envelope{
				SchemaVersion: appcollaboration.SchemaVersionV1,
				EventID:       h.newID(),
				OperationID:   operationID,
				CorrelationID: operationID,
				ProjectID:     projectID,
				ActorID:       actorID,
				OccurredAt:    occurredAt,
			},
			SPSController: state,
		}
		if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, collaborationCommand); dispatchErr != nil {
			outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
				"dispatch created SPSController for project %s: %w",
				projectID,
				dispatchErr,
			))
		}
	}
	return outcome, nil
}

func executeCreateTransaction(
	ctx context.Context,
	workflow CreateWorkflow,
	command CreateCommand,
	operationID uuid.UUID,
	historyBatch HistoryBatchContext,
) (committedCreate, error) {
	if workflow == nil {
		return committedCreate{}, ErrCreateTransactionNotConfigured
	}

	writeCtx := ctx
	batched := historyBatch != nil
	if batched {
		writeCtx = historyBatch(ctx, operationID)
	}
	controller := command.toDomain()
	systemTypes := cloneSystemTypes(command.SystemTypes)
	if err := workflow.CreateWithSystemTypes(writeCtx, controller, systemTypes); err != nil {
		return committedCreate{}, err
	}
	if controller.ID == uuid.Nil {
		return committedCreate{}, domain.ErrInvalidArgument
	}
	after, err := workflow.GetByID(writeCtx, controller.ID)
	if err != nil {
		return committedCreate{}, err
	}
	if after == nil {
		return committedCreate{}, domain.ErrNotFound
	}
	if after.ControlCabinetID != command.ControlCabinetID {
		return committedCreate{}, domain.ErrInvalidArgument
	}
	change, err := buildCreateChange(after)
	if err != nil {
		return committedCreate{}, err
	}
	return committedCreate{
		controller: cloneSPSController(after),
		change:     change,
		batched:    batched,
	}, nil
}
