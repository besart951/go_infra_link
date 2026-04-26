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
	outsiderID := uuid.New()

	projectRepo := newProjectRepo()
	projectRepo.items[projectID] = &domainProject.Project{Base: domain.Base{ID: projectID}, CreatorID: creatorID}
	if err := projectRepo.AddUser(ctx, projectID, memberID); err != nil {
		t.Fatalf("expected member setup to succeed, got %v", err)
	}

	userRepo := newProjectUserRepo()
	userRepo.items[adminID] = &domainUser.User{Base: domain.Base{ID: adminID}, Role: domainUser.RoleAdminFZAG}
	userRepo.items[outsiderID] = &domainUser.User{Base: domain.Base{ID: outsiderID}, Role: domainUser.RolePlaner}

	svc := NewServices(Dependencies{Projects: projectRepo, Users: userRepo}).AccessPolicy

	testCases := []struct {
		name        string
		requesterID uuid.UUID
		wantAccess  bool
	}{
		{name: "creator", requesterID: creatorID, wantAccess: true},
		{name: "member", requesterID: memberID, wantAccess: true},
		{name: "admin", requesterID: adminID, wantAccess: true},
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

	svc := NewServices(Dependencies{Projects: projectRepo}).AccessPolicy
	adminRole := domainUser.RoleAdminFZAG

	hasAccess, err := svc.CanAccessProject(ctx, requesterID, projectID, &adminRole)
	if err != nil {
		t.Fatalf("expected access check to succeed, got %v", err)
	}
	if !hasAccess {
		t.Fatal("expected access for admin role hint")
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
