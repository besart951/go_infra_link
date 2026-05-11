package user

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestDeleteByIdsCleansUserDependencies(t *testing.T) {
	ctx := context.Background()
	db := newUserRepoTestDB(t)
	repo := NewUserRepository(db)
	now := time.Now().UTC()

	actor := seedUserRepoUser(t, db, "actor@example.com", nil)
	target := seedUserRepoUser(t, db, "target@example.com", &actor.ID)
	child := seedUserRepoUser(t, db, "child@example.com", &target.ID)

	seedUserRepoRecord(t, db, &domainAuth.RefreshToken{
		UserID:    target.ID,
		TokenHash: "target-refresh-token",
		ExpiresAt: now.Add(time.Hour),
	})
	seedUserRepoRecord(t, db, &domainUser.BusinessDetails{
		UserID:      target.ID,
		CompanyName: "Target AG",
	})
	seedUserRepoRecord(t, db, &domainUser.UserTeam{
		UserID:   target.ID,
		TeamID:   uuid.New(),
		JoinedAt: now,
	})
	seedUserRepoRecord(t, db, &domainTeam.TeamMember{
		TeamID:   uuid.New(),
		UserID:   target.ID,
		Role:     domainTeam.MemberRoleMember,
		JoinedAt: now,
	})
	seedUserRepoRecord(t, db, &domainUser.UserInvitation{
		UserID:      target.ID,
		CreatedByID: &target.ID,
		TokenHash:   "target-invitation",
		ExpiresAt:   now.Add(time.Hour),
	})
	seedUserRepoRecord(t, db, &domainNotification.UserPreference{
		UserID:    target.ID,
		Channel:   domainNotification.DeliveryChannelBoth,
		Frequency: domainNotification.DeliveryFrequencyImmediate,
	})
	seedUserRepoRecord(t, db, &domainNotification.SystemNotification{
		RecipientID: target.ID,
		EventKey:    "user.deleted",
		Title:       "Delete target",
		Body:        "target notification",
	})
	seedUserRepoRecord(t, db, &domainNotification.SystemNotification{
		RecipientID: actor.ID,
		ActorID:     &target.ID,
		EventKey:    "user.updated",
		Title:       "Actor target",
		Body:        "actor notification",
	})
	seedUserRepoRecord(t, db, &domainNotification.EmailOutbox{
		RecipientID:    target.ID,
		RecipientEmail: target.EmailValue(),
		EventKey:       "user.invited",
		Subject:        "Invitation",
		Body:           "Invitation body",
		Frequency:      domainNotification.DeliveryFrequencyImmediate,
		Status:         domainNotification.EmailOutboxStatusPending,
		NextAttemptAt:  now,
	})
	seedUserRepoRecord(t, db, &domainNotification.NotificationRule{
		Name:          "Target rule",
		Enabled:       true,
		EventKey:      "user.updated",
		RecipientType: domainNotification.RuleRecipientUsers,
		CreatedByID:   &target.ID,
	})
	if err := db.Exec(
		"INSERT INTO project_users (project_id, user_id) VALUES (?, ?)",
		uuid.New(), target.ID,
	).Error; err != nil {
		t.Fatalf("expected project membership setup to succeed, got %v", err)
	}

	if err := repo.DeleteByIds(ctx, []uuid.UUID{target.ID}); err != nil {
		t.Fatalf("expected dependent user delete to succeed, got %v", err)
	}

	assertUserRepoCount(t, db, &domainUser.User{}, "id = ?", 0, target.ID)
	assertUserRepoCount(t, db, &domainAuth.RefreshToken{}, "user_id = ?", 0, target.ID)
	assertUserRepoCount(t, db, &domainUser.BusinessDetails{}, "user_id = ?", 0, target.ID)
	assertUserRepoCount(t, db, &domainUser.UserTeam{}, "user_id = ?", 0, target.ID)
	assertUserRepoCount(t, db, &domainTeam.TeamMember{}, "user_id = ?", 0, target.ID)
	assertUserRepoCount(t, db, &domainUser.UserInvitation{}, "user_id = ?", 0, target.ID)
	assertUserRepoCount(t, db, &domainNotification.UserPreference{}, "user_id = ?", 0, target.ID)
	assertUserRepoCount(t, db, &domainNotification.SystemNotification{}, "recipient_id = ?", 0, target.ID)
	assertUserRepoCount(t, db, &domainNotification.EmailOutbox{}, "recipient_id = ?", 0, target.ID)
	assertUserRepoTableCount(t, db, "project_users", "user_id = ?", 0, target.ID)

	var reloadedChild domainUser.User
	if err := db.First(&reloadedChild, "id = ?", child.ID).Error; err != nil {
		t.Fatalf("expected child user to remain, got %v", err)
	}
	if reloadedChild.CreatedByID != nil {
		t.Fatalf("expected child creator reference to be cleared, got %s", *reloadedChild.CreatedByID)
	}

	var actorNotification domainNotification.SystemNotification
	if err := db.First(&actorNotification, "recipient_id = ?", actor.ID).Error; err != nil {
		t.Fatalf("expected actor notification to remain, got %v", err)
	}
	if actorNotification.ActorID != nil {
		t.Fatalf("expected notification actor reference to be cleared, got %s", *actorNotification.ActorID)
	}

	var rule domainNotification.NotificationRule
	if err := db.First(&rule, "name = ?", "Target rule").Error; err != nil {
		t.Fatalf("expected notification rule to remain, got %v", err)
	}
	if rule.CreatedByID != nil {
		t.Fatalf("expected rule creator reference to be cleared, got %s", *rule.CreatedByID)
	}
}

func TestUsersAllowMultipleAnonymizedNullEmails(t *testing.T) {
	db := newUserRepoTestDB(t)
	now := time.Now().UTC()

	first := &domainUser.User{
		FirstName:    "Deleted",
		LastName:     "User",
		Email:        nil,
		Password:     "hashed-password",
		IsActive:     false,
		Role:         domainUser.RolePlaner,
		AnonymizedAt: &now,
	}
	second := &domainUser.User{
		FirstName:    "Deleted",
		LastName:     "User",
		Email:        nil,
		Password:     "hashed-password",
		IsActive:     false,
		Role:         domainUser.RolePlaner,
		AnonymizedAt: &now,
	}

	seedUserRepoRecord(t, db, first)
	seedUserRepoRecord(t, db, second)
}

func newUserRepoTestDB(t *testing.T) *gorm.DB {
	t.Helper()
	name := strings.NewReplacer("/", "_", " ", "_", "#", "_").Replace(t.Name())
	db, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared&_foreign_keys=on", name)), &gorm.Config{})
	if err != nil {
		t.Fatalf("expected sqlite db to open, got %v", err)
	}
	if err := db.AutoMigrate(
		&domainUser.User{},
		&domainUser.BusinessDetails{},
		&domainUser.UserTeam{},
		&domainUser.UserInvitation{},
		&domainAuth.RefreshToken{},
		&domainTeam.TeamMember{},
		&domainNotification.UserPreference{},
		&domainNotification.SystemNotification{},
		&domainNotification.EmailOutbox{},
		&domainNotification.NotificationRule{},
	); err != nil {
		t.Fatalf("expected user repository schema migration to succeed, got %v", err)
	}
	if err := db.Exec(`
		CREATE TABLE project_users (
			project_id text NOT NULL,
			user_id text NOT NULL,
			PRIMARY KEY (project_id, user_id)
		)
	`).Error; err != nil {
		t.Fatalf("expected project user table setup to succeed, got %v", err)
	}
	return db
}

func seedUserRepoUser(t *testing.T, db *gorm.DB, email string, createdByID *uuid.UUID) *domainUser.User {
	t.Helper()
	usr := &domainUser.User{
		FirstName:   "Test",
		LastName:    "User",
		Email:       domainUser.EmailPtr(email),
		Password:    "hashed-password",
		IsActive:    true,
		Role:        domainUser.RolePlaner,
		CreatedByID: createdByID,
	}
	seedUserRepoRecord(t, db, usr)
	return usr
}

func seedUserRepoRecord[T interface{ GetBase() *domain.Base }](t *testing.T, db *gorm.DB, entity T) T {
	t.Helper()
	if err := entity.GetBase().InitForCreate(time.Now().UTC()); err != nil {
		t.Fatalf("expected entity init to succeed, got %v", err)
	}
	if err := db.Create(entity).Error; err != nil {
		t.Fatalf("expected entity setup to succeed, got %v", err)
	}
	return entity
}

func assertUserRepoCount(t *testing.T, db *gorm.DB, model any, where string, want int64, args ...any) {
	t.Helper()
	var got int64
	if err := db.Model(model).Where(where, args...).Count(&got).Error; err != nil {
		t.Fatalf("expected count query to succeed, got %v", err)
	}
	if got != want {
		t.Fatalf("expected %s count %d, got %d", where, want, got)
	}
}

func assertUserRepoTableCount(t *testing.T, db *gorm.DB, table string, where string, want int64, args ...any) {
	t.Helper()
	var got int64
	if err := db.Table(table).Where(where, args...).Count(&got).Error; err != nil {
		t.Fatalf("expected table count query to succeed, got %v", err)
	}
	if got != want {
		t.Fatalf("expected %s.%s count %d, got %d", table, where, want, got)
	}
}
