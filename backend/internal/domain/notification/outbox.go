package notification

import (
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type EmailOutboxStatus string

const (
	EmailOutboxStatusPending EmailOutboxStatus = "pending"
	EmailOutboxStatusSent    EmailOutboxStatus = "sent"
	EmailOutboxStatusFailed  EmailOutboxStatus = "failed"

	MaxEmailOutboxAttempts = 5
)

type EmailOutbox struct {
	domain.Base
	RecipientID    uuid.UUID         `gorm:"type:uuid;not null;index:idx_email_outbox_recipient_status"`
	RecipientEmail string            `gorm:"type:varchar(320);not null"`
	EventKey       string            `gorm:"type:varchar(128);not null;index"`
	Subject        string            `gorm:"not null"`
	Body           string            `gorm:"type:text"`
	Frequency      DeliveryFrequency `gorm:"type:varchar(16);not null;index:idx_email_outbox_due"`
	ResourceType   string            `gorm:"type:varchar(64)"`
	ResourceID     *uuid.UUID        `gorm:"type:uuid;index"`
	Metadata       map[string]string `gorm:"serializer:json;type:text"`
	Status         EmailOutboxStatus `gorm:"type:varchar(16);not null;index:idx_email_outbox_due"`
	Attempts       int               `gorm:"not null;default:0"`
	NextAttemptAt  time.Time         `gorm:"not null;index:idx_email_outbox_due"`
	SentAt         *time.Time        `gorm:"index"`
	LastError      string            `gorm:"type:text"`
}

func (o *EmailOutbox) GetBase() *domain.Base {
	return &o.Base
}

func (s EmailOutboxStatus) Valid() bool {
	switch s {
	case EmailOutboxStatusPending, EmailOutboxStatusSent, EmailOutboxStatusFailed:
		return true
	default:
		return false
	}
}

func NormalizeEmailOutboxStatus(value EmailOutboxStatus) EmailOutboxStatus {
	return EmailOutboxStatus(strings.ToLower(strings.TrimSpace(string(value))))
}
