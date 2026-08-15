import { PERMISSIONS, type ProjectPermissionName } from '$lib/api/generated/permissions.js';

export interface ProjectCapabilities {
  permissions: readonly ProjectPermissionName[];
}

export function toProjectCapabilities(permissions: readonly string[] | undefined): ProjectCapabilities {
  return {
    permissions: (permissions ?? []).filter(isProjectPermissionName)
  };
}

function isProjectPermissionName(permission: string): permission is ProjectPermissionName {
  return (
    permission !== 'project.create' &&
    permission !== 'project.listAll' &&
    permission.startsWith('project.') &&
    (PERMISSIONS as readonly string[]).includes(permission)
  );
}
