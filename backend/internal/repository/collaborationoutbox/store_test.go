package collaborationoutbox

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	domainCollaboration "github.com/besart951/go_infra_link/backend/internal/domain/collaboration"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestStoreClaimsRetriesAndRecordsIdempotentDelivery(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	store := NewStore(db)
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	event := &domainCollaboration.OutboxEvent{
		EventType:     "field_device_updated",
		SchemaVersion: 2,
		OperationID:   uuid.New(),
		ProjectID:     uuid.New(),
		Payload:       json.RawMessage(`{"id":"event"}`),
		NextAttemptAt: now,
	}
	if err := store.Enqueue(context.Background(), event); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	if event.Sequence != 1 {
		t.Fatalf("expected first project sequence 1, got %d", event.Sequence)
	}
	second := &domainCollaboration.OutboxEvent{
		EventType: "field_device_updated", SchemaVersion: 2,
		OperationID: uuid.New(), ProjectID: event.ProjectID,
		Payload: json.RawMessage(`{"id":"second"}`), NextAttemptAt: now,
	}
	if err := store.Enqueue(context.Background(), second); err != nil {
		t.Fatalf("enqueue second: %v", err)
	}
	if second.Sequence != 2 {
		t.Fatalf("expected second project sequence 2, got %d", second.Sequence)
	}

	claimed, err := store.ClaimDue(context.Background(), now, 100)
	if err != nil {
		t.Fatalf("first claim: %v", err)
	}
	if len(claimed) != 1 || claimed[0].Attempts != 1 || claimed[0].Status != domainCollaboration.OutboxStatusDelivering {
		t.Fatalf("unexpected first claim: %#v", claimed)
	}
	if claimed[0].Sequence != 1 {
		t.Fatalf("events were not claimed in project sequence order: %#v", claimed)
	}
	if err := store.MarkFailed(context.Background(), claimed[0], "bus unavailable", now, now.Add(time.Second)); err != nil {
		t.Fatalf("record failure: %v", err)
	}

	claimed, err = store.ClaimDue(context.Background(), now.Add(time.Second), 100)
	if err != nil {
		t.Fatalf("second claim: %v", err)
	}
	if len(claimed) != 1 || claimed[0].Attempts != 2 {
		t.Fatalf("unexpected retry claim: %#v", claimed)
	}
	if err := store.MarkDelivered(context.Background(), "websocket-v2", claimed[0], now.Add(2*time.Second)); err != nil {
		t.Fatalf("record delivery: %v", err)
	}
	claimed, err = store.ClaimDue(context.Background(), now.Add(2*time.Second), 100)
	if err != nil || len(claimed) != 1 || claimed[0].ID != second.ID {
		t.Fatalf("second event was not released after first delivery: %#v, %v", claimed, err)
	}
	if err := store.MarkDelivered(context.Background(), "websocket-v2", claimed[0], now.Add(2*time.Second)); err != nil {
		t.Fatalf("record second delivery: %v", err)
	}
	processed, err := store.WasProcessed(context.Background(), "websocket-v2", event.EventID)
	if err != nil {
		t.Fatalf("check idempotency: %v", err)
	}
	if !processed {
		t.Fatal("expected processed event record")
	}
	if claimed, err := store.ClaimDue(context.Background(), now.Add(time.Hour), 100); err != nil || len(claimed) != 0 {
		t.Fatalf("delivered event was claimed again: %#v, %v", claimed, err)
	}
}

func TestStoreReclaimsExpiredDeliveryLease(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	store := NewStore(db)
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	event := &domainCollaboration.OutboxEvent{
		EventType: "field_device_updated", SchemaVersion: 2,
		OperationID: uuid.New(), ProjectID: uuid.New(),
		Payload: json.RawMessage(`{"id":"leased"}`), NextAttemptAt: now,
	}
	if err := store.Enqueue(context.Background(), event); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	first, err := store.ClaimDue(context.Background(), now, 1)
	if err != nil || len(first) != 1 {
		t.Fatalf("first claim: %#v, %v", first, err)
	}
	if early, err := store.ClaimDue(context.Background(), now.Add(29*time.Second), 1); err != nil || len(early) != 0 {
		t.Fatalf("active lease was reclaimed: %#v, %v", early, err)
	}
	reclaimed, err := store.ClaimDue(context.Background(), now.Add(30*time.Second), 1)
	if err != nil || len(reclaimed) != 1 || reclaimed[0].Attempts != 2 {
		t.Fatalf("expired lease was not reclaimed: %#v, %v", reclaimed, err)
	}
}

func TestStaleLeaseCompletionCannotOverwriteAReclaimedAttempt(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	store := NewStore(db)
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	event := &domainCollaboration.OutboxEvent{
		EventType: "field_device_updated", SchemaVersion: 2,
		OperationID: uuid.New(), ProjectID: uuid.New(),
		Payload: json.RawMessage(`{"id":"leased"}`), NextAttemptAt: now,
	}
	if err := store.Enqueue(context.Background(), event); err != nil {
		t.Fatalf("enqueue: %v", err)
	}
	first, err := store.ClaimDue(context.Background(), now, 1)
	if err != nil || len(first) != 1 {
		t.Fatalf("first claim: %#v, %v", first, err)
	}
	second, err := store.ClaimDue(context.Background(), now.Add(30*time.Second), 1)
	if err != nil || len(second) != 1 || second[0].Attempts != 2 {
		t.Fatalf("reclaimed claim: %#v, %v", second, err)
	}

	if err := store.MarkFailed(
		context.Background(),
		first[0],
		"late failure",
		now.Add(31*time.Second),
		now.Add(32*time.Second),
	); err == nil {
		t.Fatal("expected stale failure completion to be rejected")
	}
	if err := store.MarkDelivered(
		context.Background(),
		"websocket-v2",
		first[0],
		now.Add(31*time.Second),
	); err == nil {
		t.Fatal("expected stale delivery completion to be rejected")
	}

	var current domainCollaboration.OutboxEvent
	if err := db.First(&current, "id = ?", event.ID).Error; err != nil {
		t.Fatalf("load event: %v", err)
	}
	if current.Status != domainCollaboration.OutboxStatusDelivering || current.Attempts != 2 {
		t.Fatalf("stale completion changed current claim: %+v", current)
	}
	processed, err := store.WasProcessed(context.Background(), "websocket-v2", event.EventID)
	if err != nil {
		t.Fatalf("check idempotency: %v", err)
	}
	if processed {
		t.Fatal("stale delivery recorded the event as processed")
	}
}

func TestTerminalFailureReleasesTheNextProjectSequence(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	store := NewStore(db)
	now := time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC)
	projectID := uuid.New()
	first := &domainCollaboration.OutboxEvent{
		EventType: "field_device_updated", SchemaVersion: 2,
		OperationID: uuid.New(), ProjectID: projectID,
		Payload: json.RawMessage(`{"id":"first"}`), NextAttemptAt: now,
	}
	second := &domainCollaboration.OutboxEvent{
		EventType: "field_device_updated", SchemaVersion: 2,
		OperationID: uuid.New(), ProjectID: projectID,
		Payload: json.RawMessage(`{"id":"second"}`), NextAttemptAt: now,
	}
	if err := store.Enqueue(context.Background(), first); err != nil {
		t.Fatalf("enqueue first: %v", err)
	}
	if err := store.Enqueue(context.Background(), second); err != nil {
		t.Fatalf("enqueue second: %v", err)
	}

	for attempt := 1; attempt <= domainCollaboration.MaxOutboxAttempts; attempt++ {
		claimed, err := store.ClaimDue(context.Background(), now, 1)
		if err != nil || len(claimed) != 1 || claimed[0].ID != first.ID {
			t.Fatalf("claim attempt %d: %#v, %v", attempt, claimed, err)
		}
		if err := store.MarkFailed(
			context.Background(),
			claimed[0],
			"permanent failure",
			now,
			now,
		); err != nil {
			t.Fatalf("mark attempt %d failed: %v", attempt, err)
		}
	}

	var terminal domainCollaboration.OutboxEvent
	if err := db.First(&terminal, "id = ?", first.ID).Error; err != nil {
		t.Fatalf("load terminal event: %v", err)
	}
	if terminal.Status != domainCollaboration.OutboxStatusFailed ||
		terminal.Attempts != domainCollaboration.MaxOutboxAttempts {
		t.Fatalf("terminal event: %+v", terminal)
	}
	claimed, err := store.ClaimDue(context.Background(), now, 1)
	if err != nil || len(claimed) != 1 || claimed[0].ID != second.ID {
		t.Fatalf("next sequence was not released: %#v, %v", claimed, err)
	}
}

func TestEnqueueRollsBackEventAndProjectSequenceWithOuterTransaction(t *testing.T) {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	if err := AutoMigrate(db); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	projectID := uuid.New()
	rollback := errors.New("rollback")
	err = db.Transaction(func(tx *gorm.DB) error {
		event := &domainCollaboration.OutboxEvent{
			EventType: "field_device_updated", SchemaVersion: 2,
			OperationID: uuid.New(), ProjectID: projectID,
			Payload: json.RawMessage(`{"id":"rollback"}`),
		}
		if err := NewStore(tx).Enqueue(context.Background(), event); err != nil {
			return err
		}
		return rollback
	})
	if !errors.Is(err, rollback) {
		t.Fatalf("transaction error = %v", err)
	}
	var eventCount, streamCount int64
	if err := db.Model(&domainCollaboration.OutboxEvent{}).Count(&eventCount).Error; err != nil {
		t.Fatalf("count events: %v", err)
	}
	if err := db.Model(&domainCollaboration.ProjectStream{}).Count(&streamCount).Error; err != nil {
		t.Fatalf("count streams: %v", err)
	}
	if eventCount != 0 || streamCount != 0 {
		t.Fatalf("rollback left events=%d streams=%d", eventCount, streamCount)
	}
}
