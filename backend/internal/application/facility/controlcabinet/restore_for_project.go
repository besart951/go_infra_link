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
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

var (
	ErrProjectRestoreNotConfigured = errors.New(
		"project ControlCabinet restore is not configured",
	)
	ErrProjectRestoreAccessDenied = errors.New(
		"project ControlCabinet restore access denied",
	)
)

// ProjectRestoreScope is the consumer-owned authorization Interface. Its wire
// Adapter must validate both actor access to the project and the cabinet's
// current or historical association with that project.
type ProjectRestoreScope interface {
	RequireControlCabinetRestoreScope(
		context.Context,
		uuid.UUID,
		uuid.UUID,
		uuid.UUID,
	) error
}

// ControlCabinetHistoryRestorer owns the existing restore transaction. A nil
// error means both restored rows and their restore history have committed.
type ControlCabinetHistoryRestorer interface {
	RestoreControlCabinet(
		context.Context,
		uuid.UUID,
		domainHistory.RestoreControlCabinetRequest,
	) (*domainHistory.RestoreResult, error)
}

type RestoreForProjectWorkflow interface {
	ControlCabinetHistoryRestorer
	ProjectLinkReader
}

type RestoreForProjectCommand struct {
	ProjectID        uuid.UUID
	ControlCabinetID uuid.UUID
	AsOf             *time.Time
	EventID          *uuid.UUID
}

func (c RestoreForProjectCommand) validate() error {
	if c.ProjectID == uuid.Nil || c.ControlCabinetID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	return nil
}

type RestoreForProjectDependencies struct {
	TransactionRunner   apptransaction.Runner
	TransactionWorkflow apptransaction.Factory[RestoreForProjectWorkflow]
	Scope               ProjectRestoreScope
	Restorer            ControlCabinetHistoryRestorer
	ProjectLinks        ProjectLinkReader
	Dispatcher          appcollaboration.CommandDispatcher
	Actor               ActorProvider
	NewID               IDGenerator
	Now                 Clock
	ReportError         ErrorReporter
}

type RestoreForProjectHandler struct {
	operation             apptransaction.Operation[RestoreForProjectWorkflow, RestoreForProjectWorkflow]
	transactionConfigured bool
	scope                 ProjectRestoreScope
	restorer              ControlCabinetHistoryRestorer
	projectLinks          ProjectLinkReader
	dispatcher            appcollaboration.CommandDispatcher
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
	reportError           ErrorReporter
}

type RestoreForProjectOutcome struct {
	Restore        *domainHistory.RestoreResult
	Mutation       mutation.Result
	DispatchErrors []error
}

func NewRestoreForProjectHandler(
	deps RestoreForProjectDependencies,
) *RestoreForProjectHandler {
	newID := deps.NewID
	if newID == nil {
		newID = uuid.New
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	boundary := apptransaction.NewBoundary[RestoreForProjectWorkflow](
		deps.TransactionRunner,
		deps.TransactionWorkflow,
	)
	var noNonTransactionalWorkflow RestoreForProjectWorkflow
	operation := apptransaction.Bind(
		boundary,
		noNonTransactionalWorkflow,
		func(workflow RestoreForProjectWorkflow) RestoreForProjectWorkflow { return workflow },
	)
	return &RestoreForProjectHandler{
		operation:             operation,
		transactionConfigured: deps.TransactionRunner != nil && deps.TransactionWorkflow != nil,
		scope:                 deps.Scope,
		restorer:              deps.Restorer,
		projectLinks:          deps.ProjectLinks,
		dispatcher:            deps.Dispatcher,
		actor:                 deps.Actor,
		newID:                 newID,
		now:                   now,
		reportError:           deps.ReportError,
	}
}

// RestoreForProject preserves the existing HTTP response while routing the
// mutation through project isolation and the after-commit collaboration gate.
func (h *RestoreForProjectHandler) RestoreForProject(
	ctx context.Context,
	command RestoreForProjectCommand,
) (*domainHistory.RestoreResult, error) {
	outcome, err := h.Execute(ctx, command)
	if err != nil {
		return nil, err
	}
	for _, dispatchErr := range outcome.DispatchErrors {
		if h.reportError != nil {
			h.reportError(dispatchErr)
		}
	}
	return outcome.Restore, nil
}

func (h *RestoreForProjectHandler) Execute(
	ctx context.Context,
	command RestoreForProjectCommand,
) (RestoreForProjectOutcome, error) {
	if h == nil || h.scope == nil || (!h.transactionConfigured && h.restorer == nil) {
		return RestoreForProjectOutcome{}, ErrProjectRestoreNotConfigured
	}
	if err := command.validate(); err != nil {
		return RestoreForProjectOutcome{}, err
	}

	actorID := actorFromContext(h.actor, ctx)
	if actorID == nil {
		return RestoreForProjectOutcome{}, ErrProjectRestoreAccessDenied
	}
	if err := h.scope.RequireControlCabinetRestoreScope(
		ctx,
		*actorID,
		command.ProjectID,
		command.ControlCabinetID,
	); err != nil {
		return RestoreForProjectOutcome{}, err
	}

	occurredAt := h.now().UTC()
	var (
		restore             *domainHistory.RestoreResult
		operationID         uuid.UUID
		projectIDs          []uuid.UUID
		fallbackScopeErrors []error
	)
	if h.transactionConfigured {
		committed, err := apptransaction.RunResult(
			ctx,
			h.operation,
			func(
				txCtx context.Context,
				workflow RestoreForProjectWorkflow,
			) (committedProjectRestore, error) {
				return executeTransactionalProjectRestore(
					txCtx,
					workflow,
					command,
					actorID,
					occurredAt,
					h.newID,
				)
			},
		)
		if err != nil {
			return RestoreForProjectOutcome{}, err
		}
		restore = committed.restore
		operationID = committed.operationID
		projectIDs = committed.projectIDs
	} else {
		projectID := command.ProjectID
		var err error
		restore, err = h.restorer.RestoreControlCabinet(
			ctx,
			command.ControlCabinetID,
			domainHistory.RestoreControlCabinetRequest{
				AsOf: command.AsOf, EventID: command.EventID, ProjectID: &projectID,
			},
		)
		if err != nil {
			return RestoreForProjectOutcome{}, err
		}
		if restore == nil {
			return RestoreForProjectOutcome{}, domain.ErrInvalidArgument
		}
		operationID = restore.BatchID
		if operationID == uuid.Nil {
			operationID = h.newID()
		}
		projectIDs = []uuid.UUID{command.ProjectID}
		if h.projectLinks != nil {
			links, scopeErr := h.projectLinks.GetByControlCabinetIDs(
				ctx,
				[]uuid.UUID{command.ControlCabinetID},
			)
			if scopeErr != nil {
				fallbackScopeErrors = append(fallbackScopeErrors, fmt.Errorf(
					"resolve restored ControlCabinet collaboration projects: %w",
					scopeErr,
				))
			} else {
				projectIDs = restoreProjectIDs(
					command.ProjectID,
					linkedProjectIDs(links, command.ControlCabinetID),
				)
			}
		}
	}
	outcome := RestoreForProjectOutcome{
		Restore:        restore,
		DispatchErrors: fallbackScopeErrors,
	}

	result := mutation.Result{
		OperationID: operationID,
		ActorID:     actorID,
		OccurredAt:  occurredAt,
		ProjectIDs:  append([]uuid.UUID(nil), projectIDs...),
	}
	if restore.BatchID != uuid.Nil {
		batchID := restore.BatchID
		result.BatchID = &batchID
	}
	if restore.RestoredCount+restore.DeletedCount > 0 {
		result.Changes = []mutation.EntityChange{{
			EntityType: mutation.EntityTypeControlCabinet,
			EntityID:   command.ControlCabinetID,
			Action:     domainHistory.ActionRestore,
		}}
	}
	outcome.Mutation = result

	if len(result.Changes) == 0 || h.dispatcher == nil {
		return outcome, nil
	}
	dispatchCtx := context.WithoutCancel(ctx)
	for _, affectedProjectID := range projectIDs {
		refresh := appcollaboration.FacilityHierarchyRefreshRequired{
			Envelope: appcollaboration.Envelope{
				SchemaVersion: appcollaboration.SchemaVersionV1,
				EventID:       h.newID(),
				OperationID:   operationID,
				CorrelationID: operationID,
				ProjectID:     affectedProjectID,
				ActorID:       actorID,
				OccurredAt:    occurredAt,
			},
			Scope:       appcollaboration.FacilityScopeProject,
			FullRefresh: true,
		}
		if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, refresh); dispatchErr != nil {
			outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
				"dispatch restored ControlCabinet for project %s: %w",
				affectedProjectID,
				dispatchErr,
			))
		}
	}
	return outcome, nil
}

type committedProjectRestore struct {
	restore     *domainHistory.RestoreResult
	operationID uuid.UUID
	projectIDs  []uuid.UUID
}

func executeTransactionalProjectRestore(
	ctx context.Context,
	workflow RestoreForProjectWorkflow,
	command RestoreForProjectCommand,
	actorID *uuid.UUID,
	occurredAt time.Time,
	newID IDGenerator,
) (committedProjectRestore, error) {
	if workflow == nil {
		return committedProjectRestore{}, ErrProjectRestoreNotConfigured
	}
	projectID := command.ProjectID
	restore, err := workflow.RestoreControlCabinet(
		ctx,
		command.ControlCabinetID,
		domainHistory.RestoreControlCabinetRequest{
			AsOf: command.AsOf, EventID: command.EventID, ProjectID: &projectID,
		},
	)
	if err != nil {
		return committedProjectRestore{}, err
	}
	if restore == nil {
		return committedProjectRestore{}, domain.ErrInvalidArgument
	}
	operationID := restore.BatchID
	if operationID == uuid.Nil {
		operationID = newID()
	}
	links, err := workflow.GetByControlCabinetIDs(
		ctx,
		[]uuid.UUID{command.ControlCabinetID},
	)
	if err != nil {
		return committedProjectRestore{}, fmt.Errorf(
			"resolve restored ControlCabinet collaboration projects for outbox: %w",
			err,
		)
	}
	projectIDs := restoreProjectIDs(
		command.ProjectID,
		linkedProjectIDs(links, command.ControlCabinetID),
	)
	if restore.RestoredCount+restore.DeletedCount > 0 {
		for _, affectedProjectID := range projectIDs {
			event := appcollaboration.FacilityHierarchyRefreshRequired{
				Envelope: appcollaboration.Envelope{
					SchemaVersion: appcollaboration.SchemaVersionV2,
					EventID:       newID(),
					OperationID:   operationID,
					CorrelationID: operationID,
					ProjectID:     affectedProjectID,
					ActorID:       actorID,
					OccurredAt:    occurredAt,
				},
				Scope:       appcollaboration.FacilityScopeProject,
				FullRefresh: true,
			}
			if _, err := appcollaboration.EnqueueCommand(ctx, event); err != nil {
				return committedProjectRestore{}, fmt.Errorf(
					"enqueue restored ControlCabinet for project %s: %w",
					affectedProjectID,
					err,
				)
			}
		}
	}
	return committedProjectRestore{
		restore: restore, operationID: operationID, projectIDs: projectIDs,
	}, nil
}

func restoreProjectIDs(requested uuid.UUID, linked []uuid.UUID) []uuid.UUID {
	set := make(map[uuid.UUID]struct{}, len(linked)+1)
	if requested != uuid.Nil {
		set[requested] = struct{}{}
	}
	for _, projectID := range linked {
		if projectID != uuid.Nil {
			set[projectID] = struct{}{}
		}
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
