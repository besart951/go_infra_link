package auth

import (
	"context"
	"time"
)

type RefreshTokenRepository interface {
	Create(ctx context.Context, token *RefreshToken) error
	GetByTokenHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevokeByTokenHash(ctx context.Context, tokenHash string, revokedAt time.Time) error
	DeleteExpired(ctx context.Context, before time.Time) error
}
