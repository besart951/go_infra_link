package user

import (
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type InvitationEmailStatus string

const (
	InvitationEmailStatusPending InvitationEmailStatus = "pending"
	InvitationEmailStatusSent    InvitationEmailStatus = "sent"
	InvitationEmailStatusFailed  InvitationEmailStatus = "failed"
)

type UserInvitation struct {
	domain.Base
	UserID         uuid.UUID             `gorm:"type:uuid;not null;uniqueIndex"`
	CreatedByID    *uuid.UUID            `gorm:"type:uuid;index"`
	TokenHash      string                `gorm:"type:varchar(128);index"`
	ExpiresAt      time.Time             `gorm:"not null;index"`
	AcceptedAt     *time.Time            `gorm:"index"`
	PrivacyAckAt   *time.Time            `gorm:"index"`
	EmailStatus    InvitationEmailStatus `gorm:"type:varchar(16);not null;default:'pending';index"`
	LatestOutboxID *uuid.UUID            `gorm:"type:uuid;index"`
	SendCount      int                   `gorm:"not null;default:0"`
	LastSentAt     *time.Time            `gorm:"index"`
	LastError      string                `gorm:"type:text"`
}

func (i *UserInvitation) GetBase() *domain.Base {
	return &i.Base
}

func (s InvitationEmailStatus) Valid() bool {
	switch s {
	case InvitationEmailStatusPending, InvitationEmailStatusSent, InvitationEmailStatusFailed:
		return true
	default:
		return false
	}
}

func NormalizeInvitationEmailStatus(value InvitationEmailStatus) InvitationEmailStatus {
	normalized := InvitationEmailStatus(strings.ToLower(strings.TrimSpace(string(value))))
	if !normalized.Valid() {
		return InvitationEmailStatusPending
	}
	return normalized
}
