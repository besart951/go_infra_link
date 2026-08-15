import {
  AppNavigationDataCache,
  APP_NAVIGATION_DATA_CACHE_TTL_MS,
  type AppNavigationData
} from './appNavigationDataCache.js';
import type { User } from '$lib/domain/user/index.js';

const user: User = {
  id: 'user-1',
  first_name: 'Ada',
  last_name: 'Lovelace',
  email: 'ada@example.test',
  role: 'planer',
  permissions: ['project.listAll'],
  is_active: true,
  failed_login_attempts: 0,
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z'
};

const navigationData: AppNavigationData = {
  teams: [],
  projects: []
};

describe('AppNavigationDataCache', () => {
  it('caches teams and projects for 30 minutes while callers may still refresh authentication', async () => {
    let now = 0;
    const cache = new AppNavigationDataCache({ now: () => now });
    const fetchData = vi.fn().mockResolvedValue(navigationData);

    await cache.load(user, fetchData);
    now = APP_NAVIGATION_DATA_CACHE_TTL_MS - 1;
    await cache.load(user, fetchData);

    expect(fetchData).toHaveBeenCalledOnce();

    now += 1;
    await cache.load(user, fetchData);
    expect(fetchData).toHaveBeenCalledTimes(2);
  });

  it('does not reuse navigation data after a permission change', async () => {
    const cache = new AppNavigationDataCache();
    const fetchData = vi.fn().mockResolvedValue(navigationData);

    await cache.load(user, fetchData);
    await cache.load({ ...user, permissions: ['project.listAll', 'team.read'] }, fetchData);

    expect(fetchData).toHaveBeenCalledTimes(2);
  });

  it('clears the session data before a subsequent sign-in', async () => {
    const cache = new AppNavigationDataCache();
    const fetchData = vi.fn().mockResolvedValue(navigationData);

    await cache.load(user, fetchData);
    cache.clear();
    await cache.load(user, fetchData);

    expect(fetchData).toHaveBeenCalledTimes(2);
  });
});
