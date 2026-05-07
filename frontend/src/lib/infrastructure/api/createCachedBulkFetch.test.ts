import { afterEach, describe, expect, it, vi } from 'vitest';

import { clearCachedBulkFetchByIds, createCachedBulkFetchByIds } from './createCachedBulkFetch.js';

describe('createCachedBulkFetchByIds', () => {
  afterEach(() => {
    clearCachedBulkFetchByIds();
  });

  it('deduplicates ids and reuses cached values', async () => {
    const fetchByIds = vi.fn(async (ids: string[]) =>
      ids.map((id) => ({ id, label: `item-${id}` }))
    );
    const cachedFetch = createCachedBulkFetchByIds('test-items', fetchByIds);

    const first = await cachedFetch(['1', '2', '1']);
    const second = await cachedFetch(['2', '3']);

    expect(first).toEqual([
      { id: '1', label: 'item-1' },
      { id: '2', label: 'item-2' }
    ]);
    expect(second).toEqual([
      { id: '2', label: 'item-2' },
      { id: '3', label: 'item-3' }
    ]);
    expect(fetchByIds).toHaveBeenCalledTimes(2);
    expect(fetchByIds).toHaveBeenNthCalledWith(1, ['1', '2'], undefined);
    expect(fetchByIds).toHaveBeenNthCalledWith(2, ['3'], undefined);
  });

  it('deduplicates in-flight overlapping requests', async () => {
    let resolveFetch: (items: Array<{ id: string }>) => void = () => {};
    const fetchByIds = vi.fn(
      () =>
        new Promise<Array<{ id: string }>>((resolve) => {
          resolveFetch = resolve;
        })
    );
    const cachedFetch = createCachedBulkFetchByIds('test-in-flight', fetchByIds);

    const first = cachedFetch(['1']);
    const second = cachedFetch(['1']);
    resolveFetch([{ id: '1' }]);

    await expect(Promise.all([first, second])).resolves.toEqual([[{ id: '1' }], [{ id: '1' }]]);
    expect(fetchByIds).toHaveBeenCalledTimes(1);
  });

  it('returns an empty array for empty ids without calling the underlying fetcher', async () => {
    const fetchByIds = vi.fn(async (ids: string[]) => ids.map((id) => ({ id })));
    const cachedFetch = createCachedBulkFetchByIds('test-empty', fetchByIds);

    await expect(cachedFetch(['', ' '])).resolves.toEqual([]);
    expect(fetchByIds).not.toHaveBeenCalled();
  });
});
