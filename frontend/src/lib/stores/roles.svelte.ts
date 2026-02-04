/**
 * Role store - Svelte 5 runes-based store for role management
 * Manages roles, permissions, and their relationships
 */

import type { Role, Permission } from '$lib/domain/role/index.js';
import type { UserRole } from '$lib/domain/user/index.js';
import { listRoles, listPermissions } from '$lib/infrastructure/api/role.adapter.js';

interface RoleStoreState {
	roles: Role[];
	permissions: Permission[];
	isLoading: boolean;
	error: string | null;
}

const state = $state<RoleStoreState>({
	roles: [],
	permissions: [],
	isLoading: false,
	error: null
});

/**
 * Load all roles and permissions
 */
export async function loadRolesAndPermissions(): Promise<void> {
	state.isLoading = true;
	state.error = null;

	try {
		const [roles, permissions] = await Promise.all([listRoles(), listPermissions()]);

		state.roles = roles;
		state.permissions = permissions;
	} catch (error) {
		state.error = error instanceof Error ? error.message : 'Failed to load roles';
	} finally {
		state.isLoading = false;
	}
}

/**
 * Get a role by name
 */
export function getRoleByName(name: UserRole): Role | undefined {
	return state.roles.find((r) => r.name === name);
}

/**
 * Get all permissions for a role
 */
export function getPermissionsForRole(roleName: UserRole): string[] {
	const role = getRoleByName(roleName);
	return role?.permissions || [];
}

/**
 * Check if a role has a specific permission
 */
export function roleHasPermission(roleName: UserRole, permission: string): boolean {
	const role = getRoleByName(roleName);
	if (!role) return false;
	return role.permissions.includes('*') || role.permissions.includes(permission);
}

/**
 * Get permissions grouped by resource
 */
export function getPermissionsByResource(): Record<string, Permission[]> {
	const grouped: Record<string, Permission[]> = {};

	for (const perm of state.permissions) {
		if (!grouped[perm.resource]) {
			grouped[perm.resource] = [];
		}
		grouped[perm.resource].push(perm);
	}

	return grouped;
}

/**
 * Get unique resources from permissions
 */
export function getUniqueResources(): string[] {
	const resources = new Set(state.permissions.map((p) => p.resource));
	return Array.from(resources).sort();
}

/**
 * Get unique actions from permissions
 */
export function getUniqueActions(): string[] {
	const actions = new Set(state.permissions.map((p) => p.action));
	return Array.from(actions).sort();
}

/**
 * Reload store data
 */
export function reloadRoles(): void {
	loadRolesAndPermissions();
}

// Export reactive getters using Svelte 5 runes
export const rolesStore = {
	get roles() {
		return state.roles;
	},
	get permissions() {
		return state.permissions;
	},
	get isLoading() {
		return state.isLoading;
	},
	get error() {
		return state.error;
	}
};
