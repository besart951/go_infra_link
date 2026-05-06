type BulkFetchByIds<T extends { id: string }> = (
  ids: string[],
  signal?: AbortSignal
) => Promise<T[]>;

interface CachedBulkFetchOptions {
  ttlMs?: number;
}

interface CacheEntry<T> {
  expiresAt: number;
  value?: T | null;
  promise?: Promise<T | null>;
}

const cacheByNamespace = new Map<string, Map<string, CacheEntry<unknown>>>();

function cacheFor(namespace: string): Map<string, CacheEntry<unknown>> {
  const existing = cacheByNamespace.get(namespace);
  if (existing) {
    return existing;
  }

  const created = new Map<string, CacheEntry<unknown>>();
  cacheByNamespace.set(namespace, created);
  return created;
}

function normalizeIds(ids: string[]): string[] {
  return [...new Set(ids.map((id) => id.trim()).filter((id) => id.length > 0))];
}

export function createCachedBulkFetchByIds<T extends { id: string }>(
  namespace: string,
  fetchByIds: BulkFetchByIds<T>,
  options: CachedBulkFetchOptions = {}
): BulkFetchByIds<T> {
  const ttlMs = options.ttlMs ?? 60_000;

  return async (ids: string[], signal?: AbortSignal) => {
    const normalizedIds = normalizeIds(ids);
    if (normalizedIds.length === 0) {
      return [];
    }

    const cache = cacheFor(namespace) as Map<string, CacheEntry<T>>;
    const now = Date.now();
    const pending: Array<Promise<T | null>> = [];
    const missingIds: string[] = [];

    for (const id of normalizedIds) {
      const existing = cache.get(id);
      if (existing && existing.expiresAt > now) {
        pending.push(existing.promise ?? Promise.resolve(existing.value ?? null));
        continue;
      }

      missingIds.push(id);
    }

    if (missingIds.length > 0) {
      const fetchedByIdPromise: Promise<Map<string, T>> = fetchByIds(missingIds, signal)
        .then((items) => {
          const fetchedById = new Map(items.map((item) => [item.id, item] as const));
          for (const id of missingIds) {
            cache.set(id, {
              expiresAt: Date.now() + ttlMs,
              value: fetchedById.get(id) ?? null
            });
          }
          return fetchedById;
        })
        .catch((error) => {
          for (const id of missingIds) {
            cache.delete(id);
          }
          throw error;
        });

      for (const id of missingIds) {
        const promise = fetchedByIdPromise.then((fetchedById) => fetchedById.get(id) ?? null);
        cache.set(id, {
          expiresAt: now + ttlMs,
          promise
        });
        pending.push(promise);
      }
    }

    const values = (await Promise.all(pending)) as Array<T | null>;
    return values.filter((value): value is T => value !== null);
  };
}

export function clearCachedBulkFetchByIds(namespace?: string): void {
  if (namespace) {
    cacheByNamespace.delete(namespace);
    return;
  }

  cacheByNamespace.clear();
}
