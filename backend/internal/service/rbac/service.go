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

func GlobalRoleLevel(r user.Role) int {
	switch r {
	case user.RoleSuperAdmin:
		return 100
	case user.RoleAdmin:
		return 50
	case user.RoleUser, "":
		return 10
	default:
		return 10
	}
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
