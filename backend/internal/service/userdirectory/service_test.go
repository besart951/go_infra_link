package userdirectory

import (
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

func TestPermissionSetContainsAll(t *testing.T) {
	granted := permissionSetFromRolePermissions([]domainUser.RolePermission{{Permission: domainUser.PermissionUserRead}, {Permission: domainUser.PermissionUserUpdate}})
	required := permissionSetFromRolePermissions([]domainUser.RolePermission{{Permission: domainUser.PermissionUserRead}})
	if !permissionSetContainsAll(granted, required) {
		t.Fatal("expected superset permissions to satisfy subset")
	}
}

func TestPermissionSetForRole_SuperadminActsAsWildcard(t *testing.T) {
	permissions := permissionSetForRole(domainUser.RoleSuperAdmin, nil)
	if !permissions.has(domainUser.PermissionUserRead) {
		t.Fatal("expected superadmin wildcard to satisfy user.read")
	}
	if !permissionSetContainsAll(permissions, permissionSetFromRolePermissions([]domainUser.RolePermission{{Permission: domainUser.PermissionTeamRead}})) {
		t.Fatal("expected superadmin wildcard to satisfy subset comparisons")
	}
}

func TestCanSeeUserScopedRolesRequireSharedTeam(t *testing.T) {
	requesterID := uuid.New()
	candidate := &domainUser.User{Base: mustBase(), Role: domainUser.RolePlaner}
	requesterPermissions := permissionSetFromRolePermissions([]domainUser.RolePermission{{Permission: domainUser.PermissionUserRead}})
	candidatePermissions := permissionSetFromRolePermissions([]domainUser.RolePermission{{Permission: domainUser.PermissionUserRead}, {Permission: domainUser.PermissionUserUpdate}})

	if canSeeUser(requesterID, requesterPermissions, candidate, map[uuid.UUID]struct{}{}, candidatePermissions) {
		t.Fatal("expected user to not see stronger candidate without shared team")
	}

	if !canSeeUser(requesterID, requesterPermissions, candidate, map[uuid.UUID]struct{}{uuid.New(): {}}, candidatePermissions) {
		t.Fatal("expected shared team visibility to allow candidate")
	}
}

func TestCanSeeUserPermissionSupersetAllowsVisibility(t *testing.T) {
	requesterID := uuid.New()
	candidate := &domainUser.User{Base: mustBase(), Role: domainUser.RolePlaner}
	requesterPermissions := permissionSetFromRolePermissions([]domainUser.RolePermission{{Permission: domainUser.PermissionUserRead}, {Permission: domainUser.PermissionUserUpdate}, {Permission: domainUser.PermissionTeamRead}})
	candidatePermissions := permissionSetFromRolePermissions([]domainUser.RolePermission{{Permission: domainUser.PermissionUserRead}})

	if !canSeeUser(requesterID, requesterPermissions, candidate, map[uuid.UUID]struct{}{}, candidatePermissions) {
		t.Fatal("expected permission superset to allow visibility")
	}
}

func TestBuildCapabilitiesSelfAndLastSuperadminProtection(t *testing.T) {
	requesterID := uuid.New()
	requesterPermissions := permissionSetFromRolePermissions([]domainUser.RolePermission{
		{Permission: domainUser.PermissionUserRead},
		{Permission: domainUser.PermissionUserUpdate},
		{Permission: domainUser.PermissionUserDelete},
	})
	superadminPermissions := permissionSetFromRolePermissions([]domainUser.RolePermission{
		{Permission: domainUser.PermissionUserRead},
		{Permission: domainUser.PermissionUserUpdate},
		{Permission: domainUser.PermissionUserDelete},
	})

	selfCaps := buildCapabilities(requesterID, requesterPermissions, domainUser.User{Base: mustBaseWithID(requesterID), Role: domainUser.RoleSuperAdmin, IsActive: true}, superadminPermissions, 2)
	if selfCaps.CanUpdate || selfCaps.CanDelete || selfCaps.CanDisable || selfCaps.CanEnable || selfCaps.CanChangeRole {
		t.Fatal("expected no self-management capabilities")
	}

	lastSuperadminCaps := buildCapabilities(
		requesterID,
		requesterPermissions,
		domainUser.User{Base: mustBase(), Role: domainUser.RoleSuperAdmin, IsActive: true},
		superadminPermissions,
		1,
	)
	if !lastSuperadminCaps.CanUpdate || !lastSuperadminCaps.CanChangeRole {
		t.Fatal("expected update and role-change for superadmin target")
	}
	if lastSuperadminCaps.CanDelete || lastSuperadminCaps.CanDisable {
		t.Fatal("expected delete/disable blocked for last superadmin")
	}
}

func TestIntersectVisibleTeamIDsScoped(t *testing.T) {
	teamA := uuid.New()
	teamB := uuid.New()

	requesterTeams := map[uuid.UUID]struct{}{teamA: {}}
	candidateTeams := map[uuid.UUID]struct{}{teamA: {}, teamB: {}}

	result := intersectVisibleTeamIDs(false, requesterTeams, candidateTeams)
	if len(result) != 1 {
		t.Fatalf("expected one shared team, got %d", len(result))
	}
	if _, ok := result[teamA]; !ok {
		t.Fatal("expected shared team to be included")
	}
	if _, ok := result[teamB]; ok {
		t.Fatal("expected non-shared team to be excluded")
	}
}

func TestIntersectVisibleTeamIDsWithTeamReadShowsAllTeams(t *testing.T) {
	teamA := uuid.New()
	teamB := uuid.New()
	requesterTeams := map[uuid.UUID]struct{}{teamA: {}}
	candidateTeams := map[uuid.UUID]struct{}{teamA: {}, teamB: {}}

	result := intersectVisibleTeamIDs(true, requesterTeams, candidateTeams)
	if len(result) != 2 {
		t.Fatalf("expected both candidate teams, got %d", len(result))
	}
}

func mustBase() domain.Base {
	return domain.Base{ID: uuid.New()}
}

func mustBaseWithID(id uuid.UUID) domain.Base {
	return domain.Base{ID: id}
}
