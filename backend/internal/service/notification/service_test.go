package notification

import (
	"context"
	"errors"
	"testing"
	"time"

	domain "github.com/besart951/go_infra_link/backend/internal/domain"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
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

type userPreferenceRepoStub struct {
	preference *domainNotification.UserPreference
	saveErr    error
}

func (r *userPreferenceRepoStub) GetByUserID(_ context.Context, userID uuid.UUID) (*domainNotification.UserPreference, error) {
	if r.preference == nil || r.preference.UserID != userID {
		return nil, domain.ErrNotFound
	}
	copy := *r.preference
	return &copy, nil
}

func (r *userPreferenceRepoStub) Save(_ context.Context, preference *domainNotification.UserPreference) error {
	if r.saveErr != nil {
		return r.saveErr
	}
	copy := *preference
	r.preference = &copy
	return nil
}

type systemNotificationRepoStub struct {
	created []domainNotification.SystemNotification
}

func (r *systemNotificationRepoStub) Create(_ context.Context, notification *domainNotification.SystemNotification) error {
	r.created = append(r.created, *notification)
	return nil
}

func (r *systemNotificationRepoStub) GetPaginatedListForUser(context.Context, uuid.UUID, domain.PaginationParams, bool) (*domain.PaginatedList[domainNotification.SystemNotification], error) {
	return &domain.PaginatedList[domainNotification.SystemNotification]{}, nil
}

func (r *systemNotificationRepoStub) CountUnreadForUser(context.Context, uuid.UUID) (int64, error) {
	return 0, nil
}

func (r *systemNotificationRepoStub) MarkReadForUser(context.Context, uuid.UUID, uuid.UUID) (*domainNotification.SystemNotification, error) {
	return nil, domain.ErrNotFound
}

func (r *systemNotificationRepoStub) MarkAllReadForUser(context.Context, uuid.UUID) error {
	return nil
}

type emailOutboxRepoStub struct {
	created        []domainNotification.EmailOutbox
	due            []domainNotification.EmailOutbox
	markedSent     []uuid.UUID
	markedFailed   []uuid.UUID
	failedAttempts int
}

func (r *emailOutboxRepoStub) Create(_ context.Context, item *domainNotification.EmailOutbox) error {
	r.created = append(r.created, *item)
	return nil
}

func (r *emailOutboxRepoStub) GetDue(context.Context, time.Time, int) ([]domainNotification.EmailOutbox, error) {
	return r.due, nil
}

func (r *emailOutboxRepoStub) MarkSent(_ context.Context, ids []uuid.UUID, _ time.Time) error {
	r.markedSent = append(r.markedSent, ids...)
	return nil
}

func (r *emailOutboxRepoStub) MarkFailed(_ context.Context, ids []uuid.UUID, attempts int, _ string, _ time.Time) error {
	r.markedFailed = append(r.markedFailed, ids...)
	r.failedAttempts = attempts
	return nil
}

type userRepoStub struct {
	users map[uuid.UUID]*domainUser.User
}

func (r *userRepoStub) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainUser.User, error) {
	users := make([]*domainUser.User, 0, len(ids))
	for _, id := range ids {
		if user := r.users[id]; user != nil {
			users = append(users, user)
		}
	}
	return users, nil
}

func (r *userRepoStub) Create(context.Context, *domainUser.User) error { return nil }
func (r *userRepoStub) Update(context.Context, *domainUser.User) error { return nil }
func (r *userRepoStub) DeleteByIds(context.Context, []uuid.UUID) error { return nil }
func (r *userRepoStub) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainUser.User], error) {
	return &domain.PaginatedList[domainUser.User]{}, nil
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
	service := New(repo, nil, nil, nil, nil, nil, nil, nil, secretCipherStub{}, "test-secret", strategy)

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
	service := New(repo, nil, nil, nil, nil, nil, nil, nil, secretCipherStub{}, "test-secret", strategy)

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
	service := New(repo, nil, nil, nil, nil, nil, nil, nil, secretCipherStub{}, "test-secret", &strategyStub{})

	err := service.SendNotification(context.Background(), domainNotification.SendNotificationInput{
		To:      []string{"person@example.com"},
		Subject: "Hello",
		Body:    "World",
	})
	if !errors.Is(err, domainNotification.ErrProviderDisabled) {
		t.Fatalf("expected ErrProviderDisabled, got %v", err)
	}
}

func TestGetUserPreferenceReturnsDefaultWhenMissing(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	service := New(nil, &userPreferenceRepoStub{}, nil, nil, nil, nil, nil, nil, secretCipherStub{}, "test-secret")

	preference, err := service.GetUserPreference(context.Background(), userID)
	if err != nil {
		t.Fatalf("GetUserPreference returned error: %v", err)
	}

	if preference.UserID != userID {
		t.Fatalf("expected user id %s, got %s", userID, preference.UserID)
	}
	if preference.Channel != domainNotification.DeliveryChannelBoth {
		t.Fatalf("expected default channel both, got %q", preference.Channel)
	}
	if preference.Frequency != domainNotification.DeliveryFrequencyImmediate {
		t.Fatalf("expected default frequency immediate, got %q", preference.Frequency)
	}
}

func TestUpsertUserPreferenceNormalizesAndSaves(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	repo := &userPreferenceRepoStub{}
	service := New(nil, repo, nil, nil, nil, nil, nil, nil, secretCipherStub{}, "test-secret")

	preference, err := service.UpsertUserPreference(context.Background(), domainNotification.UpsertUserPreferenceInput{
		UserID:            userID,
		NotificationEmail: "person@example.com",
		Channel:           domainNotification.DeliveryChannel(" EMAIL "),
		Frequency:         domainNotification.DeliveryFrequency(" DAILY "),
	})
	if err != nil {
		t.Fatalf("UpsertUserPreference returned error: %v", err)
	}

	if preference.Channel != domainNotification.DeliveryChannelEmail {
		t.Fatalf("expected normalized email channel, got %q", preference.Channel)
	}
	if preference.Frequency != domainNotification.DeliveryFrequencyDaily {
		t.Fatalf("expected normalized daily frequency, got %q", preference.Frequency)
	}
	if repo.preference == nil || repo.preference.Channel != domainNotification.DeliveryChannelEmail {
		t.Fatalf("expected preference to be saved, got %#v", repo.preference)
	}
}

func TestDispatchQueuesSystemNotificationAndEmailOutbox(t *testing.T) {
	userID := uuid.Must(uuid.NewV7())
	verifiedAt := time.Now().UTC()
	preferenceRepo := &userPreferenceRepoStub{
		preference: &domainNotification.UserPreference{
			UserID:                      userID,
			NotificationEmail:           "person@example.com",
			NotificationEmailVerifiedAt: &verifiedAt,
			Channel:                     domainNotification.DeliveryChannelBoth,
			Frequency:                   domainNotification.DeliveryFrequencyHourly,
		},
	}
	systemRepo := &systemNotificationRepoStub{}
	outboxRepo := &emailOutboxRepoStub{}
	userRepo := &userRepoStub{
		users: map[uuid.UUID]*domainUser.User{
			userID: {Base: domain.Base{ID: userID}, Email: "login@example.com"},
		},
	}
	service := New(nil, preferenceRepo, systemRepo, outboxRepo, nil, nil, nil, userRepo, secretCipherStub{}, "test-secret")

	err := service.Dispatch(context.Background(), domainNotification.DispatchNotificationInput{
		RecipientIDs: []uuid.UUID{userID},
		EventKey:     "project.phase.changed",
		Metadata: map[string]string{
			"project_name": "Testprojekt",
			"old":          "Planung",
			"new":          "Ausführung",
		},
	})
	if err != nil {
		t.Fatalf("Dispatch returned error: %v", err)
	}

	if len(systemRepo.created) != 1 {
		t.Fatalf("expected one system notification, got %d", len(systemRepo.created))
	}
	if systemRepo.created[0].Title != "Projektphase geändert" {
		t.Fatalf("expected localized title, got %q", systemRepo.created[0].Title)
	}
	if len(outboxRepo.created) != 1 {
		t.Fatalf("expected one email outbox item, got %d", len(outboxRepo.created))
	}
	outbox := outboxRepo.created[0]
	if outbox.RecipientEmail != "person@example.com" {
		t.Fatalf("expected notification email recipient, got %q", outbox.RecipientEmail)
	}
	if outbox.Frequency != domainNotification.DeliveryFrequencyHourly {
		t.Fatalf("expected hourly outbox frequency, got %q", outbox.Frequency)
	}
	if outbox.NextAttemptAt.IsZero() {
		t.Fatalf("expected next attempt to be scheduled")
	}
}

func TestProcessDueEmailOutboxGroupsDigestItems(t *testing.T) {
	recipientID := uuid.Must(uuid.NewV7())
	firstID := uuid.Must(uuid.NewV7())
	secondID := uuid.Must(uuid.NewV7())
	outboxRepo := &emailOutboxRepoStub{
		due: []domainNotification.EmailOutbox{
			{
				Base:           domain.Base{ID: firstID},
				RecipientID:    recipientID,
				RecipientEmail: "person@example.com",
				Subject:        "Erste Meldung",
				Body:           "Erster Text",
				Frequency:      domainNotification.DeliveryFrequencyDaily,
			},
			{
				Base:           domain.Base{ID: secondID},
				RecipientID:    recipientID,
				RecipientEmail: "person@example.com",
				Subject:        "Zweite Meldung",
				Body:           "Zweiter Text",
				Frequency:      domainNotification.DeliveryFrequencyDaily,
			},
		},
	}
	strategy := &strategyStub{}
	smtpRepo := &smtpSettingsRepoStub{
		settings: &domainNotification.SMTPSettings{
			Provider:          domainNotification.ProviderSMTP,
			Enabled:           true,
			Host:              "smtp.example.com",
			Port:              587,
			PasswordEncrypted: "enc:secret",
			FromEmail:         "noreply@example.com",
			Security:          domainNotification.SecuritySTARTTLS,
			AuthMode:          domainNotification.AuthModeNone,
		},
	}
	service := New(smtpRepo, nil, nil, outboxRepo, nil, nil, nil, nil, secretCipherStub{}, "test-secret", strategy)

	if err := service.ProcessDueEmailOutbox(context.Background(), time.Now().UTC(), 100); err != nil {
		t.Fatalf("ProcessDueEmailOutbox returned error: %v", err)
	}

	if len(outboxRepo.markedSent) != 2 {
		t.Fatalf("expected both outbox items marked sent, got %d", len(outboxRepo.markedSent))
	}
	if len(strategy.lastMessage.To) != 1 || strategy.lastMessage.To[0] != "person@example.com" {
		t.Fatalf("unexpected digest recipients: %#v", strategy.lastMessage.To)
	}
	if strategy.lastMessage.Subject != "Infra Link: 2 Benachrichtigungen" {
		t.Fatalf("unexpected digest subject: %q", strategy.lastMessage.Subject)
	}
}
