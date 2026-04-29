package project

import (
	"context"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainProject "github.com/besart951/go_infra_link/backend/internal/domain/project"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

func TestProjectAccessPolicyService_CanAccessProject_CharacterizesAccessSources(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	creatorID := uuid.New()
	memberID := uuid.New()
	adminID := uuid.New()
	listAllID := uuid.New()
	outsiderID := uuid.New()

	projectRepo := newProjectRepo()
	projectRepo.items[projectID] = &domainProject.Project{Base: domain.Base{ID: projectID}, CreatorID: creatorID}
	if err := projectRepo.AddUser(ctx, projectID, memberID); err != nil {
		t.Fatalf("expected member setup to succeed, got %v", err)
	}

	userRepo := newProjectUserRepo()
	userRepo.items[adminID] = &domainUser.User{Base: domain.Base{ID: adminID}, Role: domainUser.RoleAdminFZAG}
	listAllRole := domainUser.Role("project_list_all")
	userRepo.items[listAllID] = &domainUser.User{Base: domain.Base{ID: listAllID}, Role: listAllRole}
	userRepo.items[outsiderID] = &domainUser.User{Base: domain.Base{ID: outsiderID}, Role: domainUser.RolePlaner}

	rolePermissionRepo := newProjectRolePermissionRepo()
	rolePermissionRepo.grant(listAllRole, domainUser.PermissionProjectListAll)

	svc := NewServices(Dependencies{
		Projects:        projectRepo,
		Users:           userRepo,
		RolePermissions: rolePermissionRepo,
	}).AccessPolicy

	testCases := []struct {
		name        string
		requesterID uuid.UUID
		wantAccess  bool
	}{
		{name: "creator without membership", requesterID: creatorID, wantAccess: false},
		{name: "member", requesterID: memberID, wantAccess: true},
		{name: "admin without list all", requesterID: adminID, wantAccess: false},
		{name: "list all permission", requesterID: listAllID, wantAccess: true},
		{name: "outsider", requesterID: outsiderID, wantAccess: false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			hasAccess, err := svc.CanAccessProject(ctx, tc.requesterID, projectID, nil)
			if err != nil {
				t.Fatalf("expected access check to succeed, got %v", err)
			}
			if hasAccess != tc.wantAccess {
				t.Fatalf("expected access=%t, got %t", tc.wantAccess, hasAccess)
			}
		})
	}
}

func TestProjectAccessPolicyService_CanAccessProject_UsesRequesterRoleHint(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	requesterID := uuid.New()

	projectRepo := newProjectRepo()
	projectRepo.items[projectID] = &domainProject.Project{Base: domain.Base{ID: projectID}, CreatorID: uuid.New()}

	rolePermissionRepo := newProjectRolePermissionRepo()
	projectViewerRole := domainUser.Role("project_viewer")
	rolePermissionRepo.grant(projectViewerRole, domainUser.PermissionProjectListAll)

	svc := NewServices(Dependencies{Projects: projectRepo, RolePermissions: rolePermissionRepo}).AccessPolicy

	hasAccess, err := svc.CanAccessProject(ctx, requesterID, projectID, &projectViewerRole)
	if err != nil {
		t.Fatalf("expected access check to succeed, got %v", err)
	}
	if !hasAccess {
		t.Fatal("expected access for role hint with project.listAll")
	}
}

func TestProjectAccessPolicyService_CanUseProjectPermission(t *testing.T) {
	ctx := context.Background()
	requesterID := uuid.New()
	projectEditorRole := domainUser.Role("project_editor")

	rolePermissionRepo := newProjectRolePermissionRepo()
	rolePermissionRepo.grant(projectEditorRole, domainUser.PermissionProjectFieldDeviceEdit)

	svc := NewServices(Dependencies{RolePermissions: rolePermissionRepo}).AccessPolicy

	hasPermission, err := svc.CanUseProjectPermission(ctx, requesterID, &projectEditorRole, domainUser.PermissionProjectFieldDeviceEdit)
	if err != nil {
		t.Fatalf("expected permission check to succeed, got %v", err)
	}
	if !hasPermission {
		t.Fatal("expected role hint with project field device edit permission to pass")
	}

	hasPermission, err = svc.CanUseProjectPermission(ctx, requesterID, &projectEditorRole, domainUser.PermissionProjectSPSControllerEdit)
	if err != nil {
		t.Fatalf("expected permission check to succeed, got %v", err)
	}
	if hasPermission {
		t.Fatal("expected missing project SPS controller edit permission to fail")
	}
}

func TestProjectAccessPolicyService_SuperadminHasProjectPermissionsWithoutStoredGrant(t *testing.T) {
	ctx := context.Background()
	requesterID := uuid.New()
	superadminRole := domainUser.RoleSuperAdmin

	svc := NewServices(Dependencies{RolePermissions: newProjectRolePermissionRepo()}).AccessPolicy

	hasPermission, err := svc.CanUseProjectPermission(ctx, requesterID, &superadminRole, domainUser.PermissionProjectFieldDeviceEdit)
	if err != nil {
		t.Fatalf("expected superadmin permission check to succeed, got %v", err)
	}
	if !hasPermission {
		t.Fatal("expected superadmin to have project permissions without stored grant")
	}
}

func TestProjectMembershipService_InviteListRemoveUser_CharacterizesMembershipBoundary(t *testing.T) {
	ctx := context.Background()
	projectID := uuid.New()
	userID := uuid.New()

	projectRepo := newProjectRepo()
	projectRepo.items[projectID] = &domainProject.Project{Base: domain.Base{ID: projectID}}
	projectRepo.listedUsers[projectID] = []domainUser.User{{Base: domain.Base{ID: userID}, Email: "member@example.com"}}

	userRepo := newProjectUserRepo()
	userRepo.items[userID] = &domainUser.User{Base: domain.Base{ID: userID}, Email: "member@example.com"}

	svc := NewServices(Dependencies{Projects: projectRepo, Users: userRepo}).Membership

	if err := svc.InviteUser(ctx, projectID, userID); err != nil {
		t.Fatalf("expected invite to succeed, got %v", err)
	}
	if !projectRepo.hasUser(projectID, userID) {
		t.Fatal("expected invited user to be attached to the project")
	}

	users, err := svc.ListUsers(ctx, projectID)
	if err != nil {
		t.Fatalf("expected list users to succeed, got %v", err)
	}
	if len(users) != 1 || users[0].ID != userID {
		t.Fatalf("expected listed users to contain invited user, got %+v", users)
	}

	if err := svc.RemoveUser(ctx, projectID, userID); err != nil {
		t.Fatalf("expected remove user to succeed, got %v", err)
	}
	if projectRepo.hasUser(projectID, userID) {
		t.Fatal("expected removed user to be detached from the project")
	}
}

type projectUserRepoFake struct {
	items map[uuid.UUID]*domainUser.User
}

func newProjectUserRepo() *projectUserRepoFake {
	return &projectUserRepoFake{items: map[uuid.UUID]*domainUser.User{}}
}

func (r *projectUserRepoFake) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainUser.User, error) {
	out := make([]*domainUser.User, 0, len(ids))
	for _, id := range ids {
		if item, ok := r.items[id]; ok {
			clone := *item
			out = append(out, &clone)
		}
	}
	return out, nil
}

func (r *projectUserRepoFake) Create(_ context.Context, entity *domainUser.User) error {
	if entity.ID == uuid.Nil {
		entity.ID = uuid.New()
	}
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectUserRepoFake) Update(_ context.Context, entity *domainUser.User) error {
	clone := *entity
	r.items[entity.ID] = &clone
	return nil
}

func (r *projectUserRepoFake) DeleteByIds(_ context.Context, ids []uuid.UUID) error {
	for _, id := range ids {
		delete(r.items, id)
	}
	return nil
}

func (r *projectUserRepoFake) GetPaginatedList(_ context.Context, _ domain.PaginationParams) (*domain.PaginatedList[domainUser.User], error) {
	items := make([]domainUser.User, 0, len(r.items))
	for _, item := range r.items {
		items = append(items, *item)
	}
	return &domain.PaginatedList[domainUser.User]{Items: items, Total: int64(len(items)), Page: 1, TotalPages: 1}, nil
}

type projectRolePermissionRepoFake struct {
	items map[domainUser.Role][]domainUser.RolePermission
}

func newProjectRolePermissionRepo() *projectRolePermissionRepoFake {
	return &projectRolePermissionRepoFake{items: map[domainUser.Role][]domainUser.RolePermission{}}
}

func (r *projectRolePermissionRepoFake) grant(role domainUser.Role, permission string) {
	r.items[role] = append(r.items[role], domainUser.RolePermission{Role: role, Permission: permission})
}

func (r *projectRolePermissionRepoFake) Create(_ context.Context, entity *domainUser.RolePermission) error {
	r.items[entity.Role] = append(r.items[entity.Role], *entity)
	return nil
}

func (r *projectRolePermissionRepoFake) GetByIds(_ context.Context, _ []uuid.UUID) ([]*domainUser.RolePermission, error) {
	return nil, nil
}

func (r *projectRolePermissionRepoFake) Update(_ context.Context, entity *domainUser.RolePermission) error {
	r.grant(entity.Role, entity.Permission)
	return nil
}

func (r *projectRolePermissionRepoFake) DeleteByIds(_ context.Context, _ []uuid.UUID) error {
	return nil
}

func (r *projectRolePermissionRepoFake) GetPaginatedList(_ context.Context, _ domain.PaginationParams) (*domain.PaginatedList[domainUser.RolePermission], error) {
	items := make([]domainUser.RolePermission, 0)
	for _, permissions := range r.items {
		items = append(items, permissions...)
	}
	return &domain.PaginatedList[domainUser.RolePermission]{Items: items, Total: int64(len(items)), Page: 1, TotalPages: 1}, nil
}

func (r *projectRolePermissionRepoFake) ListByRole(_ context.Context, role domainUser.Role) ([]domainUser.RolePermission, error) {
	items := r.items[role]
	out := make([]domainUser.RolePermission, len(items))
	copy(out, items)
	return out, nil
}

func (r *projectRolePermissionRepoFake) ListByRoles(ctx context.Context, roles []domainUser.Role) ([]domainUser.RolePermission, error) {
	out := make([]domainUser.RolePermission, 0)
	for _, role := range roles {
		items, err := r.ListByRole(ctx, role)
		if err != nil {
			return nil, err
		}
		out = append(out, items...)
	}
	return out, nil
}

func (r *projectRolePermissionRepoFake) ReplaceRolePermissions(_ context.Context, role domainUser.Role, permissions []string) error {
	next := make([]domainUser.RolePermission, 0, len(permissions))
	for _, permission := range permissions {
		next = append(next, domainUser.RolePermission{Role: role, Permission: permission})
	}
	r.items[role] = next
	return nil
}

func (r *projectRolePermissionRepoFake) AddPermissionToRole(_ context.Context, role domainUser.Role, permission string) (*domainUser.RolePermission, error) {
	grant := domainUser.RolePermission{Role: role, Permission: permission}
	r.items[role] = append(r.items[role], grant)
	return &grant, nil
}

func (r *projectRolePermissionRepoFake) RemovePermissionFromRole(_ context.Context, role domainUser.Role, permission string) error {
	next := r.items[role][:0]
	for _, grant := range r.items[role] {
		if grant.Permission != permission {
			next = append(next, grant)
		}
	}
	r.items[role] = next
	return nil
}

func (r *projectRolePermissionRepoFake) DeleteByPermissionName(_ context.Context, permission string) error {
	for role, grants := range r.items {
		next := grants[:0]
		for _, grant := range grants {
			if grant.Permission != permission {
				next = append(next, grant)
			}
		}
		r.items[role] = next
	}
	return nil
}
