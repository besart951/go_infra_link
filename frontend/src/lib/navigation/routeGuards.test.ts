import { describe, expect, it } from 'vitest';
import {
  canAccessProtectedRoute,
  forbiddenRouteRedirect,
  getRequiredPermissionForRoute
} from './routeGuards.js';
import type { User } from '$lib/domain/user/index.js';

const user = (permissions: string[] = []): User =>
  ({
    id: 'user-1',
    role: 'planer',
    permissions
  }) as User;

describe('route guards', () => {
  it('requires objectdata.create for the Excel importer', () => {
    expect(getRequiredPermissionForRoute('/excel')).toBe('objectdata.create');
    expect(canAccessProtectedRoute(user(['objectdata.read']), '/excel')).toBe(false);
    expect(canAccessProtectedRoute(user(['objectdata.create']), '/excel')).toBe(true);
  });

  it('requires team.create for the dedicated team creation route', () => {
    expect(getRequiredPermissionForRoute('/teams/new')).toBe('team.create');
    expect(canAccessProtectedRoute(user(['team.read']), '/teams/new')).toBe(false);
    expect(canAccessProtectedRoute(user(['team.create']), '/teams/new')).toBe(true);
  });

  it('leaves the role directory to its role-based page guard', () => {
    expect(getRequiredPermissionForRoute('/users/roles')).toBeUndefined();
  });

  it('maps facility routes to canonical read permissions', () => {
    expect(getRequiredPermissionForRoute('/facility/alarm-catalog')).toBe('alarmtype.read');
    expect(getRequiredPermissionForRoute('/facility/sps-controller-system-type/123')).toBe(
      'spscontrollersystemtype.read'
    );
  });

  it('treats superadmin as allowed without stored route permissions', () => {
    expect(canAccessProtectedRoute({ ...user(), role: 'superadmin' }, '/facility/buildings')).toBe(
      true
    );
  });

  it('builds a forbidden redirect with the original route', () => {
    expect(forbiddenRouteRedirect(new URL('https://app.test/excel?tab=preview'))).toBe(
      '/errors/403?from=%2Fexcel%3Ftab%3Dpreview'
    );
  });
});
