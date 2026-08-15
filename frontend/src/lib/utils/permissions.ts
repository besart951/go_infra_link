import {
  PERMISSIONS,
  type PermissionName,
  type ProjectPermissionName
} from '$lib/api/generated/permissions.js';
import type { ProjectCapabilities } from '$lib/domain/project/capabilities.js';
import { auth } from '$lib/stores/auth.svelte.js';

interface PermissionedUser {
  role?: string;
  permissions?: readonly string[];
}

export function isSuperAdminUser(user: PermissionedUser | null | undefined): boolean {
  return user?.role === 'superadmin';
}

export function hasUserPermission(
  user: PermissionedUser | null | undefined,
  permission: PermissionName
): boolean {
  if (!user) return false;
  if (isSuperAdminUser(user)) return true;
  return Boolean(user.permissions?.includes(permission));
}

export function can(permission: PermissionName): boolean {
  return hasUserPermission(auth.user, permission);
}

export function canProject(
  capabilities: ProjectCapabilities | null | undefined,
  permission: ProjectPermissionName
): boolean {
  return Boolean(capabilities?.permissions.includes(permission));
}

/**
 * Transitional adapter for existing components. New code should use `can` or
 * `canProject` with a generated permission name directly.
 */
export function canPerform(action: string, resource: string): boolean {
  const permission = `${resource}.${action}`;
  if (!(PERMISSIONS as readonly string[]).includes(permission)) {
    return false;
  }
  return can(permission as PermissionName);
}
