import type { User } from '$lib/domain/user/index.js';
import type { PermissionName } from '$lib/api/generated/permissions.js';
import { APP_ROUTE_ACCESS_RULES } from '$lib/navigation/routeAccess.js';
import { hasUserPermission } from '$lib/utils/permissions.js';

interface RoutePermissionRule {
  pattern: RegExp;
  permissions: readonly PermissionName[];
}

const routePermissionRules: RoutePermissionRule[] = APP_ROUTE_ACCESS_RULES.map((rule) =>
  route(rule.path, [rule.permission, ...(rule.requiredPermissions ?? [])])
);

export function getRequiredPermissionForRoute(pathname: string): PermissionName | undefined {
  return routePermissionRules.find((rule) => rule.pattern.test(pathname))?.permissions[0];
}

export function canAccessProtectedRoute(user: User | null | undefined, pathname: string): boolean {
  const permissions = routePermissionRules.find((rule) => rule.pattern.test(pathname))?.permissions;
  return !permissions || permissions.every((permission) => hasUserPermission(user, permission));
}

export function forbiddenRouteRedirect(url: URL): string {
  return `/errors/403?from=${encodeURIComponent(url.pathname + url.search)}`;
}

function route(path: string, permissions: readonly PermissionName[]): RoutePermissionRule {
  return {
    pattern: routePattern(path),
    permissions
  };
}

function routePattern(path: string): RegExp {
  const pattern = path
    .split('/')
    .map((part) => (part === '*' ? '[^/]+' : escapeRegExp(part)))
    .join('/');
  return new RegExp(`^${pattern}(?:/|$)`);
}

function escapeRegExp(value: string): string {
  return value.replace(/[.*+?^${}()|[\]\\]/g, '\\$&');
}
