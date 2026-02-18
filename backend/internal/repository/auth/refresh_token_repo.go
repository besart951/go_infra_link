package auth

import (
	"time"

	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	"gorm.io/gorm"
)

type refreshTokenRepo struct {
	db *gorm.DB
}

func NewRefreshTokenRepository(db *gorm.DB) domainAuth.RefreshTokenRepository {
	return &refreshTokenRepo{db: db}
}

func (r *refreshTokenRepo) Create(token *domainAuth.RefreshToken) error {
	now := time.Now().UTC()
	if err := token.Base.InitForCreate(now); err != nil {
		return err
	}
	return r.db.Create(token).Error
}

func (r *refreshTokenRepo) GetByTokenHash(tokenHash string) (*domainAuth.RefreshToken, error) {
	var token domainAuth.RefreshToken
	err := r.db.Where("token_hash = ?", tokenHash).First(&token).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil
		}
		return nil, err
	}
	return &token, nil
}

func (r *refreshTokenRepo) RevokeByTokenHash(tokenHash string, revokedAt time.Time) error {
	return r.db.Model(&domainAuth.RefreshToken{}).
		Where("token_hash = ?", tokenHash).
		Updates(map[string]any{"revoked_at": revokedAt, "updated_at": time.Now().UTC()}).Error
}

func (r *refreshTokenRepo) DeleteExpired(before time.Time) error {
	return r.db.
		Where("expires_at <= ?", before).
		Delete(&domainAuth.RefreshToken{}).Error
}
