import { describe, expect, it, vi } from 'vitest';
import { UserCache, USER_CACHE_TTL_MS } from './userCache.js';
import type { User } from '$lib/domain/user/index.js';

function makeUser(overrides: Partial<User> = {}): User {
  return {
    id: 'user-1',
    first_name: 'Test',
    last_name: 'User',
    email: 'test@example.com',
    is_active: true,
    role: 'planer',
    permissions: ['project.read'],
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
    failed_login_attempts: 0,
    ...overrides
  };
}

describe('UserCache', () => {
  it('fetches on first access and serves the cached user within the TTL', async () => {
    const cache = new UserCache({ now: () => 1000 });
    const fetchUser = vi.fn().mockResolvedValue(makeUser());

    const first = await cache.getCurrentUser(fetchUser);
    const second = await cache.getCurrentUser(fetchUser);

    expect(fetchUser).toHaveBeenCalledTimes(1);
    expect(first).toEqual(makeUser());
    expect(second).toEqual(makeUser());
  });

  it('refetches after the TTL expires', async () => {
    let now = 1000;
    const cache = new UserCache({ now: () => now });
    const fetchUser = vi
      .fn()
      .mockResolvedValueOnce(makeUser())
      .mockResolvedValueOnce(makeUser({ email: 'updated@example.com' }));

    await cache.getCurrentUser(fetchUser);
    now += USER_CACHE_TTL_MS + 1;
    const refreshed = await cache.getCurrentUser(fetchUser);

    expect(fetchUser).toHaveBeenCalledTimes(2);
    expect(refreshed.email).toBe('updated@example.com');
  });

  it('deduplicates concurrent fetches', async () => {
    const cache = new UserCache({ now: () => 1000 });
    let resolveFetch: (user: User) => void = () => undefined;
    const fetchUser = vi.fn(
      () =>
        new Promise<User>((resolve) => {
          resolveFetch = resolve;
        })
    );

    const first = cache.getCurrentUser(fetchUser);
    const second = cache.getCurrentUser(fetchUser);
    resolveFetch(makeUser());

    const [a, b] = await Promise.all([first, second]);
    expect(fetchUser).toHaveBeenCalledTimes(1);
    expect(a).toEqual(b);
  });

  it('clears the cached user on clear()', async () => {
    const cache = new UserCache({ now: () => 1000 });
    const fetchUser = vi.fn().mockResolvedValue(makeUser());

    await cache.getCurrentUser(fetchUser);
    cache.clear();
    await cache.getCurrentUser(fetchUser);

    expect(fetchUser).toHaveBeenCalledTimes(2);
  });

  it('returns a clone so callers cannot mutate the cached identity', async () => {
    const cache = new UserCache({ now: () => 1000 });
    const fetchUser = vi.fn().mockResolvedValue(makeUser());

    const first = await cache.getCurrentUser(fetchUser);
    first.permissions = [];
    const second = await cache.getCurrentUser(fetchUser);

    expect(second.permissions).toEqual(['project.read']);
  });
});
