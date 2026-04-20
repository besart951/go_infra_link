import { describe, expect, it } from 'vitest';

import {
  configuredRoutePaths,
  getRouteAuditSummary,
  routeAudits
} from '../../fixtures/route-matrix.js';
import { discoverRoutePages } from '../../helpers/route-discovery.js';

describe('route permission audit inventory', () => {
  it('covers every current frontend page route', () => {
    const discoveredPaths = discoverRoutePages().map((route) => route.path);
    const auditedPaths = routeAudits.map((route) => route.path);

    expect(auditedPaths).toEqual(discoveredPaths);
  });

  it('keeps the audited file mapping aligned with the route tree', () => {
    const discoveredByPath = new Map(
      discoverRoutePages().map((route) => [route.path, route.files] as const)
    );

    for (const route of routeAudits) {
      expect(discoveredByPath.get(route.path)).toEqual(route.files);
    }
  });

  it('summarizes which routes are currently configured correctly', () => {
    const summary = getRouteAuditSummary();

    expect(summary.totalRoutes).toBe(34);
    expect(summary.configured).toEqual([
      '/',
      '/account',
      '/admin/notifications/smtp',
      '/login',
      '/logout',
      '/projects/new'
    ]);
    expect(summary.misconfigured).toHaveLength(28);
    expect(configuredRoutePaths).toHaveLength(6);
  });
});
