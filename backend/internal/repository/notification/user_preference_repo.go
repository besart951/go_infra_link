package notification

import (
	"context"
	"errors"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

type userPreferenceRepo struct {
	db *gorm.DB
}

func NewUserPreferenceRepository(db *gorm.DB) domainNotification.UserPreferenceRepository {
	return &userPreferenceRepo{db: db}
}

func (r *userPreferenceRepo) GetByUserID(ctx context.Context, userID uuid.UUID) (*domainNotification.UserPreference, error) {
	var preference domainNotification.UserPreference
	err := r.db.WithContext(ctx).Where("user_id = ?", userID).First(&preference).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, domain.ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &preference, nil
}

func (r *userPreferenceRepo) Save(ctx context.Context, preference *domainNotification.UserPreference) error {
	now := time.Now().UTC()
	existing, err := r.GetByUserID(ctx, preference.UserID)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return err
	}

	if existing == nil {
		if err := preference.InitForCreate(now); err != nil {
			return err
		}
		return r.db.WithContext(ctx).Create(preference).Error
	}

	preference.ID = existing.ID
	preference.CreatedAt = existing.CreatedAt
	preference.TouchForUpdate(now)

	return r.db.WithContext(ctx).Model(&domainNotification.UserPreference{}).
		Where("id = ?", existing.ID).
		Updates(map[string]any{
			"updated_at":                     preference.UpdatedAt,
			"notification_email":             preference.NotificationEmail,
			"notification_email_verified_at": preference.NotificationEmailVerifiedAt,
			"email_verification_code_hash":   preference.EmailVerificationCodeHash,
			"email_verification_expires_at":  preference.EmailVerificationExpiresAt,
			"email_verification_sent_at":     preference.EmailVerificationSentAt,
			"channel":                        preference.Channel,
			"frequency":                      preference.Frequency,
		}).Error
}
