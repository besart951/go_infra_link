package auth

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type RefreshToken struct {
	domain.Base
	UserID      uuid.UUID `gorm:"index;not null"`
	User        user.User `gorm:"foreignKey:UserID"`
	TokenHash   string    `gorm:"size:64;uniqueIndex;not null"`
	ExpiresAt   time.Time `gorm:"index;not null"`
	RevokedAt   *time.Time
	CreatedByIP *string `gorm:"size:45"`
	UserAgent   *string `gorm:"size:255"`
}
