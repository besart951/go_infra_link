import type { RoleRepository } from '$lib/domain/ports/role/roleRepository.js';
import type {
    Role,
    Permission,
    CreatePermissionRequest,
    UpdatePermissionRequest,
    UpdateRolePermissionsRequest,
    RolePermission
} from '$lib/domain/role/index.js';
import type { UserRole } from '$lib/domain/user/index.js';
import { api } from '$lib/api/client.js';

export const roleRepository: RoleRepository = {
    // ── Roles ────────────────────────────────────────────────────────────
    async listRoles(signal?: AbortSignal): Promise<Role[]> {
        return api<Role[]>('/roles', { signal });
    },

    async getRole(roleName: UserRole, signal?: AbortSignal): Promise<Role | null> {
        const roles = await this.listRoles(signal);
        return roles.find((r) => r.name === roleName) || null;
    },

    // ── Permissions ──────────────────────────────────────────────────────
    async listPermissions(signal?: AbortSignal): Promise<Permission[]> {
        return api<Permission[]>('/permissions', { signal });
    },

    async createPermission(
        data: CreatePermissionRequest,
        signal?: AbortSignal
    ): Promise<Permission> {
        return api<Permission>('/permissions', {
            method: 'POST',
            body: JSON.stringify(data),
            signal
        });
    },

    async updatePermission(
        id: string,
        data: UpdatePermissionRequest,
        signal?: AbortSignal
    ): Promise<Permission> {
        return api<Permission>(`/permissions/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
            signal
        });
    },

    async deletePermission(id: string, signal?: AbortSignal): Promise<void> {
        return api<void>(`/permissions/${id}`, {
            method: 'DELETE',
            signal
        });
    },

    // ── Role-Permission assignments ──────────────────────────────────────
    async updateRolePermissions(
        roleName: UserRole,
        data: UpdateRolePermissionsRequest,
        signal?: AbortSignal
    ): Promise<Role> {
        return api<Role>(`/roles/${roleName}/permissions`, {
            method: 'PUT',
            body: JSON.stringify(data),
            signal
        });
    },

    async assignPermissionToRole(
        roleName: UserRole,
        permissionName: string,
        signal?: AbortSignal
    ): Promise<RolePermission> {
        return api<RolePermission>(`/roles/${roleName}/permissions`, {
            method: 'POST',
            body: JSON.stringify({ permission: permissionName }),
            signal
        });
    },

    async removePermissionFromRole(
        roleName: UserRole,
        permissionName: string,
        signal?: AbortSignal
    ): Promise<void> {
        return api<void>(
            `/roles/${roleName}/permissions/${encodeURIComponent(permissionName)}`,
            {
                method: 'DELETE',
                signal
            }
        );
    }
};
