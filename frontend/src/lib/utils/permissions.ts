/**
 * Permission Guard Utilities
 *
 * Helper functions and components for controlling UI visibility
 * based on user permissions and backend-provided capabilities
 */

import { auth } from '$lib/stores/auth.svelte';

interface PermissionedUser {
  role?: string;
  permissions?: string[];
}

export function isSuperAdminUser(user: PermissionedUser | null | undefined): boolean {
  return user?.role === 'superadmin';
}

export function hasUserPermission(
  user: PermissionedUser | null | undefined,
  permission: string
): boolean {
  if (!user) return false;
  if (isSuperAdminUser(user)) return true;
  return Boolean(user.permissions?.includes(permission));
}

function hasPermission(permission: string): boolean {
  return hasUserPermission(auth.user, permission);
}

export function canPerform(action: string, resource: string): boolean {
  return hasPermission(`${resource}.${action}`);
}
