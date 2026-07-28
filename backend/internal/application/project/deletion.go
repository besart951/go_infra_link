package project

import (
	"context"
	"errors"
	"fmt"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

const projectOwnedDataDeleteBatchSize = 100

var (
	ErrDeletionTransactionNotConfigured = errors.New("project deletion transaction is not configured")
	ErrDeletionForbidden                = errors.New("project deletion is forbidden")
	ErrHierarchyLinksRemain             = errors.New("project hierarchy links remain")
)

// Snapshot contains only the project state needed by the deletion policy. The
// transaction-scoped infrastructure adapter locks the backing row before
// returning it so a concurrent project-link insert cannot race the eligibility
// check once the database foreign keys are installed.
type Snapshot struct {
	ID     uuid.UUID
	Status domainProject.ProjectStatus
}

// DeletionWorkflow is a consumer-owned port implemented by one transaction-
// scoped PostgreSQL adapter. It deliberately exposes no facility-entity delete
// capability: project deletion can remove project-owned ObjectData and project
// associations, but never global hierarchy entities.
type DeletionWorkflow interface {
	GetProjectForDeletion(context.Context, uuid.UUID) (*Snapshot, error)
	GetActiveUserRole(context.Context, uuid.UUID) (domainUser.Role, error)
	HasHierarchyLinks(context.Context, uuid.UUID) (bool, error)
	ListProjectObjectDataIDs(
		context.Context,
		uuid.UUID,
		uuid.UUID,
		int,
	) ([]uuid.UUID, error)
	DeleteObjectData(context.Context, []uuid.UUID) error
	DeleteProjectMemberships(context.Context, uuid.UUID) error
	DeleteProject(context.Context, uuid.UUID) error
}

type HistoryBatchContext func(context.Context, uuid.UUID) context.Context
type ActorProvider func(context.Context) *uuid.UUID
type IDGenerator func() uuid.UUID
type Clock func() time.Time

type DeleteCommand struct {
	ProjectID uuid.UUID
}

type DeleteOutcome struct {
	Project           Snapshot
	DeletedObjectData int
	OperationID       uuid.UUID
	EventID           uuid.UUID
	OccurredAt        time.Time
}

type DeleteDependencies struct {
	TransactionRunner   apptransaction.Runner
	TransactionWorkflow apptransaction.Factory[DeletionWorkflow]
	HistoryBatch        HistoryBatchContext
	Actor               ActorProvider
	NewID               IDGenerator
	Now                 Clock
}

type DeleteHandler struct {
	operation             apptransaction.Operation[DeletionWorkflow, DeletionWorkflow]
	transactionConfigured bool
	historyBatch          HistoryBatchContext
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
}

func NewDeleteHandler(deps DeleteDependencies) *DeleteHandler {
	newID := deps.NewID
	if newID == nil {
		newID = uuid.New
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	boundary := apptransaction.NewBoundary[DeletionWorkflow](
		deps.TransactionRunner,
		deps.TransactionWorkflow,
	)
	var noNonTransactionalWorkflow DeletionWorkflow
	return &DeleteHandler{
		operation: apptransaction.Bind(
			boundary,
			noNonTransactionalWorkflow,
			func(workflow DeletionWorkflow) DeletionWorkflow { return workflow },
		),
		transactionConfigured: deps.TransactionRunner != nil &&
			deps.TransactionWorkflow != nil,
		historyBatch: deps.HistoryBatch,
		actor:        deps.Actor,
		newID:        newID,
		now:          now,
	}
}

func (h *DeleteHandler) Delete(ctx context.Context, command DeleteCommand) error {
	_, err := h.Execute(ctx, command)
	return err
}

func (h *DeleteHandler) Execute(
	ctx context.Context,
	command DeleteCommand,
) (DeleteOutcome, error) {
	if h == nil || !h.transactionConfigured {
		return DeleteOutcome{}, ErrDeletionTransactionNotConfigured
	}
	if command.ProjectID == uuid.Nil {
		return DeleteOutcome{}, domain.ErrInvalidArgument
	}

	actorID := actorFromContext(h.actor, ctx)
	if actorID == nil {
		return DeleteOutcome{}, ErrDeletionForbidden
	}

	operationID := h.newID()
	eventID := h.newID()
	occurredAt := h.now().UTC()

	type transactionResult struct {
		project           Snapshot
		deletedObjectData int
	}
	result, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(txCtx context.Context, workflow DeletionWorkflow) (transactionResult, error) {
			if h.historyBatch != nil {
				txCtx = h.historyBatch(txCtx, operationID)
			}

			project, err := workflow.GetProjectForDeletion(txCtx, command.ProjectID)
			if err != nil {
				return transactionResult{}, err
			}
			if project == nil || project.ID != command.ProjectID {
				return transactionResult{}, domain.ErrNotFound
			}

			role, err := workflow.GetActiveUserRole(txCtx, *actorID)
			if err != nil {
				if errors.Is(err, domain.ErrNotFound) {
					return transactionResult{}, ErrDeletionForbidden
				}
				return transactionResult{}, err
			}
			if !canDeleteProject(role) {
				return transactionResult{}, ErrDeletionForbidden
			}

			hasLinks, err := workflow.HasHierarchyLinks(txCtx, command.ProjectID)
			if err != nil {
				return transactionResult{}, err
			}
			if hasLinks {
				return transactionResult{}, ErrHierarchyLinksRemain
			}

			deletedObjectData, err := deleteProjectObjectData(
				txCtx,
				workflow,
				command.ProjectID,
			)
			if err != nil {
				return transactionResult{}, err
			}
			if err := workflow.DeleteProjectMemberships(txCtx, command.ProjectID); err != nil {
				return transactionResult{}, err
			}
			if err := workflow.DeleteProject(txCtx, command.ProjectID); err != nil {
				return transactionResult{}, err
			}

			durableCommand := appcollaboration.FacilityHierarchyRefreshRequired{
				Envelope: appcollaboration.Envelope{
					SchemaVersion: appcollaboration.SchemaVersionV2,
					EventID:       eventID,
					OperationID:   operationID,
					CorrelationID: operationID,
					ProjectID:     command.ProjectID,
					ActorID:       actorID,
					OccurredAt:    occurredAt,
				},
				Scope:       appcollaboration.FacilityScopeProject,
				EntityIDs:   []uuid.UUID{command.ProjectID},
				FullRefresh: true,
			}
			if _, err := appcollaboration.EnqueueCommand(txCtx, durableCommand); err != nil {
				return transactionResult{}, fmt.Errorf("enqueue project deletion: %w", err)
			}

			return transactionResult{
				project:           *project,
				deletedObjectData: deletedObjectData,
			}, nil
		},
	)
	if err != nil {
		return DeleteOutcome{}, err
	}

	return DeleteOutcome{
		Project:           result.project,
		DeletedObjectData: result.deletedObjectData,
		OperationID:       operationID,
		EventID:           eventID,
		OccurredAt:        occurredAt,
	}, nil
}

func canDeleteProject(role domainUser.Role) bool {
	return role == domainUser.RoleSuperAdmin || role == domainUser.RoleAdminFZAG
}

func deleteProjectObjectData(
	ctx context.Context,
	workflow DeletionWorkflow,
	projectID uuid.UUID,
) (int, error) {
	deleted := 0
	var after uuid.UUID
	for {
		ids, err := workflow.ListProjectObjectDataIDs(
			ctx,
			projectID,
			after,
			projectOwnedDataDeleteBatchSize,
		)
		if err != nil {
			return 0, err
		}
		if len(ids) == 0 {
			return deleted, nil
		}
		if len(ids) > projectOwnedDataDeleteBatchSize {
			return 0, fmt.Errorf(
				"project object data batch exceeds %d items",
				projectOwnedDataDeleteBatchSize,
			)
		}
		if err := workflow.DeleteObjectData(ctx, ids); err != nil {
			return 0, err
		}
		deleted += len(ids)
		after = ids[len(ids)-1]
	}
}

func actorFromContext(provider ActorProvider, ctx context.Context) *uuid.UUID {
	if provider == nil {
		return nil
	}
	actorID := provider(ctx)
	if actorID == nil || *actorID == uuid.Nil {
		return nil
	}
	copyID := *actorID
	return &copyID
}
