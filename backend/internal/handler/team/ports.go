package team

import (
	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	"github.com/google/uuid"
)

type TeamService interface {
	Create(team *domainTeam.Team) error
	GetByID(id uuid.UUID) (*domainTeam.Team, error)
	List(page, limit int, search string) (*domain.PaginatedList[domainTeam.Team], error)
	Update(team *domainTeam.Team) error
	DeleteByID(id uuid.UUID) error

	AddMember(teamID, userID uuid.UUID, role domainTeam.MemberRole) error
	RemoveMember(teamID, userID uuid.UUID) error
	ListMembers(teamID uuid.UUID, page, limit int) (*domain.PaginatedList[domainTeam.TeamMember], error)
}
