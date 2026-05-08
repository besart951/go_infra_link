import type { User } from '$lib/domain/user/index.js';
import { APP_ROUTE_ACCESS_RULES } from '$lib/navigation/routeAccess.js';
import { hasUserPermission } from '$lib/utils/permissions.js';

interface RoutePermissionRule {
  pattern: RegExp;
  permission: string;
}

const routePermissionRules: RoutePermissionRule[] = APP_ROUTE_ACCESS_RULES.map((rule) =>
  route(rule.path, rule.permission)
);

export function getRequiredPermissionForRoute(pathname: string): string | undefined {
  return routePermissionRules.find((rule) => rule.pattern.test(pathname))?.permission;
}

export function canAccessProtectedRoute(user: User | null | undefined, pathname: string): boolean {
  const permission = getRequiredPermissionForRoute(pathname);
  return !permission || hasUserPermission(user, permission);
}

export function forbiddenRouteRedirect(url: URL): string {
  return `/errors/403?from=${encodeURIComponent(url.pathname + url.search)}`;
}

function route(path: string, permission: string): RoutePermissionRule {
  return {
    pattern: routePattern(path),
    permission
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
