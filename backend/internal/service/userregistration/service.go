package userregistration

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"log/slog"
	"net/mail"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

const (
	defaultInvitationTTL            = 7 * 24 * time.Hour
	defaultInvitationResendCooldown = 2 * time.Minute
	defaultStaleInvitationRetention = 30 * 24 * time.Hour
	defaultAppPublicURL             = "http://localhost:5173"
	registrationEventKey            = "user.registration.invitation"
	pendingPasswordMarker           = "pending_registration"
)

type Store interface {
	CreatePendingRegistration(ctx context.Context, usr *domainUser.User, invitation *domainUser.UserInvitation, outbox *domainNotification.EmailOutbox) error
	GetInvitationByUserID(ctx context.Context, userID uuid.UUID) (*domainUser.UserInvitation, error)
	GetInvitationByTokenHash(ctx context.Context, tokenHash string) (*domainUser.UserInvitation, error)
	ListInvitationsByUserIDs(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]*domainUser.UserInvitation, error)
	GetUserByID(ctx context.Context, userID uuid.UUID) (*domainUser.User, error)
	GetUserByEmail(ctx context.Context, email string) (*domainUser.User, error)
	GetEmailOutboxByID(ctx context.Context, id uuid.UUID) (*domainNotification.EmailOutbox, error)
	ListEmailOutboxByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domainNotification.EmailOutbox, error)
	ResendInvitation(ctx context.Context, invitation *domainUser.UserInvitation, outbox *domainNotification.EmailOutbox, now time.Time, cooldown time.Duration) error
	CompleteRegistration(ctx context.Context, invitation *domainUser.UserInvitation, usr *domainUser.User) error
	InvalidateInvitationToken(ctx context.Context, invitationID uuid.UUID) error
	ClearExpiredTokenHashes(ctx context.Context, now time.Time) error
	DeleteStaleUnaccepted(ctx context.Context, cutoff time.Time) error
}

type MutationPolicy interface {
	CanInviteUser(ctx context.Context, actorID uuid.UUID, targetRole domainUser.Role) error
	CanReadRegistrationProcess(ctx context.Context, actorID uuid.UUID, target domainUser.User) error
}

type Service struct {
	store          Store
	policy         MutationPolicy
	passwords      domainUser.PasswordHasher
	emailBuilder   InvitationEmailBuilder
	invitationTTL  time.Duration
	resendCooldown time.Duration
	staleRetention time.Duration
	now            func() time.Time
}

type InviteInput struct {
	ActorID uuid.UUID
	Email   string
	Role    domainUser.Role
}

type CompleteInput struct {
	Token      string
	FirstName  string
	LastName   string
	Password   string
	PrivacyAck bool
}

type PublicRegistrationView struct {
	Email     string
	Role      domainUser.Role
	ExpiresAt time.Time
}

type ProcessStep struct {
	Key       string
	Label     string
	Status    string
	Timestamp *time.Time
}

type Process struct {
	Status             string
	EmailStatus        string
	Steps              []ProcessStep
	CanResend          bool
	BlocksManualEnable bool
	ExpiresAt          *time.Time
	AcceptedAt         *time.Time
	LastSentAt         *time.Time
	ResendAvailableAt  *time.Time
	SendCount          int
	LastError          string
}

func New(store Store, policy MutationPolicy, passwords domainUser.PasswordHasher, appPublicURL string) *Service {
	return &Service{
		store:          store,
		policy:         policy,
		passwords:      passwords,
		emailBuilder:   NewInvitationEmailBuilder(appPublicURL),
		invitationTTL:  defaultInvitationTTL,
		resendCooldown: defaultInvitationResendCooldown,
		staleRetention: defaultStaleInvitationRetention,
		now:            func() time.Time { return time.Now().UTC() },
	}
}

func (s *Service) CreateInvitation(ctx context.Context, input InviteInput) (*domainUser.User, *Process, error) {
	email, err := normalizeEmail(input.Email)
	if err != nil {
		return nil, nil, err
	}
	if s.policy == nil {
		return nil, nil, domainUser.ErrRoleNotAssignable
	}
	if err := s.policy.CanInviteUser(ctx, input.ActorID, input.Role); err != nil {
		return nil, nil, err
	}
	if existing, err := s.store.GetUserByEmail(ctx, email); err == nil && existing != nil {
		return nil, nil, domain.ErrConflict
	} else if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, nil, err
	}

	token, tokenHash, err := generateInvitationToken()
	if err != nil {
		return nil, nil, err
	}
	now := s.now()
	expiresAt := now.Add(s.invitationTTL)

	usr := &domainUser.User{
		FirstName:   "",
		LastName:    "",
		Email:       domainUser.EmailPtr(email),
		Password:    pendingPasswordMarker,
		IsActive:    false,
		Role:        input.Role,
		CreatedByID: &input.ActorID,
	}
	invitation := &domainUser.UserInvitation{
		CreatedByID: &input.ActorID,
		TokenHash:   tokenHash,
		ExpiresAt:   expiresAt,
		EmailStatus: domainUser.InvitationEmailStatusPending,
		SendCount:   1,
	}
	outbox := s.emailBuilder.Build(usr.ID, email, token, expiresAt, now)

	if err := s.store.CreatePendingRegistration(ctx, usr, invitation, outbox); err != nil {
		return nil, nil, err
	}
	process := s.buildProcess(usr, invitation, outbox)
	return usr, process, nil
}

func (s *Service) GetProcess(ctx context.Context, actorID, userID uuid.UUID) (*Process, error) {
	usr, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if s.policy == nil {
		return nil, domainUser.ErrRoleNotAssignable
	}
	if err := s.policy.CanReadRegistrationProcess(ctx, actorID, *usr); err != nil {
		return nil, err
	}

	invitation, err := s.store.GetInvitationByUserID(ctx, userID)
	if errors.Is(err, domain.ErrNotFound) {
		return buildDirectProcess(usr), nil
	}
	if err != nil {
		return nil, err
	}

	outbox, err := s.latestOutbox(ctx, invitation)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	return s.buildProcess(usr, invitation, outbox), nil
}

func (s *Service) ListProcessesForUsers(ctx context.Context, users []domainUser.User) (map[uuid.UUID]*Process, error) {
	result := make(map[uuid.UUID]*Process, len(users))
	if len(users) == 0 {
		return result, nil
	}
	userIDs := make([]uuid.UUID, 0, len(users))
	for _, usr := range users {
		userIDs = append(userIDs, usr.ID)
	}
	invitations, err := s.store.ListInvitationsByUserIDs(ctx, userIDs)
	if err != nil {
		return nil, err
	}
	outboxIDs := make([]uuid.UUID, 0, len(invitations))
	for _, invitation := range invitations {
		if invitation.LatestOutboxID != nil && *invitation.LatestOutboxID != uuid.Nil {
			outboxIDs = append(outboxIDs, *invitation.LatestOutboxID)
		}
	}
	outboxes, err := s.store.ListEmailOutboxByIDs(ctx, outboxIDs)
	if err != nil {
		return nil, err
	}
	for i := range users {
		usr := users[i]
		invitation := invitations[usr.ID]
		if invitation == nil {
			result[usr.ID] = buildDirectProcess(&usr)
			continue
		}
		var outbox *domainNotification.EmailOutbox
		if invitation.LatestOutboxID != nil {
			outbox = outboxes[*invitation.LatestOutboxID]
		}
		result[usr.ID] = s.buildProcess(&usr, invitation, outbox)
	}
	return result, nil
}

func (s *Service) ResendInvitation(ctx context.Context, actorID, userID uuid.UUID) (*Process, error) {
	usr, err := s.store.GetUserByID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if s.policy == nil {
		return nil, domainUser.ErrRoleNotAssignable
	}
	if err := s.policy.CanInviteUser(ctx, actorID, usr.Role); err != nil {
		return nil, err
	}
	invitation, err := s.store.GetInvitationByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	if invitation.AcceptedAt != nil {
		return nil, domainUser.ErrRegistrationAlreadyAccepted
	}
	outbox, err := s.latestOutbox(ctx, invitation)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return nil, err
	}
	if !s.canResendInvitation(invitation, outbox) {
		return nil, domainUser.ErrRegistrationResendTooSoon
	}

	token, tokenHash, err := generateInvitationToken()
	if err != nil {
		return nil, err
	}
	now := s.now()
	expiresAt := now.Add(s.invitationTTL)
	invitation.TokenHash = tokenHash
	invitation.ExpiresAt = expiresAt
	invitation.SendCount++

	outbox = s.emailBuilder.Build(usr.ID, usr.EmailValue(), token, expiresAt, now)
	if err := s.store.ResendInvitation(ctx, invitation, outbox, now, s.resendCooldown); err != nil {
		return nil, err
	}
	return s.buildProcess(usr, invitation, outbox), nil
}

func (s *Service) GetPublicRegistration(ctx context.Context, token string) (*PublicRegistrationView, error) {
	invitation, usr, err := s.lookupValidInvitation(ctx, token)
	if err != nil {
		return nil, err
	}
	return &PublicRegistrationView{
		Email:     usr.EmailValue(),
		Role:      usr.Role,
		ExpiresAt: invitation.ExpiresAt,
	}, nil
}

func (s *Service) CompleteRegistration(ctx context.Context, input CompleteInput) (*domainUser.User, error) {
	if !input.PrivacyAck {
		return nil, domain.NewValidationError().Add("privacy_ack", "is required")
	}
	firstName := strings.TrimSpace(input.FirstName)
	lastName := strings.TrimSpace(input.LastName)
	password := input.Password
	ve := domain.NewValidationError()
	if firstName == "" {
		ve = ve.Add("first_name", "is required")
	}
	if lastName == "" {
		ve = ve.Add("last_name", "is required")
	}
	if len(password) < 8 {
		ve = ve.Add("password", "must be at least 8 characters")
	}
	if password != strings.TrimSpace(password) {
		ve = ve.Add("password", "must not start or end with whitespace")
	}
	if len(ve.Fields) > 0 {
		return nil, ve
	}

	invitation, usr, err := s.lookupValidInvitation(ctx, input.Token)
	if err != nil {
		return nil, err
	}

	hashedPassword, err := s.passwords.Hash(password)
	if err != nil {
		return nil, domainUser.ErrPasswordHashingFailed
	}
	now := s.now()
	invitation.AcceptedAt = &now
	invitation.PrivacyAckAt = &now
	usr.FirstName = firstName
	usr.LastName = lastName
	usr.Password = hashedPassword
	usr.IsActive = true
	usr.DisabledAt = nil
	usr.LockedUntil = nil

	if err := s.store.CompleteRegistration(ctx, invitation, usr); err != nil {
		return nil, err
	}
	return usr, nil
}

func (s *Service) CleanupExpired(ctx context.Context) error {
	now := s.now()
	if err := s.store.ClearExpiredTokenHashes(ctx, now); err != nil {
		return err
	}
	return s.store.DeleteStaleUnaccepted(ctx, now.Add(-s.staleRetention))
}

func (s *Service) StartCleanupWorker(interval time.Duration) func() {
	if interval <= 0 {
		interval = 24 * time.Hour
	}
	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		ticker := time.NewTicker(interval)
		defer ticker.Stop()
		s.cleanupExpiredWithLog(ctx)
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				s.cleanupExpiredWithLog(ctx)
			}
		}
	}()
	return cancel
}

func (s *Service) cleanupExpiredWithLog(ctx context.Context) {
	if err := s.CleanupExpired(ctx); err != nil {
		slog.Warn("user registration cleanup failed", "err", err)
	}
}

func (s *Service) lookupValidInvitation(ctx context.Context, token string) (*domainUser.UserInvitation, *domainUser.User, error) {
	tokenHash := hashToken(strings.TrimSpace(token))
	if tokenHash == "" {
		return nil, nil, domainUser.ErrRegistrationTokenInvalid
	}
	invitation, err := s.store.GetInvitationByTokenHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, nil, domainUser.ErrRegistrationTokenInvalid
		}
		return nil, nil, err
	}
	if invitation.AcceptedAt != nil {
		return nil, nil, domainUser.ErrRegistrationAlreadyAccepted
	}
	if s.now().After(invitation.ExpiresAt) {
		if err := s.store.InvalidateInvitationToken(ctx, invitation.ID); err != nil {
			return nil, nil, err
		}
		return nil, nil, domainUser.ErrRegistrationTokenExpired
	}
	usr, err := s.store.GetUserByID(ctx, invitation.UserID)
	if err != nil {
		return nil, nil, err
	}
	// Prevent registration completion if user is deleted
	if usr.IsDeleted() {
		return nil, nil, domainUser.ErrRegistrationUserDeleted
	}
	return invitation, usr, nil
}

func (s *Service) latestOutbox(ctx context.Context, invitation *domainUser.UserInvitation) (*domainNotification.EmailOutbox, error) {
	if invitation == nil || invitation.LatestOutboxID == nil || *invitation.LatestOutboxID == uuid.Nil {
		return nil, domain.ErrNotFound
	}
	return s.store.GetEmailOutboxByID(ctx, *invitation.LatestOutboxID)
}

func (s *Service) buildProcess(usr *domainUser.User, invitation *domainUser.UserInvitation, outbox *domainNotification.EmailOutbox) *Process {
	now := s.now()
	emailStatus, lastSentAt, lastError := deriveEmailState(invitation, outbox)
	expired := invitation.AcceptedAt == nil && now.After(invitation.ExpiresAt)
	if expired {
		emailStatus = "expired"
	}

	status := "pending"
	switch {
	case usr.LastLoginAt != nil:
		status = "first_login"
	case invitation.AcceptedAt != nil:
		status = "registered"
	case expired:
		status = "expired"
	case emailStatus == string(domainUser.InvitationEmailStatusFailed):
		status = "email_failed"
	}

	canResend := s.canResendInvitation(invitation, outbox)
	expiresAt := invitation.ExpiresAt
	return &Process{
		Status:             status,
		EmailStatus:        emailStatus,
		CanResend:          canResend,
		BlocksManualEnable: invitation.AcceptedAt == nil,
		ExpiresAt:          &expiresAt,
		AcceptedAt:         invitation.AcceptedAt,
		LastSentAt:         lastSentAt,
		ResendAvailableAt:  s.resendAvailableAt(invitation, outbox, emailStatus, lastSentAt),
		SendCount:          invitation.SendCount,
		LastError:          lastError,
		Steps: []ProcessStep{
			{Key: "created", Label: "Angelegt", Status: "completed", Timestamp: &usr.CreatedAt},
			emailStep(emailStatus, lastSentAt),
			registrationStep(invitation.AcceptedAt, emailStatus),
			firstLoginStep(usr.LastLoginAt, invitation.AcceptedAt),
		},
	}
}

func (s *Service) canResendInvitation(invitation *domainUser.UserInvitation, outbox *domainNotification.EmailOutbox) bool {
	if invitation == nil || invitation.AcceptedAt != nil {
		return false
	}
	now := s.now()
	if now.After(invitation.ExpiresAt) {
		return true
	}
	emailStatus, sentAt, _ := deriveEmailState(invitation, outbox)
	if emailStatus == string(domainUser.InvitationEmailStatusFailed) || emailStatus == "expired" {
		return true
	}
	if emailStatus != string(domainUser.InvitationEmailStatusPending) && emailStatus != string(domainUser.InvitationEmailStatusSent) {
		return false
	}
	lastAttemptAt := resendActivityAt(invitation, outbox, sentAt)
	if lastAttemptAt == nil {
		return true
	}
	return !now.Before(lastAttemptAt.Add(s.resendCooldown))
}

func (s *Service) resendAvailableAt(invitation *domainUser.UserInvitation, outbox *domainNotification.EmailOutbox, emailStatus string, sentAt *time.Time) *time.Time {
	if invitation == nil || invitation.AcceptedAt != nil {
		return nil
	}
	now := s.now()
	if now.After(invitation.ExpiresAt) {
		return nil
	}
	if emailStatus == string(domainUser.InvitationEmailStatusFailed) || emailStatus == "expired" {
		return nil
	}
	if emailStatus != string(domainUser.InvitationEmailStatusPending) && emailStatus != string(domainUser.InvitationEmailStatusSent) {
		return nil
	}
	lastAttemptAt := resendActivityAt(invitation, outbox, sentAt)
	if lastAttemptAt == nil {
		return nil
	}
	availableAt := lastAttemptAt.Add(s.resendCooldown)
	if !now.Before(availableAt) {
		return nil
	}
	return &availableAt
}

func buildDirectProcess(usr *domainUser.User) *Process {
	status := "registered"
	if usr.LastLoginAt != nil {
		status = "first_login"
	}
	return &Process{
		Status:      status,
		EmailStatus: "not_applicable",
		Steps: []ProcessStep{
			{Key: "created", Label: "Angelegt", Status: "completed", Timestamp: &usr.CreatedAt},
			{Key: "email_sent", Label: "E-Mail versendet", Status: "skipped"},
			{Key: "registered", Label: "Registriert", Status: "completed"},
			firstLoginStep(usr.LastLoginAt, &usr.CreatedAt),
		},
	}
}

func resendActivityAt(invitation *domainUser.UserInvitation, outbox *domainNotification.EmailOutbox, sentAt *time.Time) *time.Time {
	if sentAt != nil {
		return sentAt
	}
	if outbox != nil {
		return &outbox.CreatedAt
	}
	if invitation == nil {
		return nil
	}
	if invitation.LastSentAt != nil {
		return invitation.LastSentAt
	}
	return &invitation.UpdatedAt
}

func deriveEmailState(invitation *domainUser.UserInvitation, outbox *domainNotification.EmailOutbox) (string, *time.Time, string) {
	if outbox != nil {
		switch outbox.Status {
		case domainNotification.EmailOutboxStatusSent:
			return string(domainUser.InvitationEmailStatusSent), outbox.SentAt, ""
		case domainNotification.EmailOutboxStatusFailed:
			return string(domainUser.InvitationEmailStatusFailed), nil, outbox.LastError
		default:
			return string(domainUser.InvitationEmailStatusPending), nil, outbox.LastError
		}
	}
	return string(domainUser.NormalizeInvitationEmailStatus(invitation.EmailStatus)), invitation.LastSentAt, invitation.LastError
}

func emailStep(emailStatus string, sentAt *time.Time) ProcessStep {
	status := "pending"
	if emailStatus == "sent" {
		status = "completed"
	}
	if emailStatus == "failed" || emailStatus == "expired" {
		status = "failed"
	}
	return ProcessStep{Key: "email_sent", Label: "E-Mail versendet", Status: status, Timestamp: sentAt}
}

func registrationStep(acceptedAt *time.Time, emailStatus string) ProcessStep {
	status := "pending"
	if acceptedAt != nil {
		status = "completed"
	} else if emailStatus == "sent" {
		status = "current"
	} else if emailStatus == "failed" || emailStatus == "expired" {
		status = "blocked"
	}
	return ProcessStep{Key: "registered", Label: "Registriert", Status: status, Timestamp: acceptedAt}
}

func firstLoginStep(lastLoginAt *time.Time, acceptedAt *time.Time) ProcessStep {
	status := "pending"
	if lastLoginAt != nil {
		status = "completed"
	} else if acceptedAt != nil {
		status = "current"
	}
	return ProcessStep{Key: "first_login", Label: "Erste Anmeldung", Status: status, Timestamp: lastLoginAt}
}

func normalizeEmail(value string) (string, error) {
	parsed, err := mail.ParseAddress(strings.TrimSpace(value))
	if err != nil {
		return "", domain.NewValidationError().Add("email", "must be a valid email")
	}
	return strings.ToLower(strings.TrimSpace(parsed.Address)), nil
}

func generateInvitationToken() (string, string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", "", err
	}
	token := hex.EncodeToString(b)
	return token, hashToken(token), nil
}

func hashToken(token string) string {
	token = strings.TrimSpace(token)
	if token == "" {
		return ""
	}
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}
