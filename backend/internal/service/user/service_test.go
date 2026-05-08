package user

import (
	"context"
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainAuth "github.com/besart951/go_infra_link/backend/internal/domain/auth"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

func TestCreateWithPasswordForActorRejectsUnassignableRole(t *testing.T) {
	ctx := context.Background()
	repo := &userRepoStub{}
	service := New(repo, userPasswordHasherStub{}, &userMutationPolicyStub{
		directCreateErr: domainUser.ErrRoleNotAssignable,
	})

	err := service.CreateWithPasswordForActor(ctx, uuid.New(), &domainUser.User{
		Email: "planer@example.com",
		Role:  domainUser.RolePlaner,
	}, "password123")

	if !errors.Is(err, domainUser.ErrRoleNotAssignable) {
		t.Fatalf("expected role not assignable, got %v", err)
	}
	if repo.created != nil {
		t.Fatalf("unassignable role must not create user")
	}
}

func TestCreateWithPasswordForActorAllowsSuperadminPolicy(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	repo := &userRepoStub{}
	service := New(repo, userPasswordHasherStub{}, &userMutationPolicyStub{})

	err := service.CreateWithPasswordForActor(ctx, actorID, &domainUser.User{
		Email: "admin@example.com",
		Role:  domainUser.RoleAdminFZAG,
	}, "password123")

	if err != nil {
		t.Fatalf("expected user creation to succeed, got %v", err)
	}
	if repo.created == nil {
		t.Fatalf("expected user to be created")
	}
	if repo.created.CreatedByID == nil || *repo.created.CreatedByID != actorID {
		t.Fatalf("expected creator to be actor, got %v", repo.created.CreatedByID)
	}
	if repo.created.Password != "hashed:password123" {
		t.Fatalf("expected hashed password, got %q", repo.created.Password)
	}
}

func TestUpdatePasswordRejectsWrongCurrentPassword(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := &userRepoStub{
		user: &domainUser.User{Base: domain.Base{ID: userID}, Email: "person@example.com", Password: "hashed:old-password"},
	}
	service := New(repo, userPasswordHasherStub{}, &userMutationPolicyStub{})

	_, err := service.UpdatePassword(ctx, userID, "wrong-password", "new-password")

	if !errors.Is(err, domainAuth.ErrInvalidCredentials) {
		t.Fatalf("expected invalid credentials, got %v", err)
	}
	if repo.updated {
		t.Fatalf("wrong current password must not update user")
	}
}

func TestUpdatePasswordValidatesCurrentPasswordBeforeHashingNewPassword(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := &userRepoStub{
		user: &domainUser.User{Base: domain.Base{ID: userID}, Email: "person@example.com", Password: "hashed:old-password"},
	}
	service := New(repo, userPasswordHasherStub{}, &userMutationPolicyStub{})

	usr, err := service.UpdatePassword(ctx, userID, "old-password", "new-password")

	if err != nil {
		t.Fatalf("expected password update, got %v", err)
	}
	if !repo.updated {
		t.Fatalf("expected user update")
	}
	if usr.Password != "hashed:new-password" {
		t.Fatalf("expected hashed new password, got %q", usr.Password)
	}
}

type userRepoStub struct {
	created *domainUser.User
	user    *domainUser.User
	updated bool
}

func (s *userRepoStub) Create(_ context.Context, entity *domainUser.User) error {
	s.created = entity
	return nil
}

func (s *userRepoStub) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainUser.User, error) {
	if len(ids) == 0 || s.user == nil || s.user.ID != ids[0] {
		return []*domainUser.User{}, nil
	}
	return []*domainUser.User{s.user}, nil
}

func (s *userRepoStub) Update(_ context.Context, user *domainUser.User) error {
	s.updated = true
	s.user = user
	return nil
}

func (s *userRepoStub) DeleteByIds(context.Context, []uuid.UUID) error {
	return nil
}

func (s *userRepoStub) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainUser.User], error) {
	return &domain.PaginatedList[domainUser.User]{}, nil
}

type userMutationPolicyStub struct {
	directCreateErr  error
	updateProfileErr error
	deleteErr        error
}

func (s *userMutationPolicyStub) CanDirectCreateUser(context.Context, uuid.UUID, domainUser.Role) error {
	return s.directCreateErr
}

func (s *userMutationPolicyStub) CanUpdateProfile(context.Context, uuid.UUID, domainUser.User) error {
	return s.updateProfileErr
}

func (s *userMutationPolicyStub) CanDeleteUser(context.Context, uuid.UUID, domainUser.User) error {
	return s.deleteErr
}

type userPasswordHasherStub struct{}

func (userPasswordHasherStub) Hash(plain string) (string, error) {
	return "hashed:" + plain, nil
}

func (userPasswordHasherStub) Compare(hash, plain string) error {
	if hash != "hashed:"+plain {
		return errors.New("password_mismatch")
	}
	return nil
}
