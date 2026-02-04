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
import { ROLE_PERMISSIONS, ROLE_LABELS } from '$lib/utils/permissions.js';
import { ROLE_LEVELS } from '$lib/stores/auth.svelte.js';

/**
 * All available roles in the system
 */
const ALL_ROLES: UserRole[] = [
	'superadmin',
	'admin_fzag',
	'fzag',
	'admin_planer',
	'planer',
	'admin_entrepreneur',
	'entrepreneur'
];

/**
 * Role descriptions
 */
const ROLE_DESCRIPTIONS: Record<UserRole, string> = {
	superadmin: 'Full system access with all administrative capabilities',
	admin_fzag: 'FZAG administrator with user and team management capabilities',
	fzag: 'FZAG user with project and limited user management',
	admin_planer: 'Planner administrator with user and project management',
	planer: 'Planner with project access and limited management',
	admin_entrepreneur: 'Entrepreneur administrator with limited user creation',
	entrepreneur: 'Basic read-only access to teams and projects'
};

/**
 * Get roles that can be managed by a given role
 */
function getCanManageRoles(role: UserRole): UserRole[] {
	const level = ROLE_LEVELS[role];
	return ALL_ROLES.filter((r) => ROLE_LEVELS[r] < level);
}

/**
 * List all roles with their permissions
 * Note: This uses client-side data as the backend doesn't have a roles endpoint yet
 */
export async function listRoles(): Promise<Role[]> {
	// For now, we generate roles from the client-side definitions
	// In a real implementation, this would call the backend API
	return ALL_ROLES.map((roleName, index) => ({
		id: `role-${index + 1}`,
		name: roleName,
		display_name: ROLE_LABELS[roleName],
		description: ROLE_DESCRIPTIONS[roleName],
		level: ROLE_LEVELS[roleName],
		permissions: ROLE_PERMISSIONS[roleName],
		can_manage: getCanManageRoles(roleName),
		created_at: new Date().toISOString(),
		updated_at: new Date().toISOString()
	}));
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
	// Collect all unique permissions from roles
	const permissionSet = new Set<string>();
	for (const role of ALL_ROLES) {
		const perms = ROLE_PERMISSIONS[role];
		if (!perms.includes('*')) {
			perms.forEach((p) => permissionSet.add(p));
		}
	}

	// Add facility permissions
	const facilityResources = ['building', 'systempart', 'systemtype', 'specification', 'apparat'];
	const actions = ['create', 'read', 'update', 'delete'];
	for (const resource of facilityResources) {
		for (const action of actions) {
			permissionSet.add(`${resource}.${action}`);
		}
	}

	// Add project-level resource permissions
	const projectResources = [
		'controlcabinet',
		'spscontroller',
		'spscontrollersystemtype',
		'fielddevice',
		'bacnetobject'
	];
	for (const resource of projectResources) {
		for (const action of actions) {
			permissionSet.add(`project.${resource}.${action}`);
		}
	}

	// Generate permission objects
	return Array.from(permissionSet)
		.sort()
		.map((permName, index) => {
			const parts = permName.split('.');
			let resource: string;
			let action: string;
			let description: string;

			if (parts.length === 3 && parts[0] === 'project') {
				// project.subresource.action format
				resource = `project.${parts[1]}`;
				action = parts[2];
				description = `${parts[2].charAt(0).toUpperCase() + parts[2].slice(1)} ${parts[1]}s in projects`;
			} else {
				// resource.action format
				resource = parts[0] || 'unknown';
				action = parts[1] || 'unknown';
				description = `${action.charAt(0).toUpperCase() + action.slice(1)} ${resource}s`;
			}

			return {
				id: `perm-${index + 1}`,
				name: permName,
				description,
				resource,
				action,
				created_at: new Date().toISOString(),
				updated_at: new Date().toISOString()
			};
		});
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
