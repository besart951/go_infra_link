package bacnetobject

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

var ErrCreateTransactionNotConfigured = errors.New("BACnet object create transaction is not configured")

type CreateWorkflow interface {
	CreateWithParent(
		context.Context,
		*domainFacility.BacnetObject,
		*uuid.UUID,
		*uuid.UUID,
	) error
	GetByID(context.Context, uuid.UUID) (*domainFacility.BacnetObject, error)
	GetObjectDataByID(context.Context, uuid.UUID) (*domainFacility.ObjectData, error)
}

// CreateInput contains only client-controlled BACnet state. Identity,
// timestamps, relationships, and persistence metadata cannot be injected via
// this command.
type CreateInput struct {
	TextFix             string
	Description         *string
	GMSVisible          bool
	Optional            bool
	TextIndividual      *string
	SoftwareType        domainFacility.BacnetSoftwareType
	SoftwareNumber      uint16
	HardwareType        domainFacility.BacnetHardwareType
	HardwareQuantity    uint8
	SoftwareReferenceID *uuid.UUID
	StateTextID         *uuid.UUID
	NotificationClassID *uuid.UUID
	AlarmDefinitionID   *uuid.UUID
	AlarmTypeID         *uuid.UUID
}

func (i CreateInput) toDomain() *domainFacility.BacnetObject {
	return &domainFacility.BacnetObject{
		TextFix:             i.TextFix,
		Description:         clonePointer(i.Description),
		GMSVisible:          i.GMSVisible,
		Optional:            i.Optional,
		TextIndividual:      clonePointer(i.TextIndividual),
		SoftwareType:        i.SoftwareType,
		SoftwareNumber:      i.SoftwareNumber,
		HardwareType:        i.HardwareType,
		HardwareQuantity:    i.HardwareQuantity,
		SoftwareReferenceID: clonePointer(i.SoftwareReferenceID),
		StateTextID:         clonePointer(i.StateTextID),
		NotificationClassID: clonePointer(i.NotificationClassID),
		AlarmDefinitionID:   clonePointer(i.AlarmDefinitionID),
		AlarmTypeID:         clonePointer(i.AlarmTypeID),
	}
}

type CreateForFieldDeviceCommand struct {
	FieldDeviceID uuid.UUID
	Input         CreateInput
}

func (c CreateForFieldDeviceCommand) validate() error {
	if c.FieldDeviceID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	return nil
}

type CreateForObjectDataCommand struct {
	ObjectDataID uuid.UUID
	Input        CreateInput
}

func (c CreateForObjectDataCommand) validate() error {
	if c.ObjectDataID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	return nil
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
	BacnetObject   *domainFacility.BacnetObject
	Mutation       mutation.Result
	DispatchErrors []error
}

type committedCreate struct {
	object    *domainFacility.BacnetObject
	change    mutation.EntityChange
	projectID *uuid.UUID
	batched   bool
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

func (h *CreateHandler) CreateForFieldDevice(
	ctx context.Context,
	command CreateForFieldDeviceCommand,
) (*domainFacility.BacnetObject, error) {
	outcome, err := h.ExecuteForFieldDevice(ctx, command)
	if err != nil {
		return nil, err
	}
	for _, dispatchErr := range outcome.DispatchErrors {
		if h.reportError != nil {
			h.reportError(dispatchErr)
		}
	}
	return outcome.BacnetObject, nil
}

func (h *CreateHandler) CreateForObjectData(
	ctx context.Context,
	command CreateForObjectDataCommand,
) (*domainFacility.BacnetObject, error) {
	outcome, err := h.ExecuteForObjectData(ctx, command)
	if err != nil {
		return nil, err
	}
	for _, dispatchErr := range outcome.DispatchErrors {
		if h.reportError != nil {
			h.reportError(dispatchErr)
		}
	}
	return outcome.BacnetObject, nil
}

func (h *CreateHandler) ExecuteForFieldDevice(
	ctx context.Context,
	command CreateForFieldDeviceCommand,
) (CreateOutcome, error) {
	if h == nil || !h.transactionConfigured {
		return CreateOutcome{}, ErrCreateTransactionNotConfigured
	}
	if err := command.validate(); err != nil {
		return CreateOutcome{}, err
	}

	operationID := h.newID()
	actorID := actorFromContext(h.actor, ctx)
	committed, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(txCtx context.Context, workflow CreateWorkflow) (committedCreate, error) {
			return executeCreateForFieldDeviceTransaction(
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
		BacnetObject: committed.object,
		Mutation:     result,
	}
	fieldDeviceID := *committed.object.FieldDeviceID
	if h.projectLinks == nil || h.dispatcher == nil {
		return outcome, nil
	}

	dispatchCtx := context.WithoutCancel(ctx)
	links, err := h.projectLinks.GetByFieldDeviceIDs(
		dispatchCtx,
		[]uuid.UUID{fieldDeviceID},
	)
	if err != nil {
		outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
			"resolve created BACnet object collaboration projects: %w",
			err,
		))
		return outcome, nil
	}
	grouped := groupLinkedFieldDevices(links, []uuid.UUID{fieldDeviceID})
	projectIDs := sortedProjectIDs(grouped)
	outcome.Mutation.ProjectIDs = append([]uuid.UUID(nil), projectIDs...)
	for _, projectID := range projectIDs {
		collaborationCommand := appcollaboration.BacnetObjectCreated{
			Envelope: appcollaboration.Envelope{
				SchemaVersion: appcollaboration.SchemaVersionV1,
				EventID:       h.newID(),
				OperationID:   operationID,
				CorrelationID: operationID,
				ProjectID:     projectID,
				ActorID:       actorID,
				OccurredAt:    occurredAt,
			},
			BacnetObjectID: committed.object.ID,
			FieldDeviceID:  fieldDeviceID,
		}
		if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, collaborationCommand); dispatchErr != nil {
			outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
				"dispatch created BACnet object for project %s: %w",
				projectID,
				dispatchErr,
			))
		}
	}
	return outcome, nil
}

func (h *CreateHandler) ExecuteForObjectData(
	ctx context.Context,
	command CreateForObjectDataCommand,
) (CreateOutcome, error) {
	if h == nil || !h.transactionConfigured {
		return CreateOutcome{}, ErrCreateTransactionNotConfigured
	}
	if err := command.validate(); err != nil {
		return CreateOutcome{}, err
	}

	operationID := h.newID()
	actorID := actorFromContext(h.actor, ctx)
	committed, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(txCtx context.Context, workflow CreateWorkflow) (committedCreate, error) {
			return executeCreateForObjectDataTransaction(
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
		BacnetObject: committed.object,
		Mutation:     result,
	}
	if committed.projectID == nil {
		return outcome, nil
	}

	projectID := *committed.projectID
	outcome.Mutation.ProjectIDs = []uuid.UUID{projectID}
	if h.dispatcher == nil {
		return outcome, nil
	}
	dispatchCtx := context.WithoutCancel(ctx)
	collaborationCommand := appcollaboration.FacilityHierarchyRefreshRequired{
		Envelope: appcollaboration.Envelope{
			SchemaVersion: appcollaboration.SchemaVersionV1,
			EventID:       h.newID(),
			OperationID:   operationID,
			CorrelationID: operationID,
			ProjectID:     projectID,
			ActorID:       actorID,
			OccurredAt:    occurredAt,
		},
		Scope:       appcollaboration.FacilityScopeProject,
		FullRefresh: true,
	}
	if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, collaborationCommand); dispatchErr != nil {
		outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
			"dispatch ObjectData BACnet object create for project %s: %w",
			projectID,
			dispatchErr,
		))
	}
	return outcome, nil
}

func executeCreateForFieldDeviceTransaction(
	ctx context.Context,
	workflow CreateWorkflow,
	command CreateForFieldDeviceCommand,
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
	object := command.Input.toDomain()
	fieldDeviceID := command.FieldDeviceID
	if err := workflow.CreateWithParent(writeCtx, object, &fieldDeviceID, nil); err != nil {
		return committedCreate{}, err
	}
	if object.ID == uuid.Nil {
		return committedCreate{}, domain.ErrInvalidArgument
	}
	after, err := workflow.GetByID(writeCtx, object.ID)
	if err != nil {
		return committedCreate{}, err
	}
	if after == nil {
		return committedCreate{}, domain.ErrNotFound
	}
	if after.FieldDeviceID == nil || *after.FieldDeviceID != command.FieldDeviceID {
		return committedCreate{}, domain.ErrInvalidArgument
	}
	change, err := buildCreateChange(after)
	if err != nil {
		return committedCreate{}, err
	}
	return committedCreate{
		object:  cloneBacnetObject(after),
		change:  change,
		batched: batched,
	}, nil
}

func executeCreateForObjectDataTransaction(
	ctx context.Context,
	workflow CreateWorkflow,
	command CreateForObjectDataCommand,
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
	object := command.Input.toDomain()
	objectDataID := command.ObjectDataID
	if err := workflow.CreateWithParent(writeCtx, object, nil, &objectDataID); err != nil {
		return committedCreate{}, err
	}
	if object.ID == uuid.Nil {
		return committedCreate{}, domain.ErrInvalidArgument
	}
	after, err := workflow.GetByID(writeCtx, object.ID)
	if err != nil {
		return committedCreate{}, err
	}
	if after == nil {
		return committedCreate{}, domain.ErrNotFound
	}
	if after.FieldDeviceID != nil {
		return committedCreate{}, domain.ErrInvalidArgument
	}
	objectData, err := workflow.GetObjectDataByID(writeCtx, command.ObjectDataID)
	if err != nil {
		return committedCreate{}, err
	}
	if objectData == nil || objectData.ID != command.ObjectDataID {
		return committedCreate{}, domain.ErrNotFound
	}
	change, err := buildCreateChangeForParent(after, &objectDataID)
	if err != nil {
		return committedCreate{}, err
	}
	projectID := clonePointer(objectData.ProjectID)
	if projectID != nil && *projectID == uuid.Nil {
		projectID = nil
	}
	return committedCreate{
		object:    cloneBacnetObject(after),
		change:    change,
		projectID: projectID,
		batched:   batched,
	}, nil
}
