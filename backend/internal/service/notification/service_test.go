package notification

import (
	"context"
	"errors"
	"testing"

	domain "github.com/besart951/go_infra_link/backend/internal/domain"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	"github.com/google/uuid"
)

type smtpSettingsRepoStub struct {
	settings *domainNotification.SMTPSettings
	saveErr  error
}

func (r *smtpSettingsRepoStub) GetByProvider(_ context.Context, provider domainNotification.Provider) (*domainNotification.SMTPSettings, error) {
	if r.settings == nil || r.settings.Provider != provider {
		return nil, domain.ErrNotFound
	}
	copy := *r.settings
	return &copy, nil
}

func (r *smtpSettingsRepoStub) Save(_ context.Context, settings *domainNotification.SMTPSettings) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	copy := *settings
	r.settings = &copy
	return nil
}

type secretCipherStub struct{}

func (secretCipherStub) Encrypt(plain string) (string, error)   { return "enc:" + plain, nil }
func (secretCipherStub) Decrypt(encoded string) (string, error) { return encoded[4:], nil }

type strategyStub struct {
	validateErr  error
	sendErr      error
	lastPassword string
	lastMessage  domainNotification.EmailMessage
}

func (s *strategyStub) Provider() domainNotification.Provider                   { return domainNotification.ProviderSMTP }
func (s *strategyStub) ValidateSettings(*domainNotification.SMTPSettings) error { return s.validateErr }
func (s *strategyStub) Send(_ context.Context, _ *domainNotification.SMTPSettings, password string, message domainNotification.EmailMessage) error {
	s.lastPassword = password
	s.lastMessage = message
	return s.sendErr
}

func TestUpsertSMTPSettingsPreservesExistingPassword(t *testing.T) {
	repo := &smtpSettingsRepoStub{
		settings: &domainNotification.SMTPSettings{
			Provider:          domainNotification.ProviderSMTP,
			PasswordEncrypted: "enc:existing",
			Host:              "smtp.example.com",
			Port:              587,
			FromEmail:         "noreply@example.com",
			Security:          domainNotification.SecuritySTARTTLS,
			AuthMode:          domainNotification.AuthModePlain,
		},
	}
	strategy := &strategyStub{}
	service := New(repo, secretCipherStub{}, strategy)

	actorID := uuid.Must(uuid.NewV7())
	settings, err := service.UpsertSMTPSettings(context.Background(), domainNotification.UpsertSMTPSettingsInput{
		ActorID:          actorID,
		Enabled:          true,
		Host:             "smtp.example.com",
		Port:             587,
		Username:         "mailer",
		FromEmail:        "noreply@example.com",
		Security:         domainNotification.SecuritySTARTTLS,
		AuthMode:         domainNotification.AuthModePlain,
		AllowInsecureTLS: false,
	})
	if err != nil {
		t.Fatalf("UpsertSMTPSettings returned error: %v", err)
	}

	if settings.PasswordEncrypted != "enc:existing" {
		t.Fatalf("expected existing encrypted password to be preserved, got %q", settings.PasswordEncrypted)
	}
	if repo.settings == nil || repo.settings.PasswordEncrypted != "enc:existing" {
		t.Fatalf("expected repository save to preserve encrypted password")
	}
}

func TestSendNotificationUsesDecryptedPassword(t *testing.T) {
	repo := &smtpSettingsRepoStub{
		settings: &domainNotification.SMTPSettings{
			Provider:          domainNotification.ProviderSMTP,
			Enabled:           true,
			Host:              "smtp.example.com",
			Port:              587,
			Username:          "mailer",
			PasswordEncrypted: "enc:secret",
			FromEmail:         "noreply@example.com",
			Security:          domainNotification.SecuritySTARTTLS,
			AuthMode:          domainNotification.AuthModePlain,
		},
	}
	strategy := &strategyStub{}
	service := New(repo, secretCipherStub{}, strategy)

	err := service.SendNotification(context.Background(), domainNotification.SendNotificationInput{
		To:      []string{"person@example.com"},
		Subject: "Hello",
		Body:    "World",
	})
	if err != nil {
		t.Fatalf("SendNotification returned error: %v", err)
	}

	if strategy.lastPassword != "secret" {
		t.Fatalf("expected decrypted password, got %q", strategy.lastPassword)
	}
	if len(strategy.lastMessage.To) != 1 || strategy.lastMessage.To[0] != "person@example.com" {
		t.Fatalf("unexpected message recipients: %#v", strategy.lastMessage.To)
	}
}

func TestSendNotificationFailsWhenDisabled(t *testing.T) {
	repo := &smtpSettingsRepoStub{
		settings: &domainNotification.SMTPSettings{
			Provider:  domainNotification.ProviderSMTP,
			Enabled:   false,
			Host:      "smtp.example.com",
			Port:      587,
			FromEmail: "noreply@example.com",
			Security:  domainNotification.SecuritySTARTTLS,
			AuthMode:  domainNotification.AuthModeNone,
		},
	}
	service := New(repo, secretCipherStub{}, &strategyStub{})

	err := service.SendNotification(context.Background(), domainNotification.SendNotificationInput{
		To:      []string{"person@example.com"},
		Subject: "Hello",
		Body:    "World",
	})
	if !errors.Is(err, domainNotification.ErrProviderDisabled) {
		t.Fatalf("expected ErrProviderDisabled, got %v", err)
	}
}
