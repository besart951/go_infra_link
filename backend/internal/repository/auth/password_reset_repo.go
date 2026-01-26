package auth

import (
	"time"

	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	"gorm.io/gorm"
)

type passwordResetRepo struct {
	db *gorm.DB
}

func NewPasswordResetTokenRepository(db *gorm.DB) domainAuth.PasswordResetTokenRepository {
	return &passwordResetRepo{db: db}
}

func (r *passwordResetRepo) Create(token *domainAuth.PasswordResetToken) error {
	now := time.Now().UTC()
	if err := token.Base.InitForCreate(now); err != nil {
		return err
	}
	return r.db.Create(token).Error
}

func (r *passwordResetRepo) GetByTokenHash(tokenHash string) (*domainAuth.PasswordResetToken, error) {
	var token domainAuth.PasswordResetToken
	err := r.db.Where("deleted_at IS NULL").Where("token_hash = ?", tokenHash).First(&token).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func (r *passwordResetRepo) MarkUsedByTokenHash(tokenHash string, usedAt time.Time) (bool, error) {
	now := time.Now().UTC()
	result := r.db.Model(&domainAuth.PasswordResetToken{}).
		Where("deleted_at IS NULL AND token_hash = ? AND used_at IS NULL AND expires_at > ?", tokenHash, usedAt).
		Updates(map[string]any{"used_at": usedAt, "updated_at": now})
	if result.Error != nil {
		return false, result.Error
	}
	return result.RowsAffected > 0, nil
}
