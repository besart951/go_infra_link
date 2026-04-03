package notification

import (
	"context"
	"errors"
	"fmt"
	"net/mail"
	"strings"

	domain "github.com/besart951/go_infra_link/backend/internal/domain"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
)

type Service struct {
	repo       domainNotification.SMTPSettingsRepository
	cipher     SecretCipher
	strategies map[domainNotification.Provider]EmailStrategy
}

func New(repo domainNotification.SMTPSettingsRepository, cipher SecretCipher, strategies ...EmailStrategy) *Service {
	registry := make(map[domainNotification.Provider]EmailStrategy, len(strategies))
	for _, strategy := range strategies {
		registry[strategy.Provider()] = strategy
	}

	return &Service{
		repo:       repo,
		cipher:     cipher,
		strategies: registry,
	}
}

func (s *Service) GetSMTPSettings() (*domainNotification.SMTPSettings, error) {
	return s.repo.GetByProvider(domainNotification.ProviderSMTP)
}

func (s *Service) UpsertSMTPSettings(input domainNotification.UpsertSMTPSettingsInput) (*domainNotification.SMTPSettings, error) {
	existing, err := s.repo.GetByProvider(domainNotification.ProviderSMTP)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if errors.Is(err, domain.ErrNotFound) {
		existing = nil
	}

	settings := &domainNotification.SMTPSettings{
		Provider:         domainNotification.ProviderSMTP,
		Enabled:          input.Enabled,
		Host:             strings.TrimSpace(input.Host),
		Port:             input.Port,
		Username:         strings.TrimSpace(input.Username),
		FromEmail:        strings.TrimSpace(input.FromEmail),
		FromName:         strings.TrimSpace(input.FromName),
		ReplyTo:          strings.TrimSpace(input.ReplyTo),
		Security:         input.Security,
		AuthMode:         input.AuthMode,
		AllowInsecureTLS: input.AllowInsecureTLS,
		UpdatedByID:      &input.ActorID,
	}

	if existing != nil {
		settings.ID = existing.ID
		settings.CreatedAt = existing.CreatedAt
		settings.PasswordEncrypted = existing.PasswordEncrypted
	}

	encryptedPassword, err := s.resolveEncryptedPassword(existing, input.Password, input.AuthMode)
	if err != nil {
		return nil, err
	}
	settings.PasswordEncrypted = encryptedPassword

	strategy, ok := s.strategies[settings.Provider]
	if !ok {
		return nil, fmt.Errorf("missing strategy for provider %s", settings.Provider)
	}
	if err := strategy.ValidateSettings(settings); err != nil {
		return nil, err
	}

	if err := s.repo.Save(settings); err != nil {
		return nil, err
	}
	return settings, nil
}

func (s *Service) SendTestEmail(ctx context.Context, input domainNotification.SendTestEmailInput) error {
	if _, err := mail.ParseAddress(strings.TrimSpace(input.To)); err != nil {
		return domain.NewValidationError().Add("to", "must be a valid email")
	}

	settings, strategy, password, err := s.resolveSMTPStrategy(false)
	if err != nil {
		return err
	}

	subject := strings.TrimSpace(input.Subject)
	if subject == "" {
		subject = "SMTP configuration test"
	}
	body := strings.TrimSpace(input.Body)
	if body == "" {
		body = "This is a test email sent from go_infra_link."
	}

	return strategy.Send(ctx, settings, password, domainNotification.EmailMessage{
		To:       []string{strings.TrimSpace(input.To)},
		Subject:  subject,
		TextBody: body,
	})
}

func (s *Service) SendNotification(ctx context.Context, input domainNotification.SendNotificationInput) error {
	settings, strategy, password, err := s.resolveSMTPStrategy(true)
	if err != nil {
		return err
	}
	return strategy.Send(ctx, settings, password, domainNotification.EmailMessage{
		To:       input.To,
		Subject:  input.Subject,
		TextBody: input.Body,
	})
}

func (s *Service) resolveSMTPStrategy(requireEnabled bool) (*domainNotification.SMTPSettings, EmailStrategy, string, error) {
	settings, err := s.repo.GetByProvider(domainNotification.ProviderSMTP)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, nil, "", domainNotification.ErrProviderNotConfigured
		}
		return nil, nil, "", err
	}
	if requireEnabled && !settings.Enabled {
		return nil, nil, "", domainNotification.ErrProviderDisabled
	}

	strategy, ok := s.strategies[settings.Provider]
	if !ok {
		return nil, nil, "", fmt.Errorf("missing strategy for provider %s", settings.Provider)
	}

	password, err := s.cipher.Decrypt(settings.PasswordEncrypted)
	if err != nil {
		return nil, nil, "", err
	}

	return settings, strategy, password, nil
}

func (s *Service) resolveEncryptedPassword(existing *domainNotification.SMTPSettings, rawPassword string, authMode domainNotification.AuthMode) (string, error) {
	if authMode == domainNotification.AuthModeNone {
		return "", nil
	}

	if strings.TrimSpace(rawPassword) != "" {
		return s.cipher.Encrypt(strings.TrimSpace(rawPassword))
	}

	if existing != nil && existing.PasswordEncrypted != "" {
		return existing.PasswordEncrypted, nil
	}

	return "", domain.NewValidationError().Add("password", "is required when auth_mode is plain")
}
