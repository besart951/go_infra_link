package controlcabinet

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

var ErrCreateTransactionNotConfigured = errors.New("control cabinet create transaction is not configured")

// CreateWorkflow is the transaction-scoped Interface needed by cabinet
// creation. The legacy service remains its compatibility Implementation so
// building existence and scoped-number uniqueness rules stay unchanged.
type CreateWorkflow interface {
	Create(context.Context, *domainFacility.ControlCabinet) error
	GetByID(context.Context, uuid.UUID) (*domainFacility.ControlCabinet, error)
}

type CreateCommand struct {
	BuildingID       uuid.UUID
	ControlCabinetNr *string
}

func (c CreateCommand) toDomain() *domainFacility.ControlCabinet {
	return &domainFacility.ControlCabinet{
		BuildingID:       c.BuildingID,
		ControlCabinetNr: clonePointer(c.ControlCabinetNr),
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
	ControlCabinet *domainFacility.ControlCabinet
	Mutation       mutation.Result
	DispatchErrors []error
}

type committedCreate struct {
	cabinet *domainFacility.ControlCabinet
	change  mutation.EntityChange
	batched bool
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
		ControlCabinet: committed.cabinet,
		Mutation:       result,
	}
	if h.projectLinks == nil || h.dispatcher == nil {
		return outcome, nil
	}

	dispatchCtx := context.WithoutCancel(ctx)
	links, err := h.projectLinks.GetByControlCabinetIDs(
		dispatchCtx,
		[]uuid.UUID{committed.cabinet.ID},
	)
	if err != nil {
		outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
			"resolve created ControlCabinet collaboration projects: %w",
			err,
		))
		return outcome, nil
	}

	projectIDs := linkedProjectIDs(links, committed.cabinet.ID)
	outcome.Mutation.ProjectIDs = append([]uuid.UUID(nil), projectIDs...)
	state := toCollaborationState(committed.cabinet)
	for _, projectID := range projectIDs {
		collaborationCommand := appcollaboration.ControlCabinetCreated{
			Envelope: appcollaboration.Envelope{
				SchemaVersion: appcollaboration.SchemaVersionV1,
				EventID:       h.newID(),
				OperationID:   operationID,
				CorrelationID: operationID,
				ProjectID:     projectID,
				ActorID:       actorID,
				OccurredAt:    occurredAt,
			},
			ControlCabinet: state,
		}
		if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, collaborationCommand); dispatchErr != nil {
			outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
				"dispatch created ControlCabinet for project %s: %w",
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
	cabinet := command.toDomain()
	if err := workflow.Create(writeCtx, cabinet); err != nil {
		return committedCreate{}, err
	}
	if cabinet.ID == uuid.Nil {
		return committedCreate{}, domain.ErrInvalidArgument
	}
	after, err := workflow.GetByID(writeCtx, cabinet.ID)
	if err != nil {
		return committedCreate{}, err
	}
	if after == nil {
		return committedCreate{}, domain.ErrNotFound
	}
	if after.BuildingID != command.BuildingID {
		return committedCreate{}, domain.ErrInvalidArgument
	}
	change, err := buildCreateChange(after)
	if err != nil {
		return committedCreate{}, err
	}
	return committedCreate{
		cabinet: cloneControlCabinet(after),
		change:  change,
		batched: batched,
	}, nil
}
