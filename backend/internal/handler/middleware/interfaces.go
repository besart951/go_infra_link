package middleware

import (
	domainTeam "github.com/besart951/go_infra_link/backend/internal/modules/team/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/modules/user/domain"
	"github.com/google/uuid"
)

// AuthorizationChecker provides role-based authorization queries.
type AuthorizationChecker interface {
	GetGlobalRole(userID uuid.UUID) (domainUser.Role, error)
	GetTeamRole(teamID, userID uuid.UUID) (*domainTeam.MemberRole, error)
	HasPermission(role domainUser.Role, permission string) (bool, error)
}
