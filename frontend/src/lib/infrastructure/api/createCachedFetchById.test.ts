import { afterEach, describe, expect, it, vi } from 'vitest';

import { clearCachedFetchById, createCachedFetchById } from './createCachedFetchById.js';

describe('createCachedFetchById', () => {
  afterEach(() => {
    clearCachedFetchById();
  });

  it('deduplicates in-flight requests and reuses cached values', async () => {
    const fetchById = vi.fn(async (id: string) => ({ id, label: `item-${id}` }));
    const cachedFetch = createCachedFetchById('test-items', fetchById);

    const [first, second] = await Promise.all([cachedFetch('1'), cachedFetch('1')]);
    const third = await cachedFetch('1');

    expect(first).toEqual({ id: '1', label: 'item-1' });
    expect(second).toEqual({ id: '1', label: 'item-1' });
    expect(third).toEqual({ id: '1', label: 'item-1' });
    expect(fetchById).toHaveBeenCalledTimes(1);
  });

  it('returns null for empty ids without calling the underlying fetcher', async () => {
    const fetchById = vi.fn(async (id: string) => ({ id }));
    const cachedFetch = createCachedFetchById('test-empty', fetchById);

    await expect(cachedFetch('')).resolves.toBeNull();
    expect(fetchById).not.toHaveBeenCalled();
  });
});
