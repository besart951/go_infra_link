package userdirectory

import (
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

func TestCanAccessUserDirectory(t *testing.T) {
	tests := []struct {
		name string
		role domainUser.Role
		want bool
	}{
		{name: "superadmin", role: domainUser.RoleSuperAdmin, want: true},
		{name: "admin_fzag", role: domainUser.RoleAdminFZAG, want: true},
		{name: "fzag", role: domainUser.RoleFZAG, want: true},
		{name: "admin_planer", role: domainUser.RoleAdminPlaner, want: true},
		{name: "admin_entrepreneur", role: domainUser.RoleAdminEnterpreneur, want: true},
		{name: "planer", role: domainUser.RolePlaner, want: false},
		{name: "entrepreneur", role: domainUser.RoleEnterpreneur, want: false},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			got := CanAccessUserDirectory(testCase.role)
			if got != testCase.want {
				t.Fatalf("CanAccessUserDirectory(%s) = %v, want %v", testCase.role, got, testCase.want)
			}
		})
	}
}

func TestCanSeeUserScopedRolesRequireSharedTeam(t *testing.T) {
	requester := &domainUser.User{Base: mustBase(), Role: domainUser.RoleAdminPlaner}
	candidate := &domainUser.User{Base: mustBase(), Role: domainUser.RolePlaner}

	if canSeeUser(requester, candidate, map[uuid.UUID]struct{}{}) {
		t.Fatal("expected admin_planer to not see candidate without shared team")
	}

	if !canSeeUser(requester, candidate, map[uuid.UUID]struct{}{uuid.New(): {}}) {
		t.Fatal("expected admin_planer to see candidate with shared team")
	}
}

func TestCanSeeUserFZAGCannotSeeSuperadmin(t *testing.T) {
	requester := &domainUser.User{Base: mustBase(), Role: domainUser.RoleFZAG}
	candidate := &domainUser.User{Base: mustBase(), Role: domainUser.RoleSuperAdmin}

	if canSeeUser(requester, candidate, map[uuid.UUID]struct{}{uuid.New(): {}}) {
		t.Fatal("expected fzag to be blocked from seeing superadmin")
	}
}

func TestBuildCapabilitiesSelfAndLastSuperadminProtection(t *testing.T) {
	requesterID := uuid.New()
	requester := &domainUser.User{Base: mustBaseWithID(requesterID), Role: domainUser.RoleSuperAdmin}

	selfCaps := buildCapabilities(requester, domainUser.User{Base: mustBaseWithID(requesterID), Role: domainUser.RoleSuperAdmin, IsActive: true}, 2)
	if selfCaps.CanUpdate || selfCaps.CanDelete || selfCaps.CanDisable || selfCaps.CanEnable || selfCaps.CanChangeRole {
		t.Fatal("expected no self-management capabilities")
	}

	lastSuperadminCaps := buildCapabilities(
		requester,
		domainUser.User{Base: mustBase(), Role: domainUser.RoleSuperAdmin, IsActive: true},
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

	requester := &domainUser.User{Base: mustBase(), Role: domainUser.RoleAdminEnterpreneur}
	candidate := &domainUser.User{Base: mustBase(), Role: domainUser.RoleEnterpreneur}

	requesterTeams := map[uuid.UUID]struct{}{teamA: {}}
	candidateTeams := map[uuid.UUID]struct{}{teamA: {}, teamB: {}}

	result := intersectVisibleTeamIDs(requester, requesterTeams, candidate, candidateTeams)
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

func mustBase() domain.Base {
	return domain.Base{ID: uuid.New()}
}

func mustBaseWithID(id uuid.UUID) domain.Base {
	return domain.Base{ID: id}
}
