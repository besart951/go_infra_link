package userregistration

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestStoreResendInvitationRejectsStaleConcurrentSnapshotInsideCooldown(t *testing.T) {
	ctx := context.Background()
	db := newStoreTestDB(t)
	store := NewStore(db)
	now := time.Date(2026, 5, 8, 10, 30, 0, 0, time.UTC)
	userID, invitationID, initialOutboxID := uuid.New(), uuid.New(), uuid.New()
	initialAt := now.Add(-time.Hour)

	seedUserRegistration(t, db, userID, invitationID, initialOutboxID, initialAt, now.Add(time.Hour))

	staleInvitation := domainUser.UserInvitation{
		Base:      domain.Base{ID: invitationID},
		TokenHash: "new-token-hash-1",
		ExpiresAt: now.Add(7 * 24 * time.Hour),
		SendCount: 2,
	}
	firstOutbox := registrationOutbox(userID, "person@example.com", now)
	if err := store.ResendInvitation(ctx, &staleInvitation, firstOutbox, now, 2*time.Minute); err != nil {
		t.Fatalf("expected first stale snapshot resend to succeed, got %v", err)
	}

	secondStaleInvitation := domainUser.UserInvitation{
		Base:      domain.Base{ID: invitationID},
		TokenHash: "new-token-hash-2",
		ExpiresAt: now.Add(7 * 24 * time.Hour),
		SendCount: 2,
	}
	secondOutbox := registrationOutbox(userID, "person@example.com", now.Add(time.Second))
	err := store.ResendInvitation(ctx, &secondStaleInvitation, secondOutbox, now.Add(time.Second), 2*time.Minute)

	if !errors.Is(err, domainUser.ErrRegistrationResendTooSoon) {
		t.Fatalf("expected stale concurrent resend to hit cooldown, got %v", err)
	}

	var outboxCount int64
	if err := db.Model(&domainNotification.EmailOutbox{}).Count(&outboxCount).Error; err != nil {
		t.Fatalf("expected outbox count query to succeed, got %v", err)
	}
	if outboxCount != 2 {
		t.Fatalf("expected only initial + first resend outbox, got %d", outboxCount)
	}

	var invitation domainUser.UserInvitation
	if err := db.First(&invitation, "id = ?", invitationID).Error; err != nil {
		t.Fatalf("expected invitation reload to succeed, got %v", err)
	}
	if invitation.TokenHash != "new-token-hash-1" {
		t.Fatalf("expected first token hash to remain active, got %q", invitation.TokenHash)
	}
	if invitation.SendCount != 2 {
		t.Fatalf("expected send_count to increment once, got %d", invitation.SendCount)
	}
	if invitation.LatestOutboxID == nil || *invitation.LatestOutboxID != firstOutbox.ID {
		t.Fatalf("expected latest outbox to remain first resend, got %v want %s", invitation.LatestOutboxID, firstOutbox.ID)
	}
}

func newStoreTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", name)), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		t.Fatalf("expected sqlite db to open, got %v", err)
	}
	if err := db.AutoMigrate(&domainUser.User{}, &domainNotification.EmailOutbox{}, &domainUser.UserInvitation{}); err != nil {
		t.Fatalf("expected test schema migration to succeed, got %v", err)
	}
	return db
}

func seedUserRegistration(t *testing.T, db *gorm.DB, userID, invitationID, outboxID uuid.UUID, initialAt, expiresAt time.Time) {
	t.Helper()
	user := domainUser.User{
		Base:     domain.Base{ID: userID, CreatedAt: initialAt, UpdatedAt: initialAt},
		Email:    domainUser.EmailPtr("person@example.com"),
		Password: "pending_registration",
		IsActive: false,
		Role:     domainUser.RolePlaner,
	}
	outbox := domainNotification.EmailOutbox{
		Base:           domain.Base{ID: outboxID, CreatedAt: initialAt, UpdatedAt: initialAt},
		RecipientID:    userID,
		RecipientEmail: "person@example.com",
		EventKey:       "user.registration.invitation",
		Subject:        "Invite",
		Body:           "Invite",
		Frequency:      domainNotification.DeliveryFrequencyImmediate,
		Status:         domainNotification.EmailOutboxStatusPending,
		NextAttemptAt:  initialAt,
	}
	invitation := domainUser.UserInvitation{
		Base:           domain.Base{ID: invitationID, CreatedAt: initialAt, UpdatedAt: initialAt},
		UserID:         userID,
		TokenHash:      "old-token-hash",
		ExpiresAt:      expiresAt,
		EmailStatus:    domainUser.InvitationEmailStatusPending,
		LatestOutboxID: &outboxID,
		SendCount:      1,
	}
	for _, entity := range []any{&user, &outbox, &invitation} {
		if err := db.Create(entity).Error; err != nil {
			t.Fatalf("expected seed create to succeed, got %v", err)
		}
	}
}

func registrationOutbox(userID uuid.UUID, email string, nextAttemptAt time.Time) *domainNotification.EmailOutbox {
	return &domainNotification.EmailOutbox{
		RecipientID:    userID,
		RecipientEmail: email,
		EventKey:       "user.registration.invitation",
		Subject:        "Invite",
		Body:           "Invite",
		Frequency:      domainNotification.DeliveryFrequencyImmediate,
		Status:         domainNotification.EmailOutboxStatusPending,
		NextAttemptAt:  nextAttemptAt,
	}
}
