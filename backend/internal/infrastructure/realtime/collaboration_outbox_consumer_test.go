package realtime

import (
	"context"
	"testing"
	"time"

	appcollaboration "github.com/besart951/go_infra_link/backend/internal/application/collaboration"
	"github.com/google/uuid"
)

func TestCollaborationOutboxConsumerEmitsVersionTwoCommittedEvent(t *testing.T) {
	hub := NewProjectCollaborationHub()
	defer hub.Close()
	projectID := uuid.New()
	fieldDeviceID := uuid.New()
	eventID := uuid.New()
	operationID := uuid.New()
	client := registerProjectTestClient(hub, projectID, uuid.New(), 8)
	drainSocket(client.socket)
	command := appcollaboration.FieldDeviceUpdated{
		Envelope: appcollaboration.Envelope{
			SchemaVersion: appcollaboration.SchemaVersionV2,
			EventID:       eventID, OperationID: operationID, CorrelationID: operationID,
			ProjectID: projectID, OccurredAt: time.Date(2026, 7, 23, 12, 0, 0, 0, time.UTC),
		},
		FieldDeviceID: fieldDeviceID,
	}
	event, err := appcollaboration.NewOutboxEvent(command)
	if err != nil {
		t.Fatalf("new event: %v", err)
	}
	event.Sequence = 7

	if err := NewCollaborationOutboxConsumer(hub).Deliver(context.Background(), *event); err != nil {
		t.Fatalf("deliver: %v", err)
	}

	message := receiveSocketMessageOfType(t, client.socket, projectCollaborationMessageCommittedEvent)
	if message["schema_version"] != float64(2) || message["event_id"] != eventID.String() ||
		message["operation_id"] != operationID.String() || message["sequence"] != float64(7) ||
		message["scope"] != projectCollaborationRefreshScopeFieldDevice {
		t.Fatalf("unexpected committed event: %#v", message)
	}
	entityIDs, ok := message["entity_ids"].([]any)
	if !ok || len(entityIDs) != 1 || entityIDs[0] != fieldDeviceID.String() {
		t.Fatalf("entity_ids = %#v", message["entity_ids"])
	}
}

func TestCollaborationOutboxConsumerRejectsMismatchedPersistedEnvelope(t *testing.T) {
	hub := NewProjectCollaborationHub()
	defer hub.Close()
	command := appcollaboration.FieldDeviceUpdated{
		Envelope: appcollaboration.Envelope{
			SchemaVersion: appcollaboration.SchemaVersionV2,
			EventID:       uuid.New(), OperationID: uuid.New(), CorrelationID: uuid.New(),
			ProjectID: uuid.New(), OccurredAt: time.Now().UTC(),
		},
		FieldDeviceID: uuid.New(),
	}
	event, err := appcollaboration.NewOutboxEvent(command)
	if err != nil {
		t.Fatalf("new event: %v", err)
	}
	event.ProjectID = uuid.New()
	event.Sequence = 1
	if err := NewCollaborationOutboxConsumer(hub).Deliver(context.Background(), *event); err == nil {
		t.Fatal("expected persisted-envelope mismatch")
	}
}
