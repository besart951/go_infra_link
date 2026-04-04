package team

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	"github.com/google/uuid"
)

type TeamService interface {
	Create(ctx context.Context, team *domainTeam.Team) error
	GetByID(ctx context.Context, id uuid.UUID) (*domainTeam.Team, error)
	List(ctx context.Context, page, limit int, search string) (*domain.PaginatedList[domainTeam.Team], error)
	Update(ctx context.Context, team *domainTeam.Team) error
	DeleteByID(ctx context.Context, id uuid.UUID) error

	AddMember(ctx context.Context, teamID, userID uuid.UUID, role domainTeam.MemberRole) error
	RemoveMember(ctx context.Context, teamID, userID uuid.UUID) error
	ListMembers(ctx context.Context, teamID uuid.UUID, page, limit int) (*domain.PaginatedList[domainTeam.TeamMember], error)
}
