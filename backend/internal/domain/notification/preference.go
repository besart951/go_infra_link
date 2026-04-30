package notification

import (
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type DeliveryChannel string

const (
	DeliveryChannelEmail  DeliveryChannel = "email"
	DeliveryChannelSystem DeliveryChannel = "system"
	DeliveryChannelBoth   DeliveryChannel = "both"
)

type DeliveryFrequency string

const (
	DeliveryFrequencyImmediate DeliveryFrequency = "immediate"
	DeliveryFrequencyHourly    DeliveryFrequency = "hourly"
	DeliveryFrequencyDaily     DeliveryFrequency = "daily"
	DeliveryFrequencyWeekly    DeliveryFrequency = "weekly"
)

type UserPreference struct {
	domain.Base
	UserID                      uuid.UUID         `gorm:"type:uuid;uniqueIndex;not null"`
	NotificationEmail           string            `gorm:"type:varchar(320)"`
	NotificationEmailVerifiedAt *time.Time        `gorm:"index"`
	EmailVerificationCodeHash   string            `gorm:"type:varchar(128)"`
	EmailVerificationExpiresAt  *time.Time        `gorm:"index"`
	EmailVerificationSentAt     *time.Time        `gorm:"index"`
	Channel                     DeliveryChannel   `gorm:"type:varchar(16);not null;default:'both'"`
	Frequency                   DeliveryFrequency `gorm:"type:varchar(16);not null;default:'immediate'"`
}

func (p *UserPreference) GetBase() *domain.Base {
	return &p.Base
}

func DefaultUserPreference(userID uuid.UUID) *UserPreference {
	return &UserPreference{
		UserID:    userID,
		Channel:   DeliveryChannelBoth,
		Frequency: DeliveryFrequencyImmediate,
	}
}

func (p DeliveryChannel) Valid() bool {
	switch p {
	case DeliveryChannelEmail, DeliveryChannelSystem, DeliveryChannelBoth:
		return true
	default:
		return false
	}
}

func (p DeliveryChannel) AllowsEmail() bool {
	return p == DeliveryChannelEmail || p == DeliveryChannelBoth
}

func (p DeliveryChannel) AllowsSystem() bool {
	return p == DeliveryChannelSystem || p == DeliveryChannelBoth
}

func (p DeliveryFrequency) Valid() bool {
	switch p {
	case DeliveryFrequencyImmediate, DeliveryFrequencyHourly, DeliveryFrequencyDaily, DeliveryFrequencyWeekly:
		return true
	default:
		return false
	}
}

func NormalizeDeliveryChannel(value DeliveryChannel) DeliveryChannel {
	return DeliveryChannel(strings.ToLower(strings.TrimSpace(string(value))))
}

func NormalizeDeliveryFrequency(value DeliveryFrequency) DeliveryFrequency {
	return DeliveryFrequency(strings.ToLower(strings.TrimSpace(string(value))))
}

func NormalizeNotificationEmail(value string) string {
	return strings.TrimSpace(value)
}

func (p *UserPreference) NotificationEmailVerified() bool {
	return p != nil && p.NotificationEmail != "" && p.NotificationEmailVerifiedAt != nil
}
