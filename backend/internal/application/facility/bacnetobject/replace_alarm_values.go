package bacnetobject

import (
	"context"
	"encoding/json"
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
	"github.com/google/uuid"
)

var ErrReplaceAlarmValuesTransactionNotConfigured = errors.New(
	"BACnet alarm-value replacement transaction is not configured",
)

type AlarmValueStore interface {
	GetValues(context.Context, uuid.UUID) ([]domainFacility.BacnetObjectAlarmValue, error)
	PutValues(context.Context, uuid.UUID, []domainFacility.BacnetObjectAlarmValue) error
}

// ReplaceAlarmValuesWorkflow is the transaction-scoped Interface consumed by
// the application Module. The compatibility alarm-value service provides
// persistence while the wire Adapter provides authoritative schema reads.
type ReplaceAlarmValuesWorkflow interface {
	AlarmValueStore
	GetAlarmSelection(context.Context, uuid.UUID) (*AlarmSelection, error)
	GetAlarmTypeFields(
		context.Context,
		[]uuid.UUID,
	) ([]*domainFacility.AlarmTypeField, error)
}

type AlarmSelection struct {
	BacnetObjectID uuid.UUID
	AlarmTypeID    *uuid.UUID
}

// BacnetObjectStateReader resolves the current direct owner after commit. It is
// deliberately separate from the transaction workflow because publication is
// best effort and must not alter the committed HTTP result.
type BacnetObjectStateReader interface {
	GetByIds(context.Context, []uuid.UUID) ([]*domainFacility.BacnetObject, error)
}

type transactionalAlarmValuesOutbox interface {
	ReplaceAlarmValuesWorkflow
	BacnetObjectStateReader
	transactionalCollaborationResolver
}

// AlarmValueInput contains only client-controlled state. IDs, timestamps, and
// the BACnet parent are assigned within the application/persistence boundary.
type AlarmValueInput struct {
	AlarmTypeFieldID uuid.UUID
	ValueNumber      *float64
	ValueInteger     *int64
	ValueBoolean     *bool
	ValueString      *string
	ValueJSON        *string
	UnitID           *uuid.UUID
	Source           string
}

type ReplaceAlarmValuesCommand struct {
	BacnetObjectID uuid.UUID
	Values         []AlarmValueInput
}

func (c ReplaceAlarmValuesCommand) validate() error {
	if c.BacnetObjectID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	return nil
}

func (c ReplaceAlarmValuesCommand) domainValues() []domainFacility.BacnetObjectAlarmValue {
	values := make([]domainFacility.BacnetObjectAlarmValue, len(c.Values))
	for i := range c.Values {
		source := c.Values[i].Source
		if source == "" {
			source = domainFacility.AlarmValueSourceUser
		}
		values[i] = domainFacility.BacnetObjectAlarmValue{
			BacnetObjectID:   c.BacnetObjectID,
			AlarmTypeFieldID: c.Values[i].AlarmTypeFieldID,
			ValueNumber:      clonePointer(c.Values[i].ValueNumber),
			ValueInteger:     clonePointer(c.Values[i].ValueInteger),
			ValueBoolean:     clonePointer(c.Values[i].ValueBoolean),
			ValueString:      clonePointer(c.Values[i].ValueString),
			ValueJSON:        clonePointer(c.Values[i].ValueJSON),
			UnitID:           clonePointer(c.Values[i].UnitID),
			Source:           source,
		}
	}
	return values
}

type AlarmValuesReloadError struct {
	Err error
}

func (e *AlarmValuesReloadError) Error() string {
	if e == nil || e.Err == nil {
		return "reload BACnet alarm values"
	}
	return "reload BACnet alarm values: " + e.Err.Error()
}

func (e *AlarmValuesReloadError) Unwrap() error {
	if e == nil {
		return nil
	}
	return e.Err
}

type ReplaceAlarmValuesDependencies struct {
	TransactionRunner   apptransaction.Runner
	TransactionWorkflow apptransaction.Factory[ReplaceAlarmValuesWorkflow]
	HistoryBatch        HistoryBatchContext
	BacnetObjects       BacnetObjectStateReader
	ProjectLinks        ProjectLinkReader
	ObjectDataOwners    ObjectDataOwnerReader
	Dispatcher          appcollaboration.CommandDispatcher
	Actor               ActorProvider
	NewID               IDGenerator
	Now                 Clock
	ReportError         ErrorReporter
}

type ReplaceAlarmValuesHandler struct {
	operation             apptransaction.Operation[ReplaceAlarmValuesWorkflow, ReplaceAlarmValuesWorkflow]
	transactionConfigured bool
	historyBatch          HistoryBatchContext
	bacnetObjects         BacnetObjectStateReader
	projectLinks          ProjectLinkReader
	objectDataOwners      ObjectDataOwnerReader
	dispatcher            appcollaboration.CommandDispatcher
	actor                 ActorProvider
	newID                 IDGenerator
	now                   Clock
	reportError           ErrorReporter
}

type ReplaceAlarmValuesOutcome struct {
	Values         []domainFacility.BacnetObjectAlarmValue
	Mutation       mutation.Result
	DispatchErrors []error
}

type committedAlarmValueReplacement struct {
	values     []domainFacility.BacnetObjectAlarmValue
	changes    []mutation.EntityChange
	projectIDs []uuid.UUID
	batched    bool
}

func NewReplaceAlarmValuesHandler(
	deps ReplaceAlarmValuesDependencies,
) *ReplaceAlarmValuesHandler {
	newID := deps.NewID
	if newID == nil {
		newID = uuid.New
	}
	now := deps.Now
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	boundary := apptransaction.NewBoundary[ReplaceAlarmValuesWorkflow](
		deps.TransactionRunner,
		deps.TransactionWorkflow,
	)
	var noNonTransactionalWorkflow ReplaceAlarmValuesWorkflow
	operation := apptransaction.Bind(
		boundary,
		noNonTransactionalWorkflow,
		func(workflow ReplaceAlarmValuesWorkflow) ReplaceAlarmValuesWorkflow {
			return workflow
		},
	)
	return &ReplaceAlarmValuesHandler{
		operation:             operation,
		transactionConfigured: deps.TransactionRunner != nil && deps.TransactionWorkflow != nil,
		historyBatch:          deps.HistoryBatch,
		bacnetObjects:         deps.BacnetObjects,
		projectLinks:          deps.ProjectLinks,
		objectDataOwners:      deps.ObjectDataOwners,
		dispatcher:            deps.Dispatcher,
		actor:                 deps.Actor,
		newID:                 newID,
		now:                   now,
		reportError:           deps.ReportError,
	}
}

func (h *ReplaceAlarmValuesHandler) ReplaceAlarmValues(
	ctx context.Context,
	command ReplaceAlarmValuesCommand,
) ([]domainFacility.BacnetObjectAlarmValue, error) {
	outcome, err := h.Execute(ctx, command)
	if err != nil {
		return nil, err
	}
	for _, dispatchErr := range outcome.DispatchErrors {
		if h.reportError != nil {
			h.reportError(dispatchErr)
		}
	}
	return outcome.Values, nil
}

func (h *ReplaceAlarmValuesHandler) Execute(
	ctx context.Context,
	command ReplaceAlarmValuesCommand,
) (ReplaceAlarmValuesOutcome, error) {
	if h == nil || !h.transactionConfigured {
		return ReplaceAlarmValuesOutcome{}, ErrReplaceAlarmValuesTransactionNotConfigured
	}
	if err := command.validate(); err != nil {
		return ReplaceAlarmValuesOutcome{}, err
	}

	operationID := h.newID()
	actorID := actorFromContext(h.actor, ctx)
	occurredAt := h.now().UTC()
	committed, err := apptransaction.RunResult(
		ctx,
		h.operation,
		func(
			txCtx context.Context,
			workflow ReplaceAlarmValuesWorkflow,
		) (committedAlarmValueReplacement, error) {
			return executeReplaceAlarmValuesTransaction(
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
		return ReplaceAlarmValuesOutcome{}, err
	}

	result := mutation.Result{
		OperationID: operationID,
		ActorID:     actorID,
		OccurredAt:  occurredAt,
		Changes:     committed.changes,
	}
	if committed.batched {
		batchID := operationID
		result.BatchID = &batchID
	}
	outcome := ReplaceAlarmValuesOutcome{
		Values:   committed.values,
		Mutation: result,
	}
	if h.dispatcher == nil || len(committed.changes) == 0 {
		return outcome, nil
	}

	dispatchCtx := context.WithoutCancel(ctx)
	fieldDeviceIDs := make([]uuid.UUID, 0, 1)
	if h.bacnetObjects != nil {
		objects, readErr := h.bacnetObjects.GetByIds(
			dispatchCtx,
			[]uuid.UUID{command.BacnetObjectID},
		)
		if readErr != nil {
			outcome.DispatchErrors = append(outcome.DispatchErrors, fmt.Errorf(
				"resolve committed BACnet object owner: %w",
				readErr,
			))
		} else {
			fieldDeviceIDs = currentFieldDeviceIDs(command.BacnetObjectID, objects)
		}
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
		fieldDeviceIDs,
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

func executeReplaceAlarmValuesTransaction(
	ctx context.Context,
	workflow ReplaceAlarmValuesWorkflow,
	command ReplaceAlarmValuesCommand,
	operationID uuid.UUID,
	actorID *uuid.UUID,
	occurredAt time.Time,
	newID IDGenerator,
	historyBatch HistoryBatchContext,
) (committedAlarmValueReplacement, error) {
	if workflow == nil {
		return committedAlarmValueReplacement{}, ErrReplaceAlarmValuesTransactionNotConfigured
	}

	if err := validateAlarmValueFields(ctx, workflow, command); err != nil {
		return committedAlarmValueReplacement{}, err
	}
	before, err := workflow.GetValues(ctx, command.BacnetObjectID)
	if err != nil {
		return committedAlarmValueReplacement{}, err
	}
	writeCtx := ctx
	batched := historyBatch != nil
	if batched {
		writeCtx = historyBatch(ctx, operationID)
	}
	if err := workflow.PutValues(
		writeCtx,
		command.BacnetObjectID,
		command.domainValues(),
	); err != nil {
		return committedAlarmValueReplacement{}, err
	}
	after, err := workflow.GetValues(writeCtx, command.BacnetObjectID)
	if err != nil {
		return committedAlarmValueReplacement{}, &AlarmValuesReloadError{Err: err}
	}
	changes, err := buildAlarmValueReplacementChanges(
		command.BacnetObjectID,
		before,
		after,
	)
	if err != nil {
		return committedAlarmValueReplacement{}, err
	}
	var projectIDs []uuid.UUID
	if len(changes) > 0 {
		if outbox, ok := workflow.(transactionalAlarmValuesOutbox); ok {
			objects, err := outbox.GetByIds(writeCtx, []uuid.UUID{command.BacnetObjectID})
			if err != nil {
				return committedAlarmValueReplacement{}, fmt.Errorf(
					"resolve BACnet object owner for outbox: %w",
					err,
				)
			}
			projectIDs, err = enqueueTransactionalMutation(
				writeCtx,
				outbox,
				command.BacnetObjectID,
				0,
				currentFieldDeviceIDs(command.BacnetObjectID, objects),
				operationID,
				actorID,
				occurredAt,
				newID,
			)
			if err != nil {
				return committedAlarmValueReplacement{}, err
			}
		}
	}
	return committedAlarmValueReplacement{
		values:     cloneAlarmValues(after),
		changes:    changes,
		projectIDs: projectIDs,
		batched:    batched,
	}, nil
}

func validateAlarmValueFields(
	ctx context.Context,
	workflow ReplaceAlarmValuesWorkflow,
	command ReplaceAlarmValuesCommand,
) error {
	selection, err := workflow.GetAlarmSelection(ctx, command.BacnetObjectID)
	if err != nil {
		return err
	}
	if selection == nil || selection.BacnetObjectID != command.BacnetObjectID {
		return domain.ErrNotFound
	}
	if len(command.Values) == 0 {
		return nil
	}

	validation := domain.NewValidationError()
	uniqueIDs := make([]uuid.UUID, 0, len(command.Values))
	seen := make(map[uuid.UUID]int, len(command.Values))
	for index, value := range command.Values {
		field := fmt.Sprintf("values.%d.alarm_type_field_id", index)
		if value.AlarmTypeFieldID == uuid.Nil {
			validation.Add(field, "alarm_type_field_id is required")
			continue
		}
		if firstIndex, exists := seen[value.AlarmTypeFieldID]; exists {
			validation.Add(
				field,
				fmt.Sprintf(
					"alarm type field is duplicated from values.%d.alarm_type_field_id",
					firstIndex,
				),
			)
			continue
		}
		seen[value.AlarmTypeFieldID] = index
		uniqueIDs = append(uniqueIDs, value.AlarmTypeFieldID)
	}
	if selection.AlarmTypeID == nil || *selection.AlarmTypeID == uuid.Nil {
		for index := range command.Values {
			validation.Add(
				fmt.Sprintf("values.%d.alarm_type_field_id", index),
				"BACnet object has no selected alarm type",
			)
		}
	}
	if len(validation.Fields) > 0 {
		return validation
	}

	fields, err := workflow.GetAlarmTypeFields(ctx, uniqueIDs)
	if err != nil {
		return err
	}
	byID := make(map[uuid.UUID]*domainFacility.AlarmTypeField, len(fields))
	for _, field := range fields {
		if field != nil && field.ID != uuid.Nil {
			byID[field.ID] = field
		}
	}
	for index, value := range command.Values {
		fieldPath := fmt.Sprintf("values.%d.alarm_type_field_id", index)
		field := byID[value.AlarmTypeFieldID]
		if field == nil {
			validation.Add(fieldPath, "alarm type field does not exist")
			continue
		}
		if field.AlarmTypeID != *selection.AlarmTypeID {
			validation.Add(
				fieldPath,
				"alarm type field does not belong to the BACnet object's selected alarm type",
			)
		}
	}
	if len(validation.Fields) > 0 {
		return validation
	}
	return nil
}

type alarmValueSnapshot struct {
	ID               uuid.UUID  `json:"id"`
	CreatedAt        time.Time  `json:"created_at"`
	UpdatedAt        time.Time  `json:"updated_at"`
	BacnetObjectID   uuid.UUID  `json:"bacnet_object_id"`
	AlarmTypeFieldID uuid.UUID  `json:"alarm_type_field_id"`
	ValueNumber      *float64   `json:"value_number"`
	ValueInteger     *int64     `json:"value_integer"`
	ValueBoolean     *bool      `json:"value_boolean"`
	ValueString      *string    `json:"value_string"`
	ValueJSON        *string    `json:"value_json"`
	UnitID           *uuid.UUID `json:"unit_id"`
	Source           string     `json:"source"`
}

func buildAlarmValueReplacementChanges(
	bacnetObjectID uuid.UUID,
	before []domainFacility.BacnetObjectAlarmValue,
	after []domainFacility.BacnetObjectAlarmValue,
) ([]mutation.EntityChange, error) {
	before = cloneAlarmValues(before)
	after = cloneAlarmValues(after)
	sort.Slice(before, func(i, j int) bool { return before[i].ID.String() < before[j].ID.String() })
	sort.Slice(after, func(i, j int) bool { return after[i].ID.String() < after[j].ID.String() })
	changes := make([]mutation.EntityChange, 0, len(before)+len(after))
	for i := range before {
		snapshot, err := json.Marshal(toAlarmValueSnapshot(before[i]))
		if err != nil {
			return nil, err
		}
		parentID := bacnetObjectID
		changes = append(changes, mutation.EntityChange{
			EntityType: mutation.EntityTypeBacnetAlarmValue,
			EntityID:   before[i].ID,
			ParentID:   &parentID,
			Action:     domainHistory.ActionDelete,
			Before:     snapshot,
		})
	}
	for i := range after {
		snapshot, err := json.Marshal(toAlarmValueSnapshot(after[i]))
		if err != nil {
			return nil, err
		}
		parentID := bacnetObjectID
		changes = append(changes, mutation.EntityChange{
			EntityType: mutation.EntityTypeBacnetAlarmValue,
			EntityID:   after[i].ID,
			ParentID:   &parentID,
			Action:     domainHistory.ActionCreate,
			After:      snapshot,
		})
	}
	return changes, nil
}

func toAlarmValueSnapshot(value domainFacility.BacnetObjectAlarmValue) alarmValueSnapshot {
	return alarmValueSnapshot{
		ID:               value.ID,
		CreatedAt:        value.CreatedAt,
		UpdatedAt:        value.UpdatedAt,
		BacnetObjectID:   value.BacnetObjectID,
		AlarmTypeFieldID: value.AlarmTypeFieldID,
		ValueNumber:      clonePointer(value.ValueNumber),
		ValueInteger:     clonePointer(value.ValueInteger),
		ValueBoolean:     clonePointer(value.ValueBoolean),
		ValueString:      clonePointer(value.ValueString),
		ValueJSON:        clonePointer(value.ValueJSON),
		UnitID:           clonePointer(value.UnitID),
		Source:           value.Source,
	}
}

func cloneAlarmValues(
	values []domainFacility.BacnetObjectAlarmValue,
) []domainFacility.BacnetObjectAlarmValue {
	clones := make([]domainFacility.BacnetObjectAlarmValue, len(values))
	for i := range values {
		clones[i] = values[i]
		clones[i].ValueNumber = clonePointer(values[i].ValueNumber)
		clones[i].ValueInteger = clonePointer(values[i].ValueInteger)
		clones[i].ValueBoolean = clonePointer(values[i].ValueBoolean)
		clones[i].ValueString = clonePointer(values[i].ValueString)
		clones[i].ValueJSON = clonePointer(values[i].ValueJSON)
		clones[i].UnitID = clonePointer(values[i].UnitID)
	}
	return clones
}

func currentFieldDeviceIDs(
	bacnetObjectID uuid.UUID,
	objects []*domainFacility.BacnetObject,
) []uuid.UUID {
	unique := make(map[uuid.UUID]struct{}, 1)
	for _, object := range objects {
		if object == nil || object.ID != bacnetObjectID || object.FieldDeviceID == nil ||
			*object.FieldDeviceID == uuid.Nil {
			continue
		}
		unique[*object.FieldDeviceID] = struct{}{}
	}
	ids := make([]uuid.UUID, 0, len(unique))
	for id := range unique {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i].String() < ids[j].String() })
	return ids
}
