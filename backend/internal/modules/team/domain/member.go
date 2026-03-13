package team

import (
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type MemberRole string

const (
	MemberRoleMember  MemberRole = "member"
	MemberRoleManager MemberRole = "manager"
	MemberRoleOwner   MemberRole = "owner"
)

type TeamMember struct {
	domain.Base
	TeamID   uuid.UUID  `gorm:"type:uuid;not null;index:idx_team_user,unique"`
	UserID   uuid.UUID  `gorm:"type:uuid;not null;index:idx_team_user,unique"`
	Role     MemberRole `gorm:"type:varchar(50);not null"`
	JoinedAt time.Time  `gorm:"not null"`
}
