import type {
    Role,
    Permission,
    CreatePermissionRequest,
    UpdatePermissionRequest,
    UpdateRolePermissionsRequest,
    RolePermission
} from '$lib/domain/role/index.js';
import type { UserRole } from '$lib/domain/user/index.js';

/**
 * RoleRepository does NOT extend ListRepository<T> because roles
 * use flat arrays (not paginated lists) and are keyed by UserRole name, not UUID.
 */
export interface RoleRepository {
    // Roles
    listRoles(signal?: AbortSignal): Promise<Role[]>;
    getRole(roleName: UserRole, signal?: AbortSignal): Promise<Role | null>;

    // Permissions
    listPermissions(signal?: AbortSignal): Promise<Permission[]>;
    createPermission(data: CreatePermissionRequest, signal?: AbortSignal): Promise<Permission>;
    updatePermission(
        id: string,
        data: UpdatePermissionRequest,
        signal?: AbortSignal
    ): Promise<Permission>;
    deletePermission(id: string, signal?: AbortSignal): Promise<void>;

    // Role-Permission assignments
    updateRolePermissions(
        roleName: UserRole,
        data: UpdateRolePermissionsRequest,
        signal?: AbortSignal
    ): Promise<Role>;
    assignPermissionToRole(
        roleName: UserRole,
        permissionName: string,
        signal?: AbortSignal
    ): Promise<RolePermission>;
    removePermissionFromRole(
        roleName: UserRole,
        permissionName: string,
        signal?: AbortSignal
    ): Promise<void>;
}
