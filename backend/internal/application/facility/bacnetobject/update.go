package bacnetobject

import (
	"context"
	"errors"
	"sort"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	"github.com/besart951/go_infra_link/backend/internal/application/facility/mutation"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainObjectData "github.com/besart951/go_infra_link/backend/internal/domain/facility/objectdata"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

var ErrUpdateTransactionNotConfigured = errors.New("BACnet object update transaction is not configured")

// UpdateWorkflow is the transaction-scoped Interface consumed by this Module.
// The existing BacnetObjectService is the first Implementation so its mature
// validation, uniqueness, alarm-binding, and ObjectData-link behavior remains
// unchanged during migration.
type UpdateWorkflow interface {
	GetByID(context.Context, uuid.UUID) (*domainFacility.BacnetObject, error)
	Update(context.Context, *domainFacility.BacnetObject, *uuid.UUID) error
}

type ProjectLinkReader interface {
	GetByFieldDeviceIDs(context.Context, []uuid.UUID) ([]*domainProject.ProjectFieldDevice, error)
}

type ObjectDataOwnerReader interface {
	GetByBacnetObjectIDs(
		context.Context,
		[]uuid.UUID,
	) ([]domainObjectData.BacnetObjectOwner, error)
}

type transactionalUpdateOutbox interface {
	UpdateWorkflow
	transactionalCollaborationResolver
}

type HistoryBatchContext func(context.Context, uuid.UUID) context.Context
type ActorProvider func(context.Context) *uuid.UUID
type IDGenerator func() uuid.UUID
type Clock func() time.Time
type ErrorReporter func(error)

type UpdateCommand struct {
	BacnetObjectID  uuid.UUID
	ExpectedVersion uint64
	FieldDeviceID   *uuid.UUID
	FieldDeviceSet  bool
	ObjectDataID    *uuid.UUID
	Patch           domainFacility.BacnetObjectPatch
}

func (c UpdateCommand) validate() error {
	if c.BacnetObjectID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	if c.Patch.ID != uuid.Nil && c.Patch.ID != c.BacnetObjectID {
		return domain.ErrInvalidArgument
	}
	fieldDeviceSet := c.FieldDeviceSet || c.FieldDeviceID != nil
	if fieldDeviceSet && c.FieldDeviceID != nil && c.ObjectDataID != nil {
		return domain.ErrInvalidArgument
	}
	if fieldDeviceSet && c.FieldDeviceID != nil && *c.FieldDeviceID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	if c.ObjectDataID != nil && *c.ObjectDataID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	return nil
}

type LoadError struct {
	Err error
}

func (e *LoadError) Error() string {
	if e == nil || e.Err == nil {
		return "load BACnet object"
	}
	return "load BACnet object: " + e.Err.Error()
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
	ObjectDataOwners    ObjectDataOwnerReader
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
	objectDataOwners      ObjectDataOwnerReader
	dispatcher            appcollaboration.CommandDispatcher
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
	reportError           ErrorReporter
}

type UpdateOutcome struct {
	BacnetObject   *domainFacility.BacnetObject
	Mutation       mutation.Result
	DispatchErrors []error
}

type committedUpdate struct {
	object                 *domainFacility.BacnetObject
	change                 mutation.EntityChange
	affectedFieldDeviceIDs []uuid.UUID
	projectIDs             []uuid.UUID
	batched                bool
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
		objectDataOwners:      deps.ObjectDataOwners,
		dispatcher:            deps.Dispatcher,
		actor:                 deps.Actor,
		newID:                 newID,
		now:                   now,
		reportError:           deps.ReportError,
	}
}

// Update preserves the existing HTTP-facing result. Collaboration delivery is
// best effort after commit and cannot turn a committed mutation into failure.
func (h *UpdateHandler) Update(
	ctx context.Context,
	command UpdateCommand,
) (*domainFacility.BacnetObject, error) {
	outcome, err := h.Execute(ctx, command)
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
	if committed.batched {
		batchID := operationID
		result.BatchID = &batchID
	}
	outcome := UpdateOutcome{
		BacnetObject: committed.object,
		Mutation:     result,
	}

	projectIDs, dispatchErrors := dispatchCommittedMutation(
		ctx,
		collaborationDependencies{
			projectLinks:     h.projectLinks,
			objectDataOwners: h.objectDataOwners,
			dispatcher:       h.dispatcher,
			newID:            h.newID,
		},
		command.BacnetObjectID,
		committed.affectedFieldDeviceIDs,
		operationID,
		actorID,
		occurredAt,
	)
	if len(committed.projectIDs) > 0 {
		outcome.Mutation.ProjectIDs = append([]uuid.UUID(nil), committed.projectIDs...)
	} else {
		outcome.Mutation.ProjectIDs = append([]uuid.UUID(nil), projectIDs...)
	}
	outcome.DispatchErrors = append(outcome.DispatchErrors, dispatchErrors...)

	return outcome, nil
}

func objectDataProjectIDs(
	bacnetObjectID uuid.UUID,
	owners []domainObjectData.BacnetObjectOwner,
) map[uuid.UUID]struct{} {
	projects := make(map[uuid.UUID]struct{})
	for _, owner := range owners {
		if owner.BacnetObjectID != bacnetObjectID || owner.ProjectID == nil ||
			*owner.ProjectID == uuid.Nil {
			continue
		}
		projects[*owner.ProjectID] = struct{}{}
	}
	return projects
}

func sortedUpdateProjectIDs(
	fieldDevices map[uuid.UUID][]uuid.UUID,
	objectData map[uuid.UUID]struct{},
) []uuid.UUID {
	unique := make(map[uuid.UUID]struct{}, len(fieldDevices)+len(objectData))
	for projectID := range fieldDevices {
		unique[projectID] = struct{}{}
	}
	for projectID := range objectData {
		unique[projectID] = struct{}{}
	}
	ids := make([]uuid.UUID, 0, len(unique))
	for projectID := range unique {
		ids = append(ids, projectID)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })
	return ids
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

	before, err := workflow.GetByID(ctx, command.BacnetObjectID)
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

	updated := cloneBacnetObject(before)
	patch := command.Patch
	patch.ID = command.BacnetObjectID
	if err := updated.ApplyPatch(patch); err != nil {
		return committedUpdate{}, err
	}
	move := newMoveCommand(before, command)
	if move != nil {
		err = move.applyTo(updated)
	} else {
		err = command.applyAssignmentTo(updated)
	}
	if err != nil {
		return committedUpdate{}, err
	}

	writeCtx := ctx
	batched := historyBatch != nil
	if batched {
		writeCtx = historyBatch(ctx, operationID)
	}
	if err := workflow.Update(writeCtx, updated, command.ObjectDataID); err != nil {
		return committedUpdate{}, err
	}

	after, err := workflow.GetByID(writeCtx, command.BacnetObjectID)
	if err != nil {
		return committedUpdate{}, err
	}
	if after == nil {
		return committedUpdate{}, domain.ErrNotFound
	}
	change, err := buildUpdateChange(before, after, command.ObjectDataID)
	if err != nil {
		return committedUpdate{}, err
	}
	fieldDeviceIDs := affectedFieldDeviceIDs(before, after)
	var projectIDs []uuid.UUID
	if outbox, ok := workflow.(transactionalUpdateOutbox); ok {
		projectIDs, err = enqueueTransactionalMutation(
			writeCtx,
			outbox,
			command.BacnetObjectID,
			after.Revision,
			fieldDeviceIDs,
			operationID,
			actorID,
			occurredAt,
			newID,
		)
		if err != nil {
			return committedUpdate{}, err
		}
	}

	return committedUpdate{
		object:                 cloneBacnetObject(after),
		change:                 change,
		affectedFieldDeviceIDs: fieldDeviceIDs,
		projectIDs:             projectIDs,
		batched:                batched,
	}, nil
}

func actorFromContext(provider ActorProvider, ctx context.Context) *uuid.UUID {
	if provider == nil {
		return nil
	}
	return clonePointer(provider(ctx))
}

func affectedFieldDeviceIDs(
	before *domainFacility.BacnetObject,
	after *domainFacility.BacnetObject,
) []uuid.UUID {
	unique := make(map[uuid.UUID]struct{}, 2)
	if before != nil && before.FieldDeviceID != nil && *before.FieldDeviceID != uuid.Nil {
		unique[*before.FieldDeviceID] = struct{}{}
	}
	if after != nil && after.FieldDeviceID != nil && *after.FieldDeviceID != uuid.Nil {
		unique[*after.FieldDeviceID] = struct{}{}
	}
	ids := make([]uuid.UUID, 0, len(unique))
	for id := range unique {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })
	return ids
}

func groupLinkedFieldDevices(
	links []*domainProject.ProjectFieldDevice,
	affected []uuid.UUID,
) map[uuid.UUID][]uuid.UUID {
	allowed := make(map[uuid.UUID]struct{}, len(affected))
	for _, id := range affected {
		allowed[id] = struct{}{}
	}
	sets := make(map[uuid.UUID]map[uuid.UUID]struct{})
	for _, link := range links {
		if link == nil || link.ProjectID == uuid.Nil {
			continue
		}
		if _, ok := allowed[link.FieldDeviceID]; !ok {
			continue
		}
		if sets[link.ProjectID] == nil {
			sets[link.ProjectID] = make(map[uuid.UUID]struct{})
		}
		sets[link.ProjectID][link.FieldDeviceID] = struct{}{}
	}

	grouped := make(map[uuid.UUID][]uuid.UUID, len(sets))
	for projectID, ids := range sets {
		items := make([]uuid.UUID, 0, len(ids))
		for id := range ids {
			items = append(items, id)
		}
		sort.Slice(items, func(i, j int) bool { return items[i].String() < items[j].String() })
		grouped[projectID] = items
	}
	return grouped
}

func sortedProjectIDs(grouped map[uuid.UUID][]uuid.UUID) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(grouped))
	for id := range grouped {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })
	return ids
}

func cloneBacnetObject(source *domainFacility.BacnetObject) *domainFacility.BacnetObject {
	if source == nil {
		return nil
	}
	clone := *source
	clone.Description = clonePointer(source.Description)
	clone.TextIndividual = clonePointer(source.TextIndividual)
	clone.FieldDeviceID = clonePointer(source.FieldDeviceID)
	clone.SoftwareReferenceID = clonePointer(source.SoftwareReferenceID)
	clone.StateTextID = clonePointer(source.StateTextID)
	clone.NotificationClassID = clonePointer(source.NotificationClassID)
	clone.AlarmTypeID = clonePointer(source.AlarmTypeID)
	clone.AlarmDefinitionID = clonePointer(source.AlarmDefinitionID)
	return &clone
}

func clonePointer[T any](value *T) *T {
	if value == nil {
		return nil
	}
	clone := *value
	return &clone
}
