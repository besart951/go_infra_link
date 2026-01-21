package auth

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type RefreshToken struct {
	domain.Base
	UserID      uuid.UUID
	User        user.User
	TokenHash   string
	ExpiresAt   time.Time
	RevokedAt   *time.Time
	CreatedByIP *string
	UserAgent   *string
}
