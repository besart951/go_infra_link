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

  it('filters facility entries by the same permission metadata used for navigation', () => {
    const items = buildAppNavItems({
      pathname: '/facility/buildings',
      user: baseUser,
      translate,
      canPerform: (action, resource) => action === 'read' && resource === 'building'
    });

    const facility = items.find((item) => item.url === '/facility');
    expect(facility?.items?.map((item) => item.url)).toEqual(['/facility', '/facility/buildings']);
  });
});
