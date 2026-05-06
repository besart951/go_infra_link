import type { User } from '$lib/domain/user/index.js';

type CanPerform = (action: string, resource: string) => boolean;

export function canAccessUserDirectory(user: User | null | undefined): boolean {
  return Boolean(user?.can_access_user_directory);
}

export function canAccessTeamDirectory(canPerform: CanPerform): boolean {
  return canPerform('read', 'team');
}

export function canAccessRoleDirectory(canPerform: CanPerform): boolean {
  return canPerform('read', 'role');
}

export function canAccessUserHub(
  user: User | null | undefined,
  canPerform: CanPerform
): boolean {
  return (
    canAccessUserDirectory(user) ||
    canAccessTeamDirectory(canPerform) ||
    canAccessRoleDirectory(canPerform)
  );
}
