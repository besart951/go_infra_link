package team

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	"github.com/google/uuid"
)

type TeamRepository = domain.Repository[Team]

// TeamMemberRepository is used for authorization and membership management.
//
// Note: Keep this interface small; add methods only as needed by services.
// Consumers should define their own interface when possible.

type TeamMemberRepository interface {
	GetUserRole(ctx context.Context, teamID, userID uuid.UUID) (*MemberRole, error)
	ListByTeam(ctx context.Context, teamID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[TeamMember], error)
	ListByUser(ctx context.Context, userID uuid.UUID, params domain.PaginationParams) (*domain.PaginatedList[TeamMember], error)
	Upsert(ctx context.Context, member *TeamMember) error
	Delete(ctx context.Context, teamID, userID uuid.UUID) error
}
