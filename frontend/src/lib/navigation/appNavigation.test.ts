import { describe, expect, it } from 'vitest';
import { buildAppNavItems, getBreadcrumbForPath } from './appNavigation.js';
import type { User } from '$lib/domain/user/index.js';

const translate = (key: string) => key;

const baseUser = {
  id: 'user-1',
  email: 'user@example.test',
  name: 'User',
  role: 'user',
  permissions: [],
  can_access_user_directory: false
} as unknown as User;

describe('app navigation', () => {
  it('resolves facility breadcrumbs from the shared route registry', () => {
    expect(getBreadcrumbForPath('/facility/buildings/123', translate)).toEqual({
      parent: {
        title: 'navigation.facility',
        href: '/facility'
      },
      current: 'navigation.buildings'
    });
  });

  it('keeps project detail routes active under the project list entry', () => {
    const items = buildAppNavItems({
      pathname: '/projects/123',
      user: baseUser,
      translate,
      canPerform: () => false
    });

    const projects = items.find((item) => item.url === '/projects');
    expect(projects?.isActive).toBe(true);
    expect(projects?.items?.find((item) => item.url === '/projects/list')?.isActive).toBe(true);
  });

  it('omits specifications from facility navigation even with specification.read', () => {
    const items = buildAppNavItems({
      pathname: '/facility/buildings',
      user: baseUser,
      translate,
      canPerform: (action, resource) =>
        action === 'read' && (resource === 'building' || resource === 'specification')
    });

    const facility = items.find((item) => item.url === '/facility');
    expect(facility?.items?.map((item) => item.url)).toEqual(['/facility', '/facility/buildings']);
  });

  it('keeps the roles child hidden when the user hub is visible through directory access only', () => {
    const items = buildAppNavItems({
      pathname: '/users',
      user: {
        ...baseUser,
        can_access_user_directory: true
      },
      translate,
      canPerform: () => false
    });

    const users = items.find((item) => item.url === '/users');
    expect(users?.items?.map((item) => item.url)).toEqual(['/users', '/users/directory']);
  });

  it('shows role management to FZAG admins without relying on role.read', () => {
    const items = buildAppNavItems({
      pathname: '/users/roles',
      user: {
        ...baseUser,
        role: 'admin_fzag'
      } as User,
      translate,
      canPerform: () => false
    });

    const users = items.find((item) => item.url === '/users');
    expect(users?.items?.map((item) => item.url)).toContain('/users/roles');
  });

  it('hides role management from non-admin FZAG users even with role.read', () => {
    const items = buildAppNavItems({
      pathname: '/users',
      user: {
        ...baseUser,
        role: 'fzag'
      } as User,
      translate,
      canPerform: (action, resource) => action === 'read' && resource === 'role'
    });

    const users = items.find((item) => item.url === '/users');
    expect(users?.items?.map((item) => item.url) ?? []).not.toContain('/users/roles');
  });

  it('places the global timeline directly below the Excel importer', () => {
    const items = buildAppNavItems({
      pathname: '/timeline',
      user: baseUser,
      translate,
      canPerform: (action, resource) =>
        (action === 'create' && resource === 'objectdata') ||
        (action === 'read' && resource === 'timeline')
    });

    const urls = items.map((item) => item.url);
    expect(urls[urls.indexOf('/excel') + 1]).toBe('/timeline');
    expect(items.find((item) => item.url === '/timeline')?.isActive).toBe(true);
  });

  it('hides the global timeline without timeline.read', () => {
    const items = buildAppNavItems({
      pathname: '/timeline',
      user: baseUser,
      translate,
      canPerform: () => false
    });

    expect(items.find((item) => item.url === '/timeline')).toBeUndefined();
  });
});
