package team

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type TeamRepository = domain.Repository[Team]

// TeamMemberRepository is used for authorization and membership management.
//
// Note: Keep this interface small; add methods only as needed by services.
// Consumers should define their own interface when possible.

type TeamMemberRepository interface {
	GetUserRole(teamID, userID uuid.UUID) (*MemberRole, error)
	ListByTeam(teamID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[TeamMember], error)
	Upsert(member *TeamMember) error
	Delete(teamID, userID uuid.UUID) error
}
