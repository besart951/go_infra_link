type FetchById<T> = (id: string) => Promise<T | null | undefined>;

interface CachedFetchByIdOptions {
  ttlMs?: number;
}

interface CacheEntry<T> {
  expiresAt: number;
  value?: T | null;
  promise?: Promise<T | null | undefined>;
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

export function createCachedFetchById<T>(
  namespace: string,
  fetchById: FetchById<T>,
  options: CachedFetchByIdOptions = {}
): FetchById<T> {
  const ttlMs = options.ttlMs ?? 60_000;

  return async (id: string) => {
    if (!id) {
      return null;
    }

    const cache = cacheFor(namespace) as Map<string, CacheEntry<T>>;
    const now = Date.now();
    const existing = cache.get(id);

    if (existing && existing.expiresAt > now) {
      if (existing.promise) {
        return existing.promise;
      }
      return existing.value;
    }

    const promise = fetchById(id)
      .then((value) => {
        cache.set(id, {
          expiresAt: Date.now() + ttlMs,
          value: value ?? null
        });
        return value ?? null;
      })
      .catch((error) => {
        cache.delete(id);
        throw error;
      });

    cache.set(id, {
      expiresAt: now + ttlMs,
      promise
    });

    return promise;
  };
}

export function clearCachedFetchById(namespace?: string): void {
  if (namespace) {
    cacheByNamespace.delete(namespace);
    return;
  }

  cacheByNamespace.clear();
}
