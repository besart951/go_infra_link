import { browser } from '$app/environment';
import type { HistoryListResponse, HistoryTimelineParams } from '$lib/domain/history.js';

export const ACTIVITY_CACHE_TTL_MS = 30 * 60 * 1000;

interface ActivityCacheEntry {
  expiresAt: number;
  response: HistoryListResponse;
}

const entries = new Map<string, ActivityCacheEntry>();

export function activityCacheKey(namespace: string, params: HistoryTimelineParams): string {
  const query = new URLSearchParams();
  const values: Array<[string, string | undefined]> = [
    ['scope_type', params.scopeType],
    ['scope_id', params.scopeId],
    ['entity_table', params.entityTable],
    ['entity_id', params.entityId],
    ['actor_id', params.actorId],
    ['occurred_from', params.occurredFrom],
    ['occurred_to', params.occurredTo],
    ['cursor', params.cursor],
    ['limit', params.limit?.toString()]
  ];
  for (const [key, value] of values) {
    if (value) query.set(key, value);
  }
  for (const action of [...(params.actions ?? [])].sort()) query.append('action', action);
  for (const field of [...(params.fields ?? [])].sort()) query.append('field', field);
  return `${namespace}?${query.toString()}`;
}

export function readActivityCache(key: string): HistoryListResponse | undefined {
  if (!browser) return undefined;
  const entry = entries.get(key);
  if (!entry) return undefined;
  if (entry.expiresAt <= Date.now()) {
    entries.delete(key);
    return undefined;
  }
  return entry.response;
}

export function writeActivityCache(key: string, response: HistoryListResponse): void {
  if (!browser) return;
  entries.set(key, { response, expiresAt: Date.now() + ACTIVITY_CACHE_TTL_MS });
}

export function invalidateActivityCache(namespace: string): void {
  if (!browser) return;
  for (const key of entries.keys()) {
    if (key.startsWith(`${namespace}?`)) entries.delete(key);
  }
}

export function clearActivityCache(): void {
  entries.clear();
}

export const clearActivityCacheForTests = clearActivityCache;
