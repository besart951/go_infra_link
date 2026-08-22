import { afterEach, describe, expect, it, vi } from 'vitest';
import type { HistoryListResponse, HistoryTimelineParams } from '$lib/domain/history.js';
import {
  ACTIVITY_CACHE_TTL_MS,
  activityCacheKey,
  clearActivityCacheForTests,
  invalidateActivityCache,
  readActivityCache,
  writeActivityCache
} from './activityCache.js';

const response: HistoryListResponse = {
  items: []
};

afterEach(() => {
  clearActivityCacheForTests();
  vi.useRealTimers();
});

describe('activityCache', () => {
  it('uses a stable key for equivalent unordered action and field filters', () => {
    const first: HistoryTimelineParams = {
      fields: ['apparat_id', 'system_part_id'],
      actions: ['update', 'create'],
      cursor: 'cursor-1'
    };
    const second: HistoryTimelineParams = {
      fields: ['system_part_id', 'apparat_id'],
      actions: ['create', 'update'],
      cursor: 'cursor-1'
    };

    expect(activityCacheKey('history:global', first)).toBe(
      activityCacheKey('history:global', second)
    );
  });

  it('keeps a scope page for thirty minutes and expires it afterwards', () => {
    vi.useFakeTimers();
    vi.setSystemTime(new Date('2026-08-15T10:00:00Z'));
    const key = activityCacheKey('history:project:project-1', { cursor: 'cursor-1' });

    writeActivityCache(key, response);
    expect(readActivityCache(key)).toBe(response);

    vi.advanceTimersByTime(ACTIVITY_CACHE_TTL_MS - 1);
    expect(readActivityCache(key)).toBe(response);

    vi.advanceTimersByTime(1);
    expect(readActivityCache(key)).toBeUndefined();
  });

  it('invalidates only the live project scope that changed', () => {
    const changed = activityCacheKey('history:project:project-1', { cursor: 'cursor-1' });
    const untouched = activityCacheKey('history:project:project-2', { cursor: 'cursor-1' });
    writeActivityCache(changed, response);
    writeActivityCache(untouched, response);

    invalidateActivityCache('history:project:project-1');

    expect(readActivityCache(changed)).toBeUndefined();
    expect(readActivityCache(untouched)).toBe(response);
  });
});
