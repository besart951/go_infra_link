package auth

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type PasswordResetToken struct {
	domain.Base
	UserID           uuid.UUID  `gorm:"type:uuid;not null;index"`
	TokenHash        string     `gorm:"uniqueIndex;not null"`
	TokenSalt        string     `gorm:"not null"`
	ExpiresAt        time.Time  `gorm:"not null;index"`
	UsedAt           *time.Time `gorm:"index"`
	CreatedByAdminID *uuid.UUID `gorm:"type:uuid"`
}
