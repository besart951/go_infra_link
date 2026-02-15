package rbac

import (
	"errors"
	"sort"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

func (s *Service) ListPermissions() ([]domainUser.Permission, error) {
	return s.permissionRepo.ListAll()
}

func (s *Service) GetPermissionByID(id uuid.UUID) (*domainUser.Permission, error) {
	perms, err := s.permissionRepo.GetByIds([]uuid.UUID{id})
	if err != nil {
		return nil, err
	}
	if len(perms) == 0 {
		return nil, domain.ErrNotFound
	}
	return perms[0], nil
}

func (s *Service) CreatePermission(permission *domainUser.Permission) error {
	if permission == nil {
		return errors.New("permission_required")
	}
	if err := s.permissionRepo.Create(permission); err != nil {
		return err
	}

	// Auto-assign new permissions to superadmin by default.
	_, err := s.rolePermissionRepo.AddPermissionToRole(domainUser.RoleSuperAdmin, permission.Name)
	return err
}

func (s *Service) UpdatePermission(permission *domainUser.Permission) error {
	if permission == nil {
		return errors.New("permission_required")
	}
	return s.permissionRepo.Update(permission)
}

func (s *Service) DeletePermission(id uuid.UUID) error {
	perm, err := s.GetPermissionByID(id)
	if err != nil {
		return err
	}
	if err := s.permissionRepo.DeleteByIds([]uuid.UUID{id}); err != nil {
		return err
	}
	return s.rolePermissionRepo.DeleteByPermissionName(perm.Name)
}

func (s *Service) ListRolesWithPermissions() ([]domainUser.RoleInfo, error) {
	roles := domainUser.AllRoles()
	rolePerms, err := s.rolePermissionRepo.ListByRoles(roles)
	if err != nil {
		return nil, err
	}

	permMap := make(map[domainUser.Role][]string, len(roles))
	for _, rp := range rolePerms {
		permMap[rp.Role] = append(permMap[rp.Role], rp.Permission)
	}
	for role := range permMap {
		sort.Strings(permMap[role])
	}

	result := make([]domainUser.RoleInfo, 0, len(roles))
	for _, role := range roles {
		permissions := permMap[role]
		if permissions == nil {
			permissions = []string{}
		}
		result = append(result, domainUser.RoleInfo{
			Name:        role,
			DisplayName: domainUser.RoleDisplayName(role),
			Description: domainUser.RoleDescription(role),
			Level:       domainUser.RoleLevel(role),
			Permissions: permissions,
			CanManage:   domainUser.GetAllowedRoles(role),
		})
	}

	return result, nil
}

func (s *Service) GetRolePermissions(role domainUser.Role) ([]string, error) {
	perms, err := s.rolePermissionRepo.ListByRole(role)
	if err != nil {
		return nil, err
	}
	result := make([]string, 0, len(perms))
	for _, perm := range perms {
		result = append(result, perm.Permission)
	}
	sort.Strings(result)
	return result, nil
}

func (s *Service) UpdateRolePermissions(role domainUser.Role, permissions []string) ([]string, error) {
	if !domainUser.IsValidRole(role) {
		return nil, errors.New("invalid_role")
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

	if err := s.validatePermissionsExist(unique); err != nil {
		return nil, err
	}

	if err := s.rolePermissionRepo.ReplaceRolePermissions(role, unique); err != nil {
		return nil, err
	}

	return unique, nil
}

func (s *Service) AddRolePermission(role domainUser.Role, permission string) (*domainUser.RolePermission, error) {
	if !domainUser.IsValidRole(role) {
		return nil, errors.New("invalid_role")
	}

	if err := s.validatePermissionsExist([]string{permission}); err != nil {
		return nil, err
	}

	return s.rolePermissionRepo.AddPermissionToRole(role, permission)
}

func (s *Service) RemoveRolePermission(role domainUser.Role, permission string) error {
	if !domainUser.IsValidRole(role) {
		return errors.New("invalid_role")
	}

	return s.rolePermissionRepo.RemovePermissionFromRole(role, permission)
}

func (s *Service) HasPermission(role domainUser.Role, permission string) (bool, error) {
	perms, err := s.rolePermissionRepo.ListByRole(role)
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

func (s *Service) validatePermissionsExist(names []string) error {
	if len(names) == 0 {
		return nil
	}

	perms, err := s.permissionRepo.ListByNames(names)
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
