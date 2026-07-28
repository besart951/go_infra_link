package spscontroller

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
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

var ErrReassignProjectLinkTransactionNotConfigured = errors.New(
	"project SPS controller reassignment transaction is not configured",
)

// ReassignProjectLinkWorkflow is defined by the application consumer. Its
// transaction-scoped Adapter combines authoritative root-link reads with the
// compatibility workflow that updates the root and materializes links for the
// new SPSController descendants.
type ReassignProjectLinkWorkflow interface {
	GetByIds(context.Context, []uuid.UUID) ([]*domainProject.ProjectSPSController, error)
	UpdateSPSController(
		context.Context,
		uuid.UUID,
		uuid.UUID,
		uuid.UUID,
	) (*domainProject.ProjectSPSController, error)
}

type ReassignProjectLinkCommand struct {
	ProjectID       uuid.UUID
	LinkID          uuid.UUID
	ExpectedVersion uint64
	SPSControllerID uuid.UUID
}

type ReassignProjectLinkDependencies struct {
	TransactionRunner   apptransaction.Runner
	TransactionWorkflow apptransaction.Factory[ReassignProjectLinkWorkflow]
	HistoryBatch        HistoryBatchContext
	Dispatcher          appcollaboration.CommandDispatcher
	Actor               ActorProvider
	NewID               IDGenerator
	Now                 Clock
	ReportError         ErrorReporter
}

type ReassignProjectLinkHandler struct {
	operation             apptransaction.Operation[ReassignProjectLinkWorkflow, ReassignProjectLinkWorkflow]
	transactionConfigured bool
	historyBatch          HistoryBatchContext
	dispatcher            appcollaboration.CommandDispatcher
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
	reportError           ErrorReporter
}

type ReassignProjectLinkOutcome struct {
	Link           *domainProject.ProjectSPSController
	Mutation       mutation.Result
	DispatchErrors []error
}

type committedProjectLinkReassignment struct {
	link                    *domainProject.ProjectSPSController
	previousSPSControllerID uuid.UUID
	change                  mutation.EntityChange
	batched                 bool
}

func NewReassignProjectLinkHandler(
	deps ReassignProjectLinkDependencies,
) *ReassignProjectLinkHandler {
	newID := deps.NewID
	if newID == nil {
		newID = uuid.New
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	boundary := apptransaction.NewBoundary[ReassignProjectLinkWorkflow](
		deps.TransactionRunner,
		deps.TransactionWorkflow,
	)
	var noNonTransactionalWorkflow ReassignProjectLinkWorkflow
	operation := apptransaction.Bind(
		boundary,
		noNonTransactionalWorkflow,
		func(workflow ReassignProjectLinkWorkflow) ReassignProjectLinkWorkflow { return workflow },
	)
	return &ReassignProjectLinkHandler{
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

func (h *ReassignProjectLinkHandler) ReassignProjectLink(
	ctx context.Context,
	command ReassignProjectLinkCommand,
) (*domainProject.ProjectSPSController, error) {
	outcome, err := h.Execute(ctx, command)
	if err != nil {
		return nil, err
	}
	for _, dispatchErr := range outcome.DispatchErrors {
		if h.reportError != nil {
			h.reportError(dispatchErr)
		}
	}
	return outcome.Link, nil
}

func (h *ReassignProjectLinkHandler) Execute(
	ctx context.Context,
	command ReassignProjectLinkCommand,
) (ReassignProjectLinkOutcome, error) {
	if h == nil || !h.transactionConfigured {
		return ReassignProjectLinkOutcome{}, ErrReassignProjectLinkTransactionNotConfigured
	}

	operationID := h.newID()
	eventID := h.newID()
	actorID := actorFromContext(h.actor, ctx)
	occurredAt := h.now().UTC()
	var collaborationCommand appcollaboration.Command
	committed, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(
			txCtx context.Context,
			workflow ReassignProjectLinkWorkflow,
		) (committedProjectLinkReassignment, error) {
			result, err := executeReassignProjectLinkTransaction(
				txCtx,
				workflow,
				command,
				operationID,
				h.historyBatch,
			)
			if err != nil {
				return committedProjectLinkReassignment{}, err
			}
			collaborationCommand = appcollaboration.FacilityHierarchyRefreshRequired{
				Envelope: appcollaboration.Envelope{
					SchemaVersion: appcollaboration.SchemaVersionV2,
					EventID:       eventID, OperationID: operationID, CorrelationID: operationID,
					ProjectID: command.ProjectID, ActorID: actorID, OccurredAt: occurredAt,
				},
				Scope: appcollaboration.FacilityScopeSPSController,
				EntityIDs: reassignedSPSControllerIDs(
					result.previousSPSControllerID,
					result.link.SPSControllerID,
				),
			}
			if _, err := appcollaboration.EnqueueCommand(txCtx, collaborationCommand); err != nil {
				return committedProjectLinkReassignment{}, fmt.Errorf("enqueue ProjectSPSController reassignment: %w", err)
			}
			return result, nil
		},
	)
	if err != nil {
		return ReassignProjectLinkOutcome{}, err
	}

	result := mutation.Result{
		OperationID: operationID,
		ActorID:     actorID,
		OccurredAt:  occurredAt,
		ProjectIDs:  []uuid.UUID{command.ProjectID},
		Changes:     []mutation.EntityChange{committed.change},
	}
	if committed.batched {
		batchID := operationID
		result.BatchID = &batchID
	}
	outcome := ReassignProjectLinkOutcome{
		Link:     committed.link,
		Mutation: result,
	}
	if h.dispatcher == nil {
		return outcome, nil
	}

	dispatchCtx := context.WithoutCancel(ctx)
	if dispatchErr := h.dispatcher.Dispatch(dispatchCtx, collaborationCommand); dispatchErr != nil {
		outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
			"dispatch ProjectSPSController reassignment for project %s: %w",
			command.ProjectID,
			dispatchErr,
		))
	}
	return outcome, nil
}

func executeReassignProjectLinkTransaction(
	ctx context.Context,
	workflow ReassignProjectLinkWorkflow,
	command ReassignProjectLinkCommand,
	operationID uuid.UUID,
	historyBatch HistoryBatchContext,
) (committedProjectLinkReassignment, error) {
	if workflow == nil {
		return committedProjectLinkReassignment{}, ErrReassignProjectLinkTransactionNotConfigured
	}

	link, err := domain.GetByID(ctx, workflow, command.LinkID)
	if err != nil {
		return committedProjectLinkReassignment{}, err
	}
	if link.ProjectID != command.ProjectID {
		return committedProjectLinkReassignment{}, domain.ErrNotFound
	}
	if command.ExpectedVersion != 0 && link.Revision != command.ExpectedVersion {
		return committedProjectLinkReassignment{}, &domain.RevisionConflict{
			EntityID: link.ID,
			Expected: command.ExpectedVersion,
			Current:  link.Revision,
		}
	}
	before := cloneProjectSPSControllerLink(link)

	writeCtx := ctx
	batched := historyBatch != nil
	if batched {
		writeCtx = historyBatch(ctx, operationID)
	}
	updated, err := workflow.UpdateSPSController(
		writeCtx,
		command.LinkID,
		command.ProjectID,
		command.SPSControllerID,
	)
	if err != nil {
		return committedProjectLinkReassignment{}, err
	}
	if updated == nil || updated.ID != command.LinkID ||
		updated.ProjectID != command.ProjectID ||
		updated.SPSControllerID != command.SPSControllerID {
		return committedProjectLinkReassignment{}, errors.New(
			"invalid project SPS controller reassignment result",
		)
	}
	change, err := buildProjectSPSControllerUpdateChange(before, updated)
	if err != nil {
		return committedProjectLinkReassignment{}, err
	}
	return committedProjectLinkReassignment{
		link:                    cloneProjectSPSControllerLink(updated),
		previousSPSControllerID: before.SPSControllerID,
		change:                  change,
		batched:                 batched,
	}, nil
}

func reassignedSPSControllerIDs(previousID, currentID uuid.UUID) []uuid.UUID {
	if previousID == currentID {
		return []uuid.UUID{currentID}
	}
	return []uuid.UUID{previousID, currentID}
}

func buildProjectSPSControllerUpdateChange(
	before *domainProject.ProjectSPSController,
	after *domainProject.ProjectSPSController,
) (mutation.EntityChange, error) {
	beforeJSON, err := json.Marshal(projectSPSControllerSnapshot(before))
	if err != nil {
		return mutation.EntityChange{}, fmt.Errorf(
			"marshal ProjectSPSController before snapshot: %w",
			err,
		)
	}
	afterJSON, err := json.Marshal(projectSPSControllerSnapshot(after))
	if err != nil {
		return mutation.EntityChange{}, fmt.Errorf(
			"marshal ProjectSPSController after snapshot: %w",
			err,
		)
	}
	projectID := after.ProjectID
	var revision *uint64
	if after.Revision != 0 {
		value := after.Revision
		revision = &value
	}
	return mutation.EntityChange{
		EntityType:    mutation.EntityTypeProjectSPSController,
		EntityID:      after.ID,
		ParentID:      &projectID,
		Action:        domainHistory.ActionUpdate,
		Before:        beforeJSON,
		After:         afterJSON,
		ChangedFields: []mutation.FieldName{mutation.FieldNameSPSController},
		Revision:      revision,
	}, nil
}
