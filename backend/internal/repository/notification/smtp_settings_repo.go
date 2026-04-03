package notification

import (
	"errors"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainNotification "github.com/besart951/go_infra_link/backend/internal/domain/notification"
	"gorm.io/gorm"
)

type smtpSettingsRepo struct {
	db *gorm.DB
}

func NewSMTPSettingsRepository(db *gorm.DB) domainNotification.SMTPSettingsRepository {
	return &smtpSettingsRepo{db: db}
}

func (r *smtpSettingsRepo) GetByProvider(provider domainNotification.Provider) (*domainNotification.SMTPSettings, error) {
	var settings domainNotification.SMTPSettings
	result := r.db.Where("provider = ?", provider).Limit(1).Find(&settings)
	if result.Error != nil {
		return nil, result.Error
	}
	if result.RowsAffected == 0 {
		return nil, domain.ErrNotFound
	}
	return &settings, nil
}

func (r *smtpSettingsRepo) Save(settings *domainNotification.SMTPSettings) error {
	now := time.Now().UTC()
	existing, err := r.GetByProvider(settings.Provider)
	if err != nil && !errors.Is(err, domain.ErrNotFound) {
		return err
	}

	if existing == nil {
		if err := settings.GetBase().InitForCreate(now); err != nil {
			return err
		}
		return r.db.Create(settings).Error
	}

	settings.ID = existing.ID
	settings.CreatedAt = existing.CreatedAt
	settings.GetBase().TouchForUpdate(now)

	updates := map[string]any{
		"updated_at":         settings.UpdatedAt,
		"enabled":            settings.Enabled,
		"host":               settings.Host,
		"port":               settings.Port,
		"username":           settings.Username,
		"password_encrypted": settings.PasswordEncrypted,
		"from_email":         settings.FromEmail,
		"from_name":          settings.FromName,
		"reply_to":           settings.ReplyTo,
		"security":           settings.Security,
		"auth_mode":          settings.AuthMode,
		"allow_insecure_tls": settings.AllowInsecureTLS,
		"updated_by_id":      settings.UpdatedByID,
	}

	return r.db.Model(&domainNotification.SMTPSettings{}).
		Where("id = ?", existing.ID).
		Updates(updates).Error
}
