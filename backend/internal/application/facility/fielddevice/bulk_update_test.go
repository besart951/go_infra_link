package fielddevice

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	"github.com/besart951/go_infra_link/backend/internal/application/facility/mutation"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainFieldDevice "github.com/besart951/go_infra_link/backend/internal/domain/facility/fielddevice"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	"github.com/google/uuid"
)

type bulkUpdaterStub struct {
	result   *domainFacility.BulkOperationResult
	items    []domainFieldDevice.BulkUpdateItemExecution
	returned bool
	onUpdate func()
	batchID  uuid.UUID
	hasBatch bool
}

type bulkHistoryBatchKey struct{}

func (s *bulkUpdaterStub) ExecuteBulkUpdate(
	ctx context.Context,
	updates []domainFacility.BulkFieldDeviceUpdate,
) domainFieldDevice.BulkUpdateExecution {
	if s.onUpdate != nil {
		s.onUpdate()
	}
	s.batchID, s.hasBatch = ctx.Value(bulkHistoryBatchKey{}).(uuid.UUID)
	s.returned = true
	items := s.items
	if items == nil {
		items = legacyTestBulkExecutionItems(updates, s.result)
	}
	return domainFieldDevice.BulkUpdateExecution{
		Result: s.result,
		Items:  items,
	}
}

type projectLinkReaderStub struct {
	updater    *bulkUpdaterStub
	links      []*domainProject.ProjectFieldDevice
	err        error
	calls      int
	received   []uuid.UUID
	contextErr error
}

func (s *projectLinkReaderStub) GetByFieldDeviceIDs(
	ctx context.Context,
	ids []uuid.UUID,
) ([]*domainProject.ProjectFieldDevice, error) {
	s.calls++
	s.received = append([]uuid.UUID(nil), ids...)
	s.contextErr = ctx.Err()
	if s.updater != nil && !s.updater.returned {
		return nil, errors.New("project scope resolved before bulk updater returned")
	}
	return s.links, s.err
}

type commandDispatcherStub struct {
	commands []appcollaboration.FacilityHierarchyRefreshRequired
	err      error
}

func (s *commandDispatcherStub) Dispatch(
	_ context.Context,
	command appcollaboration.Command,
) error {
	typed, ok := command.(appcollaboration.FacilityHierarchyRefreshRequired)
	if !ok {
		return errors.New("unexpected collaboration command")
	}
	s.commands = append(s.commands, typed)
	return s.err
}

func TestBulkUpdateDispatchesBatchedProjectScopedRefreshAfterOperation(t *testing.T) {
	deviceOne := testUUID(1)
	deviceTwo := testUUID(2)
	unrequestedDevice := testUUID(3)
	projectOne := testUUID(11)
	projectTwo := testUUID(12)
	actorID := testUUID(21)
	operationID := testUUID(31)
	eventOne := testUUID(32)
	eventTwo := testUUID(33)
	occurredAt := time.Date(2026, time.July, 20, 9, 30, 0, 0, time.UTC)

	bmk := "B-1"
	specificationSupplier := "Supplier"
	updater := &bulkUpdaterStub{result: &domainFacility.BulkOperationResult{
		Results: []domainFacility.BulkOperationResultItem{
			{ID: deviceOne, Success: true},
			{ID: deviceTwo, Success: true},
		},
		TotalCount:   2,
		SuccessCount: 2,
	}}
	links := &projectLinkReaderStub{
		updater: updater,
		links: []*domainProject.ProjectFieldDevice{
			{ProjectID: projectTwo, FieldDeviceID: deviceTwo},
			{ProjectID: projectOne, FieldDeviceID: deviceTwo},
			{ProjectID: projectOne, FieldDeviceID: deviceOne},
			{ProjectID: projectOne, FieldDeviceID: deviceOne},
			{ProjectID: projectOne, FieldDeviceID: unrequestedDevice},
		},
	}
	dispatcher := &commandDispatcherStub{}
	ids := []uuid.UUID{operationID, eventOne, eventTwo}
	handler := NewBulkUpdateHandler(BulkUpdateDependencies{
		Executor: updater,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, bulkHistoryBatchKey{}, batchID)
		},
		ProjectLinks: links,
		Dispatcher:   dispatcher,
		Actor:        func(context.Context) *uuid.UUID { return &actorID },
		NewID: func() uuid.UUID {
			id := ids[0]
			ids = ids[1:]
			return id
		},
		Now: func() time.Time { return occurredAt },
	})
	updates := []domainFacility.BulkFieldDeviceUpdate{
		{ID: deviceTwo, Specification: &domainFacility.SpecificationPatch{
			SpecificationSupplier:    &specificationSupplier,
			HasSpecificationSupplier: true,
		}},
		{ID: deviceOne, BMK: &bmk, HasBMK: true},
	}

	outcome := handler.Execute(context.Background(), updates)

	if outcome.Result != updater.result {
		t.Fatal("expected the legacy bulk result to be preserved")
	}
	if !updater.hasBatch || updater.batchID != operationID {
		t.Fatalf("expected history batch %s, got %s (present=%t)", operationID, updater.batchID, updater.hasBatch)
	}
	if links.calls != 1 {
		t.Fatalf("expected one batched project-link query, got %d", links.calls)
	}
	if want := []uuid.UUID{deviceOne, deviceTwo}; !reflect.DeepEqual(links.received, want) {
		t.Fatalf("project-link IDs: got %v, want %v", links.received, want)
	}
	if len(dispatcher.commands) != 2 {
		t.Fatalf("expected one command per linked project, got %d", len(dispatcher.commands))
	}

	first := dispatcher.commands[0]
	if first.ProjectID != projectOne {
		t.Fatalf("expected deterministic first project %s, got %s", projectOne, first.ProjectID)
	}
	if want := []uuid.UUID{deviceOne, deviceTwo}; !reflect.DeepEqual(first.EntityIDs, want) {
		t.Fatalf("first project IDs: got %v, want %v", first.EntityIDs, want)
	}
	second := dispatcher.commands[1]
	if second.ProjectID != projectTwo {
		t.Fatalf("expected second project %s, got %s", projectTwo, second.ProjectID)
	}
	if want := []uuid.UUID{deviceTwo}; !reflect.DeepEqual(second.EntityIDs, want) {
		t.Fatalf("second project IDs: got %v, want %v", second.EntityIDs, want)
	}

	for _, command := range dispatcher.commands {
		if command.OperationID != operationID || command.CorrelationID != operationID {
			t.Fatalf("expected shared operation/correlation %s, got %#v", operationID, command.Envelope)
		}
		if command.ActorID == nil || *command.ActorID != actorID {
			t.Fatalf("expected actor %s, got %v", actorID, command.ActorID)
		}
		if command.OccurredAt != occurredAt {
			t.Fatalf("expected occurred_at %s, got %s", occurredAt, command.OccurredAt)
		}
		if command.Scope != appcollaboration.FacilityScopeFieldDevice {
			t.Fatalf("unexpected scope %q", command.Scope)
		}
	}

	if want := []uuid.UUID{projectOne, projectTwo}; !reflect.DeepEqual(outcome.Mutation.ProjectIDs, want) {
		t.Fatalf("mutation project IDs: got %v, want %v", outcome.Mutation.ProjectIDs, want)
	}
	if len(outcome.Mutation.Changes) != 2 {
		t.Fatalf("expected two successful changes, got %d", len(outcome.Mutation.Changes))
	}
	if got := outcome.Mutation.Changes[0].ChangedFields; !reflect.DeepEqual(got, []mutation.FieldName{mutation.FieldNameBMK}) {
		t.Fatalf("first changed fields: got %v", got)
	}
	if got := outcome.Mutation.Changes[1].ChangedFields; !reflect.DeepEqual(got, []mutation.FieldName{mutation.FieldNameSpecification}) {
		t.Fatalf("second changed fields: got %v", got)
	}
}

func TestBulkUpdateMutationUsesExplicitSucceededPhasesForPartialItem(t *testing.T) {
	deviceID := testUUID(1)
	operationID := testUUID(31)
	bmk := "B-1"
	supplier := "Supplier"
	objects := []domainFacility.BacnetObjectPatch{}
	updater := &bulkUpdaterStub{
		result: &domainFacility.BulkOperationResult{
			Results: []domainFacility.BulkOperationResultItem{{
				ID:      deviceID,
				Success: false,
			}},
			TotalCount:   1,
			FailureCount: 1,
		},
		items: []domainFieldDevice.BulkUpdateItemExecution{{
			Index: 0,
			ID:    deviceID,
			Phases: []domainFieldDevice.BulkUpdatePhaseResult{
				{
					Phase:  domainFieldDevice.BulkUpdatePhaseFieldDevice,
					Status: domainFieldDevice.BulkUpdatePhaseSucceeded,
				},
				{
					Phase:  domainFieldDevice.BulkUpdatePhaseSpecification,
					Status: domainFieldDevice.BulkUpdatePhaseFailed,
				},
				{
					Phase:  domainFieldDevice.BulkUpdatePhaseBacnetObjects,
					Status: domainFieldDevice.BulkUpdatePhaseSucceeded,
				},
			},
		}},
	}
	handler := NewBulkUpdateHandler(BulkUpdateDependencies{
		Executor: updater,
		NewID:    func() uuid.UUID { return operationID },
	})

	outcome := handler.Execute(context.Background(), []domainFacility.BulkFieldDeviceUpdate{{
		ID:  deviceID,
		BMK: &bmk,
		Specification: &domainFacility.SpecificationPatch{
			SpecificationSupplier:    &supplier,
			HasSpecificationSupplier: true,
		},
		BacnetObjects: &objects,
	}})

	if len(outcome.Mutation.Changes) != 1 {
		t.Fatalf("expected one partially committed change, got %+v", outcome.Mutation.Changes)
	}
	got := outcome.Mutation.Changes[0].ChangedFields
	want := []mutation.FieldName{
		mutation.FieldNameBacnetObjects,
		mutation.FieldNameBMK,
	}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("changed fields: got %v, want %v", got, want)
	}
}

func TestBulkUpdateUsesUncancelledContextForPostOperationResolution(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	deviceID := testUUID(1)
	projectID := testUUID(11)
	bmk := "B-1"
	updater := &bulkUpdaterStub{
		result: &domainFacility.BulkOperationResult{
			Results:      []domainFacility.BulkOperationResultItem{{ID: deviceID, Success: true}},
			TotalCount:   1,
			SuccessCount: 1,
		},
		onUpdate: cancel,
	}
	links := &projectLinkReaderStub{
		updater: updater,
		links: []*domainProject.ProjectFieldDevice{
			{ProjectID: projectID, FieldDeviceID: deviceID},
		},
	}
	dispatcher := &commandDispatcherStub{}
	handler := NewBulkUpdateHandler(BulkUpdateDependencies{
		Executor:     updater,
		ProjectLinks: links,
		Dispatcher:   dispatcher,
	})

	handler.Execute(ctx, []domainFacility.BulkFieldDeviceUpdate{
		{ID: deviceID, BMK: &bmk, HasBMK: true},
	})

	if !errors.Is(ctx.Err(), context.Canceled) {
		t.Fatal("expected request context cancellation")
	}
	if links.contextErr != nil {
		t.Fatalf("expected post-operation context without cancellation, got %v", links.contextErr)
	}
	if len(dispatcher.commands) != 1 {
		t.Fatalf("expected dispatch despite request cancellation, got %d commands", len(dispatcher.commands))
	}
}

func TestBulkUpdateFallsBackToFullRefreshAboveVersionOneIDLimit(t *testing.T) {
	projectID := testUUID(20)
	devices := []uuid.UUID{testUUID(1), testUUID(2), testUUID(3)}
	updates := make([]domainFacility.BulkFieldDeviceUpdate, 0, len(devices))
	resultItems := make([]domainFacility.BulkOperationResultItem, 0, len(devices))
	linksForProject := make([]*domainProject.ProjectFieldDevice, 0, len(devices))
	for _, deviceID := range devices {
		number := 4
		updates = append(updates, domainFacility.BulkFieldDeviceUpdate{
			ID:        deviceID,
			ApparatNr: &number,
		})
		resultItems = append(resultItems, domainFacility.BulkOperationResultItem{
			ID:      deviceID,
			Success: true,
		})
		linksForProject = append(linksForProject, &domainProject.ProjectFieldDevice{
			ProjectID:     projectID,
			FieldDeviceID: deviceID,
		})
	}

	dispatcher := &commandDispatcherStub{}
	handler := NewBulkUpdateHandler(BulkUpdateDependencies{
		Executor: &bulkUpdaterStub{result: &domainFacility.BulkOperationResult{
			Results:      resultItems,
			TotalCount:   len(devices),
			SuccessCount: len(devices),
		}},
		ProjectLinks:          &projectLinkReaderStub{links: linksForProject},
		Dispatcher:            dispatcher,
		MaxTargetedRefreshIDs: 2,
	})

	handler.Execute(context.Background(), updates)

	if len(dispatcher.commands) != 1 {
		t.Fatalf("expected one project command, got %d", len(dispatcher.commands))
	}
	command := dispatcher.commands[0]
	if !command.FullRefresh {
		t.Fatal("expected full refresh fallback")
	}
	if len(command.EntityIDs) != 0 {
		t.Fatalf("expected no IDs for full refresh, got %v", command.EntityIDs)
	}
}

func TestBulkUpdatePreservesResultWhenProjectResolutionFails(t *testing.T) {
	deviceID := testUUID(1)
	bmk := "B-1"
	expected := &domainFacility.BulkOperationResult{
		Results:      []domainFacility.BulkOperationResultItem{{ID: deviceID, Success: true}},
		TotalCount:   1,
		SuccessCount: 1,
	}
	resolveErr := errors.New("links unavailable")
	handler := NewBulkUpdateHandler(BulkUpdateDependencies{
		Executor:     &bulkUpdaterStub{result: expected},
		ProjectLinks: &projectLinkReaderStub{err: resolveErr},
		Dispatcher:   &commandDispatcherStub{},
	})

	outcome := handler.Execute(context.Background(), []domainFacility.BulkFieldDeviceUpdate{
		{ID: deviceID, BMK: &bmk, HasBMK: true},
	})

	if outcome.Result != expected {
		t.Fatal("post-operation publication failure must not replace a committed HTTP result")
	}
	if len(outcome.DispatchErrors) != 1 || !errors.Is(outcome.DispatchErrors[0], resolveErr) {
		t.Fatalf("expected wrapped resolution error, got %v", outcome.DispatchErrors)
	}
}

func TestBulkUpdateSkipsScopeResolutionForItemsWithoutMutationPhases(t *testing.T) {
	links := &projectLinkReaderStub{}
	handler := NewBulkUpdateHandler(BulkUpdateDependencies{
		Executor:     &bulkUpdaterStub{result: &domainFacility.BulkOperationResult{}},
		ProjectLinks: links,
		Dispatcher:   &commandDispatcherStub{},
	})

	outcome := handler.Execute(context.Background(), []domainFacility.BulkFieldDeviceUpdate{
		{ID: testUUID(1)},
	})

	if links.calls != 0 {
		t.Fatalf("expected no project query, got %d", links.calls)
	}
	if len(outcome.ReconciliationIDs) != 0 {
		t.Fatalf("expected no reconciliation IDs, got %v", outcome.ReconciliationIDs)
	}
}

func legacyTestBulkExecutionItems(
	updates []domainFacility.BulkFieldDeviceUpdate,
	result *domainFacility.BulkOperationResult,
) []domainFieldDevice.BulkUpdateItemExecution {
	if result == nil {
		return nil
	}

	successByID := make(map[uuid.UUID]bool, len(result.Results))
	for _, item := range result.Results {
		successByID[item.ID] = item.Success
	}

	executions := make([]domainFieldDevice.BulkUpdateItemExecution, 0, len(updates))
	for i, update := range updates {
		status := domainFieldDevice.BulkUpdatePhaseFailed
		if successByID[update.ID] {
			status = domainFieldDevice.BulkUpdatePhaseSucceeded
		}
		phases := make([]domainFieldDevice.BulkUpdatePhaseResult, 0, 3)
		for _, phase := range []domainFieldDevice.BulkUpdatePhase{
			domainFieldDevice.BulkUpdatePhaseFieldDevice,
			domainFieldDevice.BulkUpdatePhaseSpecification,
			domainFieldDevice.BulkUpdatePhaseBacnetObjects,
		} {
			if len(requestedFieldsForPhase(update, phase)) == 0 {
				continue
			}
			phases = append(phases, domainFieldDevice.BulkUpdatePhaseResult{
				Phase:  phase,
				Status: status,
			})
		}
		executions = append(executions, domainFieldDevice.BulkUpdateItemExecution{
			Index:  i,
			ID:     update.ID,
			Phases: phases,
		})
	}
	return executions
}

func testUUID(value int) uuid.UUID {
	return uuid.MustParse("00000000-0000-0000-0000-" + leftPad12(value))
}

func leftPad12(value int) string {
	const digits = "000000000000"
	raw := fmtInt(value)
	return digits[:len(digits)-len(raw)] + raw
}

func fmtInt(value int) string {
	if value == 0 {
		return "0"
	}
	buf := make([]byte, 0, 12)
	for value > 0 {
		buf = append(buf, byte('0'+value%10))
		value /= 10
	}
	for i, j := 0, len(buf)-1; i < j; i, j = i+1, j-1 {
		buf[i], buf[j] = buf[j], buf[i]
	}
	return string(buf)
}
