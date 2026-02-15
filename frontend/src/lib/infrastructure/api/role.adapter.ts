/**
 * Role API adapter
 * Infrastructure layer - implements role and permission data operations via HTTP
 */
import { api } from '$lib/api/client.js';
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
 * List all roles with their permissions
 */
export async function listRoles(): Promise<Role[]> {
	return api<Role[]>('/roles');
}

/**
 * Get a single role by name
 */
export async function getRole(roleName: UserRole): Promise<Role | null> {
	const roles = await listRoles();
	return roles.find((r) => r.name === roleName) || null;
}

/**
 * List all permissions
 */
export async function listPermissions(): Promise<Permission[]> {
	return api<Permission[]>('/permissions');
}

/**
 * Create a new permission
 * Note: This would call the backend API in a real implementation
 */
export async function createPermission(
	data: CreatePermissionRequest,
	options?: RequestInit
): Promise<Permission> {
	return api<Permission>('/permissions', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
}

/**
 * Update a permission
 */
export async function updatePermission(
	id: string,
	data: UpdatePermissionRequest,
	options?: RequestInit
): Promise<Permission> {
	return api<Permission>(`/permissions/${id}`, {
		...options,
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

/**
 * Delete a permission
 */
export async function deletePermission(id: string, options?: RequestInit): Promise<void> {
	return api<void>(`/permissions/${id}`, {
		...options,
		method: 'DELETE'
	});
}

/**
 * Update role permissions
 */
export async function updateRolePermissions(
	roleName: UserRole,
	data: UpdateRolePermissionsRequest,
	options?: RequestInit
): Promise<Role> {
	return api<Role>(`/roles/${roleName}/permissions`, {
		...options,
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

/**
 * Assign a permission to a role
 */
export async function assignPermissionToRole(
	roleName: UserRole,
	permissionName: string,
	options?: RequestInit
): Promise<RolePermission> {
	return api<RolePermission>(`/roles/${roleName}/permissions`, {
		...options,
		method: 'POST',
		body: JSON.stringify({ permission: permissionName })
	});
}

/**
 * Remove a permission from a role
 */
export async function removePermissionFromRole(
	roleName: UserRole,
	permissionName: string,
	options?: RequestInit
): Promise<void> {
	return api<void>(`/roles/${roleName}/permissions/${encodeURIComponent(permissionName)}`, {
		...options,
		method: 'DELETE'
	});
}

// Re-export types
export type { Role, Permission, CreatePermissionRequest, UpdatePermissionRequest };
