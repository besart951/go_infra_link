package notification

import (
	"context"
	"net/http"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	"github.com/google/uuid"
)

type SMTPSettingsService interface {
	GetSMTPSettings(ctx context.Context) (*domainNotification.SMTPSettings, error)
	UpsertSMTPSettings(ctx context.Context, input domainNotification.UpsertSMTPSettingsInput) (*domainNotification.SMTPSettings, error)
	SendTestEmail(ctx context.Context, input domainNotification.SendTestEmailInput) error
}

type UserNotificationPreferenceService interface {
	GetUserPreference(ctx context.Context, userID uuid.UUID) (*domainNotification.UserPreference, error)
	UpsertUserPreference(ctx context.Context, input domainNotification.UpsertUserPreferenceInput) (*domainNotification.UserPreference, error)
	SendUserPreferenceVerificationCode(ctx context.Context, input domainNotification.SendUserPreferenceVerificationCodeInput) (*domainNotification.UserPreference, error)
	VerifyUserPreferenceEmail(ctx context.Context, input domainNotification.VerifyUserPreferenceEmailInput) (*domainNotification.UserPreference, error)
}

type SystemNotificationInboxService interface {
	ListSystemNotifications(ctx context.Context, userID uuid.UUID, page, limit int, unreadOnly bool) (*domain.PaginatedList[domainNotification.SystemNotification], error)
	CountUnreadSystemNotifications(ctx context.Context, userID uuid.UUID) (int64, error)
	MarkSystemNotificationRead(ctx context.Context, notificationID, userID uuid.UUID) (*domainNotification.SystemNotification, error)
	MarkAllSystemNotificationsRead(ctx context.Context, userID uuid.UUID) error
	ToggleSystemNotificationRead(ctx context.Context, notificationID, userID uuid.UUID) (*domainNotification.SystemNotification, error)
	ToggleSystemNotificationImportant(ctx context.Context, notificationID, userID uuid.UUID) (*domainNotification.SystemNotification, error)
	DeleteSystemNotification(ctx context.Context, notificationID, userID uuid.UUID) error
}

type NotificationRuleService interface {
	ListNotificationRules(ctx context.Context, filter domainNotification.NotificationRuleFilter) ([]domainNotification.NotificationRule, error)
	UpsertNotificationRule(ctx context.Context, input domainNotification.UpsertNotificationRuleInput) (*domainNotification.NotificationRule, error)
	DeleteNotificationRule(ctx context.Context, id uuid.UUID) error
}

type NotificationSettingsService interface {
	SMTPSettingsService
	UserNotificationPreferenceService
	SystemNotificationInboxService
	NotificationRuleService
}

type SystemNotificationStreamer interface {
	Stream(w http.ResponseWriter, r *http.Request, recipientID uuid.UUID)
}
