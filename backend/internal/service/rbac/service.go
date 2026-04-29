package rbac

import (
	"context"

	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type Service struct {
	userRepo           domainUser.UserRepository
	memberRepo         domainTeam.TeamMemberRepository
	permissionRepo     domainUser.PermissionRepository
	rolePermissionRepo domainUser.RolePermissionRepository
}

func New(userRepo domainUser.UserRepository, memberRepo domainTeam.TeamMemberRepository, permissionRepo domainUser.PermissionRepository, rolePermissionRepo domainUser.RolePermissionRepository) *Service {
	return &Service{
		userRepo:           userRepo,
		memberRepo:         memberRepo,
		permissionRepo:     permissionRepo,
		rolePermissionRepo: rolePermissionRepo,
	}
}

func (s *Service) GetGlobalRole(ctx context.Context, userID uuid.UUID) (domainUser.Role, error) {
	users, err := s.userRepo.GetByIds(ctx, []uuid.UUID{userID})
	if err != nil {
		return "", err
	}
	if len(users) == 0 {
		return "", nil
	}
	return users[0].Role, nil
}

func (s *Service) GetTeamRole(ctx context.Context, teamID, userID uuid.UUID) (*domainTeam.MemberRole, error) {
	return s.memberRepo.GetUserRole(ctx, teamID, userID)
}

// GetRoleLevel returns the hierarchical level of a role (higher = more privileged)
func (s *Service) GetRoleLevel(role domainUser.Role) int {
	switch role {
	case domainUser.RoleSuperAdmin:
		return 100
	case domainUser.RoleAdminFZAG:
		return 90
	case domainUser.RoleFZAG:
		return 80
	case domainUser.RoleAdminPlaner:
		return 70
	case domainUser.RolePlaner:
		return 60
	case domainUser.RoleAdminEnterpreneur:
		return 50
	case domainUser.RoleEnterpreneur:
		return 40
	default:
		return 0
	}
}

func (s *Service) GetAllowedRoles(ctx context.Context, requesterRole domainUser.Role) ([]domainUser.Role, error) {
	roles := domainUser.AllRoles()
	permissionSets, err := s.loadRolePermissionSets(ctx, roles)
	if err != nil {
		return nil, err
	}
	requesterPermissions := permissionSets[requesterRole]
	return manageableRolesForPermissionSet(roles, requesterPermissions, permissionSets), nil
}
