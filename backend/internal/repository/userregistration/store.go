package userregistration

import (
	"context"
	"strings"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type Store struct {
	db *gorm.DB
}

func NewStore(db *gorm.DB) *Store {
	return &Store{db: db}
}

func (s *Store) CreatePendingRegistration(ctx context.Context, usr *domainUser.User, invitation *domainUser.UserInvitation, outbox *domainNotification.EmailOutbox) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		if err := usr.InitForCreate(now); err != nil {
			return err
		}
		if err := tx.Create(usr).Error; err != nil {
			return mapWriteError(err)
		}

		outbox.RecipientID = usr.ID
		if err := outbox.InitForCreate(now); err != nil {
			return err
		}
		if err := tx.Create(outbox).Error; err != nil {
			return mapWriteError(err)
		}

		invitation.UserID = usr.ID
		latestOutboxID := outbox.ID
		invitation.LatestOutboxID = &latestOutboxID
		if err := invitation.InitForCreate(now); err != nil {
			return err
		}
		return mapWriteError(tx.Create(invitation).Error)
	})
}

func (s *Store) GetInvitationByUserID(ctx context.Context, userID uuid.UUID) (*domainUser.UserInvitation, error) {
	var invitation domainUser.UserInvitation
	err := s.db.WithContext(ctx).Where("user_id = ?", userID).First(&invitation).Error
	return invitationOrError(invitation, err)
}

func (s *Store) GetInvitationByTokenHash(ctx context.Context, tokenHash string) (*domainUser.UserInvitation, error) {
	var invitation domainUser.UserInvitation
	err := s.db.WithContext(ctx).
		Where("token_hash = ? AND token_hash <> ''", strings.TrimSpace(tokenHash)).
		First(&invitation).Error
	return invitationOrError(invitation, err)
}

func (s *Store) ListInvitationsByUserIDs(ctx context.Context, userIDs []uuid.UUID) (map[uuid.UUID]*domainUser.UserInvitation, error) {
	if len(userIDs) == 0 {
		return map[uuid.UUID]*domainUser.UserInvitation{}, nil
	}
	var invitations []domainUser.UserInvitation
	if err := s.db.WithContext(ctx).Where("user_id IN ?", userIDs).Find(&invitations).Error; err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]*domainUser.UserInvitation, len(invitations))
	for i := range invitations {
		invitation := &invitations[i]
		result[invitation.UserID] = invitation
	}
	return result, nil
}

func (s *Store) GetUserByID(ctx context.Context, userID uuid.UUID) (*domainUser.User, error) {
	var usr domainUser.User
	err := s.db.WithContext(ctx).Where("id = ?", userID).First(&usr).Error
	return userOrError(usr, err)
}

func (s *Store) GetUserByEmail(ctx context.Context, email string) (*domainUser.User, error) {
	var usr domainUser.User
	err := s.db.WithContext(ctx).Where("LOWER(email) = ?", strings.ToLower(strings.TrimSpace(email))).First(&usr).Error
	return userOrError(usr, err)
}

func (s *Store) GetEmailOutboxByID(ctx context.Context, id uuid.UUID) (*domainNotification.EmailOutbox, error) {
	var outbox domainNotification.EmailOutbox
	err := s.db.WithContext(ctx).Where("id = ?", id).First(&outbox).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &outbox, nil
}

func (s *Store) ListEmailOutboxByIDs(ctx context.Context, ids []uuid.UUID) (map[uuid.UUID]*domainNotification.EmailOutbox, error) {
	if len(ids) == 0 {
		return map[uuid.UUID]*domainNotification.EmailOutbox{}, nil
	}
	var outboxes []domainNotification.EmailOutbox
	if err := s.db.WithContext(ctx).Where("id IN ?", ids).Find(&outboxes).Error; err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]*domainNotification.EmailOutbox, len(outboxes))
	for i := range outboxes {
		outbox := &outboxes[i]
		result[outbox.ID] = outbox
	}
	return result, nil
}

func (s *Store) ResendInvitation(ctx context.Context, invitation *domainUser.UserInvitation, outbox *domainNotification.EmailOutbox, now time.Time, cooldown time.Duration) error {
	if now.IsZero() {
		now = time.Now().UTC()
	} else {
		now = now.UTC()
	}
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var current domainUser.UserInvitation
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).
			Where("id = ?", invitation.ID).
			First(&current).Error
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				return domain.ErrNotFound
			}
			return err
		}
		if current.AcceptedAt != nil {
			return domainUser.ErrRegistrationAlreadyAccepted
		}

		latestOutbox, err := emailOutboxByID(tx, current.LatestOutboxID)
		if err != nil {
			return err
		}
		if !canResendInvitationAt(&current, latestOutbox, now, cooldown) {
			return domainUser.ErrRegistrationResendTooSoon
		}

		if err := outbox.InitForCreate(now); err != nil {
			return err
		}
		outbox.RecipientID = current.UserID
		if err := tx.Create(outbox).Error; err != nil {
			return mapWriteError(err)
		}

		latestOutboxID := outbox.ID
		invitation.UserID = current.UserID
		invitation.LatestOutboxID = &latestOutboxID
		invitation.EmailStatus = domainUser.InvitationEmailStatusPending
		invitation.LastSentAt = nil
		invitation.LastError = ""
		invitation.SendCount = current.SendCount + 1
		invitation.TouchForUpdate(now)

		result := tx.Model(&domainUser.UserInvitation{}).
			Where("id = ? AND accepted_at IS NULL", invitation.ID).
			Updates(map[string]any{
				"updated_at":       invitation.UpdatedAt,
				"token_hash":       invitation.TokenHash,
				"expires_at":       invitation.ExpiresAt,
				"email_status":     invitation.EmailStatus,
				"latest_outbox_id": latestOutboxID,
				"send_count":       invitation.SendCount,
				"last_sent_at":     nil,
				"last_error":       "",
			})
		if result.Error != nil {
			return result.Error
		}
		if result.RowsAffected == 0 {
			return domainUser.ErrRegistrationAlreadyAccepted
		}
		return nil
	})
}

func emailOutboxByID(tx *gorm.DB, id *uuid.UUID) (*domainNotification.EmailOutbox, error) {
	if id == nil || *id == uuid.Nil {
		return nil, nil
	}
	var outbox domainNotification.EmailOutbox
	err := tx.Where("id = ?", *id).First(&outbox).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &outbox, nil
}

func canResendInvitationAt(invitation *domainUser.UserInvitation, outbox *domainNotification.EmailOutbox, now time.Time, cooldown time.Duration) bool {
	if invitation == nil || invitation.AcceptedAt != nil {
		return false
	}
	if now.After(invitation.ExpiresAt) {
		return true
	}
	emailStatus, sentAt := deriveEmailStatus(invitation, outbox)
	if emailStatus == domainUser.InvitationEmailStatusFailed {
		return true
	}
	if emailStatus != domainUser.InvitationEmailStatusPending && emailStatus != domainUser.InvitationEmailStatusSent {
		return false
	}
	lastAttemptAt := resendActivityAt(invitation, outbox, sentAt)
	if lastAttemptAt == nil {
		return true
	}
	return !now.Before(lastAttemptAt.Add(cooldown))
}

func deriveEmailStatus(invitation *domainUser.UserInvitation, outbox *domainNotification.EmailOutbox) (domainUser.InvitationEmailStatus, *time.Time) {
	if outbox != nil {
		switch domainNotification.NormalizeEmailOutboxStatus(outbox.Status) {
		case domainNotification.EmailOutboxStatusSent:
			return domainUser.InvitationEmailStatusSent, outbox.SentAt
		case domainNotification.EmailOutboxStatusFailed:
			return domainUser.InvitationEmailStatusFailed, nil
		default:
			return domainUser.InvitationEmailStatusPending, nil
		}
	}
	return domainUser.NormalizeInvitationEmailStatus(invitation.EmailStatus), invitation.LastSentAt
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

func (s *Store) CompleteRegistration(ctx context.Context, invitation *domainUser.UserInvitation, usr *domainUser.User) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now().UTC()
		originalTokenHash := invitation.TokenHash
		invitation.TokenHash = ""
		invitation.TouchForUpdate(now)

		invitationResult := tx.Model(&domainUser.UserInvitation{}).
			Where("id = ? AND token_hash = ? AND accepted_at IS NULL", invitation.ID, originalTokenHash).
			Updates(map[string]any{
				"updated_at":     invitation.UpdatedAt,
				"token_hash":     "",
				"accepted_at":    invitation.AcceptedAt,
				"privacy_ack_at": invitation.PrivacyAckAt,
			})
		if invitationResult.Error != nil {
			return invitationResult.Error
		}
		if invitationResult.RowsAffected == 0 {
			return domainUser.ErrRegistrationTokenInvalid
		}

		usr.TouchForUpdate(now)
		return tx.Model(&domainUser.User{}).
			Where("id = ?", usr.ID).
			Updates(map[string]any{
				"updated_at":   usr.UpdatedAt,
				"first_name":   usr.FirstName,
				"last_name":    usr.LastName,
				"password":     usr.Password,
				"is_active":    usr.IsActive,
				"disabled_at":  usr.DisabledAt,
				"locked_until": usr.LockedUntil,
			}).Error
	})
}

func (s *Store) InvalidateInvitationToken(ctx context.Context, invitationID uuid.UUID) error {
	return s.db.WithContext(ctx).Model(&domainUser.UserInvitation{}).
		Where("id = ?", invitationID).
		Updates(map[string]any{
			"updated_at": time.Now().UTC(),
			"token_hash": "",
		}).Error
}

func (s *Store) ClearExpiredTokenHashes(ctx context.Context, now time.Time) error {
	return s.db.WithContext(ctx).Model(&domainUser.UserInvitation{}).
		Where("accepted_at IS NULL AND token_hash <> '' AND expires_at < ?", now.UTC()).
		Updates(map[string]any{
			"updated_at": time.Now().UTC(),
			"token_hash": "",
		}).Error
}

func (s *Store) DeleteStaleUnaccepted(ctx context.Context, cutoff time.Time) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var invitations []domainUser.UserInvitation
		if err := tx.
			Where("accepted_at IS NULL AND expires_at < ?", cutoff.UTC()).
			Find(&invitations).Error; err != nil {
			return err
		}
		if len(invitations) == 0 {
			return nil
		}

		invitationIDs := make([]uuid.UUID, 0, len(invitations))
		userIDs := make([]uuid.UUID, 0, len(invitations))
		for _, invitation := range invitations {
			invitationIDs = append(invitationIDs, invitation.ID)
			userIDs = append(userIDs, invitation.UserID)
		}

		if err := tx.Where("id IN ?", invitationIDs).Delete(&domainUser.UserInvitation{}).Error; err != nil {
			return err
		}
		return tx.
			Where("id IN ? AND is_active = ? AND last_login_at IS NULL", userIDs, false).
			Delete(&domainUser.User{}).Error
	})
}

func invitationOrError(invitation domainUser.UserInvitation, err error) (*domainUser.UserInvitation, error) {
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &invitation, nil
}

func userOrError(usr domainUser.User, err error) (*domainUser.User, error) {
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, domain.ErrNotFound
		}
		return nil, err
	}
	return &usr, nil
}
