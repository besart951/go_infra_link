package wire

import (
	"fmt"

	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	notificationservice "github.com/besart951/go_infra_link/backend/internal/service/notification"
)

func newNotificationService(repos *Repositories, cfg ServiceConfig) (*notificationservice.Service, error) {
	secretCipher, err := notificationservice.NewAESCipher(cfg.JWTSecret)
	if err != nil {
		return nil, fmt.Errorf("notification secret cipher: %w", err)
	}

	var runtime *RuntimeAdapters
	if cfg.Runtime != nil {
		runtime = cfg.Runtime
	}

	return notificationservice.NewFromDependencies(notificationservice.Dependencies{
		SMTPSettings:    repos.NotificationSMTPSettings,
		Preferences:     repos.NotificationPreferences,
		SystemInbox:     repos.SystemNotifications,
		EmailOutbox:     repos.NotificationEmailOutbox,
		Rules:           repos.NotificationRules,
		Projects:        repos.Project,
		TeamMembers:     repos.TeamMember,
		Users:           repos.User,
		Cipher:          secretCipher,
		EmailStrategies: []notificationservice.EmailStrategy{notificationservice.NewSMTPStrategy()},
	}, notificationservice.Config{
		VerificationSecret: cfg.JWTSecret,
		SystemPublisher:    systemNotificationPublisher(runtime),
	}), nil
}

func systemNotificationPublisher(runtime *RuntimeAdapters) domainNotification.SystemNotificationPublisher {
	if runtime == nil {
		return nil
	}
	return runtime.SystemNotificationStream
}
