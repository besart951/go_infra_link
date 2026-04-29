package rbac

import (
	"context"
	"sort"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

func TestService_GetAllowedRoles(t *testing.T) {
	ctx := context.Background()
	roleRepo := &rolePermissionRepoStub{items: map[domainUser.Role][]domainUser.RolePermission{
		domainUser.RoleSuperAdmin: {
			{Role: domainUser.RoleSuperAdmin, Permission: domainUser.PermissionUserCreate},
			{Role: domainUser.RoleSuperAdmin, Permission: domainUser.PermissionUserUpdate},
			{Role: domainUser.RoleSuperAdmin, Permission: domainUser.PermissionUserRead},
			{Role: domainUser.RoleSuperAdmin, Permission: domainUser.PermissionPermissionUpdate},
		},
		domainUser.RoleAdminFZAG: {
			{Role: domainUser.RoleAdminFZAG, Permission: domainUser.PermissionUserCreate},
			{Role: domainUser.RoleAdminFZAG, Permission: domainUser.PermissionUserUpdate},
			{Role: domainUser.RoleAdminFZAG, Permission: domainUser.PermissionUserRead},
			{Role: domainUser.RoleAdminFZAG, Permission: domainUser.PermissionRoleUpdate},
		},
		domainUser.RoleFZAG: {
			{Role: domainUser.RoleFZAG, Permission: domainUser.PermissionUserRead},
			{Role: domainUser.RoleFZAG, Permission: domainUser.PermissionPermissionRead},
		},
		domainUser.RoleAdminPlaner: {
			{Role: domainUser.RoleAdminPlaner, Permission: domainUser.PermissionUserCreate},
			{Role: domainUser.RoleAdminPlaner, Permission: domainUser.PermissionUserUpdate},
			{Role: domainUser.RoleAdminPlaner, Permission: domainUser.PermissionUserRead},
			{Role: domainUser.RoleAdminPlaner, Permission: domainUser.PermissionTeamRead},
		},
		domainUser.RolePlaner: {
			{Role: domainUser.RolePlaner, Permission: domainUser.PermissionUserRead},
		},
		domainUser.RoleAdminEnterpreneur: {
			{Role: domainUser.RoleAdminEnterpreneur, Permission: domainUser.PermissionUserRead},
		},
		domainUser.RoleEnterpreneur: {
			{Role: domainUser.RoleEnterpreneur, Permission: domainUser.PermissionProjectCreate},
		},
	}}
	svc := &Service{rolePermissionRepo: roleRepo, permissionRepo: &permissionRepoStub{items: []domainUser.Permission{
		{Name: domainUser.PermissionUserCreate},
		{Name: domainUser.PermissionUserUpdate},
		{Name: domainUser.PermissionUserRead},
		{Name: domainUser.PermissionPermissionUpdate},
		{Name: domainUser.PermissionRoleUpdate},
		{Name: domainUser.PermissionPermissionRead},
		{Name: domainUser.PermissionTeamRead},
		{Name: domainUser.PermissionProjectCreate},
	}}}
	roles, err := svc.GetAllowedRoles(ctx, domainUser.RoleAdminPlaner)
	if err != nil {
		t.Fatalf("expected allowed roles lookup to succeed, got %v", err)
	}

	if len(roles) != 3 {
		t.Fatalf("expected three manageable roles, got %d (%+v)", len(roles), roles)
	}
	if roles[0] != domainUser.RoleAdminPlaner {
		t.Fatalf("expected same-permission role to be manageable first, got %+v", roles)
	}
	if roles[1] != domainUser.RolePlaner || roles[2] != domainUser.RoleAdminEnterpreneur {
		t.Fatalf("unexpected manageable roles order/content: %+v", roles)
	}
}

func TestService_GetAllowedRoles_RequiresUserMutationPermission(t *testing.T) {
	ctx := context.Background()
	roleRepo := &rolePermissionRepoStub{items: map[domainUser.Role][]domainUser.RolePermission{
		domainUser.RolePlaner: {
			{Role: domainUser.RolePlaner, Permission: domainUser.PermissionUserRead},
		},
	}}
	svc := &Service{rolePermissionRepo: roleRepo, permissionRepo: &permissionRepoStub{}}
	roles, err := svc.GetAllowedRoles(ctx, domainUser.RolePlaner)
	if err != nil {
		t.Fatalf("expected allowed roles lookup to succeed, got %v", err)
	}
	if len(roles) != 0 {
		t.Fatalf("expected no assignable roles without user mutation permission, got %+v", roles)
	}
}

func TestService_GetRolePermissions_SuperadminGetsAllPermissions(t *testing.T) {
	ctx := context.Background()
	svc := &Service{
		permissionRepo: &permissionRepoStub{items: []domainUser.Permission{
			{Name: domainUser.PermissionUserRead},
			{Name: domainUser.PermissionTeamUpdate},
			{Name: domainUser.PermissionNotificationSMTPManage},
		}},
		rolePermissionRepo: &rolePermissionRepoStub{},
	}

	permissions, err := svc.GetRolePermissions(ctx, domainUser.RoleSuperAdmin)
	if err != nil {
		t.Fatalf("expected superadmin permissions lookup to succeed, got %v", err)
	}

	if len(permissions) != 3 {
		t.Fatalf("expected three permissions, got %d (%+v)", len(permissions), permissions)
	}
	sort.Strings(permissions)
	want := []string{domainUser.PermissionNotificationSMTPManage, domainUser.PermissionTeamUpdate, domainUser.PermissionUserRead}
	for i, permission := range want {
		if permissions[i] != permission {
			t.Fatalf("unexpected permissions: got %+v want %+v", permissions, want)
		}
	}
}

func TestService_HasPermission_SuperadminChecksDefinedPermissions(t *testing.T) {
	ctx := context.Background()
	svc := &Service{
		permissionRepo: &permissionRepoStub{items: []domainUser.Permission{{Name: domainUser.PermissionUserRead}}},
		rolePermissionRepo: &rolePermissionRepoStub{},
	}

	hasPermission, err := svc.HasPermission(ctx, domainUser.RoleSuperAdmin, domainUser.PermissionUserRead)
	if err != nil {
		t.Fatalf("expected permission lookup to succeed, got %v", err)
	}
	if !hasPermission {
		t.Fatal("expected superadmin to have all defined permissions")
	}

	hasPermission, err = svc.HasPermission(ctx, domainUser.RoleSuperAdmin, "missing.permission")
	if err != nil {
		t.Fatalf("expected undefined permission check to succeed, got %v", err)
	}
	if hasPermission {
		t.Fatal("expected undefined permission to be rejected")
	}
}

func TestService_UpdateRolePermissions_SuperadminRemainsSyncedToAllPermissions(t *testing.T) {
	ctx := context.Background()
	roleRepo := &rolePermissionRepoStub{}
	svc := &Service{
		permissionRepo: &permissionRepoStub{items: []domainUser.Permission{{Name: domainUser.PermissionUserRead}, {Name: domainUser.PermissionTeamUpdate}}},
		rolePermissionRepo: roleRepo,
	}

	permissions, err := svc.UpdateRolePermissions(ctx, domainUser.RoleSuperAdmin, []string{domainUser.PermissionUserRead})
	if err != nil {
		t.Fatalf("expected superadmin update to succeed, got %v", err)
	}
	if len(permissions) != 2 {
		t.Fatalf("expected full superadmin permission set, got %+v", permissions)
	}
	assigned := roleRepo.items[domainUser.RoleSuperAdmin]
	if len(assigned) != 2 {
		t.Fatalf("expected persisted superadmin permission set to stay complete, got %+v", assigned)
	}
}

type rolePermissionRepoStub struct {
	items map[domainUser.Role][]domainUser.RolePermission
}

func (r *rolePermissionRepoStub) Create(context.Context, *domainUser.RolePermission) error { return nil }
func (r *rolePermissionRepoStub) GetByIds(context.Context, []uuid.UUID) ([]*domainUser.RolePermission, error) {
	return nil, nil
}
func (r *rolePermissionRepoStub) Update(context.Context, *domainUser.RolePermission) error { return nil }
func (r *rolePermissionRepoStub) DeleteByIds(context.Context, []uuid.UUID) error { return nil }
func (r *rolePermissionRepoStub) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainUser.RolePermission], error) {
	return &domain.PaginatedList[domainUser.RolePermission]{}, nil
}
func (r *rolePermissionRepoStub) ListByRole(_ context.Context, role domainUser.Role) ([]domainUser.RolePermission, error) {
	items := r.items[role]
	out := make([]domainUser.RolePermission, len(items))
	copy(out, items)
	return out, nil
}
func (r *rolePermissionRepoStub) ListByRoles(_ context.Context, roles []domainUser.Role) ([]domainUser.RolePermission, error) {
	out := make([]domainUser.RolePermission, 0)
	for _, role := range roles {
		out = append(out, r.items[role]...)
	}
	return out, nil
}
func (r *rolePermissionRepoStub) ReplaceRolePermissions(_ context.Context, role domainUser.Role, permissions []string) error {
	if r.items == nil {
		r.items = map[domainUser.Role][]domainUser.RolePermission{}
	}
	replaced := make([]domainUser.RolePermission, 0, len(permissions))
	for _, permission := range permissions {
		replaced = append(replaced, domainUser.RolePermission{Role: role, Permission: permission})
	}
	r.items[role] = replaced
	return nil
}
func (r *rolePermissionRepoStub) AddPermissionToRole(context.Context, domainUser.Role, string) (*domainUser.RolePermission, error) {
	return nil, nil
}
func (r *rolePermissionRepoStub) RemovePermissionFromRole(context.Context, domainUser.Role, string) error {
	return nil
}
func (r *rolePermissionRepoStub) DeleteByPermissionName(context.Context, string) error { return nil }

type permissionRepoStub struct {
	items []domainUser.Permission
}

func (p *permissionRepoStub) Create(context.Context, *domainUser.Permission) error { return nil }
func (p *permissionRepoStub) GetByIds(context.Context, []uuid.UUID) ([]*domainUser.Permission, error) {
	return nil, nil
}
func (p *permissionRepoStub) Update(context.Context, *domainUser.Permission) error { return nil }
func (p *permissionRepoStub) DeleteByIds(context.Context, []uuid.UUID) error { return nil }
func (p *permissionRepoStub) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainUser.Permission], error) {
	return &domain.PaginatedList[domainUser.Permission]{}, nil
}
func (p *permissionRepoStub) GetByName(_ context.Context, name string) (*domainUser.Permission, error) {
	for _, permission := range p.items {
		if permission.Name == name {
			copy := permission
			return &copy, nil
		}
	}
	return nil, nil
}
func (p *permissionRepoStub) ListAll(context.Context) ([]domainUser.Permission, error) {
	out := make([]domainUser.Permission, len(p.items))
	copy(out, p.items)
	return out, nil
}
func (p *permissionRepoStub) ListByNames(_ context.Context, names []string) ([]domainUser.Permission, error) {
	nameSet := map[string]struct{}{}
	for _, name := range names {
		nameSet[name] = struct{}{}
	}
	out := make([]domainUser.Permission, 0)
	for _, permission := range p.items {
		if _, ok := nameSet[permission.Name]; ok {
			out = append(out, permission)
		}
	}
	return out, nil
}
