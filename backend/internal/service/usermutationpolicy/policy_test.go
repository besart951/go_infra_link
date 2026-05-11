package usermutationpolicy

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

func TestCanInviteUserUsesStrictHierarchy(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	tests := []struct {
		name       string
		actorRole  domainUser.Role
		targetRole domainUser.Role
		wantErr    error
	}{
		{"superadmin can invite superadmin", domainUser.RoleSuperAdmin, domainUser.RoleSuperAdmin, nil},
		{"admin_fzag can invite fzag", domainUser.RoleAdminFZAG, domainUser.RoleFZAG, nil},
		{"admin_fzag can invite entrepreneur admin", domainUser.RoleAdminFZAG, domainUser.RoleAdminEnterpreneur, nil},
		{"admin_fzag cannot invite itself", domainUser.RoleAdminFZAG, domainUser.RoleAdminFZAG, domainUser.ErrRoleNotAssignable},
		{"admin_planer cannot invite fzag", domainUser.RoleAdminPlaner, domainUser.RoleFZAG, domainUser.ErrRoleNotAssignable},
		{"planer cannot invite without user.create", domainUser.RolePlaner, domainUser.RoleEnterpreneur, domainUser.ErrRoleNotAssignable},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roles := &roleProviderStub{
				roleByUser: map[uuid.UUID]domainUser.Role{actorID: tt.actorRole},
				permissionsByRole: map[domainUser.Role]map[string]bool{
					domainUser.RoleAdminFZAG: {
						domainUser.PermissionUserCreate: true,
					},
					domainUser.RoleAdminPlaner: {
						domainUser.PermissionUserCreate: true,
					},
				},
			}
			policy := New(roles, nil)

			err := policy.CanInviteUser(ctx, actorID, tt.targetRole)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

func TestCanDirectCreateUserIsSuperadminOnly(t *testing.T) {
	ctx := context.Background()
	superadminID := uuid.New()
	adminID := uuid.New()
	policy := New(&roleProviderStub{
		roleByUser: map[uuid.UUID]domainUser.Role{
			superadminID: domainUser.RoleSuperAdmin,
			adminID:      domainUser.RoleAdminFZAG,
		},
	}, nil)

	if err := policy.CanDirectCreateUser(ctx, superadminID, domainUser.RoleSuperAdmin); err != nil {
		t.Fatalf("expected superadmin direct create, got %v", err)
	}
	if err := policy.CanDirectCreateUser(ctx, adminID, domainUser.RoleEnterpreneur); !errors.Is(err, domainUser.ErrRoleNotAssignable) {
		t.Fatalf("expected non-superadmin direct create reject, got %v", err)
	}
}

func TestCanAssignRoleChecksCurrentAndNextRole(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	policy := New(&roleProviderStub{
		roleByUser: map[uuid.UUID]domainUser.Role{
			actorID: domainUser.RoleAdminPlaner,
		},
		permissionsByRole: map[domainUser.Role]map[string]bool{
			domainUser.RoleAdminPlaner: {
				domainUser.PermissionUserUpdate: true,
			},
		},
	}, nil)

	if err := policy.CanAssignRole(ctx, actorID, domainUser.RolePlaner, domainUser.RoleEnterpreneur); err != nil {
		t.Fatalf("expected role change inside scope, got %v", err)
	}
	if err := policy.CanAssignRole(ctx, actorID, domainUser.RoleAdminFZAG, domainUser.RoleEnterpreneur); !errors.Is(err, domainUser.ErrRoleNotAssignable) {
		t.Fatalf("expected current role outside scope reject, got %v", err)
	}
	if err := policy.CanAssignRole(ctx, actorID, domainUser.RolePlaner, domainUser.RoleFZAG); !errors.Is(err, domainUser.ErrRoleNotAssignable) {
		t.Fatalf("expected next role outside scope reject, got %v", err)
	}
}

func TestCanReadRegistrationProcessUsesStrictHierarchy(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	targetID := uuid.New()
	policy := New(&roleProviderStub{
		roleByUser: map[uuid.UUID]domainUser.Role{
			actorID: domainUser.RoleAdminPlaner,
		},
		permissionsByRole: map[domainUser.Role]map[string]bool{
			domainUser.RoleAdminPlaner: {
				domainUser.PermissionUserRead: true,
			},
		},
	}, nil)

	if err := policy.CanReadRegistrationProcess(ctx, actorID, domainUser.User{
		Base: domain.Base{ID: targetID},
		Role: domainUser.RolePlaner,
	}); err != nil {
		t.Fatalf("expected process read inside scope, got %v", err)
	}
	if err := policy.CanReadRegistrationProcess(ctx, actorID, domainUser.User{
		Base: domain.Base{ID: targetID},
		Role: domainUser.RoleAdminFZAG,
	}); !errors.Is(err, domainUser.ErrRoleNotAssignable) {
		t.Fatalf("expected process read outside scope reject, got %v", err)
	}
}

func TestCanEnableUserRejectsPendingRegistration(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	userID := uuid.New()
	policy := New(&roleProviderStub{
		roleByUser: map[uuid.UUID]domainUser.Role{
			actorID: domainUser.RoleAdminPlaner,
		},
		permissionsByRole: map[domainUser.Role]map[string]bool{
			domainUser.RoleAdminPlaner: {
				domainUser.PermissionUserUpdate: true,
			},
		},
	}, &invitationReaderStub{
		invitation: &domainUser.UserInvitation{UserID: userID},
	})

	err := policy.CanEnableUser(ctx, actorID, domainUser.User{
		Base: domain.Base{ID: userID},
		Role: domainUser.RolePlaner,
	})
	if !errors.Is(err, domainUser.ErrRegistrationPending) {
		t.Fatalf("expected registration pending, got %v", err)
	}
}

func TestCanEnableUserAllowsAcceptedRegistration(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	userID := uuid.New()
	acceptedAt := time.Now().UTC()
	policy := New(&roleProviderStub{
		roleByUser: map[uuid.UUID]domainUser.Role{
			actorID: domainUser.RoleAdminPlaner,
		},
		permissionsByRole: map[domainUser.Role]map[string]bool{
			domainUser.RoleAdminPlaner: {
				domainUser.PermissionUserUpdate: true,
			},
		},
	}, &invitationReaderStub{
		invitation: &domainUser.UserInvitation{UserID: userID, AcceptedAt: &acceptedAt},
	})

	err := policy.CanEnableUser(ctx, actorID, domainUser.User{
		Base: domain.Base{ID: userID},
		Role: domainUser.RolePlaner,
	})
	if err != nil {
		t.Fatalf("expected accepted registration enable, got %v", err)
	}
}

func TestCanRestoreUserRequiresDeleteAndReadDeleted(t *testing.T) {
	ctx := context.Background()
	actorID := uuid.New()
	target := domainUser.User{Base: domain.Base{ID: uuid.New()}, Role: domainUser.RolePlaner}

	tests := []struct {
		name        string
		permissions map[string]bool
		wantErr     error
	}{
		{
			name: "allows delete and read deleted",
			permissions: map[string]bool{
				domainUser.PermissionUserDelete:      true,
				domainUser.PermissionUserReadDeleted: true,
			},
		},
		{
			name: "rejects missing delete",
			permissions: map[string]bool{
				domainUser.PermissionUserReadDeleted: true,
			},
			wantErr: domainUser.ErrRoleNotAssignable,
		},
		{
			name: "rejects missing read deleted",
			permissions: map[string]bool{
				domainUser.PermissionUserDelete: true,
			},
			wantErr: domainUser.ErrRoleNotAssignable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			policy := New(&roleProviderStub{
				roleByUser: map[uuid.UUID]domainUser.Role{
					actorID: domainUser.RoleAdminPlaner,
				},
				permissionsByRole: map[domainUser.Role]map[string]bool{
					domainUser.RoleAdminPlaner: tt.permissions,
				},
			}, nil)

			err := policy.CanRestoreUser(ctx, actorID, target)
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("expected %v, got %v", tt.wantErr, err)
			}
		})
	}
}

type roleProviderStub struct {
	roleByUser        map[uuid.UUID]domainUser.Role
	permissionsByRole map[domainUser.Role]map[string]bool
}

func (s *roleProviderStub) GetGlobalRole(_ context.Context, userID uuid.UUID) (domainUser.Role, error) {
	return s.roleByUser[userID], nil
}

func (s *roleProviderStub) HasPermission(_ context.Context, role domainUser.Role, permission string) (bool, error) {
	if role == domainUser.RoleSuperAdmin {
		return true, nil
	}
	return s.permissionsByRole[role][permission], nil
}

type invitationReaderStub struct {
	invitation *domainUser.UserInvitation
}

func (s *invitationReaderStub) GetInvitationByUserID(_ context.Context, userID uuid.UUID) (*domainUser.UserInvitation, error) {
	if s.invitation == nil || s.invitation.UserID != userID {
		return nil, domain.ErrNotFound
	}
	return s.invitation, nil
}
