package rbac

import (
	"testing"

	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
)

func TestService_CanManageRoleHierarchy(t *testing.T) {
	svc := &Service{}

	if !svc.CanManageRole(domainUser.RoleSuperAdmin, domainUser.RoleAdminFZAG) {
		t.Fatal("expected superadmin to manage admin_fzag")
	}
	if svc.CanManageRole(domainUser.RolePlaner, domainUser.RolePlaner) {
		t.Fatal("expected same-role management to be denied")
	}
	if svc.CanManageRole(domainUser.RoleEnterpreneur, domainUser.RolePlaner) {
		t.Fatal("expected entrepreneur to manage nobody")
	}
}

func TestService_GetAllowedRoles(t *testing.T) {
	svc := &Service{}
	roles := svc.GetAllowedRoles(domainUser.RoleAdminPlaner)

	if len(roles) != 3 {
		t.Fatalf("expected 3 allowed roles, got %d", len(roles))
	}
	if roles[0] != domainUser.RolePlaner || roles[1] != domainUser.RoleAdminEnterpreneur || roles[2] != domainUser.RoleEnterpreneur {
		t.Fatalf("unexpected allowed roles: %+v", roles)
	}
}
