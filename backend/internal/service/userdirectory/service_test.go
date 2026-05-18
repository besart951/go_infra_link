package userdirectory

import (
	"context"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainTeam "github.com/besart951/go_infra_link/backend/internal/domain/team"
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

	selfCaps := buildCapabilities(requesterID, domainUser.RoleSuperAdmin, requesterPermissions, domainUser.User{Base: mustBaseWithID(requesterID), Role: domainUser.RoleSuperAdmin, IsActive: true}, superadminPermissions, 2)
	if selfCaps.CanUpdate || selfCaps.CanDelete || selfCaps.CanDisable || selfCaps.CanEnable || selfCaps.CanRestore || selfCaps.CanChangeRole {
		t.Fatal("expected no self-management capabilities")
	}

	lastSuperadminCaps := buildCapabilities(
		requesterID,
		domainUser.RoleSuperAdmin,
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

func TestBuildCapabilitiesRequiresRoleHierarchy(t *testing.T) {
	requesterID := uuid.New()
	requesterPermissions := permissionSetFromRolePermissions([]domainUser.RolePermission{
		{Permission: domainUser.PermissionUserRead},
		{Permission: domainUser.PermissionUserUpdate},
		{Permission: domainUser.PermissionUserDelete},
	})
	targetPermissions := permissionSetFromRolePermissions([]domainUser.RolePermission{
		{Permission: domainUser.PermissionUserRead},
	})

	lowerCaps := buildCapabilities(
		requesterID,
		domainUser.RoleAdminPlaner,
		requesterPermissions,
		domainUser.User{Base: mustBase(), Role: domainUser.RolePlaner, IsActive: true},
		targetPermissions,
		2,
	)
	if !lowerCaps.CanUpdate || !lowerCaps.CanDelete || !lowerCaps.CanDisable || !lowerCaps.CanChangeRole {
		t.Fatal("expected capabilities for lower role")
	}

	sameCaps := buildCapabilities(
		requesterID,
		domainUser.RoleAdminPlaner,
		requesterPermissions,
		domainUser.User{Base: mustBase(), Role: domainUser.RoleAdminPlaner, IsActive: true},
		targetPermissions,
		2,
	)
	if sameCaps.CanUpdate || sameCaps.CanDelete || sameCaps.CanDisable || sameCaps.CanEnable || sameCaps.CanRestore || sameCaps.CanChangeRole {
		t.Fatal("expected no capabilities for same-level role")
	}

	higherCaps := buildCapabilities(
		requesterID,
		domainUser.RoleAdminPlaner,
		requesterPermissions,
		domainUser.User{Base: mustBase(), Role: domainUser.RoleAdminFZAG, IsActive: true},
		targetPermissions,
		2,
	)
	if higherCaps.CanUpdate || higherCaps.CanDelete || higherCaps.CanDisable || higherCaps.CanEnable || higherCaps.CanRestore || higherCaps.CanChangeRole {
		t.Fatal("expected no capabilities for higher role")
	}
}

func TestBuildCapabilitiesDeletedUsersUseRestoreCapability(t *testing.T) {
	requesterID := uuid.New()
	deletedAt := time.Now().UTC().Add(-time.Hour)
	restoreUntil := time.Now().UTC().Add(time.Hour)
	requesterPermissions := permissionSetFromRolePermissions([]domainUser.RolePermission{
		{Permission: domainUser.PermissionUserRead},
		{Permission: domainUser.PermissionUserUpdate},
		{Permission: domainUser.PermissionUserDelete},
		{Permission: domainUser.PermissionUserReadDeleted},
	})
	targetPermissions := permissionSetFromRolePermissions([]domainUser.RolePermission{
		{Permission: domainUser.PermissionUserRead},
	})

	caps := buildCapabilities(
		requesterID,
		domainUser.RoleAdminPlaner,
		requesterPermissions,
		domainUser.User{
			Base:         mustBase(),
			Role:         domainUser.RolePlaner,
			IsActive:     false,
			DeletedAt:    &deletedAt,
			RestoreUntil: &restoreUntil,
		},
		targetPermissions,
		2,
	)

	if !caps.CanRestore {
		t.Fatal("expected deleted restorable user to expose restore capability")
	}
	if caps.CanEnable || caps.CanUpdate || caps.CanDelete || caps.CanDisable || caps.CanChangeRole {
		t.Fatalf("expected deleted user to hide normal actions, got %+v", caps)
	}

	withoutReadDeleted := requesterPermissions
	delete(withoutReadDeleted, domainUser.PermissionUserReadDeleted)
	caps = buildCapabilities(
		requesterID,
		domainUser.RoleAdminPlaner,
		withoutReadDeleted,
		domainUser.User{
			Base:         mustBase(),
			Role:         domainUser.RolePlaner,
			IsActive:     false,
			DeletedAt:    &deletedAt,
			RestoreUntil: &restoreUntil,
		},
		targetPermissions,
		2,
	)
	if caps.CanRestore {
		t.Fatal("expected restore capability to require user.read_deleted")
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

func TestListBuildsTeamFiltersFromCandidateTeamNames(t *testing.T) {
	requesterID := uuid.New()
	candidateID := uuid.New()
	teamID := uuid.New()
	users := []*domainUser.User{
		testUser(requesterID, domainUser.RoleSuperAdmin, "Root", "Admin", "root@example.com"),
		testUser(candidateID, domainUser.RolePlaner, "Ada", "Lovelace", "ada@example.com"),
	}
	service := New(
		&directoryUserReaderStub{users: users},
		&directoryTeamReaderStub{teams: map[uuid.UUID]*domainTeam.Team{
			teamID: {Base: mustBaseWithID(teamID), Name: "Team Alpha"},
		}},
		&directoryMembershipReaderStub{memberships: map[uuid.UUID][]domainTeam.TeamMember{
			candidateID: {{TeamID: teamID, UserID: candidateID, Role: domainTeam.MemberRoleMember}},
		}},
		&directoryRolePermissionReaderStub{},
	)

	result, err := service.List(context.Background(), requesterID, 1, 10, "", "", "", "", "", false)
	if err != nil {
		t.Fatalf("expected list to succeed, got %v", err)
	}
	if len(result.Teams) != 1 {
		t.Fatalf("expected one team filter, got %d", len(result.Teams))
	}
	if result.Teams[0].ID != teamID || result.Teams[0].Name != "Team Alpha" {
		t.Fatalf("expected team filter to use candidate team name, got %+v", result.Teams[0])
	}
}

func TestListFiltersByRole(t *testing.T) {
	requesterID := uuid.New()
	planerID := uuid.New()
	entrepreneurID := uuid.New()
	users := []*domainUser.User{
		testUser(requesterID, domainUser.RoleSuperAdmin, "Root", "Admin", "root@example.com"),
		testUser(planerID, domainUser.RolePlaner, "Ada", "Lovelace", "ada@example.com"),
		testUser(entrepreneurID, domainUser.RoleEnterpreneur, "Grace", "Hopper", "grace@example.com"),
	}
	service := New(
		&directoryUserReaderStub{users: users},
		&directoryTeamReaderStub{},
		&directoryMembershipReaderStub{},
		&directoryRolePermissionReaderStub{},
	)

	result, err := service.List(context.Background(), requesterID, 1, 10, "", "", string(domainUser.RolePlaner), "", "", false)
	if err != nil {
		t.Fatalf("expected list to succeed, got %v", err)
	}
	if result.Total != 1 {
		t.Fatalf("expected one role-filtered user, got %d", result.Total)
	}
	if got := result.Items[0].User.Role; got != domainUser.RolePlaner {
		t.Fatalf("expected planer user, got %s", got)
	}
}

func mustBase() domain.Base {
	return domain.Base{ID: uuid.New()}
}

func mustBaseWithID(id uuid.UUID) domain.Base {
	return domain.Base{ID: id}
}

func testUser(id uuid.UUID, role domainUser.Role, firstName, lastName, email string) *domainUser.User {
	return &domainUser.User{
		Base:      mustBaseWithID(id),
		FirstName: firstName,
		LastName:  lastName,
		Email:     domainUser.EmailPtr(email),
		IsActive:  true,
		Role:      role,
	}
}

type directoryUserReaderStub struct {
	users []*domainUser.User
}

func (s *directoryUserReaderStub) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainUser.User, error) {
	lookup := map[uuid.UUID]*domainUser.User{}
	for _, user := range s.users {
		lookup[user.ID] = user
	}
	result := make([]*domainUser.User, 0, len(ids))
	for _, id := range ids {
		if user, ok := lookup[id]; ok {
			result = append(result, user)
		}
	}
	return result, nil
}

func (s *directoryUserReaderStub) GetPaginatedList(_ context.Context, _ domain.PaginationParams) (*domain.PaginatedList[domainUser.User], error) {
	items := make([]domainUser.User, 0, len(s.users))
	for _, user := range s.users {
		items = append(items, *user)
	}
	return &domain.PaginatedList[domainUser.User]{
		Items:      items,
		Total:      int64(len(items)),
		Page:       1,
		TotalPages: 1,
	}, nil
}

type directoryTeamReaderStub struct {
	teams map[uuid.UUID]*domainTeam.Team
}

func (s *directoryTeamReaderStub) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainTeam.Team, error) {
	result := make([]*domainTeam.Team, 0, len(ids))
	for _, id := range ids {
		if team, ok := s.teams[id]; ok {
			result = append(result, team)
		}
	}
	return result, nil
}

type directoryMembershipReaderStub struct {
	memberships map[uuid.UUID][]domainTeam.TeamMember
}

func (s *directoryMembershipReaderStub) ListByUser(_ context.Context, userID uuid.UUID, _ domain.PaginationParams) (*domain.PaginatedList[domainTeam.TeamMember], error) {
	items := s.memberships[userID]
	return &domain.PaginatedList[domainTeam.TeamMember]{
		Items:      items,
		Total:      int64(len(items)),
		Page:       1,
		TotalPages: 1,
	}, nil
}

type directoryRolePermissionReaderStub struct{}

func (s *directoryRolePermissionReaderStub) ListByRole(_ context.Context, role domainUser.Role) ([]domainUser.RolePermission, error) {
	if role == domainUser.RoleSuperAdmin {
		return nil, nil
	}
	return []domainUser.RolePermission{{Role: role, Permission: domainUser.PermissionUserRead}}, nil
}
