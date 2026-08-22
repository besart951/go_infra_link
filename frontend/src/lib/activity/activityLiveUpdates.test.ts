import { beforeEach, describe, expect, it, vi } from 'vitest';
import {
  clearActivityLiveUpdatesForTests,
  notifyProjectActivityChanged,
  subscribeToProjectActivity
} from './activityLiveUpdates.js';
import {
  activityCacheKey,
  clearActivityCacheForTests,
  readActivityCache,
  writeActivityCache
} from './activityCache.js';

describe('project activity live updates', () => {
  beforeEach(() => {
    clearActivityCacheForTests();
    clearActivityLiveUpdatesForTests();
  });

  it('invalidates project and global activity caches and notifies only that project', () => {
    const response = { items: [], next_cursor: 'next' };
    const projectKey = activityCacheKey('history:project:project-1', { limit: 25 });
    const globalKey = activityCacheKey('history:global', { limit: 25 });
    const otherProjectKey = activityCacheKey('history:project:project-2', { limit: 25 });
    writeActivityCache(projectKey, response);
    writeActivityCache(globalKey, response);
    writeActivityCache(otherProjectKey, response);

    const listener = vi.fn();
    const otherListener = vi.fn();
    subscribeToProjectActivity('project-1', listener);
    subscribeToProjectActivity('project-2', otherListener);

    notifyProjectActivityChanged('project-1');

    expect(listener).toHaveBeenCalledOnce();
    expect(otherListener).not.toHaveBeenCalled();
    expect(readActivityCache(projectKey)).toBeUndefined();
    expect(readActivityCache(globalKey)).toBeUndefined();
    expect(readActivityCache(otherProjectKey)).toEqual(response);
  });

  it('unsubscribes cleanly when a dialog closes', () => {
    const listener = vi.fn();
    const unsubscribe = subscribeToProjectActivity('project-1', listener);

    unsubscribe();
    notifyProjectActivityChanged('project-1');

    expect(listener).not.toHaveBeenCalled();
  });
});
