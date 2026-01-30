/**
 * Permission Guard Utilities
 * 
 * Helper functions and components for controlling UI visibility
 * based on user permissions and role hierarchy
 */

import type { UserRole } from '$lib/api/users';
import { canManageRole, hasRole, hasMinRole, auth } from '$lib/stores/auth.svelte';

/**
 * Check if the current user can perform an action on a resource
 */
export function canPerform(action: string, resource: string): boolean {
	if (!auth.user) return false;

	const role = auth.user.role;

	// Define permission mappings
	const permissions: Record<UserRole, string[]> = {
		superadmin: ['*'], // Can do everything
		admin_fzag: [
			'user.create',
			'user.read',
			'user.update',
			'user.delete',
			'team.create',
			'team.read',
			'team.update',
			'team.delete',
			'project.create',
			'project.read',
			'project.update',
			'project.delete'
		],
		fzag: [
			'user.create',
			'user.read',
			'user.update',
			'team.read',
			'team.update',
			'project.create',
			'project.read',
			'project.update',
			'project.delete'
		],
		admin_planer: [
			'user.create',
			'user.read',
			'user.update',
			'team.read',
			'project.create',
			'project.read',
			'project.update'
		],
		planer: ['user.read', 'team.read', 'project.read', 'project.update'],
		admin_entrepreneur: ['user.create', 'user.read', 'team.read', 'project.read'],
		entrepreneur: ['team.read', 'project.read'],
		admin: [
			'user.create',
			'user.read',
			'user.update',
			'team.create',
			'team.read',
			'team.update',
			'project.create',
			'project.read',
			'project.update'
		],
		user: ['team.read', 'project.read']
	};

	const rolePerms = permissions[role] || [];
	const permission = `${resource}.${action}`;

	// Superadmin has all permissions
	if (rolePerms.includes('*')) return true;

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

/**
 * Role display names for UI
 */
export const ROLE_LABELS: Record<UserRole, string> = {
	superadmin: 'Super Administrator',
	admin_fzag: 'FZAG Administrator',
	fzag: 'FZAG',
	admin_planer: 'Planner Administrator',
	planer: 'Planner',
	admin_entrepreneur: 'Entrepreneur Administrator',
	entrepreneur: 'Entrepreneur',
	admin: 'Administrator',
	user: 'User'
};

/**
 * Get display name for a role
 */
export function getRoleLabel(role: UserRole): string {
	return ROLE_LABELS[role] || role;
}
