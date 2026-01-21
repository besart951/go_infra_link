package auth

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type PasswordResetToken struct {
	domain.Base
	UserID           uuid.UUID
	TokenHash        string
	TokenSalt        string
	ExpiresAt        time.Time
	UsedAt           *time.Time
	CreatedByAdminID *uuid.UUID
}
