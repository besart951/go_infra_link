/**
 * Role domain types
 * Defines the structure for role hierarchy management
 */

import type { UserRole } from '../user/index.js';

/**
 * Permission represents a specific action that can be performed on a resource
 */
export interface Permission {
	id: string;
	name: string;
	description: string;
	resource: string; // e.g., "user", "team", "project"
	action: string; // e.g., "create", "read", "update", "delete"
	created_at: string;
	updated_at: string;
}

/**
 * Role with its associated permissions and hierarchy info
 */
export interface Role {
	id: string;
	name: UserRole;
	display_name: string;
	description: string;
	level: number;
	permissions: string[]; // List of permission names like "user.create"
	can_manage: UserRole[]; // Roles this role can manage
	created_at: string;
	updated_at: string;
}

/**
 * Request to create a new custom permission
 */
export interface CreatePermissionRequest {
	name: string;
	description: string;
	resource: string;
	action: string;
}

/**
 * Request to update a permission
 */
export interface UpdatePermissionRequest {
	description?: string;
}

/**
 * Request to update role permissions
 */
export interface UpdateRolePermissionsRequest {
	permissions: string[];
}

/**
 * Role permission assignment
 */
export interface RolePermission {
	id: string;
	role: UserRole;
	permission: string;
	created_at: string;
}

/**
 * Predefined permission actions
 */
export const PERMISSION_ACTIONS = ['create', 'read', 'update', 'delete', 'manage'] as const;
export type PermissionAction = (typeof PERMISSION_ACTIONS)[number];

/**
 * Facility resources (direct: resource.action)
 */
export const FACILITY_RESOURCES = [
	'building',
	'systempart',
	'systemtype',
	'specification',
	'apparat'
] as const;
export type FacilityResource = (typeof FACILITY_RESOURCES)[number];

/**
 * Project sub-resources (nested: project.subresource.action)
 */
export const PROJECT_SUB_RESOURCES = [
	'controlcabinet',
	'spscontroller',
	'spscontrollersystemtype',
	'fielddevice',
	'bacnetobject'
] as const;
export type ProjectSubResource = (typeof PROJECT_SUB_RESOURCES)[number];

/**
 * General resources
 */
export const GENERAL_RESOURCES = [
	'user',
	'team',
	'project',
	'phase',
	'role',
	'permission'
] as const;
export type GeneralResource = (typeof GENERAL_RESOURCES)[number];

/**
 * All predefined resources (combined)
 */
export const PERMISSION_RESOURCES = [...GENERAL_RESOURCES, ...FACILITY_RESOURCES] as const;
export type PermissionResource = (typeof PERMISSION_RESOURCES)[number];

/**
 * Permission category for UI organization
 */
export type PermissionCategory = 'general' | 'facility' | 'project';

/**
 * Generate permission name from resource and action
 * Supports both simple (resource.action) and nested (project.subresource.action)
 */
export function createPermissionName(
	resource: string,
	action: string,
	subResource?: string
): string {
	if (subResource) {
		return `${resource}.${subResource}.${action}`;
	}
	return `${resource}.${action}`;
}

/**
 * Parse permission name into resource, action, and optional sub-resource
 * Handles both "resource.action" and "project.subresource.action"
 */
export function parsePermissionName(permissionName: string): {
	resource: string;
	action: string;
	subResource?: string;
	category: PermissionCategory;
} {
	const parts = permissionName.split('.');

	if (parts.length === 3 && parts[0] === 'project') {
		// project.subresource.action format
		return {
			resource: parts[0],
			subResource: parts[1],
			action: parts[2],
			category: 'project'
		};
	}

	if (parts.length >= 2) {
		const resource = parts[0];
		const action = parts[parts.length - 1];

		// Check if it's a facility resource
		if (FACILITY_RESOURCES.includes(resource as FacilityResource)) {
			return { resource, action, category: 'facility' };
		}

		return { resource, action, category: 'general' };
	}

	return { resource: parts[0] || '', action: '', category: 'general' };
}
