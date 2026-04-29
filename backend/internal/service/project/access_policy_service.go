package project

import (
	"context"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type ProjectAccessPolicyService struct {
	repo               domainProject.ProjectRepository
	userRepo           domainUser.UserRepository
	rolePermissionRepo domainUser.RolePermissionRepository
}

func (s *ProjectAccessPolicyService) CanAccessProject(ctx context.Context, requesterID, projectID uuid.UUID, requesterRole *domainUser.Role) (bool, error) {
	if _, err := domain.GetByID(ctx, s.repo, projectID); err != nil {
		return false, err
	}

	if requesterRole, ok, err := s.resolveRequesterRole(ctx, requesterID, requesterRole); err != nil {
		return false, err
	} else if ok {
		canListAll, err := s.roleHasPermission(ctx, requesterRole, domainUser.PermissionProjectListAll)
		if err != nil {
			return false, err
		}
		if canListAll {
			return true, nil
		}
	}

	return s.repo.HasUser(ctx, projectID, requesterID)
}

func (s *ProjectAccessPolicyService) CanUseProjectPermission(ctx context.Context, requesterID uuid.UUID, requesterRole *domainUser.Role, permission string) (bool, error) {
	role, ok, err := s.resolveRequesterRole(ctx, requesterID, requesterRole)
	if err != nil || !ok {
		return false, err
	}
	return s.roleHasPermission(ctx, role, permission)
}

func (s *ProjectAccessPolicyService) resolveRequesterRole(ctx context.Context, requesterID uuid.UUID, requesterRole *domainUser.Role) (domainUser.Role, bool, error) {
	if requesterRole != nil {
		return *requesterRole, true, nil
	}
	if s.userRepo == nil {
		return "", false, nil
	}

	users, err := s.userRepo.GetByIds(ctx, []uuid.UUID{requesterID})
	if err != nil {
		return "", false, err
	}
	if len(users) == 0 {
		return "", false, nil
	}
	return users[0].Role, true, nil
}

func (s *ProjectAccessPolicyService) roleHasPermission(ctx context.Context, role domainUser.Role, permission string) (bool, error) {
	if role == domainUser.RoleSuperAdmin {
		return true, nil
	}
	if s.rolePermissionRepo == nil {
		return false, nil
	}

	rolePermissions, err := s.rolePermissionRepo.ListByRole(ctx, role)
	if err != nil {
		return false, err
	}
	for _, rolePermission := range rolePermissions {
		if rolePermission.Permission == permission {
			return true, nil
		}
	}
	return false, nil
}
