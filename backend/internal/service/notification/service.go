package notification

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"net/mail"
	"strings"
	"time"

	domain "github.com/besart951/go_infra_link/backend/internal/domain"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type SettingsService struct {
	smtpSettingsRepo domainNotification.SMTPSettingsRepository
	cipher           SecretCipher
	strategies       map[domainNotification.Provider]EmailStrategy
}

type PreferencesService struct {
	preferenceRepo  domainNotification.UserPreferenceRepository
	settings        *SettingsService
	verificationKey []byte
}

type InboxService struct {
	systemRepo      domainNotification.SystemNotificationRepository
	systemPublisher domainNotification.SystemNotificationPublisher
}

type notificationRecipientResolver struct {
	projectReader    domainNotification.ProjectMembershipReader
	teamMemberReader domainNotification.TeamMemberReader
}

type RuleDispatcher struct {
	emailOutboxRepo domainNotification.EmailOutboxRepository
	ruleRepo        domainNotification.NotificationRuleRepository
	userRepo        domainUser.UserRepository
	preferences     *PreferencesService
	inbox           *InboxService
	settings        *SettingsService
	recipients      notificationRecipientResolver
	renderer        *TemplateRenderer
}

type OutboxProcessor struct {
	emailOutboxRepo domainNotification.EmailOutboxRepository
	settings        *SettingsService
}

type TemplateRenderer struct{}

type Services struct {
	Settings    *SettingsService
	Preferences *PreferencesService
	Inbox       *InboxService
	Rules       *RuleDispatcher
	Outbox      *OutboxProcessor
	Templates   *TemplateRenderer
}

type Dependencies struct {
	SMTPSettings    domainNotification.SMTPSettingsRepository
	Preferences     domainNotification.UserPreferenceRepository
	SystemInbox     domainNotification.SystemNotificationRepository
	EmailOutbox     domainNotification.EmailOutboxRepository
	Rules           domainNotification.NotificationRuleRepository
	Projects        domainNotification.ProjectMembershipReader
	TeamMembers     domainNotification.TeamMemberReader
	Users           domainUser.UserRepository
	Cipher          SecretCipher
	EmailStrategies []EmailStrategy
}

type Config struct {
	VerificationSecret string
	SystemPublisher    domainNotification.SystemNotificationPublisher
}

func NewServices(deps Dependencies, cfg Config) *Services {
	registry := make(map[domainNotification.Provider]EmailStrategy, len(deps.EmailStrategies))
	for _, strategy := range deps.EmailStrategies {
		if strategy == nil {
			continue
		}
		registry[strategy.Provider()] = strategy
	}

	keySeed := "notification-email-verification:" + cfg.VerificationSecret
	key := sha256.Sum256([]byte(keySeed))

	settings := &SettingsService{
		smtpSettingsRepo: deps.SMTPSettings,
		cipher:           deps.Cipher,
		strategies:       registry,
	}
	preferences := &PreferencesService{
		preferenceRepo:  deps.Preferences,
		settings:        settings,
		verificationKey: key[:],
	}
	inbox := &InboxService{systemRepo: deps.SystemInbox, systemPublisher: cfg.SystemPublisher}
	renderer := &TemplateRenderer{}
	rules := &RuleDispatcher{
		emailOutboxRepo: deps.EmailOutbox,
		ruleRepo:        deps.Rules,
		userRepo:        deps.Users,
		preferences:     preferences,
		inbox:           inbox,
		settings:        settings,
		recipients: notificationRecipientResolver{
			projectReader: deps.Projects, teamMemberReader: deps.TeamMembers,
		},
		renderer: renderer,
	}
	return &Services{
		Settings: settings, Preferences: preferences, Inbox: inbox, Rules: rules,
		Outbox:    &OutboxProcessor{emailOutboxRepo: deps.EmailOutbox, settings: settings},
		Templates: renderer,
	}
}

func (s *InboxService) SetSystemNotificationPublisher(publisher domainNotification.SystemNotificationPublisher) {
	s.systemPublisher = publisher
}

func (s *SettingsService) GetSMTPSettings(ctx context.Context) (*domainNotification.SMTPSettings, error) {
	return s.smtpSettingsRepo.GetByProvider(ctx, domainNotification.ProviderSMTP)
}

func (s *SettingsService) UpsertSMTPSettings(ctx context.Context, input domainNotification.UpsertSMTPSettingsInput) (*domainNotification.SMTPSettings, error) {
	existing, err := s.smtpSettingsRepo.GetByProvider(ctx, domainNotification.ProviderSMTP)
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

	if err := s.smtpSettingsRepo.Save(ctx, settings); err != nil {
		return nil, err
	}
	return settings, nil
}

func (s *PreferencesService) GetUserPreference(ctx context.Context, userID uuid.UUID) (*domainNotification.UserPreference, error) {
	if userID == uuid.Nil {
		return nil, domain.ErrInvalidArgument
	}
	preference, err := s.preferenceRepo.GetByUserID(ctx, userID)
	if errors.Is(err, domain.ErrNotFound) {
		return domainNotification.DefaultUserPreference(userID), nil
	}
	return preference, err
}

func (s *PreferencesService) UpsertUserPreference(ctx context.Context, input domainNotification.UpsertUserPreferenceInput) (*domainNotification.UserPreference, error) {
	if input.UserID == uuid.Nil {
		return nil, domain.ErrInvalidArgument
	}

	preference := &domainNotification.UserPreference{
		UserID:            input.UserID,
		NotificationEmail: normalizeVerifiedRecipient(input.NotificationEmail),
		Channel:           domainNotification.NormalizeDeliveryChannel(input.Channel),
		Frequency:         domainNotification.NormalizeDeliveryFrequency(input.Frequency),
	}

	ve := domain.NewValidationError()
	existing, err := s.preferenceRepo.GetByUserID(ctx, input.UserID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}

	if !preference.Channel.Valid() {
		ve = ve.Add("channel", "must be one of: email system both")
	}
	if !preference.Frequency.Valid() {
		ve = ve.Add("frequency", "must be one of: immediate hourly daily weekly")
	}
	if preference.NotificationEmail != "" {
		if _, err := mail.ParseAddress(preference.NotificationEmail); err != nil {
			ve = ve.Add("notification_email", "must be a valid email")
		}
	}
	if preference.Channel.AllowsEmail() && preference.NotificationEmail == "" {
		ve = ve.Add("notification_email", "is required")
	}
	if len(ve.Fields) > 0 {
		return nil, ve
	}

	if existing != nil && strings.EqualFold(existing.NotificationEmail, preference.NotificationEmail) {
		preference.NotificationEmailVerifiedAt = existing.NotificationEmailVerifiedAt
		preference.EmailVerificationCodeHash = existing.EmailVerificationCodeHash
		preference.EmailVerificationExpiresAt = existing.EmailVerificationExpiresAt
		preference.EmailVerificationSentAt = existing.EmailVerificationSentAt
	}

	if err := s.preferenceRepo.Save(ctx, preference); err != nil {
		return nil, err
	}
	return preference, nil
}

func (s *PreferencesService) SendUserPreferenceVerificationCode(ctx context.Context, input domainNotification.SendUserPreferenceVerificationCodeInput) (*domainNotification.UserPreference, error) {
	if input.UserID == uuid.Nil {
		return nil, domain.ErrInvalidArgument
	}

	preference, err := s.GetUserPreference(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if preference.NotificationEmail == "" {
		return nil, domain.NewValidationError().Add("notification_email", "is required")
	}
	if _, err := mail.ParseAddress(preference.NotificationEmail); err != nil {
		return nil, domain.NewValidationError().Add("notification_email", "must be a valid email")
	}

	code, err := generateVerificationCode()
	if err != nil {
		return nil, err
	}

	now := time.Now().UTC()
	expiresAt := now.Add(15 * time.Minute)
	preference.EmailVerificationCodeHash = s.hashVerificationCode(input.UserID, preference.NotificationEmail, code)
	preference.EmailVerificationExpiresAt = &expiresAt
	preference.EmailVerificationSentAt = &now

	if err := s.preferenceRepo.Save(ctx, preference); err != nil {
		return nil, err
	}

	settings, strategy, password, err := s.settings.resolveSMTPStrategy(ctx, false)
	if err != nil {
		return nil, err
	}
	if err := strategy.Send(ctx, settings, password, domainNotification.EmailMessage{
		To:       []string{preference.NotificationEmail},
		Subject:  "Infra Link E-Mail bestaetigen",
		TextBody: fmt.Sprintf("Ihr Infra Link Bestaetigungscode lautet: %s\n\nDer Code ist 15 Minuten gueltig.", code),
	}); err != nil {
		return nil, err
	}

	return preference, nil
}

func (s *PreferencesService) VerifyUserPreferenceEmail(ctx context.Context, input domainNotification.VerifyUserPreferenceEmailInput) (*domainNotification.UserPreference, error) {
	if input.UserID == uuid.Nil {
		return nil, domain.ErrInvalidArgument
	}

	code := strings.TrimSpace(input.Code)
	if len(code) != 6 {
		return nil, domain.NewValidationError().Add("code", "length 6")
	}

	preference, err := s.GetUserPreference(ctx, input.UserID)
	if err != nil {
		return nil, err
	}
	if preference.NotificationEmail == "" || preference.EmailVerificationCodeHash == "" || preference.EmailVerificationExpiresAt == nil {
		return nil, domain.NewValidationError().Add("code", "invalid")
	}
	if time.Now().UTC().After(*preference.EmailVerificationExpiresAt) {
		return nil, domain.NewValidationError().Add("code", "expired")
	}

	expectedHash := s.hashVerificationCode(input.UserID, preference.NotificationEmail, code)
	if !hmac.Equal([]byte(expectedHash), []byte(preference.EmailVerificationCodeHash)) {
		return nil, domain.NewValidationError().Add("code", "invalid")
	}

	now := time.Now().UTC()
	preference.NotificationEmailVerifiedAt = &now
	preference.EmailVerificationCodeHash = ""
	preference.EmailVerificationExpiresAt = nil

	if err := s.preferenceRepo.Save(ctx, preference); err != nil {
		return nil, err
	}
	return preference, nil
}

func (s *InboxService) ListSystemNotifications(ctx context.Context, userID uuid.UUID, page, limit int, unreadOnly bool) (*domain.PaginatedList[domainNotification.SystemNotification], error) {
	if userID == uuid.Nil {
		return nil, domain.ErrInvalidArgument
	}
	return s.systemRepo.GetPaginatedListForUser(ctx, userID, domain.PaginationParams{
		Page:  page,
		Limit: limit,
	}, unreadOnly)
}

func (s *InboxService) CountUnreadSystemNotifications(ctx context.Context, userID uuid.UUID) (int64, error) {
	if userID == uuid.Nil {
		return 0, domain.ErrInvalidArgument
	}
	return s.systemRepo.CountUnreadForUser(ctx, userID)
}

func (s *InboxService) MarkSystemNotificationRead(ctx context.Context, notificationID, userID uuid.UUID) (*domainNotification.SystemNotification, error) {
	if notificationID == uuid.Nil || userID == uuid.Nil {
		return nil, domain.ErrInvalidArgument
	}
	notification, err := s.systemRepo.MarkReadForUser(ctx, notificationID, userID)
	if err != nil {
		return nil, err
	}
	s.publishSystemNotificationUpdated(ctx, notification)
	return notification, nil
}

func (s *InboxService) MarkAllSystemNotificationsRead(ctx context.Context, userID uuid.UUID) error {
	if userID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	if err := s.systemRepo.MarkAllReadForUser(ctx, userID); err != nil {
		return err
	}
	s.publishSystemNotificationsReadAll(ctx, userID)
	return nil
}

func (s *InboxService) ToggleSystemNotificationRead(ctx context.Context, notificationID, userID uuid.UUID) (*domainNotification.SystemNotification, error) {
	if notificationID == uuid.Nil || userID == uuid.Nil {
		return nil, domain.ErrInvalidArgument
	}
	notification, err := s.systemRepo.ToggleReadForUser(ctx, notificationID, userID)
	if err != nil {
		return nil, err
	}
	s.publishSystemNotificationUpdated(ctx, notification)
	return notification, nil
}

func (s *InboxService) ToggleSystemNotificationImportant(ctx context.Context, notificationID, userID uuid.UUID) (*domainNotification.SystemNotification, error) {
	if notificationID == uuid.Nil || userID == uuid.Nil {
		return nil, domain.ErrInvalidArgument
	}
	notification, err := s.systemRepo.ToggleImportantForUser(ctx, notificationID, userID)
	if err != nil {
		return nil, err
	}
	s.publishSystemNotificationUpdated(ctx, notification)
	return notification, nil
}

func (s *InboxService) DeleteSystemNotification(ctx context.Context, notificationID, userID uuid.UUID) error {
	if notificationID == uuid.Nil || userID == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	if err := s.systemRepo.DeleteForUser(ctx, notificationID, userID); err != nil {
		return err
	}
	s.publishSystemNotificationDeleted(ctx, userID, notificationID)
	return nil
}

func (s *InboxService) publishSystemNotificationCreated(ctx context.Context, notification *domainNotification.SystemNotification) {
	if notification == nil {
		return
	}
	s.publishSystemNotificationChange(ctx, domainNotification.SystemNotificationChange{
		Type:         domainNotification.SystemNotificationChangeCreated,
		RecipientID:  notification.RecipientID,
		Notification: notification,
	})
}

func (s *InboxService) publishSystemNotificationUpdated(ctx context.Context, notification *domainNotification.SystemNotification) {
	if notification == nil {
		return
	}
	s.publishSystemNotificationChange(ctx, domainNotification.SystemNotificationChange{
		Type:         domainNotification.SystemNotificationChangeUpdated,
		RecipientID:  notification.RecipientID,
		Notification: notification,
	})
}

func (s *InboxService) publishSystemNotificationDeleted(ctx context.Context, recipientID, notificationID uuid.UUID) {
	s.publishSystemNotificationChange(ctx, domainNotification.SystemNotificationChange{
		Type:           domainNotification.SystemNotificationChangeDeleted,
		RecipientID:    recipientID,
		NotificationID: notificationID,
	})
}

func (s *InboxService) publishSystemNotificationsReadAll(ctx context.Context, recipientID uuid.UUID) {
	s.publishSystemNotificationChange(ctx, domainNotification.SystemNotificationChange{
		Type:        domainNotification.SystemNotificationChangeReadAll,
		RecipientID: recipientID,
		UnreadCount: 0,
	})
}

func (s *InboxService) publishSystemNotificationChange(ctx context.Context, change domainNotification.SystemNotificationChange) {
	if s.systemPublisher == nil || change.Type == "" {
		return
	}
	if change.Notification != nil {
		change.RecipientID = change.Notification.RecipientID
		change.NotificationID = change.Notification.ID
	}
	if change.RecipientID == uuid.Nil {
		return
	}
	if change.OccurredAt.IsZero() {
		change.OccurredAt = time.Now().UTC()
	}
	if s.systemRepo != nil && change.Type != domainNotification.SystemNotificationChangeReadAll {
		unreadCount, err := s.systemRepo.CountUnreadForUser(ctx, change.RecipientID)
		if err != nil {
			return
		}
		change.UnreadCount = unreadCount
	}
	s.systemPublisher.PublishSystemNotificationChange(ctx, change)
}

func (s *RuleDispatcher) ListNotificationRules(ctx context.Context, filter domainNotification.NotificationRuleFilter) ([]domainNotification.NotificationRule, error) {
	if s.ruleRepo == nil {
		return []domainNotification.NotificationRule{}, nil
	}
	return s.ruleRepo.List(ctx, filter)
}

func (s *RuleDispatcher) UpsertNotificationRule(ctx context.Context, input domainNotification.UpsertNotificationRuleInput) (*domainNotification.NotificationRule, error) {
	eventKey, err := input.EventKey.Normalize()
	if err != nil {
		return nil, domain.NewValidationError().Add("event_key", "is required")
	}
	resourceType := domainNotification.ResourceKind(strings.TrimSpace(input.ResourceType.String()))
	rule := &domainNotification.NotificationRule{
		Base:             domain.Base{ID: input.ID},
		Name:             strings.TrimSpace(input.Name),
		Enabled:          input.Enabled,
		EventKey:         eventKey,
		ProjectID:        input.ProjectID,
		ResourceType:     resourceType,
		ResourceID:       input.ResourceID,
		RecipientType:    domainNotification.NormalizeRuleRecipientType(input.RecipientType),
		RecipientUserIDs: dedupeUUIDs(input.RecipientUserIDs),
		RecipientTeamID:  input.RecipientTeamID,
		RecipientRole:    input.RecipientRole,
	}
	if input.ActorID != uuid.Nil {
		rule.CreatedByID = &input.ActorID
	}

	if err := validateNotificationRule(rule); err != nil {
		return nil, err
	}
	if s.ruleRepo == nil {
		return nil, domain.ErrInvalidArgument
	}

	if rule.ID == uuid.Nil {
		if err := s.ruleRepo.Create(ctx, rule); err != nil {
			return nil, err
		}
		return rule, nil
	}

	existing, err := s.ruleRepo.GetByID(ctx, rule.ID)
	if err != nil {
		return nil, err
	}
	rule.CreatedAt = existing.CreatedAt
	if existing.CreatedByID != nil {
		rule.CreatedByID = existing.CreatedByID
	}
	if err := s.ruleRepo.Update(ctx, rule); err != nil {
		return nil, err
	}
	return rule, nil
}

func (s *RuleDispatcher) DeleteNotificationRule(ctx context.Context, id uuid.UUID) error {
	if id == uuid.Nil {
		return domain.ErrInvalidArgument
	}
	if s.ruleRepo == nil {
		return domain.ErrInvalidArgument
	}
	return s.ruleRepo.DeleteByID(ctx, id)
}

func (s *RuleDispatcher) DispatchEvent(ctx context.Context, input domainNotification.DispatchEventInput) error {
	if s.ruleRepo == nil {
		return nil
	}
	eventKey, err := input.EventKey.Normalize()
	if err != nil {
		return domain.NewValidationError().Add("event_key", "is required")
	}

	rules, err := s.ruleRepo.ListMatching(ctx, eventKey, input.ProjectID, input.ResourceType, input.ResourceID)
	if err != nil {
		return err
	}
	if len(rules) == 0 {
		return nil
	}

	recipientIDs, err := s.resolveRuleRecipients(ctx, rules, input.ProjectID)
	if err != nil {
		return err
	}
	if len(recipientIDs) == 0 {
		return nil
	}

	title, body := s.renderer.Render(input.Locale, eventKey, input.Title, input.Body, input.Metadata)
	return s.Dispatch(ctx, domainNotification.DispatchNotificationInput{
		RecipientIDs: recipientIDs,
		ActorID:      input.ActorID,
		Locale:       input.Locale,
		EventKey:     eventKey,
		Title:        title,
		Body:         body,
		ResourceType: input.ResourceType,
		ResourceID:   input.ResourceID,
		Metadata:     input.Metadata,
	})
}

func (s *RuleDispatcher) Dispatch(ctx context.Context, input domainNotification.DispatchNotificationInput) error {
	message, err := s.prepareDispatch(input)
	if err != nil {
		return err
	}
	users, err := s.userRepo.GetByIds(ctx, message.recipientIDs)
	if err != nil {
		return err
	}
	emailRecipients := make([]string, 0, len(message.recipientIDs))
	usersByID := notificationUsersByID(users)
	for _, recipientID := range message.recipientIDs {
		user := usersByID[recipientID]
		if user == nil {
			continue
		}
		email, err := s.dispatchToRecipient(ctx, recipientID, message)
		if err != nil {
			return err
		}
		if email != "" {
			emailRecipients = append(emailRecipients, email)
		}
	}
	if s.emailOutboxRepo != nil || len(emailRecipients) == 0 {
		return nil
	}
	return s.settings.SendNotification(ctx, domainNotification.SendNotificationInput{
		To: emailRecipients, Subject: message.title, Body: message.body,
	})
}

type preparedDispatch struct {
	recipientIDs []uuid.UUID
	input        domainNotification.DispatchNotificationInput
	eventKey     domainNotification.EventKey
	resourceType domainNotification.ResourceKind
	title        string
	body         string
	now          time.Time
}

func (s *RuleDispatcher) prepareDispatch(input domainNotification.DispatchNotificationInput) (preparedDispatch, error) {
	recipients := dedupeUUIDs(input.RecipientIDs)
	if len(recipients) == 0 {
		return preparedDispatch{}, domain.NewValidationError().Add("recipient_ids", "is required")
	}
	eventKey, err := input.EventKey.Normalize()
	if err != nil {
		return preparedDispatch{}, domain.NewValidationError().Add("event_key", "is required")
	}
	title, body := s.renderer.Render(input.Locale, eventKey, input.Title, input.Body, input.Metadata)
	if strings.TrimSpace(title) == "" {
		return preparedDispatch{}, domain.NewValidationError().Add("title", "is required")
	}
	return preparedDispatch{recipientIDs: recipients, input: input, eventKey: eventKey,
		resourceType: domainNotification.ResourceKind(strings.TrimSpace(input.ResourceType.String())),
		title:        title, body: body, now: time.Now().UTC()}, nil
}

func notificationUsersByID(users []*domainUser.User) map[uuid.UUID]*domainUser.User {
	indexed := make(map[uuid.UUID]*domainUser.User, len(users))
	for _, user := range users {
		indexed[user.ID] = user
	}
	return indexed
}

func (s *RuleDispatcher) dispatchToRecipient(ctx context.Context, recipientID uuid.UUID, message preparedDispatch) (string, error) {
	preference, err := s.preferences.GetUserPreference(ctx, recipientID)
	if err != nil {
		return "", err
	}
	if preference.Channel.AllowsSystem() {
		if err := s.createInboxNotification(ctx, recipientID, message); err != nil {
			return "", err
		}
	}
	if !preference.Channel.AllowsEmail() || !preference.NotificationEmailVerified() {
		return "", nil
	}
	if s.emailOutboxRepo != nil {
		if err := s.createEmailOutbox(ctx, recipientID, preference, message); err != nil {
			return "", err
		}
	}
	return preference.NotificationEmail, nil
}

func (s *RuleDispatcher) createInboxNotification(ctx context.Context, recipientID uuid.UUID, message preparedDispatch) error {
	notification := &domainNotification.SystemNotification{
		RecipientID: recipientID, ActorID: message.input.ActorID, EventKey: message.eventKey,
		Title: message.title, Body: message.body, ResourceType: message.resourceType,
		ResourceID: message.input.ResourceID, Metadata: message.input.Metadata,
	}
	if err := s.inbox.systemRepo.Create(ctx, notification); err != nil {
		return err
	}
	s.inbox.publishSystemNotificationCreated(ctx, notification)
	return nil
}

func (s *RuleDispatcher) createEmailOutbox(ctx context.Context, recipientID uuid.UUID, preference *domainNotification.UserPreference, message preparedDispatch) error {
	return s.emailOutboxRepo.Create(ctx, &domainNotification.EmailOutbox{
		RecipientID: recipientID, RecipientEmail: strings.TrimSpace(preference.NotificationEmail),
		EventKey: message.eventKey, Subject: message.title, Body: message.body, Frequency: preference.Frequency,
		ResourceType: message.resourceType, ResourceID: message.input.ResourceID, Metadata: message.input.Metadata,
		Status: domainNotification.EmailOutboxStatusPending, NextAttemptAt: nextDeliveryAttemptAt(message.now, preference.Frequency),
	})
}

func (s *OutboxProcessor) ProcessDueEmailOutbox(ctx context.Context, now time.Time, limit int) error {
	if s.emailOutboxRepo == nil {
		return nil
	}
	items, err := s.emailOutboxRepo.GetDue(ctx, now.UTC(), limit)
	if err != nil {
		return err
	}
	if len(items) == 0 {
		return nil
	}

	for _, group := range groupEmailOutboxItems(items) {
		if err := s.sendEmailOutboxGroup(ctx, group); err != nil {
			attempts := maxOutboxAttempts(group.items) + 1
			nextAttemptAt := nextRetryAttemptAt(now, attempts)
			if markErr := s.emailOutboxRepo.MarkFailed(ctx, outboxIDs(group.items), attempts, truncateError(err), nextAttemptAt); markErr != nil {
				return markErr
			}
			continue
		}
		if err := s.emailOutboxRepo.MarkSent(ctx, outboxIDs(group.items), now.UTC()); err != nil {
			return err
		}
	}

	return nil
}

func (s *OutboxProcessor) StartEmailOutboxWorker(interval time.Duration, batchSize int) func() {
	if interval <= 0 {
		interval = time.Minute
	}
	if batchSize <= 0 {
		batchSize = 100
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		_ = s.ProcessDueEmailOutbox(ctx, time.Now().UTC(), batchSize)
		for {
			select {
			case <-ctx.Done():
				return
			case now := <-ticker.C:
				_ = s.ProcessDueEmailOutbox(ctx, now.UTC(), batchSize)
			}
		}
	}()
	return cancel
}

func (s *SettingsService) SendTestEmail(ctx context.Context, input domainNotification.SendTestEmailInput) error {
	if _, err := mail.ParseAddress(strings.TrimSpace(input.To)); err != nil {
		return domain.NewValidationError().Add("to", "must be a valid email")
	}

	settings, strategy, password, err := s.resolveSMTPStrategy(ctx, false)
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

func (s *SettingsService) SendNotification(ctx context.Context, input domainNotification.SendNotificationInput) error {
	settings, strategy, password, err := s.resolveSMTPStrategy(ctx, true)
	if err != nil {
		return err
	}
	return strategy.Send(ctx, settings, password, domainNotification.EmailMessage{
		To:       input.To,
		Subject:  input.Subject,
		TextBody: input.Body,
	})
}

func (s *SettingsService) resolveSMTPStrategy(ctx context.Context, requireEnabled bool) (*domainNotification.SMTPSettings, EmailStrategy, string, error) {
	settings, err := s.smtpSettingsRepo.GetByProvider(ctx, domainNotification.ProviderSMTP)
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

func (s *RuleDispatcher) resolveRuleRecipients(ctx context.Context, rules []domainNotification.NotificationRule, projectID *uuid.UUID) ([]uuid.UUID, error) {
	recipientIDs := make([]uuid.UUID, 0)
	for _, rule := range rules {
		switch rule.RecipientType {
		case domainNotification.RuleRecipientUsers:
			recipientIDs = append(recipientIDs, rule.RecipientUserIDs...)
		case domainNotification.RuleRecipientTeam:
			if rule.RecipientTeamID == nil || s.recipients.teamMemberReader == nil {
				continue
			}
			members, err := s.recipients.teamMemberReader.ListByTeam(ctx, *rule.RecipientTeamID, domain.PaginationParams{Page: 1, Limit: 1000})
			if err != nil {
				return nil, err
			}
			for _, member := range members.Items {
				recipientIDs = append(recipientIDs, member.UserID)
			}
		case domainNotification.RuleRecipientProjectUsers:
			scopedProjectID := projectID
			if rule.ProjectID != nil {
				scopedProjectID = rule.ProjectID
			}
			if scopedProjectID == nil || s.recipients.projectReader == nil {
				continue
			}
			users, err := s.recipients.projectReader.ListUsers(ctx, *scopedProjectID)
			if err != nil {
				return nil, err
			}
			for _, user := range users {
				recipientIDs = append(recipientIDs, user.ID)
			}
		case domainNotification.RuleRecipientProjectRole:
			scopedProjectID := projectID
			if rule.ProjectID != nil {
				scopedProjectID = rule.ProjectID
			}
			if scopedProjectID == nil || s.recipients.projectReader == nil || rule.RecipientRole == "" {
				continue
			}
			users, err := s.recipients.projectReader.ListUsers(ctx, *scopedProjectID)
			if err != nil {
				return nil, err
			}
			for _, user := range users {
				if user.Role == rule.RecipientRole {
					recipientIDs = append(recipientIDs, user.ID)
				}
			}
		}
	}
	return dedupeUUIDs(recipientIDs), nil
}

func validateNotificationRule(rule *domainNotification.NotificationRule) error {
	ve := domain.NewValidationError()
	if strings.TrimSpace(rule.Name) == "" {
		ve = ve.Add("name", "is required")
	}
	if _, err := rule.EventKey.Normalize(); err != nil {
		ve = ve.Add("event_key", "is required")
	}
	if !rule.RecipientType.Valid() {
		ve = ve.Add("recipient_type", "must be one of: users team project_users project_role")
	}
	switch rule.RecipientType {
	case domainNotification.RuleRecipientUsers:
		if len(rule.RecipientUserIDs) == 0 {
			ve = ve.Add("recipient_user_ids", "is required")
		}
	case domainNotification.RuleRecipientTeam:
		if rule.RecipientTeamID == nil || *rule.RecipientTeamID == uuid.Nil {
			ve = ve.Add("recipient_team_id", "is required")
		}
	case domainNotification.RuleRecipientProjectRole:
		if rule.RecipientRole == "" {
			ve = ve.Add("recipient_role", "is required")
		}
	}
	if len(ve.Fields) == 0 {
		return nil
	}
	return ve
}

type notificationTemplate struct {
	Title string
	Body  string
}

const defaultNotificationLocale = "de_CH"

var localizedNotificationTemplates = map[string]map[string]notificationTemplate{
	defaultNotificationLocale: {
		"project.updated": {
			Title: "Projekt aktualisiert",
			Body:  "Im Projekt {{project_name}} wurden Änderungen gespeichert.",
		},
		"project.deleted": {
			Title: "Projekt gelöscht",
			Body:  "Das Projekt {{project_name}} wurde gelöscht.",
		},
		"project.user.invited": {
			Title: "Projektmitglied hinzugefügt",
			Body:  "Ein Benutzer wurde zum Projekt {{project_name}} hinzugefügt.",
		},
		"project.user.removed": {
			Title: "Projektmitglied entfernt",
			Body:  "Ein Benutzer wurde aus dem Projekt {{project_name}} entfernt.",
		},
		"project.phase.changed": {
			Title: "Projektphase geändert",
			Body:  "Die Phase von {{project_name}} wurde von {{old}} auf {{new}} geändert.",
		},
		"project.control_cabinet.created": {
			Title: "Schaltschrank hinzugefügt",
			Body:  "Im Projekt {{project_name}} wurde ein Schaltschrank hinzugefügt.",
		},
		"project.control_cabinet.updated": {
			Title: "Schaltschrank aktualisiert",
			Body:  "Im Projekt {{project_name}} wurde ein Schaltschrank aktualisiert.",
		},
		"project.control_cabinet.deleted": {
			Title: "Schaltschrank entfernt",
			Body:  "Im Projekt {{project_name}} wurde ein Schaltschrank entfernt.",
		},
		"project.sps_controller.created": {
			Title: "SPS-Regler hinzugefügt",
			Body:  "Im Projekt {{project_name}} wurde ein SPS-Regler hinzugefügt.",
		},
		"project.sps_controller.updated": {
			Title: "SPS-Regler aktualisiert",
			Body:  "Im Projekt {{project_name}} wurde ein SPS-Regler aktualisiert.",
		},
		"project.sps_controller.deleted": {
			Title: "SPS-Regler entfernt",
			Body:  "Im Projekt {{project_name}} wurde ein SPS-Regler entfernt.",
		},
		"project.sps_controller.ip_address.changed": {
			Title: "SPS-Regler {{name}}: IP-Adresse geändert",
			Body:  "Die IP-Adresse wurde von {{old}} auf {{new}} geändert.",
		},
		"project.field_device.created": {
			Title: "Feldgerät hinzugefügt",
			Body:  "Im Projekt {{project_name}} wurde ein Feldgerät hinzugefügt.",
		},
		"project.field_device.updated": {
			Title: "Feldgerät aktualisiert",
			Body:  "Im Projekt {{project_name}} wurde ein Feldgerät aktualisiert.",
		},
		"project.field_device.deleted": {
			Title: "Feldgerät entfernt",
			Body:  "Im Projekt {{project_name}} wurde ein Feldgerät entfernt.",
		},
		"project.field_device.multi_created": {
			Title: "Feldgeräte hinzugefügt",
			Body:  "Im Projekt {{project_name}} wurden {{count}} Feldgeräte hinzugefügt.",
		},
		"project.object_data.created": {
			Title: "Objektdaten verknüpft",
			Body:  "Im Projekt {{project_name}} wurden Objektdaten verknüpft.",
		},
		"project.object_data.deleted": {
			Title: "Objektdaten entfernt",
			Body:  "Im Projekt {{project_name}} wurden Objektdaten entfernt.",
		},
	},
}

func (*TemplateRenderer) Render(locale string, eventKey domainNotification.EventKey, title, body string, metadata map[string]string) (string, string) {
	title = strings.TrimSpace(title)
	body = strings.TrimSpace(body)
	normalizedEventKey := strings.TrimSpace(eventKey.String())
	if template, ok := notificationTemplateFor(strings.TrimSpace(locale), normalizedEventKey); ok {
		if title == "" {
			title = template.Title
		}
		if body == "" {
			body = template.Body
		}
	}
	if title == "" {
		title = normalizedEventKey
	}
	return applyTemplateValues(title, metadata), applyTemplateValues(body, metadata)
}

func notificationTemplateFor(locale, eventKey string) (notificationTemplate, bool) {
	if templates, ok := localizedNotificationTemplates[locale]; ok {
		if template, ok := templates[eventKey]; ok {
			return template, true
		}
	}
	if templates, ok := localizedNotificationTemplates[defaultNotificationLocale]; ok {
		template, ok := templates[eventKey]
		return template, ok
	}
	return notificationTemplate{}, false
}

func applyTemplateValues(value string, metadata map[string]string) string {
	if metadata == nil {
		return value
	}
	result := value
	for key, raw := range metadata {
		result = strings.ReplaceAll(result, "{{"+key+"}}", raw)
		result = strings.ReplaceAll(result, "{"+key+"}", raw)
	}
	return result
}

func nextDeliveryAttemptAt(now time.Time, frequency domainNotification.DeliveryFrequency) time.Time {
	now = now.UTC()
	switch frequency {
	case domainNotification.DeliveryFrequencyHourly:
		return now.Truncate(time.Hour).Add(time.Hour)
	case domainNotification.DeliveryFrequencyDaily:
		return nextClockTime(now, 8, 0)
	case domainNotification.DeliveryFrequencyWeekly:
		next := nextClockTime(now, 8, 0)
		for next.Weekday() != time.Monday {
			next = next.AddDate(0, 0, 1)
		}
		return next
	default:
		return now
	}
}

func nextClockTime(now time.Time, hour, minute int) time.Time {
	candidate := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, time.UTC)
	if !candidate.After(now) {
		candidate = candidate.AddDate(0, 0, 1)
	}
	return candidate
}

func nextRetryAttemptAt(now time.Time, attempts int) time.Time {
	if attempts >= domainNotification.MaxEmailOutboxAttempts {
		return now.UTC()
	}
	if attempts <= 0 {
		attempts = 1
	}
	return now.UTC().Add(time.Duration(attempts*5) * time.Minute)
}

type emailOutboxGroup struct {
	items []domainNotification.EmailOutbox
}

func groupEmailOutboxItems(items []domainNotification.EmailOutbox) []emailOutboxGroup {
	groups := make([]emailOutboxGroup, 0, len(items))
	digestGroups := make(map[string]int)
	for _, item := range items {
		if item.Frequency == domainNotification.DeliveryFrequencyImmediate {
			groups = append(groups, emailOutboxGroup{items: []domainNotification.EmailOutbox{item}})
			continue
		}
		key := item.RecipientID.String() + "|" + string(item.Frequency) + "|" + item.RecipientEmail
		index, ok := digestGroups[key]
		if !ok {
			digestGroups[key] = len(groups)
			groups = append(groups, emailOutboxGroup{})
			index = len(groups) - 1
		}
		groups[index].items = append(groups[index].items, item)
	}
	return groups
}

func (s *OutboxProcessor) sendEmailOutboxGroup(ctx context.Context, group emailOutboxGroup) error {
	if len(group.items) == 0 {
		return nil
	}
	first := group.items[0]
	if first.Frequency == domainNotification.DeliveryFrequencyImmediate && len(group.items) == 1 {
		return s.settings.SendNotification(ctx, domainNotification.SendNotificationInput{
			To:      []string{first.RecipientEmail},
			Subject: first.Subject,
			Body:    first.Body,
		})
	}

	subject := fmt.Sprintf("Infra Link: %d Benachrichtigungen", len(group.items))
	body := renderDigestBody(group.items)
	return s.settings.SendNotification(ctx, domainNotification.SendNotificationInput{
		To:      []string{first.RecipientEmail},
		Subject: subject,
		Body:    body,
	})
}

func renderDigestBody(items []domainNotification.EmailOutbox) string {
	var builder strings.Builder
	builder.WriteString("Ihre gesammelten Benachrichtigungen:\n\n")
	for index, item := range items {
		fmt.Fprintf(&builder, "%d. %s\n", index+1, item.Subject)
		if strings.TrimSpace(item.Body) != "" {
			builder.WriteString(strings.TrimSpace(item.Body))
			builder.WriteString("\n")
		}
		builder.WriteString("\n")
	}
	return strings.TrimSpace(builder.String())
}

func outboxIDs(items []domainNotification.EmailOutbox) []uuid.UUID {
	ids := make([]uuid.UUID, 0, len(items))
	for _, item := range items {
		ids = append(ids, item.ID)
	}
	return ids
}

func maxOutboxAttempts(items []domainNotification.EmailOutbox) int {
	maxAttempts := 0
	for _, item := range items {
		if item.Attempts > maxAttempts {
			maxAttempts = item.Attempts
		}
	}
	return maxAttempts
}

func truncateError(err error) string {
	if err == nil {
		return ""
	}
	message := err.Error()
	if len(message) <= 1000 {
		return message
	}
	return message[:1000]
}

func dedupeUUIDs(ids []uuid.UUID) []uuid.UUID {
	seen := make(map[uuid.UUID]struct{}, len(ids))
	result := make([]uuid.UUID, 0, len(ids))
	for _, id := range ids {
		if id == uuid.Nil {
			continue
		}
		if _, ok := seen[id]; ok {
			continue
		}
		seen[id] = struct{}{}
		result = append(result, id)
	}
	return result
}

func normalizeVerifiedRecipient(email string) string {
	trimmed := domainNotification.NormalizeNotificationEmail(email)
	if trimmed == "" {
		return ""
	}
	parsed, err := mail.ParseAddress(trimmed)
	if err != nil {
		return trimmed
	}
	return parsed.Address
}

func generateVerificationCode() (string, error) {
	value, err := rand.Int(rand.Reader, big.NewInt(1000000))
	if err != nil {
		return "", fmt.Errorf("generate verification code: %w", err)
	}
	return fmt.Sprintf("%06d", value.Int64()), nil
}

func (s *PreferencesService) hashVerificationCode(userID uuid.UUID, email, code string) string {
	mac := hmac.New(sha256.New, s.verificationKey)
	_, _ = mac.Write([]byte(userID.String()))
	_, _ = mac.Write([]byte(":"))
	_, _ = mac.Write([]byte(strings.ToLower(strings.TrimSpace(email))))
	_, _ = mac.Write([]byte(":"))
	_, _ = mac.Write([]byte(strings.TrimSpace(code)))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *SettingsService) resolveEncryptedPassword(existing *domainNotification.SMTPSettings, rawPassword string, authMode domainNotification.AuthMode) (string, error) {
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
