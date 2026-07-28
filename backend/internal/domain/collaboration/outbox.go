// Package collaboration contains persistence-neutral collaboration delivery
// records. They are deliberately separate from the websocket transport: a
// committed facility change is durable before any best-effort fanout begins.
package collaboration

import (
	"context"
	"encoding/json"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type OutboxStatus string

const (
	OutboxStatusPending    OutboxStatus = "pending"
	OutboxStatusDelivering OutboxStatus = "delivering"
	OutboxStatusDelivered  OutboxStatus = "delivered"
	OutboxStatusFailed     OutboxStatus = "failed"

	MaxOutboxAttempts = 5
)

// OutboxEvent is written in the same transaction as the business mutation.
// EventID is stable across retries and is the idempotency key consumed by
// delivery adapters and clients.
type OutboxEvent struct {
	domain.Base
	EventID       uuid.UUID       `gorm:"type:uuid;not null;uniqueIndex"`
	EventType     string          `gorm:"type:varchar(128);not null"`
	SchemaVersion uint16          `gorm:"not null"`
	OperationID   uuid.UUID       `gorm:"type:uuid;not null;index"`
	ProjectID     uuid.UUID       `gorm:"type:uuid;not null;index;uniqueIndex:idx_collaboration_outbox_project_sequence,priority:1"`
	Sequence      uint64          `gorm:"not null;uniqueIndex:idx_collaboration_outbox_project_sequence,priority:2"`
	Payload       json.RawMessage `gorm:"serializer:json;type:jsonb;not null"`
	Status        OutboxStatus    `gorm:"type:varchar(16);not null;index:idx_collaboration_outbox_due,priority:1"`
	Attempts      int             `gorm:"not null;default:0"`
	NextAttemptAt time.Time       `gorm:"not null;index:idx_collaboration_outbox_due,priority:2"`
	ClaimedUntil  *time.Time      `gorm:"index:idx_collaboration_outbox_claimed"`
	DeliveredAt   *time.Time      `gorm:"index"`
	LastError     string          `gorm:"type:text"`
}

func (OutboxEvent) TableName() string { return "collaboration_outbox_events" }

type ProjectStream struct {
	ProjectID    uuid.UUID `gorm:"type:uuid;primaryKey"`
	NextSequence uint64    `gorm:"not null"`
}

func (ProjectStream) TableName() string { return "collaboration_project_streams" }

type DeliveryAttempt struct {
	domain.Base
	OutboxEventID uuid.UUID  `gorm:"type:uuid;not null;uniqueIndex:idx_collaboration_outbox_attempt,priority:1"`
	Attempt       int        `gorm:"not null;uniqueIndex:idx_collaboration_outbox_attempt,priority:2"`
	StartedAt     time.Time  `gorm:"not null"`
	FinishedAt    *time.Time `gorm:"index"`
	Error         string     `gorm:"type:text"`
}

func (DeliveryAttempt) TableName() string { return "collaboration_delivery_attempts" }

// ProcessedEvent records successful handling by one named consumer. Keeping
// this separate permits independent websocket, webhook, or projection
// consumers without making an already handled event observable twice.
type ProcessedEvent struct {
	domain.Base
	ConsumerID  string    `gorm:"type:varchar(128);not null;uniqueIndex:idx_collaboration_processed_event,priority:1"`
	EventID     uuid.UUID `gorm:"type:uuid;not null;uniqueIndex:idx_collaboration_processed_event,priority:2"`
	ProcessedAt time.Time `gorm:"not null"`
}

func (ProcessedEvent) TableName() string { return "collaboration_processed_events" }

type OutboxStore interface {
	Enqueue(context.Context, *OutboxEvent) error
	ClaimDue(context.Context, time.Time, int) ([]OutboxEvent, error)
	WasProcessed(context.Context, string, uuid.UUID) (bool, error)
	MarkDelivered(context.Context, string, OutboxEvent, time.Time) error
	MarkFailed(context.Context, OutboxEvent, string, time.Time, time.Time) error
}

type outboxStoreContextKey struct{}

func WithOutboxStore(ctx context.Context, store OutboxStore) context.Context {
	if store == nil {
		return ctx
	}
	return context.WithValue(ctx, outboxStoreContextKey{}, store)
}

func OutboxStoreFromContext(ctx context.Context) (OutboxStore, bool) {
	if ctx == nil {
		return nil, false
	}
	store, ok := ctx.Value(outboxStoreContextKey{}).(OutboxStore)
	return store, ok && store != nil
}
