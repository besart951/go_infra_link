package auth

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
)

type RefreshTokenRepository interface {
	Create(token *RefreshToken) error
	GetByTokenHash(tokenHash string) (*RefreshToken, error)
	RevokeByTokenHash(tokenHash string, revokedAt time.Time) error
	DeleteExpired(before time.Time) error
}

type LoginAttemptRepository interface {
	Create(attempt *LoginAttempt) error
	GetPaginatedList(params domain.PaginationParams) (*domain.PaginatedList[LoginAttempt], error)
}

type PasswordResetTokenRepository interface {
	Create(token *PasswordResetToken) error
	GetByTokenHash(tokenHash string) (*PasswordResetToken, error)
	MarkUsedByTokenHash(tokenHash string, usedAt time.Time) (bool, error)
}
