import { afterEach, describe, expect, it, vi } from 'vitest';

import { clearCachedFetchById } from './createCachedFetchById.js';
import { spsControllerSystemTypeRepository } from './spsControllerSystemTypeRepository.js';
import { api } from '$lib/api/client.js';

vi.mock('$lib/api/client.js', () => ({
  api: vi.fn()
}));

afterEach(() => {
  vi.clearAllMocks();
  clearCachedFetchById('facility-sps-controller-system-types');
});

describe('spsControllerSystemTypeRepository', () => {
  it('deduplicates in-flight list calls with canonicalized filters', async () => {
    vi.mocked(api).mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      total_pages: 0
    });

    const first = spsControllerSystemTypeRepository.list({
      pagination: { page: 1, pageSize: 1000 },
      search: { text: '' },
      filters: {
        sps_controller_id: 'c|a|b'
      }
    });
    const second = spsControllerSystemTypeRepository.list({
      pagination: { page: 1, pageSize: 1000 },
      search: { text: '' },
      filters: {
        sps_controller_id: 'b|a|c'
      }
    });

    const [firstResult, secondResult] = await Promise.all([first, second]);
    expect(firstResult).toEqual({
      items: [],
      metadata: { total: 0, page: 1, pageSize: 1000, totalPages: 0 }
    });
    expect(secondResult).toEqual(firstResult);
    expect(api).toHaveBeenCalledTimes(1);
  });

  it('caches identical get requests', async () => {
    vi.mocked(api).mockResolvedValue({
      id: 'system-type-1',
      sps_controller_id: 'controller-1',
      system_type_id: 'type-1',
      created_at: '2026-01-01T00:00:00Z',
      updated_at: '2026-01-01T00:00:00Z'
    });

    const [first, second] = await Promise.all([
      spsControllerSystemTypeRepository.get('system-type-1'),
      spsControllerSystemTypeRepository.get('system-type-1')
    ]);

    expect(first).toEqual(second);
    expect(api).toHaveBeenCalledTimes(1);
  });
});
