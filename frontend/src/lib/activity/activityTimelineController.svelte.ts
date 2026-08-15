import type { ChangeEvent, HistoryTimelineParams } from '$lib/domain/history.js';
import { getErrorMessage } from '$lib/api/client.js';
import {
  activityCacheKey,
  invalidateActivityCache,
  readActivityCache,
  writeActivityCache
} from './activityCache.js';
import type { ActivityDataSource } from './contract.js';

export class ActivityTimelineController {
  events = $state.raw<ChangeEvent[]>([]);
  loading = $state(false);
  loadingMore = $state(false);
  error = $state<string | null>(null);
  page = $state(0);
  total = $state(0);
  totalPages = $state(1);
  pendingLiveChanges = $state(0);
  hasMore = $derived(this.page < this.totalPages);

  private requestId = 0;
  private controller: AbortController | undefined;
  private activeQueryKey = '';

  constructor(private readonly source: ActivityDataSource) {}

  async load(
    params: HistoryTimelineParams,
    options: { append?: boolean; force?: boolean } = {}
  ): Promise<void> {
    const append = options.append ?? false;
    if (append && (this.loading || this.loadingMore || !this.hasMore)) return;

    const targetPage = append ? this.page + 1 : (params.page ?? 1);
    const requestParams = { ...params, page: targetPage };
    const cacheKey = activityCacheKey(this.source.cacheNamespace, requestParams);
    const queryKey = activityCacheKey(this.source.cacheNamespace, { ...params, page: 1 });
    if (!append && this.activeQueryKey && this.activeQueryKey !== queryKey) {
      this.events = [];
      this.page = 0;
      this.total = 0;
      this.totalPages = 1;
    }
    if (!append) this.activeQueryKey = queryKey;
    if (!append && !options.force) {
      const cached = readActivityCache(cacheKey);
      if (cached) {
        this.applyResponse(
          cached.items,
          cached.total,
          cached.page || targetPage,
          cached.total_pages || 1,
          false
        );
        return;
      }
    }

    this.controller?.abort();
    const controller = new AbortController();
    this.controller = controller;
    const requestId = ++this.requestId;
    this.error = null;
    if (append) {
      this.loadingMore = true;
    } else {
      this.loading = true;
      this.loadingMore = false;
    }

    try {
      const response = await this.source.list(requestParams, controller.signal);
      if (requestId !== this.requestId) return;
      writeActivityCache(cacheKey, response);
      this.applyResponse(
        response.items,
        response.total,
        response.page || targetPage,
        response.total_pages || 1,
        append
      );
      if (!append) this.pendingLiveChanges = 0;
    } catch (error) {
      if (requestId !== this.requestId || isAbortError(error)) return;
      this.error = getErrorMessage(error);
    } finally {
      if (requestId === this.requestId) {
        this.loading = false;
        this.loadingMore = false;
      }
    }
  }

  markLiveChange(): void {
    invalidateActivityCache(this.source.cacheNamespace);
    this.pendingLiveChanges += 1;
  }

  dispose(): void {
    this.controller?.abort();
    this.controller = undefined;
  }

  private applyResponse(
    items: ChangeEvent[],
    total: number,
    page: number,
    totalPages: number,
    append: boolean
  ): void {
    this.events = append ? mergeEvents(this.events, items) : items;
    this.total = total;
    this.page = page;
    this.totalPages = Math.max(totalPages, 1);
  }
}

function mergeEvents(existing: ChangeEvent[], incoming: ChangeEvent[]): ChangeEvent[] {
  const seen = new Set(existing.map((event) => event.id));
  return [...existing, ...incoming.filter((event) => !seen.has(event.id))];
}

function isAbortError(error: unknown): boolean {
  return error instanceof DOMException && error.name === 'AbortError';
}
