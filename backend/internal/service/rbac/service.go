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

func (s *Service) GetAllowedRoles(ctx context.Context, requesterRole domainUser.Role) ([]domainUser.Role, error) {
	if requesterRole == domainUser.RoleSuperAdmin {
		return domainUser.AssignableRoles(requesterRole), nil
	}

	permissionSets, err := s.loadRolePermissionSets(ctx, []domainUser.Role{requesterRole})
	if err != nil {
		return nil, err
	}
	requesterPermissions := permissionSets[requesterRole]
	if !requesterPermissions.hasAny(domainUser.PermissionUserCreate, domainUser.PermissionUserUpdate) {
		return []domainUser.Role{}, nil
	}
	return domainUser.AssignableRoles(requesterRole), nil
}
