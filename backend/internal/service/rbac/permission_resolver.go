package rbac

import (
	"context"
	"sync"
	"time"

	domainUser "github.com/besart951/go_infra_link/backend/internal/domain/user"
	"github.com/google/uuid"
)

// RolePermissionLoader interface for loading role permissions.
type RolePermissionLoader interface {
	loadRolePermissionSets(ctx context.Context, roles []domainUser.Role) (map[domainUser.Role]permissionSet, error)
}

// RoleProvider interface for resolving role permissions.
// Implemented by *Service.
type RoleProvider interface {
	GetGlobalRole(ctx context.Context, userID uuid.UUID) (domainUser.Role, error)
	RolePermissionLoader
}

// PermissionResolver provides a single source of truth for permission checking and capability resolution.
// It caches role permissions internally to avoid repeated database queries.
type PermissionResolver struct {
	roleProvider  RoleProvider
	rolePermCache map[domainUser.Role]permissionSet
	mu            sync.RWMutex
}

// NewPermissionResolver creates a resolver with the given role provider.
func NewPermissionResolver(provider RoleProvider) *PermissionResolver {
	return &PermissionResolver{
		roleProvider:  provider,
		rolePermCache: make(map[domainUser.Role]permissionSet),
	}
}

// ActorCapabilities returns the set of capabilities an actor has over a target user.
// Capability includes: can update, can delete, can restore, can change role, can enable/disable.
func (r *PermissionResolver) ActorCapabilities(ctx context.Context, actorID uuid.UUID, target *domainUser.User) (*ActorCapabilities, error) {
	actorRole, err := r.roleProvider.GetGlobalRole(ctx, actorID)
	if err != nil {
		return nil, err
	}

	actorPerms, err := r.resolvePermissions(ctx, actorRole)
	if err != nil {
		return nil, err
	}

	caps := &ActorCapabilities{
		CanUpdate:     r.canActorUpdateUser(actorPerms, target),
		CanDelete:     r.canActorDeleteUser(actorPerms, actorRole, target),
		CanRestore:    r.canActorRestoreUser(actorPerms, actorRole, target),
		CanChangeRole: r.canActorChangeRole(actorPerms, target),
		CanEnable:     r.canActorEnableUser(actorPerms, target),
		CanDisable:    r.canActorDisableUser(actorPerms, target),
	}

	return caps, nil
}

// CanActorUpdate checks if actor can update target's profile.
func (r *PermissionResolver) CanActorUpdate(ctx context.Context, actorID uuid.UUID, target *domainUser.User) (bool, error) {
	actorRole, err := r.roleProvider.GetGlobalRole(ctx, actorID)
	if err != nil {
		return false, err
	}

	actorPerms, err := r.resolvePermissions(ctx, actorRole)
	if err != nil {
		return false, err
	}

	return r.canActorUpdateUser(actorPerms, target), nil
}

// CanActorDelete checks if actor can delete target.
func (r *PermissionResolver) CanActorDelete(ctx context.Context, actorID uuid.UUID, target *domainUser.User) (bool, error) {
	actorRole, err := r.roleProvider.GetGlobalRole(ctx, actorID)
	if err != nil {
		return false, err
	}

	actorPerms, err := r.resolvePermissions(ctx, actorRole)
	if err != nil {
		return false, err
	}

	return r.canActorDeleteUser(actorPerms, actorRole, target), nil
}

// CanActorRestore checks if actor can restore a deleted target.
func (r *PermissionResolver) CanActorRestore(ctx context.Context, actorID uuid.UUID, target *domainUser.User) (bool, error) {
	actorRole, err := r.roleProvider.GetGlobalRole(ctx, actorID)
	if err != nil {
		return false, err
	}

	actorPerms, err := r.resolvePermissions(ctx, actorRole)
	if err != nil {
		return false, err
	}

	return r.canActorRestoreUser(actorPerms, actorRole, target), nil
}

// CanActorChangeRole checks if actor can assign roles to target.
func (r *PermissionResolver) CanActorChangeRole(ctx context.Context, actorID uuid.UUID, target *domainUser.User) (bool, error) {
	actorRole, err := r.roleProvider.GetGlobalRole(ctx, actorID)
	if err != nil {
		return false, err
	}

	actorPerms, err := r.resolvePermissions(ctx, actorRole)
	if err != nil {
		return false, err
	}

	return r.canActorChangeRole(actorPerms, target), nil
}

// Private helper methods for capability computation

func (r *PermissionResolver) canActorUpdateUser(actorPerms permissionSet, target *domainUser.User) bool {
	if target.IsDeleted() || target.IsAnonymized() {
		return false
	}
	return actorPerms.has(domainUser.PermissionUserUpdate)
}

func (r *PermissionResolver) canActorDeleteUser(actorPerms permissionSet, actorRole domainUser.Role, target *domainUser.User) bool {
	if target.IsDeleted() || target.IsAnonymized() {
		return false
	}
	if !actorPerms.has(domainUser.PermissionUserDelete) {
		return false
	}
	return r.canMutateSuperAdmin(actorRole, target)
}

func (r *PermissionResolver) canActorRestoreUser(actorPerms permissionSet, _ domainUser.Role, target *domainUser.User) bool {
	if !target.IsDeleted() || target.IsAnonymized() {
		return false
	}
	if target.RestoreUntil == nil {
		return false
	}
	if time.Now().UTC().After(*target.RestoreUntil) {
		return false
	}
	if !actorPerms.has(domainUser.PermissionUserDelete) {
		return false
	}
	return actorPerms.has(domainUser.PermissionUserReadDeleted)
}

func (r *PermissionResolver) canActorChangeRole(actorPerms permissionSet, target *domainUser.User) bool {
	if target.IsDeleted() || target.IsAnonymized() {
		return false
	}
	return actorPerms.has(domainUser.PermissionUserUpdate)
}

func (r *PermissionResolver) canActorEnableUser(actorPerms permissionSet, target *domainUser.User) bool {
	if target.IsDeleted() || target.IsAnonymized() {
		return false
	}
	if target.IsActive {
		return false
	}
	return actorPerms.has(domainUser.PermissionUserUpdate)
}

func (r *PermissionResolver) canActorDisableUser(actorPerms permissionSet, target *domainUser.User) bool {
	if target.IsDeleted() || target.IsAnonymized() {
		return false
	}
	if !target.IsActive {
		return false
	}
	return actorPerms.has(domainUser.PermissionUserUpdate)
}

func (r *PermissionResolver) canMutateSuperAdmin(actorRole domainUser.Role, target *domainUser.User) bool {
	if target.Role != domainUser.RoleSuperAdmin {
		return true
	}
	// Cannot mutate if target is SuperAdmin
	return false
}

// resolvePermissions returns the cached permission set for a role, fetching if not cached.
func (r *PermissionResolver) resolvePermissions(ctx context.Context, role domainUser.Role) (permissionSet, error) {
	r.mu.RLock()
	if cached, ok := r.rolePermCache[role]; ok {
		r.mu.RUnlock()
		return cached, nil
	}
	r.mu.RUnlock()

	// Cache miss: fetch from source
	permissionSets, err := r.roleProvider.loadRolePermissionSets(ctx, []domainUser.Role{role})
	if err != nil {
		return nil, err
	}

	resolved := permissionSets[role]

	// Write to cache
	r.mu.Lock()
	r.rolePermCache[role] = resolved
	r.mu.Unlock()

	return resolved, nil
}

// ActorCapabilities describes what an actor can do to a target user.
type ActorCapabilities struct {
	CanUpdate     bool
	CanDelete     bool
	CanRestore    bool
	CanChangeRole bool
	CanEnable     bool
	CanDisable    bool
}
