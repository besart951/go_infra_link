package auth

import "time"

type RefreshTokenRepository interface {
	Create(token *RefreshToken) error
	GetByTokenHash(tokenHash string) (*RefreshToken, error)
	RevokeByTokenHash(tokenHash string, revokedAt time.Time) error
	DeleteExpired(before time.Time) error
}
