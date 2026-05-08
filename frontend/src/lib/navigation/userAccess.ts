import type { User } from '$lib/domain/user/index.js';
import type { UserRole } from '$lib/domain/user/index.js';

type CanPerform = (action: string, resource: string) => boolean;

const ROLE_DIRECTORY_ROLES = new Set<UserRole>(['superadmin', 'admin_fzag']);

export function canAccessUserDirectory(user: User | null | undefined): boolean {
  return Boolean(user?.can_access_user_directory);
}

export function canAccessTeamDirectory(canPerform: CanPerform): boolean {
  return canPerform('read', 'team');
}

export function canAccessRoleDirectory(user: User | null | undefined): boolean {
  return Boolean(user?.role && ROLE_DIRECTORY_ROLES.has(user.role));
}

export function canEditRolePermissions(
  user: User | null | undefined,
  targetRole: UserRole
): boolean {
  if (!user) return false;
  if (user.role === 'superadmin') return true;
  return user.role === 'admin_fzag' && targetRole !== 'superadmin';
}

export function canAccessUserHub(user: User | null | undefined, canPerform: CanPerform): boolean {
  return (
    canAccessUserDirectory(user) ||
    canAccessTeamDirectory(canPerform) ||
    canAccessRoleDirectory(user)
  );
}
