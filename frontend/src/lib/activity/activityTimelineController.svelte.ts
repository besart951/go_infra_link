import type {
  ChangeEvent,
  HistoryListResponse,
  HistoryTimelineParams
} from '$lib/domain/history.js';
import { getErrorMessage } from '$lib/api/client.js';
import {
  activityCacheKey,
  invalidateActivityCache,
  readActivityCache,
  writeActivityCache
} from './activityCache.js';
import type { ActivityDataSource } from './contract.js';
import { SvelteSet } from 'svelte/reactivity';

interface ActivityLoadOptions {
  append?: boolean;
  force?: boolean;
  cursor?: string;
}

export class ActivityTimelineController {
  events = $state.raw<ChangeEvent[]>([]);
  loading = $state(false);
  loadingMore = $state(false);
  error = $state<string | null>(null);
  nextCursor = $state<string | undefined>();
  previousCursor = $state<string | undefined>();
  pendingLiveChanges = $state(0);

  readonly hasMore = $derived(Boolean(this.nextCursor));
  readonly hasPrevious = $derived(Boolean(this.previousCursor));

  private requestId = 0;
  private controller: AbortController | undefined;
  private activeQueryKey = '';

  constructor(private readonly source: ActivityDataSource) {}

  async load(params: HistoryTimelineParams, options: ActivityLoadOptions = {}): Promise<void> {
    const append = options.append ?? false;
    if (append && (this.loading || this.loadingMore || !this.hasMore)) return;
    const cursor = options.cursor ?? (append ? this.nextCursor : undefined);
    const requestParams = { ...params, ...(cursor ? { cursor } : {}) };
    const queryKey = activityQueryKey(this.source.cacheNamespace, params);
    if (!append && queryKey !== this.activeQueryKey) this.resetQuery(queryKey);

    const cacheKey = activityCacheKey(this.source.cacheNamespace, requestParams);
    const cached = options.force ? undefined : readActivityCache(cacheKey);
    if (cached) {
      this.applyResponse(cached, append);
      return;
    }
    await this.fetch(requestParams, append);
  }

  loadNext(params: HistoryTimelineParams): Promise<void> {
    if (!this.nextCursor) return Promise.resolve();
    return this.load(params, { cursor: this.nextCursor, force: true });
  }

  loadPrevious(params: HistoryTimelineParams): Promise<void> {
    if (!this.previousCursor) return Promise.resolve();
    return this.load(params, { cursor: this.previousCursor, force: true });
  }

  markLiveChange(): void {
    invalidateActivityCache(this.source.cacheNamespace);
    this.pendingLiveChanges += 1;
  }

  dispose(): void {
    this.controller?.abort();
    this.controller = undefined;
  }

  private async fetch(params: HistoryTimelineParams, append: boolean): Promise<void> {
    const request = this.startRequest(append);
    try {
      const response = await this.source.list(params, request.controller.signal);
      if (request.id !== this.requestId) return;
      writeActivityCache(activityCacheKey(this.source.cacheNamespace, params), response);
      this.applyResponse(response, append);
      if (!append) this.pendingLiveChanges = 0;
    } catch (error) {
      if (request.id !== this.requestId || isAbortError(error)) return;
      this.error = getErrorMessage(error);
    } finally {
      if (request.id === this.requestId) this.finishRequest();
    }
  }

  private startRequest(append: boolean): { id: number; controller: AbortController } {
    this.controller?.abort();
    const controller = new AbortController();
    this.controller = controller;
    this.error = null;
    this.loading = !append;
    this.loadingMore = append;
    return { id: ++this.requestId, controller };
  }

  private finishRequest(): void {
    this.loading = false;
    this.loadingMore = false;
  }

  private applyResponse(response: HistoryListResponse, append: boolean): void {
    this.events = append ? mergeEvents(this.events, response.items) : response.items;
    this.nextCursor = response.next_cursor;
    this.previousCursor = response.previous_cursor;
  }

  private resetQuery(queryKey: string): void {
    this.activeQueryKey = queryKey;
    this.events = [];
    this.nextCursor = undefined;
    this.previousCursor = undefined;
  }
}

function activityQueryKey(namespace: string, params: HistoryTimelineParams): string {
  const { cursor: _cursor, ...query } = params;
  return activityCacheKey(namespace, query);
}

function mergeEvents(existing: ChangeEvent[], incoming: ChangeEvent[]): ChangeEvent[] {
  const seen = new SvelteSet(existing.map((event) => event.id));
  return [...existing, ...incoming.filter((event) => !seen.has(event.id))];
}

function isAbortError(error: unknown): boolean {
  return error instanceof DOMException && error.name === 'AbortError';
}
