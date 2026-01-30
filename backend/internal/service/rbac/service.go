package rbac

import (
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type Service struct {
	userRepo   user.UserRepository
	memberRepo domainTeam.TeamMemberRepository
}

func New(userRepo user.UserRepository, memberRepo domainTeam.TeamMemberRepository) *Service {
	return &Service{userRepo: userRepo, memberRepo: memberRepo}
}

func (s *Service) GetGlobalRole(userID uuid.UUID) (user.Role, error) {
	users, err := s.userRepo.GetByIds([]uuid.UUID{userID})
	if err != nil {
		return "", err
	}
	if len(users) == 0 {
		return "", nil
	}
	return users[0].Role, nil
}

func (s *Service) GetTeamRole(teamID, userID uuid.UUID) (*domainTeam.MemberRole, error) {
	return s.memberRepo.GetUserRole(teamID, userID)
}

// GetRoleLevel returns the hierarchical level of a role (higher = more privileged)
func (s *Service) GetRoleLevel(role user.Role) int {
	return GetRoleLevel(role)
}

// CanManageRole checks if a user with requesterRole can manage/create a user with targetRole
// Based on the hierarchy:
// superadmin > admin_fzag > fzag > admin_planer > planer > admin_entrepreneur > entrepreneur
func (s *Service) CanManageRole(requesterRole user.Role, targetRole user.Role) bool {
	return CanManageRole(requesterRole, targetRole)
}

// GetAllowedRoles returns the list of roles that a user with the given role can assign to others
func (s *Service) GetAllowedRoles(requesterRole user.Role) []user.Role {
	return GetAllowedRoles(requesterRole)
}

// GetRoleLevel returns the hierarchical level of a role (higher = more privileged)
func GetRoleLevel(role user.Role) int {
	switch role {
	case user.RoleSuperAdmin:
		return 100
	case user.RoleAdminFZAG:
		return 90
	case user.RoleFZAG:
		return 80
	case user.RoleAdminPlaner:
		return 70
	case user.RolePlaner:
		return 60
	case user.RoleAdminEnterpreneur:
		return 50
	case user.RoleEnterpreneur:
		return 40
	// Legacy roles
	case user.RoleAdmin:
		return 50 // Same level as admin_planer for backwards compatibility
	case user.RoleUser:
		return 10
	default:
		return 0
	}
}

// CanManageRole checks if a user with requesterRole can manage/create a user with targetRole
func CanManageRole(requesterRole user.Role, targetRole user.Role) bool {
	requesterLevel := GetRoleLevel(requesterRole)
	targetLevel := GetRoleLevel(targetRole)

	// Special case: entrepreneur cannot manage any users
	if requesterRole == user.RoleEnterpreneur {
		return false
	}

	// User can manage roles below their level
	return requesterLevel > targetLevel
}

// GetAllowedRoles returns the list of roles that a user with the given role can assign to others
func GetAllowedRoles(requesterRole user.Role) []user.Role {
	switch requesterRole {
	case user.RoleSuperAdmin:
		// Can manage all roles
		return []user.Role{
			user.RoleSuperAdmin,
			user.RoleAdminFZAG,
			user.RoleFZAG,
			user.RoleAdminPlaner,
			user.RolePlaner,
			user.RoleAdminEnterpreneur,
			user.RoleEnterpreneur,
			user.RoleAdmin, // Legacy
			user.RoleUser,  // Legacy
		}
	case user.RoleAdminFZAG:
		// Can manage fzag and all below
		return []user.Role{
			user.RoleFZAG,
			user.RoleAdminPlaner,
			user.RolePlaner,
			user.RoleAdminEnterpreneur,
			user.RoleEnterpreneur,
		}
	case user.RoleFZAG:
		// Can manage admin_planer, planer, admin_entrepreneur, entrepreneur
		return []user.Role{
			user.RoleAdminPlaner,
			user.RolePlaner,
			user.RoleAdminEnterpreneur,
			user.RoleEnterpreneur,
		}
	case user.RoleAdminPlaner:
		// Can manage planer, admin_entrepreneur, entrepreneur
		return []user.Role{
			user.RolePlaner,
			user.RoleAdminEnterpreneur,
			user.RoleEnterpreneur,
		}
	case user.RolePlaner:
		// Can manage admin_entrepreneur, entrepreneur
		return []user.Role{
			user.RoleAdminEnterpreneur,
			user.RoleEnterpreneur,
		}
	case user.RoleAdminEnterpreneur:
		// Can manage entrepreneur
		return []user.Role{
			user.RoleEnterpreneur,
		}
	case user.RoleEnterpreneur:
		// Cannot manage any users
		return []user.Role{}
	default:
		return []user.Role{}
	}
}

func GlobalRoleLevel(r user.Role) int {
	return GetRoleLevel(r)
}

func TeamRoleLevel(r domainTeam.MemberRole) int {
	switch r {
	case domainTeam.MemberRoleOwner:
		return 100
	case domainTeam.MemberRoleManager:
		return 50
	case domainTeam.MemberRoleMember, "":
		return 10
	default:
		return 10
	}
}

func IsGlobalAdmin(r user.Role) bool {
	return GlobalRoleLevel(r) >= GlobalRoleLevel(user.RoleAdmin)
}
