import type { Project } from '$lib/domain/project/index.js';
import type { Team } from '$lib/domain/team/index.js';
import type { User } from '$lib/domain/user/index.js';

export interface AppNavigationData {
  teams: Team[];
  projects: Project[];
}

interface AppNavigationDataCacheEntry {
  key: string;
  expiresAt: number;
  data: AppNavigationData;
}

export interface AppNavigationDataCacheOptions {
  ttlMs?: number;
  now?: () => number;
}

export const APP_NAVIGATION_DATA_CACHE_TTL_MS = 30 * 60 * 1000;

/**
 * Caches user-scoped navigation data while authentication itself remains
 * freshly checked on every navigation. A changed permission signature creates
 * a new cache key so no prior user's navigation data can be reused.
 */
export class AppNavigationDataCache {
  private readonly ttlMs: number;
  private readonly now: () => number;
  private entry: AppNavigationDataCacheEntry | null = null;
  private request: Promise<AppNavigationData> | null = null;
  private requestKey: string | null = null;
  private generation = 0;

  constructor(options: AppNavigationDataCacheOptions = {}) {
    this.ttlMs = options.ttlMs ?? APP_NAVIGATION_DATA_CACHE_TTL_MS;
    this.now = options.now ?? Date.now;
  }

  load(user: User, fetchData: () => Promise<AppNavigationData>): Promise<AppNavigationData> {
    const key = navigationDataKey(user);
    if (this.entry?.key === key && this.entry.expiresAt > this.now()) {
      return Promise.resolve(cloneNavigationData(this.entry.data));
    }

    if (this.request && this.requestKey === key) {
      return this.request.then(cloneNavigationData);
    }

    const generation = this.generation;
    const request = fetchData()
      .then((data) => {
        if (generation === this.generation && this.request === request) {
          this.entry = {
            key,
            data: cloneNavigationData(data),
            expiresAt: this.now() + this.ttlMs
          };
        }
        return data;
      })
      .finally(() => {
        if (this.request === request) {
          this.request = null;
          this.requestKey = null;
        }
      });

    this.request = request;
    this.requestKey = key;
    return request.then(cloneNavigationData);
  }

  clear(): void {
    this.generation += 1;
    this.entry = null;
    this.request = null;
    this.requestKey = null;
  }
}

function navigationDataKey(user: User): string {
  const permissions = [...(user.permissions ?? [])].sort().join(',');
  return [user.id, user.role, user.can_access_user_directory === true, permissions].join('|');
}

function cloneNavigationData(data: AppNavigationData): AppNavigationData {
  return {
    teams: data.teams.map((team) => ({ ...team })),
    projects: data.projects.map((project) => ({
      ...project,
      phase: project.phase ? { ...project.phase } : project.phase
    }))
  };
}

export const appNavigationDataCache = new AppNavigationDataCache();
