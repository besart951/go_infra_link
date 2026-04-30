package notification

import (
	"context"
	"time"

	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type emailOutboxRepo struct {
	db *gorm.DB
}

func NewEmailOutboxRepository(db *gorm.DB) domainNotification.EmailOutboxRepository {
	return &emailOutboxRepo{db: db}
}

func (r *emailOutboxRepo) Create(ctx context.Context, item *domainNotification.EmailOutbox) error {
	if item.Status == "" {
		item.Status = domainNotification.EmailOutboxStatusPending
	}
	if item.NextAttemptAt.IsZero() {
		item.NextAttemptAt = time.Now().UTC()
	}
	if err := item.InitForCreate(time.Now().UTC()); err != nil {
		return err
	}
	return r.db.WithContext(ctx).Create(item).Error
}

func (r *emailOutboxRepo) GetDue(ctx context.Context, now time.Time, limit int) ([]domainNotification.EmailOutbox, error) {
	if limit <= 0 {
		limit = 100
	}
	var items []domainNotification.EmailOutbox
	err := r.db.WithContext(ctx).
		Where("status = ? AND next_attempt_at <= ?", domainNotification.EmailOutboxStatusPending, now).
		Order("next_attempt_at ASC, created_at ASC").
		Limit(limit).
		Find(&items).Error
	return items, err
}

func (r *emailOutboxRepo) MarkSent(ctx context.Context, ids []uuid.UUID, sentAt time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).Model(&domainNotification.EmailOutbox{}).
		Where("id IN ?", ids).
		Updates(map[string]any{
			"updated_at": time.Now().UTC(),
			"status":     domainNotification.EmailOutboxStatusSent,
			"sent_at":    sentAt,
			"last_error": "",
		}).Error
}

func (r *emailOutboxRepo) MarkFailed(ctx context.Context, ids []uuid.UUID, attempts int, lastError string, nextAttemptAt time.Time) error {
	if len(ids) == 0 {
		return nil
	}
	status := domainNotification.EmailOutboxStatusPending
	if attempts >= domainNotification.MaxEmailOutboxAttempts {
		status = domainNotification.EmailOutboxStatusFailed
	}
	return r.db.WithContext(ctx).Model(&domainNotification.EmailOutbox{}).
		Where("id IN ?", ids).
		Updates(map[string]any{
			"updated_at":      time.Now().UTC(),
			"status":          status,
			"attempts":        attempts,
			"last_error":      lastError,
			"next_attempt_at": nextAttemptAt,
		}).Error
}
