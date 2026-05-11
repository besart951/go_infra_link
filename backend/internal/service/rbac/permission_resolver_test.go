package rbac

import (
	"context"
	"testing"
	"time"

	"github.com/besart951/go_infra_link/backend/internal/domain"
	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

func TestPermissionResolverCanActorUpdate(t *testing.T) {
	ctx := context.Background()
	service := &mockService{
		roles: map[uuid.UUID]domainUser.Role{
			uuid.MustParse("00000000-0000-0000-0000-000000000001"): domainUser.RoleSuperAdmin,
		},
		permissionSets: map[domainUser.Role]permissionSet{
			domainUser.RoleSuperAdmin: permissionSet{domainUser.PermissionUserUpdate: struct{}{}},
		},
	}
	resolver := NewPermissionResolver(service)

	actorID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	target := &domainUser.User{
		Base:     domain.Base{ID: uuid.New()},
		Role:     domainUser.RolePlaner,
		IsActive: true,
	}

	can, err := resolver.CanActorUpdate(ctx, actorID, target)
	if err != nil {
		t.Fatalf("CanActorUpdate returned error: %v", err)
	}

	if !can {
		t.Fatal("expected can update to be true for superadmin")
	}
}

func TestPermissionResolverCannotUpdateDeletedUser(t *testing.T) {
	ctx := context.Background()
	service := &mockService{
		roles: map[uuid.UUID]domainUser.Role{
			uuid.MustParse("00000000-0000-0000-0000-000000000001"): domainUser.RoleSuperAdmin,
		},
		permissionSets: map[domainUser.Role]permissionSet{
			domainUser.RoleSuperAdmin: permissionSet{domainUser.PermissionUserUpdate: struct{}{}},
		},
	}
	resolver := NewPermissionResolver(service)

	actorID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	deletedAt := time.Now()
	target := &domainUser.User{
		Base:      domain.Base{ID: uuid.New()},
		Role:      domainUser.RolePlaner,
		IsActive:  true,
		DeletedAt: &deletedAt,
	}

	can, err := resolver.CanActorUpdate(ctx, actorID, target)
	if err != nil {
		t.Fatalf("CanActorUpdate returned error: %v", err)
	}

	if can {
		t.Fatal("expected can update to be false for deleted user")
	}
}

func TestPermissionResolverCanRestore(t *testing.T) {
	ctx := context.Background()
	service := &mockService{
		roles: map[uuid.UUID]domainUser.Role{
			uuid.MustParse("00000000-0000-0000-0000-000000000001"): domainUser.RoleSuperAdmin,
		},
		permissionSets: map[domainUser.Role]permissionSet{
			domainUser.RoleSuperAdmin: permissionSet{
				domainUser.PermissionUserDelete:      struct{}{},
				domainUser.PermissionUserReadDeleted: struct{}{},
			},
		},
	}
	resolver := NewPermissionResolver(service)

	actorID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	now := time.Now()
	restoreUntil := now.Add(time.Hour)
	deletedAt := now.Add(-time.Minute)
	target := &domainUser.User{
		Base:         domain.Base{ID: uuid.New()},
		Role:         domainUser.RolePlaner,
		DeletedAt:    &deletedAt,
		RestoreUntil: &restoreUntil,
	}

	can, err := resolver.CanActorRestore(ctx, actorID, target)
	if err != nil {
		t.Fatalf("CanActorRestore returned error: %v", err)
	}

	if !can {
		t.Fatal("expected can restore to be true within restore window")
	}
}

func TestPermissionResolverCannotRestoreAfterWindow(t *testing.T) {
	ctx := context.Background()
	service := &mockService{
		roles: map[uuid.UUID]domainUser.Role{
			uuid.MustParse("00000000-0000-0000-0000-000000000001"): domainUser.RoleSuperAdmin,
		},
		permissionSets: map[domainUser.Role]permissionSet{
			domainUser.RoleSuperAdmin: permissionSet{
				domainUser.PermissionUserDelete:      struct{}{},
				domainUser.PermissionUserReadDeleted: struct{}{},
			},
		},
	}
	resolver := NewPermissionResolver(service)

	actorID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	now := time.Now()
	restoreUntil := now.Add(-time.Hour) // window expired
	deletedAt := now.Add(-2 * time.Hour)
	target := &domainUser.User{
		Base:         domain.Base{ID: uuid.New()},
		Role:         domainUser.RolePlaner,
		DeletedAt:    &deletedAt,
		RestoreUntil: &restoreUntil,
	}

	can, err := resolver.CanActorRestore(ctx, actorID, target)
	if err != nil {
		t.Fatalf("CanActorRestore returned error: %v", err)
	}

	if can {
		t.Fatal("expected can restore to be false after restore window expires")
	}
}

// Mock service for testing
type mockService struct {
	roles          map[uuid.UUID]domainUser.Role
	permissionSets map[domainUser.Role]permissionSet
}

func (m *mockService) GetGlobalRole(ctx context.Context, userID uuid.UUID) (domainUser.Role, error) {
	role, ok := m.roles[userID]
	if !ok {
		return "", domain.ErrNotFound
	}
	return role, nil
}

func (m *mockService) loadRolePermissionSets(ctx context.Context, roles []domainUser.Role) (map[domainUser.Role]permissionSet, error) {
	result := make(map[domainUser.Role]permissionSet)
	for _, role := range roles {
		if ps, ok := m.permissionSets[role]; ok {
			result[role] = ps
		} else {
			result[role] = permissionSet{}
		}
	}
	return result, nil
}
