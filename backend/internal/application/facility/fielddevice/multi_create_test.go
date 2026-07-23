package fielddevice

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainFacility "github.com/besart951/go_infra_link/backend/internal/domain/facility"
	domainHistory "github.com/besart951/go_infra_link/backend/internal/domain/history"
	"github.com/google/uuid"
)

type multiCreateBatchKey struct{}

func fieldDeviceTestUUID(value int) uuid.UUID {
	var id uuid.UUID
	id[14] = byte(value >> 8)
	id[15] = byte(value)
	return id
}

type multiCreateExecutorStub struct {
	result  *domainFacility.FieldDeviceMultiCreateResult
	items   []domainFacility.FieldDeviceCreateItem
	batchID *uuid.UUID
	calls   int
}

func (s *multiCreateExecutorStub) MultiCreate(
	ctx context.Context,
	items []domainFacility.FieldDeviceCreateItem,
) *domainFacility.FieldDeviceMultiCreateResult {
	s.calls++
	s.items = items
	if batchID, ok := ctx.Value(multiCreateBatchKey{}).(uuid.UUID); ok {
		clone := batchID
		s.batchID = &clone
	}
	return s.result
}

func TestMultiCreatePreservesPartialResultAndCorrelatesSuccessfulHistory(t *testing.T) {
	operationID := fieldDeviceTestUUID(201)
	actorID := fieldDeviceTestUUID(202)
	firstID := fieldDeviceTestUUID(203)
	thirdID := fieldDeviceTestUUID(204)
	firstParentID := fieldDeviceTestUUID(205)
	thirdParentID := fieldDeviceTestUUID(206)
	objectDataID := fieldDeviceTestUUID(207)
	createdAt := time.Date(2026, time.July, 20, 21, 0, 0, 0, time.UTC)
	occurredAt := createdAt.Add(time.Minute)
	firstBMK := "BMK-1"
	thirdDescription := "created explicitly"
	first := &domainFacility.FieldDevice{
		Base:                      domain.Base{ID: firstID, CreatedAt: createdAt, UpdatedAt: createdAt},
		BMK:                       &firstBMK,
		ApparatNr:                 1,
		SPSControllerSystemTypeID: firstParentID,
		SystemPartID:              fieldDeviceTestUUID(208),
		ApparatID:                 fieldDeviceTestUUID(209),
	}
	third := &domainFacility.FieldDevice{
		Base:                      domain.Base{ID: thirdID, CreatedAt: createdAt, UpdatedAt: createdAt},
		Description:               &thirdDescription,
		ApparatNr:                 3,
		SPSControllerSystemTypeID: thirdParentID,
		SystemPartID:              fieldDeviceTestUUID(210),
		ApparatID:                 fieldDeviceTestUUID(211),
	}
	legacyResult := &domainFacility.FieldDeviceMultiCreateResult{
		Results: []domainFacility.FieldDeviceCreateResult{
			{Index: 0, Success: true, FieldDevice: first},
			{Index: 1, Error: "apparatnummer ist bereits vergeben", ErrorField: "fielddevice.apparat_nr"},
			{Index: 2, Success: true, FieldDevice: third},
		},
		TotalRequests: 3,
		SuccessCount:  2,
		FailureCount:  1,
	}
	executor := &multiCreateExecutorStub{result: legacyResult}
	handler := NewMultiCreateHandler(MultiCreateDependencies{
		Executor: executor,
		HistoryBatch: func(ctx context.Context, batchID uuid.UUID) context.Context {
			return context.WithValue(ctx, multiCreateBatchKey{}, batchID)
		},
		Actor: func(context.Context) *uuid.UUID { return &actorID },
		NewID: func() uuid.UUID { return operationID },
		Now:   func() time.Time { return occurredAt },
	})
	items := []domainFacility.FieldDeviceCreateItem{
		{FieldDevice: &domainFacility.FieldDevice{ApparatNr: 1}, ObjectDataID: &objectDataID},
		{FieldDevice: &domainFacility.FieldDevice{ApparatNr: 1}},
		{
			FieldDevice: &domainFacility.FieldDevice{ApparatNr: 3},
			BacnetObjects: []domainFacility.BacnetObject{{
				TextFix:        "AI1",
				SoftwareType:   domainFacility.BacnetSoftwareTypeAI,
				SoftwareNumber: 1,
			}},
		},
	}

	outcome := handler.Execute(context.Background(), MultiCreateCommand{Items: items})

	if outcome.Result != legacyResult || executor.calls != 1 || len(executor.items) != len(items) {
		t.Fatalf("legacy result/executor: outcome=%p want=%p calls=%d items=%d",
			outcome.Result,
			legacyResult,
			executor.calls,
			len(executor.items),
		)
	}
	if executor.items[0].ObjectDataID == nil || *executor.items[0].ObjectDataID != objectDataID ||
		len(executor.items[2].BacnetObjects) != 1 || executor.items[2].BacnetObjects[0].TextFix != "AI1" {
		t.Fatalf("creation selection changed: %+v", executor.items)
	}
	if executor.batchID == nil || *executor.batchID != operationID ||
		outcome.Mutation.BatchID == nil || *outcome.Mutation.BatchID != operationID {
		t.Fatalf("history batch: executor=%v result=%v", executor.batchID, outcome.Mutation.BatchID)
	}
	if outcome.Mutation.OperationID != operationID ||
		outcome.Mutation.ActorID == nil || *outcome.Mutation.ActorID != actorID ||
		!outcome.Mutation.OccurredAt.Equal(occurredAt) || len(outcome.Mutation.ProjectIDs) != 0 {
		t.Fatalf("mutation envelope: %+v", outcome.Mutation)
	}
	if len(outcome.Mutation.Changes) != 2 {
		t.Fatalf("changes: got %d, want 2", len(outcome.Mutation.Changes))
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
			change.Action != domainHistory.ActionCreate || len(change.Before) != 0 {
			t.Fatalf("change %d: %+v", index, change)
		}
		var snapshot fieldDeviceSnapshot
		if err := json.Unmarshal(change.After, &snapshot); err != nil {
			t.Fatalf("decode change %d: %v", index, err)
		}
		if snapshot.ID != want.id || snapshot.SPSControllerSystemTypeID != want.parentID ||
			!snapshot.CreatedAt.Equal(createdAt) {
			t.Fatalf("snapshot %d: %+v", index, snapshot)
		}
	}
}

func TestMultiCreateDoesNotTurnPartialFailuresIntoRequestFailure(t *testing.T) {
	legacyResult := &domainFacility.FieldDeviceMultiCreateResult{
		Results: []domainFacility.FieldDeviceCreateResult{{
			Index:      0,
			Error:      "BACnet validation failed",
			ErrorField: "bacnet_objects",
		}},
		TotalRequests: 1,
		FailureCount:  1,
	}
	executor := &multiCreateExecutorStub{result: legacyResult}
	handler := NewMultiCreateHandler(MultiCreateDependencies{Executor: executor})

	result := handler.MultiCreate(context.Background(), MultiCreateCommand{
		Items: []domainFacility.FieldDeviceCreateItem{{FieldDevice: &domainFacility.FieldDevice{}}},
	})

	if result != legacyResult || result.FailureCount != 1 || result.SuccessCount != 0 ||
		result.Results[0].Error != "BACnet validation failed" ||
		result.Results[0].ErrorField != "bacnet_objects" {
		// FieldDeviceCreateResult has no phase/detail expansion; its exact
		// compatibility fields must pass through unchanged.
		t.Fatalf("partial result changed: %+v", result)
	}
}

func TestMultiCreateMissingExecutorReturnsIndexAlignedFailures(t *testing.T) {
	handler := NewMultiCreateHandler(MultiCreateDependencies{})
	result := handler.MultiCreate(context.Background(), MultiCreateCommand{
		Items: []domainFacility.FieldDeviceCreateItem{{}, {}},
	})

	if result.TotalRequests != 2 || result.SuccessCount != 0 || result.FailureCount != 2 ||
		len(result.Results) != 2 {
		t.Fatalf("configuration failure result: %+v", result)
	}
	for index, item := range result.Results {
		if item.Index != index || item.Success || item.Error == "" || item.ErrorField != "fielddevice" {
			t.Fatalf("configuration failure item %d: %+v", index, item)
		}
	}
}
