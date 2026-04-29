/**
 * Permission Guard Utilities
 *
 * Helper functions and components for controlling UI visibility
 * based on user permissions and backend-provided capabilities
 */

import type { UserRole } from '$lib/api/users.js';
import { auth } from '$lib/stores/auth.svelte';

function hasPermission(permission: string): boolean {
  if (!auth.user) return false;
  const rolePerms = auth.user.permissions || [];
  return rolePerms.includes(permission);
}

export function canPerform(action: string, resource: string): boolean {
  return hasPermission(`${resource}.${action}`);
}

export function canPerformAny(actions: string[], resource: string): boolean {
  return actions.some((action) => canPerform(action, resource));
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
  const allowedRoles = new Set(auth.allowedRoles.map((roleInfo) => roleInfo.role));
  return roles.filter((role) => allowedRoles.has(role));
}
