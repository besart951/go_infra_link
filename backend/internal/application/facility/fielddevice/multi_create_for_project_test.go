package fielddevice

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	apptransaction "github.com/besart951/go_infra_link/backend/internal/application/transaction"
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

type projectMultiCreateHistoryBatchKey struct{}

type projectMultiCreateTransactionState struct {
	fieldDevices map[uuid.UUID]*domainFacility.FieldDevice
	projectLinks map[uuid.UUID][]uuid.UUID
	history      []string
}

func (s projectMultiCreateTransactionState) clone() projectMultiCreateTransactionState {
	fieldDevices := make(map[uuid.UUID]*domainFacility.FieldDevice, len(s.fieldDevices))
	for id, fieldDevice := range s.fieldDevices {
		fieldDevices[id] = cloneFieldDevice(fieldDevice)
	}
	projectLinks := make(map[uuid.UUID][]uuid.UUID, len(s.projectLinks))
	for projectID, fieldDeviceIDs := range s.projectLinks {
		projectLinks[projectID] = append([]uuid.UUID(nil), fieldDeviceIDs...)
	}
	return projectMultiCreateTransactionState{
		fieldDevices: fieldDevices,
		projectLinks: projectLinks,
		history:      append([]string(nil), s.history...),
	}
}

type projectMultiCreateTransactionUnit struct {
	state *projectMultiCreateTransactionState
}

type projectMultiCreateTransactionHarness struct {
	committed       projectMultiCreateTransactionState
	result          *domainFacility.FieldDeviceMultiCreateResult
	workflowErr     error
	commitErr       error
	failedAttempt   *domainFacility.FieldDevice
	runnerCalls     int
	workflowCalls   int
	projectID       uuid.UUID
	items           []domainFacility.FieldDeviceCreateItem
	historyBatchIDs []uuid.UUID
}

func (h *projectMultiCreateTransactionHarness) runner(
	ctx context.Context,
	run func(context.Context, apptransaction.UnitOfWork) error,
) error {
	h.runnerCalls++
	staged := h.committed.clone()
	if err := run(ctx, projectMultiCreateTransactionUnit{state: &staged}); err != nil {
		return err
	}
	if h.commitErr != nil {
		return h.commitErr
	}
	h.committed = staged
	return nil
}

func (h *projectMultiCreateTransactionHarness) factory(
	unit apptransaction.UnitOfWork,
) (MultiCreateForProjectWorkflow, error) {
	typed, ok := unit.(projectMultiCreateTransactionUnit)
	if !ok || typed.state == nil {
		return nil, errors.New("unexpected project FieldDevice multi-create transaction unit")
	}
	return &projectMultiCreateWorkflowStub{harness: h, state: typed.state}, nil
}

type projectMultiCreateWorkflowStub struct {
	harness *projectMultiCreateTransactionHarness
	state   *projectMultiCreateTransactionState
}

func (s *projectMultiCreateWorkflowStub) MultiCreateAndAssignFieldDevices(
	ctx context.Context,
	projectID uuid.UUID,
	items []domainFacility.FieldDeviceCreateItem,
) (*domainFacility.FieldDeviceMultiCreateResult, error) {
	s.harness.workflowCalls++
	s.harness.projectID = projectID
	s.harness.items = items

	batchID, hasBatch := ctx.Value(projectMultiCreateHistoryBatchKey{}).(uuid.UUID)
	if s.harness.failedAttempt != nil {
		historyStart := len(s.state.history)
		failed := cloneFieldDevice(s.harness.failedAttempt)
		s.state.fieldDevices[failed.ID] = failed
		s.state.history = append(s.state.history, "field_device:create:failed-item-prefix")
		if hasBatch {
			s.harness.historyBatchIDs = append(s.harness.historyBatchIDs, batchID)
		}
		delete(s.state.fieldDevices, failed.ID)
		s.state.history = s.state.history[:historyStart]
	}
	if s.harness.result != nil {
		for _, item := range s.harness.result.Results {
			if !item.Success || item.FieldDevice == nil {
				continue
			}
			fieldDevice := cloneFieldDevice(item.FieldDevice)
			s.state.fieldDevices[fieldDevice.ID] = fieldDevice
			s.state.projectLinks[projectID] = append(
				s.state.projectLinks[projectID],
				fieldDevice.ID,
			)
			s.state.history = append(
				s.state.history,
				"field_device:create",
				"project_field_device:create",
			)
			if hasBatch {
				s.harness.historyBatchIDs = append(
					s.harness.historyBatchIDs,
					batchID,
					batchID,
				)
			}
		}
	}
	return s.harness.result, s.harness.workflowErr
}

type projectMultiCreateDispatcherStub struct {
	harness           *projectMultiCreateTransactionHarness
	projectID         uuid.UUID
	wantIDs           []uuid.UUID
	commands          []appcollaboration.Command
	err               error
	calledAfterCommit bool
}

func (s *projectMultiCreateDispatcherStub) Dispatch(
	_ context.Context,
	command appcollaboration.Command,
) error {
	s.commands = append(s.commands, command)
	links := s.harness.committed.projectLinks[s.projectID]
	committed := len(links) == len(s.wantIDs)
	for _, id := range s.wantIDs {
		committed = committed && s.harness.committed.fieldDevices[id] != nil
	}
	s.calledAfterCommit = committed && reflect.DeepEqual(links, s.wantIDs)
	return s.err
}

func TestMultiCreateForProjectCommitsPartialResultAndLinksBeforeTypedDispatch(t *testing.T) {
	projectID := fieldDeviceTestUUID(301)
	unrelatedProjectID := fieldDeviceTestUUID(302)
	firstID := fieldDeviceTestUUID(303)
	thirdID := fieldDeviceTestUUID(304)
	leakedFailedID := fieldDeviceTestUUID(305)
	firstParentID := fieldDeviceTestUUID(306)
	thirdParentID := fieldDeviceTestUUID(307)
	actorID := fieldDeviceTestUUID(308)
	operationID := fieldDeviceTestUUID(309)
	eventID := fieldDeviceTestUUID(310)
	objectDataID := fieldDeviceTestUUID(311)
	createdAt := time.Date(2026, time.July, 21, 8, 0, 0, 0, time.UTC)
	updatedAt := createdAt.Add(time.Minute)
	occurredAt := updatedAt.Add(time.Second)
	bmk := "M01"
	description := "supply air sensor"
	textFix := "AI01"
	specificationID := fieldDeviceTestUUID(312)
	first := &domainFacility.FieldDevice{
		Base:                      domain.Base{ID: firstID, CreatedAt: createdAt, UpdatedAt: updatedAt},
		BMK:                       &bmk,
		Description:               &description,
		TextIndividuell:           &textFix,
		ApparatNr:                 4,
		SPSControllerSystemTypeID: firstParentID,
		SystemPartID:              fieldDeviceTestUUID(313),
		SpecificationID:           &specificationID,
		ApparatID:                 fieldDeviceTestUUID(314),
	}
	third := &domainFacility.FieldDevice{
		Base:                      domain.Base{ID: thirdID, CreatedAt: createdAt, UpdatedAt: createdAt},
		ApparatNr:                 7,
		SPSControllerSystemTypeID: thirdParentID,
		SystemPartID:              fieldDeviceTestUUID(315),
		ApparatID:                 fieldDeviceTestUUID(316),
	}
	legacyResult := &domainFacility.FieldDeviceMultiCreateResult{
		Results: []domainFacility.FieldDeviceCreateResult{
			{Index: 0, Success: true, FieldDevice: first},
			{Index: 1, Error: "BACnet validation failed", ErrorField: "bacnet_objects"},
			{Index: 2, Success: true, FieldDevice: third},
		},
		TotalRequests: 3,
		SuccessCount:  2,
		FailureCount:  1,
	}
	harness := &projectMultiCreateTransactionHarness{
		committed: projectMultiCreateTransactionState{
			fieldDevices: map[uuid.UUID]*domainFacility.FieldDevice{},
			projectLinks: map[uuid.UUID][]uuid.UUID{
				unrelatedProjectID: {fieldDeviceTestUUID(317)},
			},
		},
		result: legacyResult,
		failedAttempt: &domainFacility.FieldDevice{
			Base: domain.Base{ID: leakedFailedID},
		},
	}
	dispatcher := &projectMultiCreateDispatcherStub{
		harness:   harness,
		projectID: projectID,
		wantIDs:   []uuid.UUID{firstID, thirdID},
	}
	generatedIDs := []uuid.UUID{operationID, eventID}
	handler := NewMultiCreateForProjectHandler(MultiCreateForProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, projectMultiCreateHistoryBatchKey{}, batchID)
		},
		Dispatcher: dispatcher,
		Actor:      func(context.Context) *uuid.UUID { return &actorID },
		NewID: func() uuid.UUID {
			id := generatedIDs[0]
			generatedIDs = generatedIDs[1:]
			return id
		},
		Now: func() time.Time { return occurredAt },
	})
	items := []domainFacility.FieldDeviceCreateItem{
		{FieldDevice: &domainFacility.FieldDevice{ApparatNr: 4}, ObjectDataID: &objectDataID},
		{FieldDevice: &domainFacility.FieldDevice{ApparatNr: 5}},
		{
			FieldDevice: &domainFacility.FieldDevice{ApparatNr: 7},
			BacnetObjects: []domainFacility.BacnetObject{{
				TextFix:        "AI7",
				SoftwareType:   domainFacility.BacnetSoftwareTypeAI,
				SoftwareNumber: 7,
			}},
		},
	}

	outcome, err := handler.Execute(context.Background(), MultiCreateForProjectCommand{
		ProjectID: projectID,
		Items:     items,
	})
	if err != nil {
		t.Fatalf("execute project FieldDevice multi-create: %v", err)
	}
	if outcome.Result != legacyResult || harness.runnerCalls != 1 || harness.workflowCalls != 1 ||
		harness.projectID != projectID || len(harness.items) != len(items) {
		t.Fatalf("workflow/result changed: outcome=%p want=%p harness=%+v",
			outcome.Result,
			legacyResult,
			harness,
		)
	}
	if harness.items[0].ObjectDataID == nil || *harness.items[0].ObjectDataID != objectDataID ||
		len(harness.items[2].BacnetObjects) != 1 || harness.items[2].BacnetObjects[0].TextFix != "AI7" {
		t.Fatalf("creation selections changed: %+v", harness.items)
	}
	if !reflect.DeepEqual(harness.committed.projectLinks[projectID], []uuid.UUID{firstID, thirdID}) ||
		harness.committed.fieldDevices[firstID] == nil ||
		harness.committed.fieldDevices[thirdID] == nil {
		t.Fatalf("successful rows/links not committed: %+v", harness.committed)
	}
	if harness.committed.fieldDevices[leakedFailedID] != nil {
		t.Fatal("failed-item prefix write escaped its savepoint")
	}
	if !reflect.DeepEqual(harness.committed.history, []string{
		"field_device:create",
		"project_field_device:create",
		"field_device:create",
		"project_field_device:create",
	}) {
		t.Fatalf("committed history includes failed item: %v", harness.committed.history)
	}
	if !reflect.DeepEqual(harness.historyBatchIDs, []uuid.UUID{
		operationID,
		operationID,
		operationID,
		operationID,
		operationID,
	}) {
		t.Fatalf("history batch IDs: %v", harness.historyBatchIDs)
	}
	if outcome.Mutation.OperationID != operationID ||
		outcome.Mutation.BatchID == nil || *outcome.Mutation.BatchID != operationID ||
		outcome.Mutation.ActorID == nil || *outcome.Mutation.ActorID != actorID ||
		!outcome.Mutation.OccurredAt.Equal(occurredAt) ||
		!reflect.DeepEqual(outcome.Mutation.ProjectIDs, []uuid.UUID{projectID}) ||
		len(outcome.Mutation.Changes) != 2 {
		t.Fatalf("mutation result: %+v", outcome.Mutation)
	}
	for index, want := range []struct {
		id       uuid.UUID
		parentID uuid.UUID
	}{
		{id: firstID, parentID: firstParentID},
		{id: thirdID, parentID: thirdParentID},
	} {
		change := outcome.Mutation.Changes[index]
		if change.EntityType != "field_device" || change.EntityID != want.id ||
			change.ParentID == nil || *change.ParentID != want.parentID ||
			change.Action != domainHistory.ActionCreate {
			t.Fatalf("change %d: %+v", index, change)
		}
		var snapshot fieldDeviceSnapshot
		if err := json.Unmarshal(change.After, &snapshot); err != nil {
			t.Fatalf("decode change %d: %v", index, err)
		}
		if snapshot.ID != want.id || snapshot.SPSControllerSystemTypeID != want.parentID {
			t.Fatalf("snapshot %d: %+v", index, snapshot)
		}
	}
	if len(dispatcher.commands) != 1 || !dispatcher.calledAfterCommit {
		t.Fatalf("dispatch: commands=%v afterCommit=%t", dispatcher.commands, dispatcher.calledAfterCommit)
	}
	command, ok := dispatcher.commands[0].(appcollaboration.FieldDevicesCreated)
	if !ok {
		t.Fatalf("command type: %T", dispatcher.commands[0])
	}
	if command.SchemaVersion != appcollaboration.SchemaVersionV1 ||
		command.EventID != eventID || command.OperationID != operationID ||
		command.CorrelationID != operationID || command.ProjectID != projectID ||
		command.ActorID == nil || *command.ActorID != actorID ||
		!command.OccurredAt.Equal(occurredAt) || len(command.FieldDevices) != 2 {
		t.Fatalf("collaboration command: %+v", command)
	}
	firstState := command.FieldDevices[0]
	if firstState.ID != firstID || firstState.BMK == nil || *firstState.BMK != bmk ||
		firstState.Description == nil || *firstState.Description != description ||
		firstState.TextFix == nil || *firstState.TextFix != textFix ||
		firstState.ApparatNumber != 4 || firstState.SPSControllerSystemTypeID != firstParentID ||
		firstState.SystemPartID != first.SystemPartID ||
		firstState.SpecificationID == nil || *firstState.SpecificationID != specificationID ||
		firstState.ApparatID != first.ApparatID || !firstState.CreatedAt.Equal(createdAt) ||
		!firstState.UpdatedAt.Equal(updatedAt) {
		t.Fatalf("first collaboration state: %+v", firstState)
	}
	if command.FieldDevices[1].ID != thirdID {
		t.Fatalf("failed item leaked into collaboration command: %+v", command.FieldDevices)
	}
}

func TestMultiCreateForProjectRollsBackAndDoesNotDispatchOnWorkflowOrCommitFailure(t *testing.T) {
	workflowErr := errors.New("association failed")
	commitErr := errors.New("commit failed")
	for _, tc := range []struct {
		name        string
		workflowErr error
		commitErr   error
		wantErr     error
	}{
		{name: "workflow", workflowErr: workflowErr, wantErr: workflowErr},
		{name: "commit", commitErr: commitErr, wantErr: commitErr},
	} {
		t.Run(tc.name, func(t *testing.T) {
			projectID := fieldDeviceTestUUID(321)
			fieldDeviceID := fieldDeviceTestUUID(322)
			harness := &projectMultiCreateTransactionHarness{
				committed: projectMultiCreateTransactionState{
					fieldDevices: map[uuid.UUID]*domainFacility.FieldDevice{},
					projectLinks: map[uuid.UUID][]uuid.UUID{},
				},
				result: &domainFacility.FieldDeviceMultiCreateResult{
					Results: []domainFacility.FieldDeviceCreateResult{{
						Index:       0,
						Success:     true,
						FieldDevice: &domainFacility.FieldDevice{Base: domain.Base{ID: fieldDeviceID}},
					}},
					TotalRequests: 1,
					SuccessCount:  1,
				},
				workflowErr: tc.workflowErr,
				commitErr:   tc.commitErr,
			}
			dispatcher := &projectMultiCreateDispatcherStub{harness: harness, projectID: projectID}
			handler := NewMultiCreateForProjectHandler(MultiCreateForProjectDependencies{
				TransactionRunner:   harness.runner,
				TransactionWorkflow: harness.factory,
				Dispatcher:          dispatcher,
				NewID:               func() uuid.UUID { return fieldDeviceTestUUID(323) },
			})

			outcome, err := handler.Execute(context.Background(), MultiCreateForProjectCommand{
				ProjectID: projectID,
				Items:     []domainFacility.FieldDeviceCreateItem{{FieldDevice: &domainFacility.FieldDevice{}}},
			})
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("error: got %v, want %v", err, tc.wantErr)
			}
			if outcome.Result != nil || len(outcome.Mutation.Changes) != 0 ||
				len(harness.committed.fieldDevices) != 0 || len(harness.committed.projectLinks) != 0 ||
				len(harness.committed.history) != 0 || len(dispatcher.commands) != 0 {
				t.Fatalf("failed transaction leaked state/dispatch: outcome=%+v state=%+v commands=%v",
					outcome,
					harness.committed,
					dispatcher.commands,
				)
			}
		})
	}
}

func TestMultiCreateForProjectAllFailuresPreserveResultWithoutDispatch(t *testing.T) {
	projectID := fieldDeviceTestUUID(331)
	leakedID := fieldDeviceTestUUID(332)
	legacyResult := &domainFacility.FieldDeviceMultiCreateResult{
		Results: []domainFacility.FieldDeviceCreateResult{{
			Index:      0,
			Error:      "apparat number conflict",
			ErrorField: "fielddevice.apparat_nr",
		}},
		TotalRequests: 1,
		FailureCount:  1,
	}
	harness := &projectMultiCreateTransactionHarness{
		committed: projectMultiCreateTransactionState{
			fieldDevices: map[uuid.UUID]*domainFacility.FieldDevice{},
			projectLinks: map[uuid.UUID][]uuid.UUID{},
		},
		result:        legacyResult,
		failedAttempt: &domainFacility.FieldDevice{Base: domain.Base{ID: leakedID}},
	}
	dispatcher := &projectMultiCreateDispatcherStub{harness: harness, projectID: projectID}
	handler := NewMultiCreateForProjectHandler(MultiCreateForProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Dispatcher:          dispatcher,
		NewID:               func() uuid.UUID { return fieldDeviceTestUUID(333) },
	})

	outcome, err := handler.Execute(context.Background(), MultiCreateForProjectCommand{
		ProjectID: projectID,
		Items:     []domainFacility.FieldDeviceCreateItem{{FieldDevice: &domainFacility.FieldDevice{}}},
	})
	if err != nil {
		t.Fatalf("execute all-failure result: %v", err)
	}
	if outcome.Result != legacyResult || outcome.Result.FailureCount != 1 ||
		len(outcome.Mutation.Changes) != 0 || len(dispatcher.commands) != 0 {
		t.Fatalf("all-failure outcome changed: %+v commands=%v", outcome, dispatcher.commands)
	}
	if harness.committed.fieldDevices[leakedID] != nil || len(harness.committed.history) != 0 {
		t.Fatal("all-failure item state escaped its savepoint")
	}
}

func TestMultiCreateForProjectReportsDispatchFailureWithoutChangingCommittedResult(t *testing.T) {
	projectID := fieldDeviceTestUUID(351)
	fieldDeviceID := fieldDeviceTestUUID(352)
	dispatchErr := errors.New("realtime unavailable")
	legacyResult := &domainFacility.FieldDeviceMultiCreateResult{
		Results: []domainFacility.FieldDeviceCreateResult{{
			Index:   0,
			Success: true,
			FieldDevice: &domainFacility.FieldDevice{
				Base:                      domain.Base{ID: fieldDeviceID},
				SPSControllerSystemTypeID: fieldDeviceTestUUID(353),
			},
		}},
		TotalRequests: 1,
		SuccessCount:  1,
	}
	harness := &projectMultiCreateTransactionHarness{
		committed: projectMultiCreateTransactionState{
			fieldDevices: map[uuid.UUID]*domainFacility.FieldDevice{},
			projectLinks: map[uuid.UUID][]uuid.UUID{},
		},
		result: legacyResult,
	}
	dispatcher := &projectMultiCreateDispatcherStub{
		harness:   harness,
		projectID: projectID,
		wantIDs:   []uuid.UUID{fieldDeviceID},
		err:       dispatchErr,
	}
	var reported []error
	handler := NewMultiCreateForProjectHandler(MultiCreateForProjectDependencies{
		TransactionRunner:   harness.runner,
		TransactionWorkflow: harness.factory,
		Dispatcher:          dispatcher,
		NewID:               func() uuid.UUID { return uuid.New() },
		ReportError:         func(err error) { reported = append(reported, err) },
	})

	result, err := handler.MultiCreateForProject(
		context.Background(),
		MultiCreateForProjectCommand{ProjectID: projectID},
	)
	if err != nil || result != legacyResult {
		t.Fatalf("committed result changed by dispatch failure: result=%p want=%p err=%v",
			result,
			legacyResult,
			err,
		)
	}
	if harness.committed.fieldDevices[fieldDeviceID] == nil ||
		!reflect.DeepEqual(harness.committed.projectLinks[projectID], []uuid.UUID{fieldDeviceID}) ||
		len(reported) != 1 || !errors.Is(reported[0], dispatchErr) {
		t.Fatalf("commit/report state: committed=%+v reported=%v", harness.committed, reported)
	}
}

func TestMultiCreateForProjectRequiresConfiguredTransaction(t *testing.T) {
	handler := NewMultiCreateForProjectHandler(MultiCreateForProjectDependencies{})
	outcome, err := handler.Execute(context.Background(), MultiCreateForProjectCommand{
		ProjectID: fieldDeviceTestUUID(341),
	})
	if !errors.Is(err, ErrMultiCreateForProjectTransactionNotConfigured) ||
		outcome.Result != nil || len(outcome.Mutation.Changes) != 0 {
		t.Fatalf("configuration result: outcome=%+v err=%v", outcome, err)
	}
}
