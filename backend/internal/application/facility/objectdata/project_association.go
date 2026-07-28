package objectdata

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	"github.com/besart951/go_infra_link/backend/internal/application/facility/mutation"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

var ErrProjectAssociationTransactionNotConfigured = errors.New(
	"project ObjectData association transaction is not configured",
)

// ProjectAssociationWorkflow is the consumer-owned transaction Interface for
// project ObjectData activation. Its infrastructure Adapter combines the
// existing project and ObjectData repositories without exposing GORM here.
type ProjectAssociationWorkflow interface {
	RequireProject(context.Context, uuid.UUID) error
	GetObjectData(context.Context, uuid.UUID) (*domainFacility.ObjectData, error)
	UpdateObjectData(context.Context, *domainFacility.ObjectData) error
}

type HistoryBatchContext func(context.Context, uuid.UUID) context.Context
type ActorProvider func(context.Context) *uuid.UUID
type IDGenerator func() uuid.UUID
type Clock func() time.Time
type ErrorReporter func(error)

type AttachToProjectCommand struct {
	ProjectID    uuid.UUID
	ObjectDataID uuid.UUID
}

func (c AttachToProjectCommand) validate() error {
	if c.ProjectID == uuid.Nil || c.ObjectDataID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	return nil
}

type DeactivateForProjectCommand struct {
	ProjectID    uuid.UUID
	ObjectDataID uuid.UUID
}

func (c DeactivateForProjectCommand) validate() error {
	if c.ProjectID == uuid.Nil || c.ObjectDataID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	return nil
}

type ProjectAssociationDependencies struct {
	TransactionRunner   apptransaction.Runner
	TransactionWorkflow apptransaction.Factory[ProjectAssociationWorkflow]
	HistoryBatch        HistoryBatchContext
	Dispatcher          appcollaboration.CommandDispatcher
	Actor               ActorProvider
	NewID               IDGenerator
	Now                 Clock
	ReportError         ErrorReporter
}

type ProjectAssociationHandler struct {
	operation             apptransaction.Operation[ProjectAssociationWorkflow, ProjectAssociationWorkflow]
	transactionConfigured bool
	historyBatch          HistoryBatchContext
	dispatcher            appcollaboration.CommandDispatcher
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
	reportError           ErrorReporter
}

type ProjectAssociationOutcome struct {
	ObjectData     *domainFacility.ObjectData
	Mutation       mutation.Result
	DispatchErrors []error
}

type committedProjectAssociation struct {
	objectData *domainFacility.ObjectData
	change     mutation.EntityChange
	batched    bool
}

type objectDataSnapshot struct {
	ID          uuid.UUID  `json:"id"`
	Description string     `json:"description"`
	Version     string     `json:"version"`
	IsActive    bool       `json:"is_active"`
	ProjectID   *uuid.UUID `json:"project_id"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

func NewProjectAssociationHandler(
	deps ProjectAssociationDependencies,
) *ProjectAssociationHandler {
	newID := deps.NewID
	if newID == nil {
		newID = uuid.New
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	boundary := apptransaction.NewBoundary[ProjectAssociationWorkflow](
		deps.TransactionRunner,
		deps.TransactionWorkflow,
	)
	var noNonTransactionalWorkflow ProjectAssociationWorkflow
	operation := apptransaction.Bind(
		boundary,
		noNonTransactionalWorkflow,
		func(workflow ProjectAssociationWorkflow) ProjectAssociationWorkflow {
			return workflow
		},
	)
	return &ProjectAssociationHandler{
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

// AttachToProject activates a global or already-owned ObjectData template for
// one server-authorized project. Realtime delivery is best effort after commit.
func (h *ProjectAssociationHandler) AttachToProject(
	ctx context.Context,
	command AttachToProjectCommand,
) (*domainFacility.ObjectData, error) {
	outcome, err := h.ExecuteAttach(ctx, command)
	if err != nil {
		return nil, err
	}
	h.reportDispatchErrors(outcome.DispatchErrors)
	return outcome.ObjectData, nil
}

func (h *ProjectAssociationHandler) ExecuteAttach(
	ctx context.Context,
	command AttachToProjectCommand,
) (ProjectAssociationOutcome, error) {
	if h == nil || !h.transactionConfigured {
		return ProjectAssociationOutcome{}, ErrProjectAssociationTransactionNotConfigured
	}
	if err := command.validate(); err != nil {
		return ProjectAssociationOutcome{}, err
	}
	return h.execute(
		ctx,
		command.ProjectID,
		command.ObjectDataID,
		func(objectData *domainFacility.ObjectData) error {
			return objectData.ActivateForProject(command.ProjectID)
		},
	)
}

// DeactivateForProject preserves the existing DELETE route behavior: the
// template remains owned by ProjectID and is marked inactive.
func (h *ProjectAssociationHandler) DeactivateForProject(
	ctx context.Context,
	command DeactivateForProjectCommand,
) (*domainFacility.ObjectData, error) {
	outcome, err := h.ExecuteDeactivate(ctx, command)
	if err != nil {
		return nil, err
	}
	h.reportDispatchErrors(outcome.DispatchErrors)
	return outcome.ObjectData, nil
}

func (h *ProjectAssociationHandler) ExecuteDeactivate(
	ctx context.Context,
	command DeactivateForProjectCommand,
) (ProjectAssociationOutcome, error) {
	if h == nil || !h.transactionConfigured {
		return ProjectAssociationOutcome{}, ErrProjectAssociationTransactionNotConfigured
	}
	if err := command.validate(); err != nil {
		return ProjectAssociationOutcome{}, err
	}
	return h.execute(
		ctx,
		command.ProjectID,
		command.ObjectDataID,
		func(objectData *domainFacility.ObjectData) error {
			return objectData.DeactivateForProject(command.ProjectID)
		},
	)
}

func (h *ProjectAssociationHandler) execute(
	ctx context.Context,
	projectID uuid.UUID,
	objectDataID uuid.UUID,
	apply func(*domainFacility.ObjectData) error,
) (ProjectAssociationOutcome, error) {
	operationID := h.newID()
	durableEventID := h.newID()
	compatibilityEventID := h.newID()
	actorID := actorFromContext(h.actor, ctx)
	occurredAt := h.now().UTC()
	var compatibilityCommand appcollaboration.Command
	committed, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(
			txCtx context.Context,
			workflow ProjectAssociationWorkflow,
		) (committedProjectAssociation, error) {
			result, err := executeProjectAssociationTransaction(
				txCtx,
				workflow,
				projectID,
				objectDataID,
				operationID,
				h.historyBatch,
				apply,
			)
			if err != nil {
				return committedProjectAssociation{}, err
			}
			durableCommand := appcollaboration.FacilityHierarchyRefreshRequired{
				Envelope: appcollaboration.Envelope{
					SchemaVersion: appcollaboration.SchemaVersionV2,
					EventID:       durableEventID, OperationID: operationID, CorrelationID: operationID,
					ProjectID: projectID, ActorID: actorID, OccurredAt: occurredAt,
				},
				Scope:     appcollaboration.FacilityScopeObjectData,
				EntityIDs: []uuid.UUID{objectDataID},
			}
			if _, err := appcollaboration.EnqueueCommand(txCtx, durableCommand); err != nil {
				return committedProjectAssociation{}, fmt.Errorf("enqueue project ObjectData refresh: %w", err)
			}
			compatibilityCommand = appcollaboration.FacilityHierarchyRefreshRequired{
				Envelope: appcollaboration.Envelope{
					SchemaVersion: appcollaboration.SchemaVersionV1,
					EventID:       compatibilityEventID,
					OperationID:   operationID,
					CorrelationID: operationID,
					ProjectID:     projectID,
					ActorID:       actorID,
					OccurredAt:    occurredAt,
				},
				Scope:       appcollaboration.FacilityScopeProject,
				FullRefresh: true,
			}
			return result, nil
		},
	)
	if err != nil {
		return ProjectAssociationOutcome{}, err
	}

	result := mutation.Result{
		OperationID: operationID,
		ActorID:     actorID,
		OccurredAt:  occurredAt,
		ProjectIDs:  []uuid.UUID{projectID},
		Changes:     []mutation.EntityChange{committed.change},
	}
	if committed.batched {
		batchID := operationID
		result.BatchID = &batchID
	}
	outcome := ProjectAssociationOutcome{
		ObjectData: committed.objectData,
		Mutation:   result,
	}
	if h.dispatcher == nil {
		return outcome, nil
	}

	dispatchCtx := context.WithoutCancel(ctx)
	if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, compatibilityCommand); dispatchErr != nil {
		outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
			"dispatch project ObjectData refresh for project %s: %w",
			projectID,
			dispatchErr,
		))
	}
	return outcome, nil
}

func executeProjectAssociationTransaction(
	ctx context.Context,
	workflow ProjectAssociationWorkflow,
	projectID uuid.UUID,
	objectDataID uuid.UUID,
	operationID uuid.UUID,
	historyBatch HistoryBatchContext,
	apply func(*domainFacility.ObjectData) error,
) (committedProjectAssociation, error) {
	if workflow == nil {
		return committedProjectAssociation{}, ErrProjectAssociationTransactionNotConfigured
	}
	if err := workflow.RequireProject(ctx, projectID); err != nil {
		return committedProjectAssociation{}, err
	}
	objectData, err := workflow.GetObjectData(ctx, objectDataID)
	if err != nil {
		return committedProjectAssociation{}, err
	}
	if objectData == nil || objectData.ID != objectDataID {
		return committedProjectAssociation{}, domain.ErrNotFound
	}
	before := cloneObjectData(objectData)
	if err := apply(objectData); err != nil {
		return committedProjectAssociation{}, err
	}

	writeCtx := ctx
	batched := historyBatch != nil
	if batched {
		writeCtx = historyBatch(ctx, operationID)
	}
	if err := workflow.UpdateObjectData(writeCtx, objectData); err != nil {
		return committedProjectAssociation{}, err
	}
	after, err := workflow.GetObjectData(writeCtx, objectDataID)
	if err != nil {
		return committedProjectAssociation{}, err
	}
	if after == nil || after.ID != objectDataID {
		return committedProjectAssociation{}, errors.New(
			"invalid ObjectData project association result",
		)
	}
	change, err := buildProjectAssociationChange(before, after)
	if err != nil {
		return committedProjectAssociation{}, err
	}
	return committedProjectAssociation{
		objectData: cloneObjectData(after),
		change:     change,
		batched:    batched,
	}, nil
}

func buildProjectAssociationChange(
	before *domainFacility.ObjectData,
	after *domainFacility.ObjectData,
) (mutation.EntityChange, error) {
	beforeJSON, err := json.Marshal(toObjectDataSnapshot(before))
	if err != nil {
		return mutation.EntityChange{}, fmt.Errorf("marshal ObjectData before snapshot: %w", err)
	}
	afterJSON, err := json.Marshal(toObjectDataSnapshot(after))
	if err != nil {
		return mutation.EntityChange{}, fmt.Errorf("marshal ObjectData after snapshot: %w", err)
	}
	changedFields := make([]mutation.FieldName, 0, 2)
	if !equalUUIDPointers(before.ProjectID, after.ProjectID) {
		changedFields = append(changedFields, mutation.FieldNameProject)
	}
	if before.IsActive != after.IsActive {
		changedFields = append(changedFields, mutation.FieldNameIsActive)
	}
	return mutation.EntityChange{
		EntityType:    mutation.EntityTypeObjectData,
		EntityID:      after.ID,
		ParentID:      cloneUUIDPointer(after.ProjectID),
		Action:        domainHistory.ActionUpdate,
		Before:        beforeJSON,
		After:         afterJSON,
		ChangedFields: changedFields,
	}, nil
}

func toObjectDataSnapshot(objectData *domainFacility.ObjectData) objectDataSnapshot {
	if objectData == nil {
		return objectDataSnapshot{}
	}
	return objectDataSnapshot{
		ID:          objectData.ID,
		Description: objectData.Description,
		Version:     objectData.Version,
		IsActive:    objectData.IsActive,
		ProjectID:   cloneUUIDPointer(objectData.ProjectID),
		CreatedAt:   objectData.CreatedAt,
		UpdatedAt:   objectData.UpdatedAt,
	}
}

func cloneObjectData(objectData *domainFacility.ObjectData) *domainFacility.ObjectData {
	if objectData == nil {
		return nil
	}
	clone := *objectData
	clone.ProjectID = cloneUUIDPointer(objectData.ProjectID)
	clone.Project = objectData.Project
	clone.BacnetObjects = append([]*domainFacility.BacnetObject(nil), objectData.BacnetObjects...)
	clone.Apparats = append([]*domainFacility.Apparat(nil), objectData.Apparats...)
	return &clone
}

func cloneUUIDPointer(value *uuid.UUID) *uuid.UUID {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}

func equalUUIDPointers(left, right *uuid.UUID) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}

func actorFromContext(provider ActorProvider, ctx context.Context) *uuid.UUID {
	if provider == nil {
		return nil
	}
	return cloneUUIDPointer(provider(ctx))
}

func (h *ProjectAssociationHandler) reportDispatchErrors(errs []error) {
	if h == nil || h.reportError == nil {
		return
	}
	for _, err := range errs {
		h.reportError(err)
	}
}
