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
	TeamID   uuid.UUID
	UserID   uuid.UUID
	Role     MemberRole
	JoinedAt time.Time
}
