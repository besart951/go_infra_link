package project

import (
	"context"
	"errors"
	"slices"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type ProjectAccessPolicyService struct {
	repo                domainProject.ProjectRepository
	userRepo            domainUser.UserRepository
	rolePermissionRepo  domainUser.RolePermissionRepository
	phasePermissionRepo domainProject.PhasePermissionRepository
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

func (s *ProjectAccessPolicyService) CanUseProjectPermissionForProject(ctx context.Context, requesterID, projectID uuid.UUID, requesterRole *domainUser.Role, permission string) (bool, error) {
	role, ok, err := s.resolveRequesterRole(ctx, requesterID, requesterRole)
	if err != nil || !ok {
		return false, err
	}
	if role == domainUser.RoleSuperAdmin {
		return true, nil
	}

	hasRolePermission, err := s.roleHasPermission(ctx, role, permission)
	if err != nil || !hasRolePermission {
		return false, err
	}

	return s.phaseAllowsProjectPermission(ctx, role, projectID, permission)
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

func (s *ProjectAccessPolicyService) phaseAllowsProjectPermission(ctx context.Context, role domainUser.Role, projectID uuid.UUID, permission string) (bool, error) {
	if s.phasePermissionRepo == nil || s.repo == nil || !isPhaseScopedProjectPermission(permission) {
		return true, nil
	}

	project, err := domain.GetByID(ctx, s.repo, projectID)
	if err != nil {
		return false, err
	}
	if project.PhaseID == uuid.Nil {
		return true, nil
	}

	rule, err := s.phasePermissionRepo.GetByPhaseAndRole(ctx, project.PhaseID, role)
	if errors.Is(err, domain.ErrNotFound) {
		return true, nil
	}
	if err != nil {
		return false, err
	}

	return slices.Contains(rule.Permissions, permission), nil
}

func isPhaseScopedProjectPermission(permission string) bool {
	if !strings.HasPrefix(permission, "project.") {
		return false
	}
	switch permission {
	case domainUser.PermissionProjectCreate, domainUser.PermissionProjectListAll:
		return false
	default:
		return !strings.HasSuffix(permission, ".edit")
	}
}
