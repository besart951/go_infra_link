import type { User } from '$lib/domain/user/index.js';

export interface UserCacheOptions {
  ttlMs?: number;
  now?: () => number;
}

export const USER_CACHE_TTL_MS = 60 * 1000;

/**
 * In-memory, session-scoped cache for the current user (identity + permissions).
 *
 * `/auth/me` is currently fetched on every SPA navigation (root layout load)
 * and once more on mount via `loadAuth()`. That is wasteful for data that
 * rarely changes within a session. This cache:
 *
 * - serves a short TTL (default 60s) so permission/account changes propagate
 *   quickly while skipping needless round trips on rapid navigation;
 * - deduplicates concurrent fetches (layout load + `loadAuth()` on mount);
 * - is explicitly cleared on login, logout, and profile updates.
 *
 * The cache is intentionally NOT persisted; a full page reload always
 * revalidates from the server.
 */
export class UserCache {
  private readonly ttlMs: number;
  private readonly now: () => number;
  private cached: User | null = null;
  private cachedAt = 0;
  private request: Promise<User> | null = null;

  constructor(options: UserCacheOptions = {}) {
    this.ttlMs = options.ttlMs ?? USER_CACHE_TTL_MS;
    this.now = options.now ?? Date.now;
  }

  getCurrentUser<T extends User>(fetchUser: () => Promise<T>): Promise<T> {
    if (this.cached && this.now() - this.cachedAt < this.ttlMs) {
      return Promise.resolve(cloneUser(this.cached as T));
    }

    if (this.request) {
      return this.request.then(cloneUser) as Promise<T>;
    }

    const request = fetchUser()
      .then((user) => {
        this.cached = cloneUser(user);
        this.cachedAt = this.now();
        return user;
      })
      .finally(() => {
        if (this.request === request) {
          this.request = null;
        }
      });

    this.request = request;
    return request.then(cloneUser) as Promise<T>;
  }

  clear(): void {
    this.cached = null;
    this.cachedAt = 0;
  }
}

function cloneUser<T extends User>(user: T): T {
  return {
    ...user,
    permissions: user.permissions ? [...user.permissions] : undefined
  };
}

export const userCache = new UserCache();
