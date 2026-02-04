package rbac

import (
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type Service struct {
	userRepo   domainUser.UserRepository
	memberRepo domainTeam.TeamMemberRepository
}

func New(userRepo domainUser.UserRepository, memberRepo domainTeam.TeamMemberRepository) *Service {
	return &Service{userRepo: userRepo, memberRepo: memberRepo}
}

func (s *Service) GetGlobalRole(userID uuid.UUID) (domainUser.Role, error) {
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
func (s *Service) GetRoleLevel(role domainUser.Role) int {
	return domainUser.RoleLevel(role)
}

// CanManageRole checks if a user with requesterRole can manage/create a user with targetRole
func (s *Service) CanManageRole(requesterRole domainUser.Role, targetRole domainUser.Role) bool {
	return domainUser.CanManageRole(requesterRole, targetRole)
}

// GetAllowedRoles returns the list of roles that a user with the given role can assign to others
func (s *Service) GetAllowedRoles(requesterRole domainUser.Role) []domainUser.Role {
	return domainUser.GetAllowedRoles(requesterRole)
}
