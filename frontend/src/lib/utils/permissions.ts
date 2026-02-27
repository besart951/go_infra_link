/**
 * Permission Guard Utilities
 *
 * Helper functions and components for controlling UI visibility
 * based on user permissions and role hierarchy
 */

import type { UserRole } from '$lib/api/users.js';
import { canManageRole, auth } from '$lib/stores/auth.svelte';

export function canPerform(action: string, resource: string): boolean {
	if (!auth.user) return false;
	const rolePerms = auth.user.permissions || [];
	const permission = `${resource}.${action}`;
	return rolePerms.includes(permission);
}

/**
 * Check if user can manage users (create/edit/delete)
 */
export function canManageUsers(): boolean {
	return canPerform('create', 'user') || canPerform('update', 'user');
}

/**
 * Get filtered roles based on what the current user can assign
 */
export function getFilteredRoles(roles: UserRole[]): UserRole[] {
	return roles.filter((role) => canManageRole(role));
}
