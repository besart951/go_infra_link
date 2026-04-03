package notification

import (
	"context"

	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
)

type EmailStrategy interface {
	Provider() domainNotification.Provider
	ValidateSettings(settings *domainNotification.SMTPSettings) error
	Send(ctx context.Context, settings *domainNotification.SMTPSettings, password string, message domainNotification.EmailMessage) error
}
