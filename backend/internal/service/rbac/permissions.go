package rbac

import (
	"context"
	"errors"
	"sort"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

func (s *Service) ListPermissions(ctx context.Context) ([]domainUser.Permission, error) {
	return s.permissionRepo.ListAll(ctx)
}

func (s *Service) GetPermissionByID(ctx context.Context, id uuid.UUID) (*domainUser.Permission, error) {
	perms, err := s.permissionRepo.GetByIds(ctx, []uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(perms) == 0 {
		return nil, domain.ErrNotFound
	}
	return perms[0], nil
}

func (s *Service) CreatePermission(ctx context.Context, permission *domainUser.Permission) error {
	if permission == nil {
		return errors.New("permission_required")
	}
	if err := s.permissionRepo.Create(ctx, permission); err != nil {
		return err
	}

	// Auto-assign new permissions to superadmin by default.
	_, err := s.rolePermissionRepo.AddPermissionToRole(ctx, domainUser.RoleSuperAdmin, permission.Name)
	return err
}

func (s *Service) UpdatePermission(ctx context.Context, permission *domainUser.Permission) error {
	if permission == nil {
		return errors.New("permission_required")
	}
	return s.permissionRepo.Update(ctx, permission)
}

func (s *Service) DeletePermission(ctx context.Context, id uuid.UUID) error {
	perm, err := s.GetPermissionByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.permissionRepo.DeleteByIds(ctx, []uuid.UUID{id}); err != nil {
		return err
	}
	return s.rolePermissionRepo.DeleteByPermissionName(ctx, perm.Name)
}

func (s *Service) ListRolesWithPermissions(ctx context.Context) ([]domainUser.RoleInfo, error) {
	roles := domainUser.AllRoles()
	permissionSets, err := s.loadRolePermissionSets(ctx, roles)
	if err != nil {
		return nil, err
	}

	result := make([]domainUser.RoleInfo, 0, len(roles))
	for _, role := range roles {
		permissions := permissionSets[role].sortedValues()
		result = append(result, domainUser.RoleInfo{
			Name:        role,
			DisplayName: domainUser.RoleDisplayName(role),
			Description: domainUser.RoleDescription(role),
			Level:       s.GetRoleLevel(role),
			Permissions: permissions,
			CanManage:   manageableRolesForPermissionSet(roles, permissionSets[role], permissionSets),
		})
	}

	return result, nil
}

func (s *Service) GetRolePermissions(ctx context.Context, role domainUser.Role) ([]string, error) {
	permissionSets, err := s.loadRolePermissionSets(ctx, []domainUser.Role{role})
	if err != nil {
		return nil, err
	}
	return permissionSets[role].sortedValues(), nil
}

func (s *Service) UpdateRolePermissions(ctx context.Context, role domainUser.Role, permissions []string) ([]string, error) {
	if !domainUser.IsValidRole(role) {
		return nil, errors.New("invalid_role")
	}
	if role == domainUser.RoleSuperAdmin {
		return s.syncSuperAdminPermissions(ctx)
	}

	unique := make([]string, 0, len(permissions))
	seen := make(map[string]struct{}, len(permissions))
	for _, perm := range permissions {
		if _, ok := seen[perm]; ok {
			continue
		}
		seen[perm] = struct{}{}
		unique = append(unique, perm)
	}

	if err := s.validatePermissionsExist(ctx, unique); err != nil {
		return nil, err
	}

	if err := s.rolePermissionRepo.ReplaceRolePermissions(ctx, role, unique); err != nil {
		return nil, err
	}

	return unique, nil
}

func (s *Service) AddRolePermission(ctx context.Context, role domainUser.Role, permission string) (*domainUser.RolePermission, error) {
	if !domainUser.IsValidRole(role) {
		return nil, errors.New("invalid_role")
	}
	if role == domainUser.RoleSuperAdmin {
		if err := s.validatePermissionsExist(ctx, []string{permission}); err != nil {
			return nil, err
		}
		_, err := s.syncSuperAdminPermissions(ctx)
		if err != nil {
			return nil, err
		}
		return &domainUser.RolePermission{Role: role, Permission: permission}, nil
	}

	if err := s.validatePermissionsExist(ctx, []string{permission}); err != nil {
		return nil, err
	}

	return s.rolePermissionRepo.AddPermissionToRole(ctx, role, permission)
}

func (s *Service) RemoveRolePermission(ctx context.Context, role domainUser.Role, permission string) error {
	if !domainUser.IsValidRole(role) {
		return errors.New("invalid_role")
	}
	if role == domainUser.RoleSuperAdmin {
		_, err := s.syncSuperAdminPermissions(ctx)
		return err
	}

	return s.rolePermissionRepo.RemovePermissionFromRole(ctx, role, permission)
}

func (s *Service) HasPermission(ctx context.Context, role domainUser.Role, permission string) (bool, error) {
	if role == domainUser.RoleSuperAdmin {
		return s.permissionExists(ctx, permission)
	}

	perms, err := s.rolePermissionRepo.ListByRole(ctx, role)
	if err != nil {
		return false, err
	}
	for _, p := range perms {
		if p.Permission == permission {
			return true, nil
		}
	}
	return false, nil
}

func (s *Service) loadRolePermissionSets(ctx context.Context, roles []domainUser.Role) (map[domainUser.Role]permissionSet, error) {
	rolePerms, err := s.rolePermissionRepo.ListByRoles(ctx, roles)
	if err != nil {
		return nil, err
	}

	sets := rolePermissionSets(roles, rolePerms)
	if containsRole(roles, domainUser.RoleSuperAdmin) && s.permissionRepo != nil {
		allPermissions, err := s.permissionRepo.ListAll(ctx)
		if err != nil {
			return nil, err
		}
		superadminPermissions := permissionSet{}
		for _, permission := range allPermissions {
			superadminPermissions[permission.Name] = struct{}{}
		}
		sets[domainUser.RoleSuperAdmin] = superadminPermissions
	}

	return sets, nil
}

func (s *Service) permissionExists(ctx context.Context, permission string) (bool, error) {
	if s.permissionRepo == nil {
		return false, nil
	}
	perms, err := s.permissionRepo.ListByNames(ctx, []string{permission})
	if err != nil {
		return false, err
	}
	return len(perms) > 0, nil
}

func (s *Service) syncSuperAdminPermissions(ctx context.Context) ([]string, error) {
	if s.permissionRepo == nil {
		return nil, errors.New("permission_repo_required")
	}
	allPermissions, err := s.permissionRepo.ListAll(ctx)
	if err != nil {
		return nil, err
	}
	permissionNames := make([]string, 0, len(allPermissions))
	for _, permission := range allPermissions {
		permissionNames = append(permissionNames, permission.Name)
	}
	if err := s.rolePermissionRepo.ReplaceRolePermissions(ctx, domainUser.RoleSuperAdmin, permissionNames); err != nil {
		return nil, err
	}
	sort.Strings(permissionNames)
	return permissionNames, nil
}

func containsRole(roles []domainUser.Role, target domainUser.Role) bool {
	for _, role := range roles {
		if role == target {
			return true
		}
	}
	return false
}

func (s *Service) validatePermissionsExist(ctx context.Context, names []string) error {
	if len(names) == 0 {
		return nil
	}

	perms, err := s.permissionRepo.ListByNames(ctx, names)
	if err != nil {
		return err
	}

	unique := map[string]struct{}{}
	for _, name := range names {
		unique[name] = struct{}{}
	}

	if len(perms) != len(unique) {
		return domain.ErrNotFound
	}

	return nil
}

type permissionSet map[string]struct{}

func rolePermissionSets(roles []domainUser.Role, rolePerms []domainUser.RolePermission) map[domainUser.Role]permissionSet {
	sets := make(map[domainUser.Role]permissionSet, len(roles))
	for _, role := range roles {
		sets[role] = permissionSet{}
	}
	for _, rp := range rolePerms {
		set := sets[rp.Role]
		if set == nil {
			set = permissionSet{}
			sets[rp.Role] = set
		}
		set[rp.Permission] = struct{}{}
	}
	return sets
}

func manageableRolesForPermissionSet(roles []domainUser.Role, requesterPermissions permissionSet, rolePermissionSets map[domainUser.Role]permissionSet) []domainUser.Role {
	if !requesterPermissions.hasAny(domainUser.PermissionUserCreate, domainUser.PermissionUserUpdate) {
		return []domainUser.Role{}
	}

	allowed := make([]domainUser.Role, 0, len(roles))
	for _, role := range roles {
		if permissionSetContainsAll(requesterPermissions, rolePermissionSets[role]) {
			allowed = append(allowed, role)
		}
	}
	return allowed
}

func permissionSetContainsAll(granted permissionSet, required permissionSet) bool {
	for permission := range required {
		if _, ok := granted[permission]; !ok {
			return false
		}
	}
	return true
}

func (s permissionSet) has(permission string) bool {
	_, ok := s[permission]
	return ok
}

func (s permissionSet) hasAny(permissions ...string) bool {
	for _, permission := range permissions {
		if s.has(permission) {
			return true
		}
	}
	return false
}

func (s permissionSet) sortedValues() []string {
	values := make([]string, 0, len(s))
	for permission := range s {
		values = append(values, permission)
	}
	sort.Strings(values)
	return values
}
