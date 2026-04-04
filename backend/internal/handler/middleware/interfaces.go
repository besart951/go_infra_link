package middleware

import (
	"context"

	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

// AuthorizationChecker provides role-based authorization queries.
type AuthorizationChecker interface {
	GetGlobalRole(ctx context.Context, userID uuid.UUID) (domainUser.Role, error)
	GetTeamRole(ctx context.Context, teamID, userID uuid.UUID) (*domainTeam.MemberRole, error)
	HasPermission(ctx context.Context, role domainUser.Role, permission string) (bool, error)
}
