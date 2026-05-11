package userregistration

import (
	"context"
	"errors"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

func TestCreateInvitationQueuesPendingUserWithHashOnlyToken(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 8, 10, 30, 0, 0, time.UTC)
	store := &registrationStoreStub{}
	service := newTestService(store, now)
	actorID := uuid.New()

	usr, process, err := service.CreateInvitation(ctx, InviteInput{
		ActorID: actorID,
		Email:   " Invited.Person@Example.COM ",
		Role:    domainUser.RolePlaner,
	})
	if err != nil {
		t.Fatalf("CreateInvitation returned error: %v", err)
	}

	if usr.EmailValue() != "invited.person@example.com" {
		t.Fatalf("expected normalized email, got %q", usr.EmailValue())
	}
	if usr.IsActive {
		t.Fatalf("invited user must stay inactive until registration is completed")
	}
	if usr.FirstName != "" || usr.LastName != "" {
		t.Fatalf("admin invite should not persist profile names")
	}
	if usr.Password != pendingPasswordMarker {
		t.Fatalf("expected pending password marker, got %q", usr.Password)
	}
	if store.invitation.TokenHash == "" || len(store.invitation.TokenHash) != 64 {
		t.Fatalf("expected sha256 token hash, got %q", store.invitation.TokenHash)
	}
	if store.invitation.SendCount != 1 {
		t.Fatalf("expected initial send count 1, got %d", store.invitation.SendCount)
	}
	if !store.invitation.ExpiresAt.Equal(now.Add(defaultInvitationTTL)) {
		t.Fatalf("expected 7 day expiry, got %s", store.invitation.ExpiresAt)
	}
	rawToken := tokenFromOutboxBody(t, store.outbox.Body)
	if rawToken == store.invitation.TokenHash {
		t.Fatalf("raw token must not be stored as token_hash")
	}
	if hashToken(rawToken) != store.invitation.TokenHash {
		t.Fatalf("outbox link token does not match stored token hash")
	}
	if store.outbox.EventKey != registrationEventKey {
		t.Fatalf("expected registration outbox event key, got %q", store.outbox.EventKey)
	}
	if store.outbox.RecipientEmail != usr.EmailValue() {
		t.Fatalf("expected outbox recipient email %q, got %q", usr.EmailValue(), store.outbox.RecipientEmail)
	}
	if process == nil || len(process.Steps) != 4 {
		t.Fatalf("expected 4-step process, got %#v", process)
	}
	if process.Status != "pending" || process.CanResend {
		t.Fatalf("expected pending process inside resend cooldown, got status=%q can_resend=%v", process.Status, process.CanResend)
	}
}

func TestCreateInvitationRejectsUnassignableRole(t *testing.T) {
	ctx := context.Background()
	store := &registrationStoreStub{}
	policy := &registrationMutationPolicyStub{err: domainUser.ErrRoleNotAssignable}
	service := New(store, policy, passwordHasherStub{}, "https://infra-link.example")
	actorID := uuid.New()

	_, _, err := service.CreateInvitation(ctx, InviteInput{
		ActorID: actorID,
		Email:   "person@example.com",
		Role:    domainUser.RoleSuperAdmin,
	})

	if !errors.Is(err, domainUser.ErrRoleNotAssignable) {
		t.Fatalf("expected ErrRoleNotAssignable, got %v", err)
	}
	if store.user != nil {
		t.Fatalf("user should not be created when role is outside requester scope")
	}
}

func TestResendInvitationRotatesTokenAndOutbox(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 8, 10, 30, 0, 0, time.UTC)
	store := seededRegistrationStore(now)
	service := newTestService(store, now)
	oldTokenHash := store.invitation.TokenHash
	oldOutboxID := *store.invitation.LatestOutboxID

	process, err := service.ResendInvitation(ctx, uuid.New(), store.user.ID)
	if err != nil {
		t.Fatalf("ResendInvitation returned error: %v", err)
	}

	if store.invitation.TokenHash == "" || store.invitation.TokenHash == oldTokenHash {
		t.Fatalf("expected rotated token hash, got %q", store.invitation.TokenHash)
	}
	if store.invitation.SendCount != 2 {
		t.Fatalf("expected incremented send count, got %d", store.invitation.SendCount)
	}
	if store.invitation.LatestOutboxID == nil || *store.invitation.LatestOutboxID == oldOutboxID {
		t.Fatalf("expected new latest outbox id")
	}
	if !store.invitation.ExpiresAt.Equal(now.Add(defaultInvitationTTL)) {
		t.Fatalf("expected refreshed expiry, got %s", store.invitation.ExpiresAt)
	}
	if hashToken(tokenFromOutboxBody(t, store.outbox.Body)) != store.invitation.TokenHash {
		t.Fatalf("new outbox link token does not match rotated token hash")
	}
	if process == nil || process.SendCount != 2 || process.ExpiresAt == nil || !process.ExpiresAt.Equal(store.invitation.ExpiresAt) {
		t.Fatalf("expected process to reflect resend state, got %#v", process)
	}
}

func TestResendInvitationRejectsPendingInsideCooldown(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 8, 10, 30, 0, 0, time.UTC)
	store := seededRegistrationStore(now)
	store.outbox.CreatedAt = now.Add(-time.Minute)
	service := newTestService(store, now)
	oldTokenHash := store.invitation.TokenHash
	oldSendCount := store.invitation.SendCount

	_, err := service.ResendInvitation(ctx, uuid.New(), store.user.ID)

	if !errors.Is(err, domainUser.ErrRegistrationResendTooSoon) {
		t.Fatalf("expected resend cooldown error, got %v", err)
	}
	if store.invitation.TokenHash != oldTokenHash {
		t.Fatalf("token hash must not rotate inside cooldown")
	}
	if store.invitation.SendCount != oldSendCount {
		t.Fatalf("send count must not change inside cooldown")
	}
}

func TestGetProcessReportsResendAvailableAtInsideCooldown(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 8, 10, 30, 0, 0, time.UTC)
	store := seededRegistrationStore(now)
	store.outbox.CreatedAt = now.Add(-time.Minute)
	service := newTestService(store, now)

	process, err := service.GetProcess(ctx, uuid.New(), store.user.ID)
	if err != nil {
		t.Fatalf("GetProcess returned error: %v", err)
	}

	if process.CanResend {
		t.Fatalf("expected resend to be blocked inside cooldown")
	}
	wantAvailableAt := store.outbox.CreatedAt.Add(defaultInvitationResendCooldown)
	if process.ResendAvailableAt == nil || !process.ResendAvailableAt.Equal(wantAvailableAt) {
		t.Fatalf("expected resend_available_at %s, got %v", wantAvailableAt, process.ResendAvailableAt)
	}
}

func TestResendInvitationAllowsFailedInsideCooldown(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 8, 10, 30, 0, 0, time.UTC)
	store := seededRegistrationStore(now)
	store.outbox.CreatedAt = now.Add(-time.Minute)
	store.outbox.Status = domainNotification.EmailOutboxStatusFailed
	service := newTestService(store, now)

	_, err := service.ResendInvitation(ctx, uuid.New(), store.user.ID)

	if err != nil {
		t.Fatalf("expected failed email resend despite cooldown, got %v", err)
	}
	if store.invitation.SendCount != 2 {
		t.Fatalf("expected send count increment, got %d", store.invitation.SendCount)
	}
}

func TestCompleteRegistrationActivatesUserAndInvalidatesToken(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 8, 10, 30, 0, 0, time.UTC)
	store := seededRegistrationStore(now)
	service := newTestService(store, now)

	usr, err := service.CompleteRegistration(ctx, CompleteInput{
		Token:      "valid-token",
		FirstName:  " Ada ",
		LastName:   " Lovelace ",
		Password:   "CorrectHorse1",
		PrivacyAck: true,
	})
	if err != nil {
		t.Fatalf("CompleteRegistration returned error: %v", err)
	}

	if !usr.IsActive {
		t.Fatalf("registered user should be active")
	}
	if usr.FirstName != "Ada" || usr.LastName != "Lovelace" {
		t.Fatalf("expected trimmed profile names, got %q %q", usr.FirstName, usr.LastName)
	}
	if usr.Password != "hashed:CorrectHorse1" {
		t.Fatalf("expected hashed password, got %q", usr.Password)
	}
	if store.invitation.AcceptedAt == nil || !store.invitation.AcceptedAt.Equal(now) {
		t.Fatalf("expected accepted_at to be set to test clock")
	}
	if store.invitation.PrivacyAckAt == nil || !store.invitation.PrivacyAckAt.Equal(now) {
		t.Fatalf("expected privacy_ack_at to be set to test clock")
	}
	if store.invitation.TokenHash != "" {
		t.Fatalf("accepted invitation token hash must be cleared")
	}

	_, err = service.CompleteRegistration(ctx, CompleteInput{
		Token:      "valid-token",
		FirstName:  "Ada",
		LastName:   "Lovelace",
		Password:   "CorrectHorse1",
		PrivacyAck: true,
	})
	if !errors.Is(err, domainUser.ErrRegistrationTokenInvalid) {
		t.Fatalf("expected used token to be invalid, got %v", err)
	}
}

func TestCompleteRegistrationRejectsOuterPasswordWhitespace(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 8, 10, 30, 0, 0, time.UTC)
	store := seededRegistrationStore(now)
	service := newTestService(store, now)

	_, err := service.CompleteRegistration(ctx, CompleteInput{
		Token:      "valid-token",
		FirstName:  "Ada",
		LastName:   "Lovelace",
		Password:   " CorrectHorse1 ",
		PrivacyAck: true,
	})

	validationErr, ok := domain.AsValidationError(err)
	if !ok {
		t.Fatalf("expected validation error, got %v", err)
	}
	if validationErr.Fields["password"] == "" {
		t.Fatalf("expected password field error, got %#v", validationErr.Fields)
	}
}

func TestExpiredRegistrationInvalidatesTokenHash(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 8, 10, 30, 0, 0, time.UTC)
	store := seededRegistrationStore(now)
	store.invitation.ExpiresAt = now.Add(-time.Minute)
	service := newTestService(store, now)

	_, err := service.GetPublicRegistration(ctx, "valid-token")

	if !errors.Is(err, domainUser.ErrRegistrationTokenExpired) {
		t.Fatalf("expected expired token error, got %v", err)
	}
	if store.invalidatedInvitationID != store.invitation.ID {
		t.Fatalf("expected expired invitation to be invalidated")
	}
	if store.invitation.TokenHash != "" {
		t.Fatalf("expired invitation token hash must be cleared")
	}
}

func TestExpiredRegistrationReturnsInvalidationError(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 8, 10, 30, 0, 0, time.UTC)
	store := seededRegistrationStore(now)
	store.invitation.ExpiresAt = now.Add(-time.Minute)
	store.invalidateErr = errors.New("db unavailable")
	service := newTestService(store, now)

	_, err := service.GetPublicRegistration(ctx, "valid-token")

	if !errors.Is(err, store.invalidateErr) {
		t.Fatalf("expected invalidation error, got %v", err)
	}
	if store.invitation.TokenHash == "" {
		t.Fatalf("token hash must not be cleared when invalidation failed")
	}
}

func newTestService(store *registrationStoreStub, now time.Time) *Service {
	service := New(
		store,
		&registrationMutationPolicyStub{},
		passwordHasherStub{},
		"https://infra-link.example",
	)
	service.now = func() time.Time { return now }
	return service
}

func seededRegistrationStore(now time.Time) *registrationStoreStub {
	userID := uuid.New()
	outboxID := uuid.New()
	invitationID := uuid.New()
	return &registrationStoreStub{
		user: &domainUser.User{
			Base:     domain.Base{ID: userID, CreatedAt: now.Add(-time.Hour), UpdatedAt: now.Add(-time.Hour)},
			Email:    domainUser.EmailPtr("person@example.com"),
			Password: pendingPasswordMarker,
			IsActive: false,
			Role:     domainUser.RolePlaner,
		},
		invitation: &domainUser.UserInvitation{
			Base:           domain.Base{ID: invitationID, CreatedAt: now.Add(-time.Hour), UpdatedAt: now.Add(-time.Hour)},
			UserID:         userID,
			TokenHash:      hashToken("valid-token"),
			ExpiresAt:      now.Add(time.Hour),
			EmailStatus:    domainUser.InvitationEmailStatusPending,
			LatestOutboxID: &outboxID,
			SendCount:      1,
		},
		outbox: &domainNotification.EmailOutbox{
			Base:           domain.Base{ID: outboxID, CreatedAt: now.Add(-time.Hour), UpdatedAt: now.Add(-time.Hour)},
			RecipientID:    userID,
			RecipientEmail: "person@example.com",
			EventKey:       registrationEventKey,
			Status:         domainNotification.EmailOutboxStatusPending,
		},
	}
}

func TestCompleteRegistrationRejectsDeletedUser(t *testing.T) {
	ctx := context.Background()
	now := time.Date(2026, 5, 8, 10, 30, 0, 0, time.UTC)
	store := seededRegistrationStore(now)
	// Mark user as deleted (soft-delete)
	deletedAt := now.Add(-time.Hour)
	store.user.DeletedAt = &deletedAt
	service := newTestService(store, now)

	_, err := service.CompleteRegistration(ctx, CompleteInput{
		Token:      "valid-token",
		FirstName:  "Ada",
		LastName:   "Lovelace",
		Password:   "CorrectHorse1",
		PrivacyAck: true,
	})

	if !errors.Is(err, domainUser.ErrRegistrationUserDeleted) {
		t.Fatalf("expected ErrRegistrationUserDeleted, got %v", err)
	}
}

func tokenFromOutboxBody(t *testing.T, body string) string {
	t.Helper()
	const marker = "/register/"
	start := strings.Index(body, marker)
	if start < 0 {
		t.Fatalf("registration link missing from outbox body: %q", body)
	}
	start += len(marker)
	end := strings.IndexAny(body[start:], "\r\n \t")
	if end < 0 {
		end = len(body)
	} else {
		end += start
	}
	token, err := url.PathUnescape(strings.TrimSpace(body[start:end]))
	if err != nil {
		t.Fatalf("failed to unescape token from body: %v", err)
	}
	return token
}

type registrationStoreStub struct {
	user                    *domainUser.User
	invitation              *domainUser.UserInvitation
	outbox                  *domainNotification.EmailOutbox
	invalidatedInvitationID uuid.UUID
	invalidateErr           error
}

func (s *registrationStoreStub) CreatePendingRegistration(_ context.Context, usr *domainUser.User, invitation *domainUser.UserInvitation, outbox *domainNotification.EmailOutbox) error {
	now := time.Now().UTC()
	if err := usr.InitForCreate(now); err != nil {
		return err
	}
	if err := invitation.InitForCreate(now); err != nil {
		return err
	}
	if err := outbox.InitForCreate(now); err != nil {
		return err
	}
	invitation.UserID = usr.ID
	outbox.RecipientID = usr.ID
	invitation.LatestOutboxID = &outbox.ID
	s.user = usr
	s.invitation = invitation
	s.outbox = outbox
	return nil
}

func (s *registrationStoreStub) GetInvitationByUserID(_ context.Context, userID uuid.UUID) (*domainUser.UserInvitation, error) {
	if s.invitation == nil || s.invitation.UserID != userID {
		return nil, domain.ErrNotFound
	}
	return s.invitation, nil
}

func (s *registrationStoreStub) GetInvitationByTokenHash(_ context.Context, tokenHash string) (*domainUser.UserInvitation, error) {
	if tokenHash == "" || s.invitation == nil || s.invitation.TokenHash != tokenHash {
		return nil, domain.ErrNotFound
	}
	return s.invitation, nil
}

func (s *registrationStoreStub) ListInvitationsByUserIDs(_ context.Context, userIDs []uuid.UUID) (map[uuid.UUID]*domainUser.UserInvitation, error) {
	result := map[uuid.UUID]*domainUser.UserInvitation{}
	if s.invitation == nil {
		return result, nil
	}
	for _, userID := range userIDs {
		if s.invitation.UserID == userID {
			result[userID] = s.invitation
		}
	}
	return result, nil
}

func (s *registrationStoreStub) GetUserByID(_ context.Context, userID uuid.UUID) (*domainUser.User, error) {
	if s.user == nil || s.user.ID != userID {
		return nil, domain.ErrNotFound
	}
	return s.user, nil
}

func (s *registrationStoreStub) GetUserByEmail(_ context.Context, email string) (*domainUser.User, error) {
	if s.user == nil || s.user.EmailValue() != email {
		return nil, domain.ErrNotFound
	}
	return s.user, nil
}

func (s *registrationStoreStub) GetEmailOutboxByID(_ context.Context, id uuid.UUID) (*domainNotification.EmailOutbox, error) {
	if s.outbox == nil || s.outbox.ID != id {
		return nil, domain.ErrNotFound
	}
	return s.outbox, nil
}

func (s *registrationStoreStub) ListEmailOutboxByIDs(_ context.Context, ids []uuid.UUID) (map[uuid.UUID]*domainNotification.EmailOutbox, error) {
	result := map[uuid.UUID]*domainNotification.EmailOutbox{}
	if s.outbox == nil {
		return result, nil
	}
	for _, id := range ids {
		if s.outbox.ID == id {
			result[id] = s.outbox
		}
	}
	return result, nil
}

func (s *registrationStoreStub) ResendInvitation(_ context.Context, invitation *domainUser.UserInvitation, outbox *domainNotification.EmailOutbox, _ time.Time, _ time.Duration) error {
	if err := outbox.InitForCreate(time.Now().UTC()); err != nil {
		return err
	}
	invitation.LatestOutboxID = &outbox.ID
	s.invitation = invitation
	s.outbox = outbox
	return nil
}

func (s *registrationStoreStub) CompleteRegistration(_ context.Context, invitation *domainUser.UserInvitation, usr *domainUser.User) error {
	invitation.TokenHash = ""
	s.invitation = invitation
	s.user = usr
	return nil
}

func (s *registrationStoreStub) InvalidateInvitationToken(_ context.Context, invitationID uuid.UUID) error {
	s.invalidatedInvitationID = invitationID
	if s.invalidateErr != nil {
		return s.invalidateErr
	}
	if s.invitation != nil && s.invitation.ID == invitationID {
		s.invitation.TokenHash = ""
	}
	return nil
}

func (s *registrationStoreStub) ClearExpiredTokenHashes(_ context.Context, _ time.Time) error {
	return nil
}

func (s *registrationStoreStub) DeleteStaleUnaccepted(_ context.Context, _ time.Time) error {
	return nil
}

type registrationMutationPolicyStub struct {
	err error
}

func (s *registrationMutationPolicyStub) CanInviteUser(context.Context, uuid.UUID, domainUser.Role) error {
	return s.err
}

func (s *registrationMutationPolicyStub) CanReadRegistrationProcess(context.Context, uuid.UUID, domainUser.User) error {
	return s.err
}

type passwordHasherStub struct{}

func (passwordHasherStub) Hash(plain string) (string, error) {
	return "hashed:" + plain, nil
}

func (passwordHasherStub) Compare(_, _ string) error {
	return nil
}
