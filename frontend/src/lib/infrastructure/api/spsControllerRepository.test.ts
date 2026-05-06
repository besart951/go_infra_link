import { afterEach, describe, expect, it, vi } from 'vitest';

import { spsControllerRepository } from './spsControllerRepository.js';
import { api } from '$lib/api/client.js';

vi.mock('$lib/api/client.js', () => ({
  api: vi.fn()
}));

afterEach(() => {
  vi.clearAllMocks();
});

describe('spsControllerRepository', () => {
  it('deduplicates listSystemTypes calls for canonicalized filters', async () => {
    vi.mocked(api).mockResolvedValue({
      items: [],
      total: 0,
      page: 1,
      total_pages: 0
    });

    const first = spsControllerRepository.listSystemTypes({
      page: 1,
      limit: 1000,
      sps_controller_id: 'c|a|b'
    });
    const second = spsControllerRepository.listSystemTypes({
      limit: 1000,
      page: 1,
      sps_controller_id: 'b|a|c'
    });

    const [firstResult, secondResult] = await Promise.all([first, second]);
    expect(firstResult).toEqual({
      items: [],
      total: 0,
      page: 1,
      limit: 1000,
      total_pages: 0
    });
    expect(secondResult).toEqual(firstResult);
    expect(api).toHaveBeenCalledTimes(1);
  });
});
