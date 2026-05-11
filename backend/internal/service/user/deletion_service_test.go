package user

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

func TestDeletionService_DeleteByID_MarksUserDeleted(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	usr := &user.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     stringPtr("john@example.com"),
		IsActive:  true,
	}
	usr.ID = userID

	repo := &userRepoStub{
		user: usr,
	}
	service := NewDeletionService(repo, nil)

	err := service.DeleteByID(ctx, usr.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if usr.DeletedAt == nil {
		t.Fatalf("DeletedAt should be set")
	}
	if usr.IsActive {
		t.Fatalf("user should be inactive after deletion")
	}
}

func TestDeletionService_DeleteByID_IsIdempotent(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	now := time.Now().UTC()
	usr := &user.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     stringPtr("john@example.com"),
		DeletedAt: &now,
	}
	usr.ID = userID

	repo := &userRepoStub{
		user: usr,
	}
	service := NewDeletionService(repo, nil)

	err := service.DeleteByID(ctx, usr.ID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should still be marked as updated (idempotent returns nil without update)
	// Actually for idempotent, we check if already deleted and skip update
	if repo.updated {
		t.Fatalf("update should not be called for already-deleted user (idempotent)")
	}
}

func TestDeletionService_DeleteByID_RejectsAnonymizedUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	now := time.Now().UTC()
	usr := &user.User{
		FirstName:    "Deleted",
		LastName:     "User",
		DeletedAt:    &now,
		AnonymizedAt: &now,
		IsActive:     false,
	}
	usr.ID = userID

	repo := &userRepoStub{
		user: usr,
	}
	service := NewDeletionService(repo, nil)

	err := service.DeleteByID(ctx, usr.ID)
	if !errors.Is(err, user.ErrUserAlreadyAnonymized) {
		t.Fatalf("expected ErrUserAlreadyAnonymized, got %v", err)
	}
}

func TestDeletionService_DeleteByIDForActor_EnforcesPermission(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	userID := uuid.New()
	usr := &user.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     stringPtr("john@example.com"),
		IsActive:  true,
	}
	usr.ID = userID

	repo := &userRepoStub{
		user: usr,
	}

	policy := &mockMutationPolicy{
		canDeleteError: user.ErrRoleNotAssignable,
	}
	service := NewDeletionService(repo, policy)

	err := service.DeleteByIDForActor(ctx, actorID, userID)
	if !errors.Is(err, user.ErrRoleNotAssignable) {
		t.Fatalf("expected ErrRoleNotAssignable, got %v", err)
	}
}

func TestDeletionService_RestoreByIDForActor_ValidatesWindow(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	userID := uuid.New()

	// Create user deleted more than 30 days ago
	expiredWindow := time.Now().UTC().Add(-35 * 24 * time.Hour)
	usr := &user.User{
		FirstName:    "John",
		LastName:     "Doe",
		Email:        stringPtr("john@example.com"),
		DeletedAt:    &expiredWindow,
		RestoreUntil: &expiredWindow,
		IsActive:     false,
	}
	usr.ID = userID

	repo := &userRepoStub{
		user: usr,
	}

	policy := &mockMutationPolicy{}
	service := NewDeletionService(repo, policy)

	err := service.RestoreByIDForActor(ctx, actorID, userID)
	if !errors.Is(err, user.ErrRestoreWindowExpired) {
		t.Fatalf("expected ErrRestoreWindowExpired, got %v", err)
	}
}

func TestDeletionService_RestoreByIDForActor_ClearsMarkers(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	userID := uuid.New()

	// Create user deleted recently (within restore window)
	now := time.Now().UTC()
	restoreUntil := now.Add(24 * time.Hour)
	usr := &user.User{
		FirstName:    "John",
		LastName:     "Doe",
		Email:        stringPtr("john@example.com"),
		DeletedAt:    &now,
		RestoreUntil: &restoreUntil,
		IsActive:     false,
	}
	usr.ID = userID

	repo := &userRepoStub{
		user: usr,
	}

	policy := &mockMutationPolicy{}
	service := NewDeletionService(repo, policy)

	err := service.RestoreByIDForActor(ctx, actorID, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if usr.DeletedAt != nil {
		t.Fatalf("DeletedAt should be cleared")
	}
	if !usr.IsActive {
		t.Fatalf("user should be active after restore")
	}
	if usr.DisabledAt != nil {
		t.Fatalf("DisabledAt should be cleared")
	}
}

func TestDeletionService_Anonymize_AnonymizesDeletedUser(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	now := time.Now().UTC()

	usr := &user.User{
		FirstName: "John",
		LastName:  "Doe",
		Email:     stringPtr("john@example.com"),
		DeletedAt: &now,
		IsActive:  false,
	}
	usr.ID = userID

	repo := &userRepoStub{
		user: usr,
	}
	service := NewDeletionService(repo, nil)

	err := service.Anonymize(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if usr.FirstName != "Deleted" || usr.LastName != "User" {
		t.Fatalf("user identity should be anonymized")
	}
	if usr.Email != nil {
		t.Fatalf("email should be cleared")
	}
	if usr.AnonymizedAt == nil {
		t.Fatalf("AnonymizedAt should be set")
	}
}

func TestDeletionService_Anonymize_IsIdempotent(t *testing.T) {
	ctx := context.Background()
	userID := uuid.New()
	now := time.Now().UTC()

	usr := &user.User{
		FirstName:    "Deleted",
		LastName:     "User",
		Email:        nil,
		DeletedAt:    &now,
		AnonymizedAt: &now,
		IsActive:     false,
	}
	usr.ID = userID

	repo := &userRepoStub{
		user: usr,
	}
	service := NewDeletionService(repo, nil)

	err := service.Anonymize(ctx, userID)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Should not update (idempotent)
	if repo.updated {
		t.Fatalf("update should not be called for already-anonymized user (idempotent)")
	}
}

// Mock mutation policy for testing
type mockMutationPolicy struct {
	canDeleteError  error
	canRestoreError error
}

func (m *mockMutationPolicy) CanDirectCreateUser(ctx context.Context, actorID uuid.UUID, targetRole user.Role) error {
	return nil
}

func (m *mockMutationPolicy) CanUpdateProfile(ctx context.Context, actorID uuid.UUID, target user.User) error {
	return nil
}

func (m *mockMutationPolicy) CanDeleteUser(ctx context.Context, actorID uuid.UUID, target user.User) error {
	return m.canDeleteError
}

func (m *mockMutationPolicy) CanRestoreUser(ctx context.Context, actorID uuid.UUID, target user.User) error {
	return m.canRestoreError
}

// Helper for tests
func stringPtr(s string) *string {
	return &s
}
