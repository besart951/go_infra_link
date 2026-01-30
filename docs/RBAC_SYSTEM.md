# RBAC System - Hierarchical Permission Management

## Overview

This system implements a hierarchical Role-Based Access Control (RBAC) with a "Can-Manage" logic for user management based on role hierarchy.

## Role Hierarchy

The system implements the following role hierarchy (from highest to lowest privilege):

```
superadmin
    └── admin_fzag
        └── fzag
            └── admin_planer
                └── planer
                    └── admin_entrepreneur
                        └── entrepreneur
```

### Role Permissions

| Role | Can Manage | Permissions |
|------|-----------|-------------|
| **superadmin** | All roles | System configuration, all user management, all resources |
| **admin_fzag** | fzag and below | User management (fzag+), team management, project management |
| **fzag** | planners and entrepreneurs | User management (planners/entrepreneurs), project management |
| **admin_planer** | planners and entrepreneurs | User management (limited), project management |
| **planer** | entrepreneurs only | Project updates, read access |
| **admin_entrepreneur** | entrepreneur only | User creation (entrepreneur), read access |
| **entrepreneur** | None | Read-only access to teams and projects |

## Backend Implementation

### Domain Layer

**Entities:**
- `User`: Enhanced with `IsActive`, `Role`, and `Teams` relationship
- `UserTeam`: Many-to-many join table between users and teams
- `Permission`: Defines specific permissions
- `RolePermission`: Links roles to permissions

**Interfaces:**
- `RBACService`: Interface for permission checking
  - `CanManageRole(requester, target Role) bool`
  - `GetAllowedRoles(requester Role) []Role`
  - `GetRoleLevel(role Role) int`

### Service Layer

Location: `backend/internal/service/rbac/service.go`

**Key Functions:**
```go
// Check if a user can manage another role
func (s *Service) CanManageRole(requesterRole, targetRole user.Role) bool

// Get list of roles a user can assign
func (s *Service) GetAllowedRoles(requesterRole user.Role) []user.Role

// Get hierarchical level of a role
func (s *Service) GetRoleLevel(role user.Role) int
```

### Middleware

Location: `backend/internal/handler/middleware/permission.go`

**Available Middleware:**
```go
// Require specific permission
RequirePermission(rbac *rbacsvc.Service, permission string)

// Require ability to manage a role
RequireCanManageRole(rbac *rbacsvc.Service, targetRole user.Role)

// Require minimum global role
RequireGlobalRole(rbac *rbacsvc.Service, minRole user.Role)

// Require minimum team role
RequireTeamRole(rbac *rbacsvc.Service, teamIDParam string, minRole team.MemberRole)
```

### API Endpoints

**New Endpoint:**
```
GET /api/v1/users/allowed-roles
```
Returns the list of roles that the authenticated user can assign to others.

**Response:**
```json
{
  "roles": ["planer", "admin_entrepreneur", "entrepreneur"]
}
```

## Frontend Implementation

### Auth Store (Svelte 5 Runes)

Location: `frontend/src/lib/stores/auth.svelte.ts`

**Usage:**
```typescript
import { auth, loadAuth, canManageRole } from '$lib/stores/auth.svelte';

// Load authentication state
await loadAuth();

// Access current user
const user = auth.user;

// Check permissions
if (canManageRole('entrepreneur')) {
  // Show UI for managing entrepreneurs
}
```

**Available Functions:**
- `loadAuth()`: Load current user and permissions
- `clearAuth()`: Clear auth state on logout
- `canManageRole(targetRole)`: Check if user can manage a role
- `hasRole(role)`: Check if user has specific role
- `hasMinRole(minRole)`: Check if user has at least minimum role level
- `isAuthenticated()`: Check if user is logged in
- `getAllowedRolesForCreation()`: Get roles user can assign

### Permission Utilities

Location: `frontend/src/lib/utils/permissions.ts`

**Functions:**
```typescript
// Check if user can perform action on resource
canPerform(action: string, resource: string): boolean

// Check if user can manage users
canManageUsers(): boolean

// Get filtered roles based on permissions
getFilteredRoles(roles: UserRole[]): UserRole[]

// Get display label for role
getRoleLabel(role: UserRole): string
```

### Permission Guard Component

Location: `frontend/src/lib/components/permission-guard.svelte`

**Usage:**
```svelte
<!-- Guard by permission -->
<PermissionGuard action="create" resource="user">
  <button>Create User</button>
</PermissionGuard>

<!-- Guard by role management ability -->
<PermissionGuard canManageRole="entrepreneur">
  <button>Assign Entrepreneur Role</button>
</PermissionGuard>
```

### User Management Form

Location: `frontend/src/lib/components/user-management-form.svelte`

A complete example implementation showing:
- Filtered role selection based on current user's permissions
- Form validation with field-level error handling
- Integration with auth store and API

## Database Schema

The system adds three new tables:

**UserTeam** (Many-to-Many between User and Team)
**Permission** (Defines available permissions)
**RolePermission** (Maps roles to permissions)

All tables are automatically created via GORM AutoMigrate.

## Security Considerations

1. **Backend Validation**: Always validate permissions on the backend. Frontend guards are for UX only.
2. **JWT Claims**: User role is stored in JWT and validated on every request.
3. **Middleware Order**: Auth middleware runs before RBAC middleware.

## Migration Guide

### Existing Users
Legacy roles (`user`, `admin`) are supported for backwards compatibility:
- `admin` → maps to same level as `admin_entrepreneur`
- `user` → lowest level (read-only)
