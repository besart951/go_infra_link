package realtime

import (
	"context"
	"errors"
	"testing"
	"time"

	apprealtime "github.com/besart951/go_infra_link/backend/internal/application/realtime"
)

func TestInMemoryBusPublishesToSubscribers(t *testing.T) {
	bus := NewInMemoryBus()
	defer bus.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	events, err := bus.Subscribe(ctx, apprealtime.TopicProjectCollaboration)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	published := apprealtime.NewEvent(apprealtime.TopicProjectCollaboration, "node-a", []byte(`{"ok":true}`))
	if err := bus.Publish(ctx, published); err != nil {
		t.Fatalf("publish: %v", err)
	}

	got := receiveBusEvent(t, events)
	if got.ID != published.ID {
		t.Fatalf("event id = %q, want %q", got.ID, published.ID)
	}
	if string(got.Payload) != `{"ok":true}` {
		t.Fatalf("payload = %s", got.Payload)
	}
}

func TestInMemoryBusReportsBackpressure(t *testing.T) {
	bus := NewInMemoryBus(InMemoryBusConfig{SubscriberBuffer: 1})
	defer bus.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	if _, err := bus.Subscribe(ctx, apprealtime.TopicSystemNotifications); err != nil {
		t.Fatalf("subscribe: %v", err)
	}

	event := apprealtime.NewEvent(apprealtime.TopicSystemNotifications, "node-a", []byte(`{"n":1}`))
	if err := bus.Publish(ctx, event); err != nil {
		t.Fatalf("first publish: %v", err)
	}
	if err := bus.Publish(ctx, event); !errors.Is(err, apprealtime.ErrBackpressure) {
		t.Fatalf("second publish err = %v, want ErrBackpressure", err)
	}
}

func TestInMemoryBusClosesSubscriptionOnContextCancel(t *testing.T) {
	bus := NewInMemoryBus()
	defer bus.Close()

	ctx, cancel := context.WithCancel(context.Background())
	events, err := bus.Subscribe(ctx, apprealtime.TopicProjectCollaboration)
	if err != nil {
		t.Fatalf("subscribe: %v", err)
	}
	cancel()

	select {
	case _, ok := <-events:
		if ok {
			t.Fatalf("expected subscription channel to close")
		}
	case <-time.After(time.Second):
		t.Fatalf("subscription channel did not close")
	}
}

func receiveBusEvent(t *testing.T, events <-chan apprealtime.Event) apprealtime.Event {
	t.Helper()
	select {
	case event := <-events:
		return event
	case <-time.After(time.Second):
		t.Fatalf("timed out waiting for bus event")
		return apprealtime.Event{}
	}
}
