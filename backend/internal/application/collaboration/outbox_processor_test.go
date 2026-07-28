package collaboration

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	domainCollaboration "github.com/besart951/go_infra_link/backend/internal/domain/collaboration"
	"github.com/google/uuid"
)

type outboxStoreSpy struct {
	events          []domainCollaboration.OutboxEvent
	processed       bool
	delivered       int
	failed          int
	deliveredEvents []uuid.UUID
	failedEvents    []uuid.UUID
}

func (s *outboxStoreSpy) Enqueue(context.Context, *domainCollaboration.OutboxEvent) error { return nil }
func (s *outboxStoreSpy) ClaimDue(context.Context, time.Time, int) ([]domainCollaboration.OutboxEvent, error) {
	return s.events, nil
}
func (s *outboxStoreSpy) WasProcessed(context.Context, string, uuid.UUID) (bool, error) {
	return s.processed, nil
}
func (s *outboxStoreSpy) MarkDelivered(_ context.Context, _ string, event domainCollaboration.OutboxEvent, _ time.Time) error {
	s.delivered++
	s.deliveredEvents = append(s.deliveredEvents, event.EventID)
	return nil
}
func (s *outboxStoreSpy) MarkFailed(_ context.Context, event domainCollaboration.OutboxEvent, _ string, _, _ time.Time) error {
	s.failed++
	s.failedEvents = append(s.failedEvents, event.EventID)
	return nil
}

type outboxConsumerSpy struct {
	deliveries      int
	err             error
	errorsByEvent   map[uuid.UUID]error
	deliveredEvents []uuid.UUID
}

func (s *outboxConsumerSpy) ConsumerID() string { return "websocket-v2" }
func (s *outboxConsumerSpy) Deliver(_ context.Context, event domainCollaboration.OutboxEvent) error {
	s.deliveries++
	s.deliveredEvents = append(s.deliveredEvents, event.EventID)
	if err := s.errorsByEvent[event.EventID]; err != nil {
		return err
	}
	return s.err
}

func TestOutboxProcessorSkipsAnAlreadyProcessedEvent(t *testing.T) {
	store := &outboxStoreSpy{processed: true, events: []domainCollaboration.OutboxEvent{{EventID: uuid.New(), Attempts: 1}}}
	consumer := &outboxConsumerSpy{}
	processor, err := NewOutboxProcessor(store, consumer, func(int) time.Duration { return time.Second })
	if err != nil {
		t.Fatalf("new processor: %v", err)
	}
	processed, err := processor.RunOnce(context.Background(), time.Now(), 100)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if processed != 1 || consumer.deliveries != 0 || store.delivered != 1 || store.failed != 0 {
		t.Fatalf("unexpected idempotent delivery: processed=%d deliveries=%d delivered=%d failed=%d", processed, consumer.deliveries, store.delivered, store.failed)
	}
}

func TestOutboxProcessorBlocksOnlyTheFailedProjectForTheCurrentPass(t *testing.T) {
	projectOne := uuid.New()
	projectTwo := uuid.New()
	first := domainCollaboration.OutboxEvent{
		EventID: uuid.New(), ProjectID: projectOne, Attempts: 1,
	}
	blocked := domainCollaboration.OutboxEvent{
		EventID: uuid.New(), ProjectID: projectOne, Attempts: 1,
	}
	independent := domainCollaboration.OutboxEvent{
		EventID: uuid.New(), ProjectID: projectTwo, Attempts: 1,
	}
	store := &outboxStoreSpy{events: []domainCollaboration.OutboxEvent{
		first, blocked, independent,
	}}
	consumer := &outboxConsumerSpy{
		errorsByEvent: map[uuid.UUID]error{first.EventID: errors.New("bus unavailable")},
	}
	processor, err := NewOutboxProcessor(store, consumer, func(int) time.Duration { return time.Second })
	if err != nil {
		t.Fatalf("new processor: %v", err)
	}

	claimed, err := processor.RunOnce(context.Background(), time.Now(), 100)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if claimed != 3 || consumer.deliveries != 2 ||
		len(consumer.deliveredEvents) != 2 ||
		consumer.deliveredEvents[0] != first.EventID ||
		consumer.deliveredEvents[1] != independent.EventID {
		t.Fatalf("unexpected deliveries: %+v", consumer.deliveredEvents)
	}
	if store.failed != 1 || store.failedEvents[0] != first.EventID ||
		store.delivered != 1 || store.deliveredEvents[0] != independent.EventID {
		t.Fatalf(
			"unexpected persistence: failed=%v delivered=%v",
			store.failedEvents,
			store.deliveredEvents,
		)
	}
}

func TestCommandCodecRoundTripsTypedCommand(t *testing.T) {
	operationID := uuid.New()
	original := FieldDevicesCreated{
		Envelope: Envelope{
			EventID: uuid.New(), OperationID: operationID, CorrelationID: operationID,
			ProjectID: uuid.New(), SchemaVersion: SchemaVersionV2, OccurredAt: time.Now().UTC(),
		},
		FieldDevices: []FieldDeviceState{{ID: uuid.New(), ApparatNumber: 4}},
	}
	encoded, err := EncodeCommand(original)
	if err != nil {
		t.Fatalf("encode: %v", err)
	}
	if !json.Valid(encoded.Payload) || encoded.Type != "field_devices_created" {
		t.Fatalf("unexpected encoded command: %#v", encoded)
	}
	decoded, err := DecodeCommand(encoded)
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	actual, ok := decoded.(FieldDevicesCreated)
	if !ok || actual.EventID != original.EventID || len(actual.FieldDevices) != 1 || actual.FieldDevices[0].ID != original.FieldDevices[0].ID {
		t.Fatalf("unexpected decoded command: %#v", decoded)
	}
	outboxEvent, err := NewOutboxEvent(original)
	if err != nil {
		t.Fatalf("new outbox event: %v", err)
	}
	if outboxEvent.EventID != original.EventID || outboxEvent.OperationID != original.OperationID || outboxEvent.ProjectID != original.ProjectID || outboxEvent.EventType != encoded.Type {
		t.Fatalf("unexpected outbox event: %#v", outboxEvent)
	}
}

func TestNewOutboxEventRejectsIncompleteOrLegacyEnvelope(t *testing.T) {
	valid := Envelope{
		SchemaVersion: SchemaVersionV2,
		EventID:       uuid.New(),
		OperationID:   uuid.New(),
		CorrelationID: uuid.New(),
		ProjectID:     uuid.New(),
		OccurredAt:    time.Now().UTC(),
	}
	tests := []struct {
		name     string
		envelope Envelope
	}{
		{name: "missing event ID", envelope: func() Envelope {
			value := valid
			value.EventID = uuid.Nil
			return value
		}()},
		{name: "missing operation ID", envelope: func() Envelope {
			value := valid
			value.OperationID = uuid.Nil
			return value
		}()},
		{name: "missing correlation ID", envelope: func() Envelope {
			value := valid
			value.CorrelationID = uuid.Nil
			return value
		}()},
		{name: "missing project ID", envelope: func() Envelope {
			value := valid
			value.ProjectID = uuid.Nil
			return value
		}()},
		{name: "missing occurred at", envelope: func() Envelope {
			value := valid
			value.OccurredAt = time.Time{}
			return value
		}()},
		{name: "legacy schema", envelope: func() Envelope {
			value := valid
			value.SchemaVersion = SchemaVersionV1
			return value
		}()},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			_, err := NewOutboxEvent(FieldDeviceUpdated{
				Envelope:      test.envelope,
				FieldDeviceID: uuid.New(),
			})
			if err == nil {
				t.Fatal("expected incomplete envelope to be rejected")
			}
		})
	}
}
