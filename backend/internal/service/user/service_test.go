package user

import (
	"context"
	"errors"
	"testing"
	"time"

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
		Email: domainUser.EmailPtr("planer@example.com"),
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
		Email: domainUser.EmailPtr("admin@example.com"),
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

func TestCreateWithPasswordReturnsRestorableConflictForDeletedUser(t *testing.T) {
	ctx := context.Background()
	restoreUntil := time.Now().UTC().Add(time.Hour)
	deletedAt := time.Now().UTC().Add(-time.Hour)
	repo := &userRepoStub{
		user: &domainUser.User{
			Base:         domain.Base{ID: uuid.New()},
			Email:        domainUser.EmailPtr("person@example.com"),
			DeletedAt:    &deletedAt,
			RestoreUntil: &restoreUntil,
		},
	}
	service := New(repo, userPasswordHasherStub{}, &userMutationPolicyStub{})

	err := service.CreateWithPassword(ctx, &domainUser.User{Email: domainUser.EmailPtr(" person@example.com ")}, "password123")

	if !errors.Is(err, domainUser.ErrDeletedUserRestorable) {
		t.Fatalf("expected restorable deleted user error, got %v", err)
	}
	if repo.created != nil {
		t.Fatalf("restorable deleted user must not create duplicate account")
	}
}

func TestUpdatePasswordRejectsWrongCurrentPassword(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	repo := &userRepoStub{
		user: &domainUser.User{Base: domain.Base{ID: userID}, Email: domainUser.EmailPtr("person@example.com"), Password: "hashed:old-password"},
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
		user: &domainUser.User{Base: domain.Base{ID: userID}, Email: domainUser.EmailPtr("person@example.com"), Password: "hashed:old-password"},
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

func TestDeleteByIDForActorMarksUserAsDeleted(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	userID := uuid.New()
	repo := &userRepoStub{
		user: &domainUser.User{Base: domain.Base{ID: userID}, Email: domainUser.EmailPtr("delete-me@example.com"), IsActive: true},
	}
	service := New(repo, userPasswordHasherStub{}, &userMutationPolicyStub{})

	err := service.DeleteByIDForActor(ctx, actorID, userID)

	if err != nil {
		t.Fatalf("expected delete to succeed, got %v", err)
	}
	if !repo.updated {
		t.Fatalf("expected user update during soft delete")
	}
	if repo.user.DeletedAt == nil {
		t.Fatalf("expected deleted_at to be set")
	}
	if repo.user.DeletedByID == nil || *repo.user.DeletedByID != actorID {
		t.Fatalf("expected deleted_by_id to match actor")
	}
	if repo.user.IsActive {
		t.Fatalf("expected user to become inactive")
	}
	if repo.user.RestoreUntil == nil || repo.user.ScheduledPurgeAt == nil {
		t.Fatalf("expected restore and purge timestamps to be set")
	}
}

func TestRestoreByIDForActorFailsWhenWindowExpired(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	userID := uuid.New()
	past := time.Now().UTC().Add(-time.Hour)
	deletedAt := time.Now().UTC().Add(-2 * time.Hour)
	repo := &userRepoStub{
		user: &domainUser.User{
			Base:         domain.Base{ID: userID},
			Email:        domainUser.EmailPtr("deleted@example.com"),
			DeletedAt:    &deletedAt,
			RestoreUntil: &past,
		},
	}
	service := New(repo, userPasswordHasherStub{}, &userMutationPolicyStub{})

	err := service.RestoreByIDForActor(ctx, actorID, userID)

	if !errors.Is(err, domainUser.ErrRestoreWindowExpired) {
		t.Fatalf("expected restore window error, got %v", err)
	}
	if repo.updated {
		t.Fatalf("expired restore window must not update user")
	}
}

func TestRestoreByIDForActorClearsDeletionMarkers(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	userID := uuid.New()
	deletedAt := time.Now().UTC().Add(-time.Hour)
	restoreUntil := time.Now().UTC().Add(time.Hour)
	deletedBy := uuid.New()
	hash := "abc"
	repo := &userRepoStub{
		user: &domainUser.User{
			Base:             domain.Base{ID: userID},
			Email:            domainUser.EmailPtr("deleted@example.com"),
			DeletedAt:        &deletedAt,
			RestoreUntil:     &restoreUntil,
			DeletedByID:      &deletedBy,
			DeletedEmailHash: &hash,
			IsActive:         false,
		},
	}
	policy := &userMutationPolicyStub{}
	service := New(repo, userPasswordHasherStub{}, policy)

	err := service.RestoreByIDForActor(ctx, actorID, userID)

	if err != nil {
		t.Fatalf("expected restore success, got %v", err)
	}
	if !repo.updated {
		t.Fatalf("expected repository update")
	}
	if !policy.restoreCalled {
		t.Fatalf("expected restore policy to be checked")
	}
	if repo.user.DeletedAt != nil || repo.user.RestoreUntil != nil || repo.user.DeletedByID != nil || repo.user.ScheduledPurgeAt != nil {
		t.Fatalf("expected deletion markers to be cleared")
	}
	if !repo.user.IsActive || repo.user.DisabledAt != nil {
		t.Fatalf("expected user to be active and enabled")
	}
}

func TestListKeepsRepositoryPaginationTotals(t *testing.T) {
	ctx := context.Background()
	repo := &userRepoStub{
		listResult: &domain.PaginatedList[domainUser.User]{
			Items: []domainUser.User{
				{Base: domain.Base{ID: uuid.New()}, Email: domainUser.EmailPtr("one@example.com")},
				{Base: domain.Base{ID: uuid.New()}, Email: domainUser.EmailPtr("two@example.com")},
			},
			Total:      25,
			Page:       2,
			TotalPages: 3,
		},
	}
	service := New(repo, userPasswordHasherStub{}, &userMutationPolicyStub{})

	result, err := service.List(ctx, 2, 10, "", "", "", false)

	if err != nil {
		t.Fatalf("expected list success, got %v", err)
	}
	if result.Total != 25 || result.TotalPages != 3 {
		t.Fatalf("expected repository totals to be preserved, got total=%d pages=%d", result.Total, result.TotalPages)
	}
}

func TestPurgeDeletedUsersAnonymizesDueUsers(t *testing.T) {
	ctx := context.Background()
	now := time.Now().UTC()
	userID := uuid.New()
	deleteAt := now.Add(-48 * time.Hour)
	scheduled := now.Add(-time.Hour)
	repo := &userRepoStub{
		due: []*domainUser.User{{
			Base:             domain.Base{ID: userID},
			FirstName:        "John",
			LastName:         "Doe",
			Email:            domainUser.EmailPtr("john@example.com"),
			DeletedAt:        &deleteAt,
			ScheduledPurgeAt: &scheduled,
		}},
	}
	service := New(repo, userPasswordHasherStub{}, &userMutationPolicyStub{})

	err := service.PurgeDeletedUsers(ctx, 10)

	if err != nil {
		t.Fatalf("expected purge success, got %v", err)
	}
	if !repo.updated {
		t.Fatalf("expected anonymized user update")
	}
	if repo.user.AnonymizedAt == nil {
		t.Fatalf("expected anonymized_at to be set")
	}
	if repo.user.FirstName != "Deleted" || repo.user.LastName != "User" {
		t.Fatalf("expected projected anonymized name, got %s %s", repo.user.FirstName, repo.user.LastName)
	}
	if repo.user.Email != nil {
		t.Fatalf("expected email to be cleared")
	}
}

type userRepoStub struct {
	created    *domainUser.User
	user       *domainUser.User
	due        []*domainUser.User
	listResult *domain.PaginatedList[domainUser.User]
	updated    bool
}

func (s *userRepoStub) Create(_ context.Context, entity *domainUser.User) error {
	s.created = entity
	return nil
}

func (s *userRepoStub) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainUser.User, error) {
	if len(ids) == 0 {
		return []*domainUser.User{}, nil
	}

	// Check if the user is in the main user field
	if s.user != nil && s.user.ID == ids[0] {
		return []*domainUser.User{s.user}, nil
	}

	// Check if the user is in the due list
	for _, dueUser := range s.due {
		if dueUser != nil && dueUser.ID == ids[0] {
			return []*domainUser.User{dueUser}, nil
		}
	}

	return []*domainUser.User{}, nil
}

func (s *userRepoStub) Update(_ context.Context, user *domainUser.User) error {
	s.updated = true
	s.user = user
	return nil
}

func (s *userRepoStub) DeleteByIds(context.Context, []uuid.UUID) error {
	return nil
}

func (s *userRepoStub) GetByEmail(_ context.Context, email string) (*domainUser.User, error) {
	if s.user == nil || s.user.EmailValue() != email {
		return nil, domain.ErrNotFound
	}
	return s.user, nil
}

func (s *userRepoStub) ListDueForAnonymization(context.Context, time.Time, int) ([]*domainUser.User, error) {
	return s.due, nil
}

func (s *userRepoStub) GetPaginatedList(context.Context, domain.PaginationParams) (*domain.PaginatedList[domainUser.User], error) {
	if s.listResult != nil {
		return s.listResult, nil
	}
	return &domain.PaginatedList[domainUser.User]{}, nil
}

type userMutationPolicyStub struct {
	directCreateErr  error
	updateProfileErr error
	deleteErr        error
	restoreErr       error
	restoreCalled    bool
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

func (s *userMutationPolicyStub) CanRestoreUser(context.Context, uuid.UUID, domainUser.User) error {
	s.restoreCalled = true
	return s.restoreErr
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
