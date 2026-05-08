package project

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

type ProjectAccessPolicyService struct {
	repo                domainProject.ProjectRepository
	phaseRepo           domainProject.PhaseRepository
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
		if roleCanAccessAllProjects(requesterRole) {
			return true, nil
		}

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

func roleCanAccessAllProjects(role domainUser.Role) bool {
	return role == domainUser.RoleSuperAdmin || role == domainUser.RoleAdminFZAG
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
	if roleCanAccessAllProjects(role) {
		return true, nil
	}

	hasRolePermission, err := s.roleHasPermission(ctx, role, permission)
	if err != nil || !hasRolePermission {
		return false, err
	}

	return s.phaseAllowsProjectPermission(ctx, role, projectID, permission)
}

func (s *ProjectAccessPolicyService) ExplainProjectPermissionDenial(ctx context.Context, requesterID uuid.UUID, requesterRole *domainUser.Role, permissions []string) (*domainProject.PermissionDenialDetails, error) {
	role, ok, err := s.resolveRequesterRole(ctx, requesterID, requesterRole)
	if err != nil {
		return nil, err
	}
	if !ok {
		return genericPermissionDenialDetails("", permissions), nil
	}

	for _, permission := range permissions {
		hasRolePermission, err := s.roleHasPermission(ctx, role, permission)
		if err != nil {
			return nil, err
		}
		if !hasRolePermission {
			return s.missingGeneralPermissionDetails(ctx, role, permission, permissions, nil, false)
		}
	}

	return genericPermissionDenialDetails(role, permissions), nil
}

func (s *ProjectAccessPolicyService) ExplainProjectScopedPermissionDenial(ctx context.Context, requesterID, projectID uuid.UUID, requesterRole *domainUser.Role, permissions []string) (*domainProject.PermissionDenialDetails, error) {
	if s.repo == nil {
		return genericPermissionDenialDetails("", permissions), nil
	}

	project, err := domain.GetByID(ctx, s.repo, projectID)
	if err != nil {
		return nil, err
	}

	role, ok, err := s.resolveRequesterRole(ctx, requesterID, requesterRole)
	if err != nil {
		return nil, err
	}
	if !ok {
		return genericPermissionDenialDetails("", permissions), nil
	}

	for _, permission := range permissions {
		if roleCanAccessAllProjects(role) {
			continue
		}

		hasRolePermission, err := s.roleHasPermission(ctx, role, permission)
		if err != nil {
			return nil, err
		}
		if !hasRolePermission {
			return s.missingGeneralPermissionDetails(ctx, role, permission, permissions, project, true)
		}

		phaseAllowed, err := s.phaseAllowsProjectPermissionForProject(ctx, role, project, permission)
		if err != nil {
			return nil, err
		}
		if !phaseAllowed {
			return s.phaseBlockedPermissionDetails(ctx, role, permission, permissions, project)
		}
	}

	return genericPermissionDenialDetails(role, permissions), nil
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
	if !isPhaseScopedProjectPermission(permission) {
		return true, nil
	}
	if s.repo == nil {
		return false, nil
	}

	project, err := domain.GetByID(ctx, s.repo, projectID)
	if err != nil {
		return false, err
	}
	return s.phaseAllowsProjectPermissionForProject(ctx, role, project, permission)
}

func (s *ProjectAccessPolicyService) phaseAllowsProjectPermissionForProject(ctx context.Context, role domainUser.Role, project *domainProject.Project, permission string) (bool, error) {
	if !isPhaseScopedProjectPermission(permission) {
		return true, nil
	}
	if project == nil {
		return false, nil
	}
	if project.PhaseID == uuid.Nil {
		return true, nil
	}
	if s.phasePermissionRepo == nil {
		return false, nil
	}

	rule, err := s.phasePermissionRepo.GetByPhaseAndRole(ctx, project.PhaseID, role)
	if errors.Is(err, domain.ErrNotFound) {
		return false, nil
	}
	if err != nil {
		return false, err
	}

	return slices.Contains(rule.Permissions, permission), nil
}

func (s *ProjectAccessPolicyService) missingGeneralPermissionDetails(ctx context.Context, role domainUser.Role, permission string, permissions []string, project *domainProject.Project, projectScoped bool) (*domainProject.PermissionDenialDetails, error) {
	requiredRoles, minimumRole, err := s.requiredRolesForPermission(ctx, permission, projectScoped)
	if err != nil {
		return nil, err
	}

	details := &domainProject.PermissionDenialDetails{
		Reason:             domainProject.PermissionDenialReasonMissingGeneral,
		Permission:         permission,
		Permissions:        clonePermissions(permissions),
		RequesterRole:      role,
		RequesterRoleLabel: domainUser.RoleDisplayName(role),
		RequiredRoles:      requiredRoles,
	}
	if minimumRole != nil {
		details.MinimumRole = minimumRole.Role
		details.MinimumRoleLabel = minimumRole.Label
	}
	if err := s.addProjectContext(ctx, details, project); err != nil {
		return nil, err
	}
	details.Message = missingGeneralPermissionMessage(details)
	return details, nil
}

func (s *ProjectAccessPolicyService) phaseBlockedPermissionDetails(ctx context.Context, role domainUser.Role, permission string, permissions []string, project *domainProject.Project) (*domainProject.PermissionDenialDetails, error) {
	details := &domainProject.PermissionDenialDetails{
		Reason:             domainProject.PermissionDenialReasonPhaseBlocked,
		Permission:         permission,
		Permissions:        clonePermissions(permissions),
		RequesterRole:      role,
		RequesterRoleLabel: domainUser.RoleDisplayName(role),
	}
	if err := s.addProjectContext(ctx, details, project); err != nil {
		return nil, err
	}
	details.Message = phaseBlockedPermissionMessage(details)
	return details, nil
}

func (s *ProjectAccessPolicyService) addProjectContext(ctx context.Context, details *domainProject.PermissionDenialDetails, project *domainProject.Project) error {
	if details == nil || project == nil {
		return nil
	}

	projectID := project.ID
	details.ProjectID = &projectID
	if project.PhaseID == uuid.Nil {
		return nil
	}

	phaseID := project.PhaseID
	details.PhaseID = &phaseID
	phaseName, err := s.projectPhaseName(ctx, project)
	if err != nil {
		return err
	}
	details.PhaseName = phaseName
	return nil
}

func (s *ProjectAccessPolicyService) projectPhaseName(ctx context.Context, project *domainProject.Project) (string, error) {
	if project == nil || project.PhaseID == uuid.Nil {
		return "", nil
	}
	if project.Phase != nil && strings.TrimSpace(project.Phase.Name) != "" {
		return project.Phase.Name, nil
	}
	if s.phaseRepo == nil {
		return "", nil
	}

	phase, err := domain.GetByID(ctx, s.phaseRepo, project.PhaseID)
	if errors.Is(err, domain.ErrNotFound) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return phase.Name, nil
}

func (s *ProjectAccessPolicyService) requiredRolesForPermission(ctx context.Context, permission string, projectScoped bool) ([]domainProject.PermissionRequiredRole, *domainProject.PermissionRequiredRole, error) {
	roles := make([]domainProject.PermissionRequiredRole, 0)
	for _, role := range domainUser.AllRoles() {
		hasPermission := projectScoped && roleCanAccessAllProjects(role)
		if !hasPermission {
			var err error
			hasPermission, err = s.roleHasPermission(ctx, role, permission)
			if err != nil {
				return nil, nil, err
			}
		}
		if !hasPermission {
			continue
		}

		requiredRole := domainProject.PermissionRequiredRole{
			Role:  role,
			Label: domainUser.RoleDisplayName(role),
		}
		roles = append(roles, requiredRole)
	}

	var minimumRole *domainProject.PermissionRequiredRole
	if len(roles) > 0 {
		minimumRole = &roles[len(roles)-1]
	}
	return roles, minimumRole, nil
}

func missingGeneralPermissionMessage(details *domainProject.PermissionDenialDetails) string {
	if details == nil {
		return "Sie haben keine Berechtigung für diese Aktion."
	}
	if details.MinimumRoleLabel != "" {
		return fmt.Sprintf("Sie haben keine allgemeine Berechtigung für \"%s\". Sie müssen mindestens die Rolle \"%s\" oder eine Rolle mit dieser Berechtigung haben.", details.Permission, details.MinimumRoleLabel)
	}
	return fmt.Sprintf("Sie haben keine allgemeine Berechtigung für \"%s\".", details.Permission)
}

func phaseBlockedPermissionMessage(details *domainProject.PermissionDenialDetails) string {
	if details == nil {
		return "Sie haben keine Berechtigung für diese Aktion."
	}

	phase := details.PhaseName
	if phase == "" && details.PhaseID != nil {
		phase = details.PhaseID.String()
	}
	if phase == "" {
		phase = "der aktuellen Phase"
	}

	role := details.RequesterRoleLabel
	if role == "" {
		role = string(details.RequesterRole)
	}

	return fmt.Sprintf("Das Projekt befindet sich in der Phase \"%s\". In dieser Phase hat Ihre Rolle \"%s\" keine Berechtigung für \"%s\".", phase, role, details.Permission)
}

func genericPermissionDenialDetails(role domainUser.Role, permissions []string) *domainProject.PermissionDenialDetails {
	details := &domainProject.PermissionDenialDetails{
		Reason:      domainProject.PermissionDenialReasonForbidden,
		Permissions: clonePermissions(permissions),
		Message:     "Sie haben keine Berechtigung für diese Aktion.",
	}
	if role != "" {
		details.RequesterRole = role
		details.RequesterRoleLabel = domainUser.RoleDisplayName(role)
	}
	if len(permissions) == 1 {
		details.Permission = permissions[0]
	}
	return details
}

func clonePermissions(permissions []string) []string {
	if len(permissions) == 0 {
		return nil
	}
	out := make([]string, len(permissions))
	copy(out, permissions)
	return out
}

func isPhaseScopedProjectPermission(permission string) bool {
	if !strings.HasPrefix(permission, "project.") {
		return false
	}
	if strings.HasSuffix(permission, ".edit") {
		return false
	}
	switch permission {
	case domainUser.PermissionProjectCreate, domainUser.PermissionProjectListAll:
		return false
	default:
		return true
	}
}
