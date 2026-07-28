package collaborationoutbox

import (
	"context"
	"fmt"
	"time"

	domain "github.com/besart951/go_infra_link/backend/internal/domain/collaboration"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Store struct{ db *gorm.DB }

func NewStore(db *gorm.DB) *Store { return &Store{db: db} }

func AutoMigrate(db *gorm.DB) error {
	return db.AutoMigrate(
		&domain.ProjectStream{},
		&domain.OutboxEvent{},
		&domain.DeliveryAttempt{},
		&domain.ProcessedEvent{},
	)
}

func (s *Store) Enqueue(ctx context.Context, event *domain.OutboxEvent) error {
	if event == nil || event.EventType == "" || event.ProjectID == uuid.Nil || event.OperationID == uuid.Nil {
		return fmt.Errorf("invalid collaboration outbox event")
	}
	now := time.Now().UTC()
	if event.EventID == uuid.Nil {
		event.EventID = uuid.Must(uuid.NewV7())
	}
	if event.Status == "" {
		event.Status = domain.OutboxStatusPending
	}
	if event.NextAttemptAt.IsZero() {
		event.NextAttemptAt = now
	}
	if err := event.InitForCreate(now); err != nil {
		return fmt.Errorf("initialize collaboration outbox event: %w", err)
	}
	if event.Sequence != 0 {
		return s.db.WithContext(ctx).Create(event).Error
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		sequence, err := allocateProjectSequence(tx, event.ProjectID)
		if err != nil {
			return err
		}
		event.Sequence = sequence
		return tx.Create(event).Error
	})
}

func (s *Store) ClaimDue(ctx context.Context, now time.Time, limit int) ([]domain.OutboxEvent, error) {
	if limit <= 0 || limit > 100 {
		limit = 100
	}
	var claimed []domain.OutboxEvent
	claimedUntil := now.UTC().Add(30 * time.Second)
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Where(
				"(status = ? AND next_attempt_at <= ?) OR (status = ? AND claimed_until <= ?)",
				domain.OutboxStatusPending,
				now.UTC(),
				domain.OutboxStatusDelivering,
				now.UTC(),
			).
			Where(`
				NOT EXISTS (
					SELECT 1
					FROM collaboration_outbox_events AS earlier
					WHERE earlier.project_id = collaboration_outbox_events.project_id
					  AND earlier.sequence < collaboration_outbox_events.sequence
					  AND earlier.status IN (?, ?)
				)
			`, domain.OutboxStatusPending, domain.OutboxStatusDelivering).
			Order("project_id ASC, sequence ASC").Limit(limit).Find(&claimed).Error; err != nil {
			return err
		}
		for index := range claimed {
			event := &claimed[index]
			attempt := event.Attempts + 1
			if err := tx.Model(&domain.OutboxEvent{}).Where("id = ?", event.ID).
				Updates(map[string]any{
					"status":        domain.OutboxStatusDelivering,
					"attempts":      attempt,
					"claimed_until": claimedUntil,
					"updated_at":    now.UTC(),
				}).Error; err != nil {
				return err
			}
			attemptRecord := domain.DeliveryAttempt{OutboxEventID: event.ID, Attempt: attempt, StartedAt: now.UTC()}
			if err := attemptRecord.InitForCreate(now.UTC()); err != nil {
				return err
			}
			if err := tx.Create(&attemptRecord).Error; err != nil {
				return err
			}
			event.Attempts = attempt
			event.Status = domain.OutboxStatusDelivering
			event.ClaimedUntil = &claimedUntil
		}
		return nil
	})
	return claimed, err
}

func (s *Store) WasProcessed(ctx context.Context, consumerID string, eventID uuid.UUID) (bool, error) {
	var count int64
	err := s.db.WithContext(ctx).Model(&domain.ProcessedEvent{}).
		Where("consumer_id = ? AND event_id = ?", consumerID, eventID).Count(&count).Error
	return count > 0, err
}

func (s *Store) MarkDelivered(ctx context.Context, consumerID string, event domain.OutboxEvent, now time.Time) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		processed := domain.ProcessedEvent{ConsumerID: consumerID, EventID: event.EventID, ProcessedAt: now.UTC()}
		if err := processed.InitForCreate(now.UTC()); err != nil {
			return err
		}
		if err := tx.Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "consumer_id"}, {Name: "event_id"}}, DoNothing: true}).Create(&processed).Error; err != nil {
			return err
		}
		result := tx.Model(&domain.OutboxEvent{}).
			Where("id = ? AND status = ? AND attempts = ?", event.ID, domain.OutboxStatusDelivering, event.Attempts).
			Updates(map[string]any{
				"status": domain.OutboxStatusDelivered, "delivered_at": now.UTC(), "claimed_until": nil, "last_error": "", "updated_at": now.UTC(),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return fmt.Errorf("collaboration outbox event %s delivery claim is stale", event.EventID)
		}
		return tx.Model(&domain.DeliveryAttempt{}).Where("outbox_event_id = ? AND attempt = ?", event.ID, event.Attempts).
			Updates(map[string]any{"finished_at": now.UTC(), "updated_at": now.UTC()}).Error
	})
}

func (s *Store) MarkFailed(ctx context.Context, event domain.OutboxEvent, deliveryError string, now, nextAttemptAt time.Time) error {
	status := domain.OutboxStatusPending
	if event.Attempts >= domain.MaxOutboxAttempts {
		status = domain.OutboxStatusFailed
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		result := tx.Model(&domain.OutboxEvent{}).
			Where("id = ? AND status = ? AND attempts = ?", event.ID, domain.OutboxStatusDelivering, event.Attempts).
			Updates(map[string]any{
				"status": status, "last_error": deliveryError, "next_attempt_at": nextAttemptAt.UTC(), "claimed_until": nil, "updated_at": now.UTC(),
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected != 1 {
			return fmt.Errorf("collaboration outbox event %s delivery claim is stale", event.EventID)
		}
		return tx.Model(&domain.DeliveryAttempt{}).Where("outbox_event_id = ? AND attempt = ?", event.ID, event.Attempts).
			Updates(map[string]any{"finished_at": now.UTC(), "error": deliveryError, "updated_at": now.UTC()}).Error
	})
}

func allocateProjectSequence(tx *gorm.DB, projectID uuid.UUID) (uint64, error) {
	var result struct {
		Sequence uint64 `gorm:"column:sequence"`
	}
	err := tx.Raw(`
		INSERT INTO collaboration_project_streams (project_id, next_sequence)
		VALUES (?, 2)
		ON CONFLICT (project_id)
		DO UPDATE SET next_sequence = collaboration_project_streams.next_sequence + 1
		RETURNING next_sequence - 1 AS sequence
	`, projectID).Scan(&result).Error
	if err != nil {
		return 0, fmt.Errorf("allocate collaboration project sequence: %w", err)
	}
	if result.Sequence == 0 {
		return 0, fmt.Errorf("allocate collaboration project sequence: empty result")
	}
	return result.Sequence, nil
}
