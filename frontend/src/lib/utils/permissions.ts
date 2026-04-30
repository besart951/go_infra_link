/**
 * Permission Guard Utilities
 *
 * Helper functions and components for controlling UI visibility
 * based on user permissions and backend-provided capabilities
 */

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
