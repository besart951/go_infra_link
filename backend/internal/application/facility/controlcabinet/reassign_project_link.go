package controlcabinet

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
	"project control cabinet reassignment transaction is not configured",
)

// ReassignProjectLinkWorkflow combines authoritative root-link reads with the
// compatibility workflow that updates the root and materializes links for the
// new ControlCabinet descendants. The application consumer owns this Interface.
type ReassignProjectLinkWorkflow interface {
	GetByIds(context.Context, []uuid.UUID) ([]*domainProject.ProjectControlCabinet, error)
	UpdateControlCabinet(
		context.Context,
		uuid.UUID,
		uuid.UUID,
		uuid.UUID,
	) (*domainProject.ProjectControlCabinet, error)
}

type ReassignProjectLinkCommand struct {
	ProjectID        uuid.UUID
	LinkID           uuid.UUID
	ExpectedVersion  uint64
	ControlCabinetID uuid.UUID
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
	Link           *domainProject.ProjectControlCabinet
	Mutation       mutation.Result
	DispatchErrors []error
}

type committedProjectLinkReassignment struct {
	link                     *domainProject.ProjectControlCabinet
	previousControlCabinetID uuid.UUID
	change                   mutation.EntityChange
	batched                  bool
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
) (*domainProject.ProjectControlCabinet, error) {
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
				Scope: appcollaboration.FacilityScopeControlCabinet,
				EntityIDs: reassignedControlCabinetIDs(
					result.previousControlCabinetID,
					result.link.ControlCabinetID,
				),
			}
			if _, err := appcollaboration.EnqueueCommand(txCtx, collaborationCommand); err != nil {
				return committedProjectLinkReassignment{}, fmt.Errorf("enqueue ProjectControlCabinet reassignment: %w", err)
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
			"dispatch ProjectControlCabinet reassignment for project %s: %w",
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
	before := cloneProjectControlCabinetLink(link)

	writeCtx := ctx
	batched := historyBatch != nil
	if batched {
		writeCtx = historyBatch(ctx, operationID)
	}
	updated, err := workflow.UpdateControlCabinet(
		writeCtx,
		command.LinkID,
		command.ProjectID,
		command.ControlCabinetID,
	)
	if err != nil {
		return committedProjectLinkReassignment{}, err
	}
	if updated == nil || updated.ID != command.LinkID ||
		updated.ProjectID != command.ProjectID ||
		updated.ControlCabinetID != command.ControlCabinetID {
		return committedProjectLinkReassignment{}, errors.New(
			"invalid project control cabinet reassignment result",
		)
	}
	change, err := buildProjectControlCabinetUpdateChange(before, updated)
	if err != nil {
		return committedProjectLinkReassignment{}, err
	}
	return committedProjectLinkReassignment{
		link:                     cloneProjectControlCabinetLink(updated),
		previousControlCabinetID: before.ControlCabinetID,
		change:                   change,
		batched:                  batched,
	}, nil
}

func reassignedControlCabinetIDs(previousID, currentID uuid.UUID) []uuid.UUID {
	if previousID == currentID {
		return []uuid.UUID{currentID}
	}
	return []uuid.UUID{previousID, currentID}
}

func buildProjectControlCabinetUpdateChange(
	before *domainProject.ProjectControlCabinet,
	after *domainProject.ProjectControlCabinet,
) (mutation.EntityChange, error) {
	beforeJSON, err := json.Marshal(projectControlCabinetSnapshot(before))
	if err != nil {
		return mutation.EntityChange{}, fmt.Errorf(
			"marshal ProjectControlCabinet before snapshot: %w",
			err,
		)
	}
	afterJSON, err := json.Marshal(projectControlCabinetSnapshot(after))
	if err != nil {
		return mutation.EntityChange{}, fmt.Errorf(
			"marshal ProjectControlCabinet after snapshot: %w",
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
		EntityType:    mutation.EntityTypeProjectControlCabinet,
		EntityID:      after.ID,
		ParentID:      &projectID,
		Action:        domainHistory.ActionUpdate,
		Before:        beforeJSON,
		After:         afterJSON,
		ChangedFields: []mutation.FieldName{mutation.FieldNameControlCabinet},
		Revision:      revision,
	}, nil
}
