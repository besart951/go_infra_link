package admin

import (
	"context"
	"errors"
	"testing"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

func TestEnableUserRejectsPendingRegistration(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	userID := uuid.New()
	userRepo := &adminUserRepoStub{
		user: &domainUser.User{Base: domain.Base{ID: userID}, Email: domainUser.EmailPtr("pending@example.com"), IsActive: false},
	}
	service := New(userRepo, &adminMutationPolicyStub{enableErr: domainUser.ErrRegistrationPending})

	err := service.EnableUser(ctx, actorID, userID)

	if !errors.Is(err, domainUser.ErrRegistrationPending) {
		t.Fatalf("expected registration pending error, got %v", err)
	}
	if userRepo.updated {
		t.Fatalf("pending registration user must not be enabled")
	}
}

func TestEnableUserAllowsAcceptedRegistration(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	userID := uuid.New()
	userRepo := &adminUserRepoStub{
		user: &domainUser.User{Base: domain.Base{ID: userID}, Email: domainUser.EmailPtr("accepted@example.com"), IsActive: false},
	}
	service := New(userRepo, &adminMutationPolicyStub{})

	err := service.EnableUser(ctx, actorID, userID)

	if err != nil {
		t.Fatalf("expected accepted registration to enable, got %v", err)
	}
	if !userRepo.user.IsActive || userRepo.user.DisabledAt != nil {
		t.Fatalf("expected user active and not disabled, got active=%v disabled=%v", userRepo.user.IsActive, userRepo.user.DisabledAt)
	}
}

func TestSetUserRoleRejectsUnassignableRole(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	userID := uuid.New()
	userRepo := &adminUserRepoStub{
		user: &domainUser.User{Base: domain.Base{ID: userID}, Email: domainUser.EmailPtr("target@example.com"), Role: domainUser.RolePlaner},
	}
	service := New(userRepo, &adminMutationPolicyStub{assignErr: domainUser.ErrRoleNotAssignable})

	err := service.SetUserRole(ctx, actorID, userID, domainUser.RoleAdminFZAG)

	if !errors.Is(err, domainUser.ErrRoleNotAssignable) {
		t.Fatalf("expected role not assignable, got %v", err)
	}
	if userRepo.updated {
		t.Fatalf("unassignable role must not update user")
	}
}

func TestSetUserRoleAllowsLowerHierarchyRole(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	userID := uuid.New()
	userRepo := &adminUserRepoStub{
		user: &domainUser.User{Base: domain.Base{ID: userID}, Email: domainUser.EmailPtr("target@example.com"), Role: domainUser.RolePlaner},
	}
	policy := &adminMutationPolicyStub{}
	service := New(userRepo, policy)

	err := service.SetUserRole(ctx, actorID, userID, domainUser.RoleEnterpreneur)

	if err != nil {
		t.Fatalf("expected lower role assignment to succeed, got %v", err)
	}
	if userRepo.user.Role != domainUser.RoleEnterpreneur {
		t.Fatalf("expected role updated, got %s", userRepo.user.Role)
	}
	if policy.assignCurrentRole != domainUser.RolePlaner || policy.assignNextRole != domainUser.RoleEnterpreneur {
		t.Fatalf("expected policy to check current+next role, got current=%s next=%s", policy.assignCurrentRole, policy.assignNextRole)
	}
}

func TestSetUserRoleRejectsTargetAboveRequesterScope(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	userID := uuid.New()
	userRepo := &adminUserRepoStub{
		user: &domainUser.User{Base: domain.Base{ID: userID}, Email: domainUser.EmailPtr("target@example.com"), Role: domainUser.RoleAdminFZAG},
	}
	service := New(userRepo, &adminMutationPolicyStub{assignErr: domainUser.ErrRoleNotAssignable})

	err := service.SetUserRole(ctx, actorID, userID, domainUser.RoleEnterpreneur)

	if !errors.Is(err, domainUser.ErrRoleNotAssignable) {
		t.Fatalf("expected role not assignable, got %v", err)
	}
	if userRepo.updated {
		t.Fatalf("target above requester scope must not be updated")
	}
}

type adminUserRepoStub struct {
	user    *domainUser.User
	updated bool
}

func (s *adminUserRepoStub) GetByIds(_ context.Context, ids []uuid.UUID) ([]*domainUser.User, error) {
	if len(ids) == 0 || s.user == nil || s.user.ID != ids[0] {
		return []*domainUser.User{}, nil
	}
	return []*domainUser.User{s.user}, nil
}

func (s *adminUserRepoStub) Update(_ context.Context, user *domainUser.User) error {
	s.updated = true
	s.user = user
	return nil
}

type adminMutationPolicyStub struct {
	disableErr        error
	enableErr         error
	assignErr         error
	assignCurrentRole domainUser.Role
	assignNextRole    domainUser.Role
}

func (s *adminMutationPolicyStub) CanDisableUser(context.Context, uuid.UUID, domainUser.User) error {
	return s.disableErr
}

func (s *adminMutationPolicyStub) CanEnableUser(context.Context, uuid.UUID, domainUser.User) error {
	return s.enableErr
}

func (s *adminMutationPolicyStub) CanAssignRole(_ context.Context, _ uuid.UUID, currentRole, nextRole domainUser.Role) error {
	s.assignCurrentRole = currentRole
	s.assignNextRole = nextRole
	return s.assignErr
}
