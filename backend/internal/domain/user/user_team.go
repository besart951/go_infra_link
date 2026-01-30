package user

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

// UserTeam represents the many-to-many relationship between users and teams
type UserTeam struct {
	domain.Base
	UserID   uuid.UUID `gorm:"type:uuid;not null;index:idx_user_team,unique"`
	TeamID   uuid.UUID `gorm:"type:uuid;not null;index:idx_user_team,unique"`
	JoinedAt time.Time `gorm:"not null"`
}
