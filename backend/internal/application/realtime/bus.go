package realtime

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
)

type Topic string

const (
	TopicProjectCollaboration Topic = "project_collaboration"
	TopicSystemNotifications  Topic = "system_notifications"
)

var (
	ErrBackpressure = errors.New("realtime bus backpressure")
	ErrClosed       = errors.New("realtime bus closed")
	ErrNoTopics     = errors.New("realtime bus subscription requires at least one topic")
)

type Event struct {
	ID          string          `json:"id"`
	Topic       Topic           `json:"topic"`
	Source      string          `json:"source"`
	Payload     json.RawMessage `json:"payload"`
	PublishedAt time.Time       `json:"published_at"`
}

type Bus interface {
	Publish(ctx context.Context, event Event) error
	Subscribe(ctx context.Context, topics ...Topic) (<-chan Event, error)
	Close() error
}

func NewEvent(topic Topic, source string, payload []byte) Event {
	return NormalizeEvent(Event{
		Topic:   topic,
		Source:  source,
		Payload: append(json.RawMessage(nil), payload...),
	})
}

func NormalizeEvent(event Event) Event {
	if event.ID == "" {
		event.ID = uuid.NewString()
	}
	if event.PublishedAt.IsZero() {
		event.PublishedAt = time.Now().UTC()
	} else {
		event.PublishedAt = event.PublishedAt.UTC()
	}
	if event.Payload != nil {
		event.Payload = append(json.RawMessage(nil), event.Payload...)
	}
	return event
}
