package notification

import (
	"context"

	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
)

type NotificationSettingsService interface {
	GetSMTPSettings(ctx context.Context) (*domainNotification.SMTPSettings, error)
	UpsertSMTPSettings(ctx context.Context, input domainNotification.UpsertSMTPSettingsInput) (*domainNotification.SMTPSettings, error)
	SendTestEmail(ctx context.Context, input domainNotification.SendTestEmailInput) error
}
