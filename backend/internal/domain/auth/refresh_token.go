package auth

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type RefreshToken struct {
	domain.Base
	UserID      uuid.UUID  `gorm:"type:uuid;not null;index"`
	User        user.User  `gorm:"foreignKey:UserID"`
	TokenHash   string     `gorm:"uniqueIndex;not null"`
	ExpiresAt   time.Time  `gorm:"not null;index"`
	RevokedAt   *time.Time `gorm:"index"`
	CreatedByIP *string
	UserAgent   *string
}
