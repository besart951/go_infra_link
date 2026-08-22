import { afterEach, describe, expect, it, vi } from 'vitest';
import { api } from '$lib/api/client.js';
import { fieldDeviceRepository } from './fieldDeviceRepository.js';

vi.mock('$lib/api/client.js', () => ({ api: vi.fn() }));

afterEach(() => vi.clearAllMocks());

describe('fieldDeviceRepository.listCursor', () => {
  it('uses cursor parameters without page or count semantics', async () => {
    vi.mocked(api).mockResolvedValue({
      items: [],
      next_cursor: 'next',
      previous_cursor: 'previous'
    });

    const page = await fieldDeviceRepository.listCursor({
      limit: 300,
      cursor: 'current',
      search: 'pump',
      orderBy: 'bmk',
      order: 'asc',
      filters: { project_id: 'project-1' }
    });

    const [url] = vi.mocked(api).mock.calls[0];
    expect(url).toContain('/facility/field-devices?');
    expect(url).toContain('cursor=current');
    expect(url).toContain('project_id=project-1');
    expect(url).not.toContain('page=');
    expect(page).toEqual({ items: [], nextCursor: 'next', previousCursor: 'previous' });
  });
});

describe('fieldDeviceRepository optimistic deletes', () => {
  it('sends base_version for a single delete', async () => {
    vi.mocked(api).mockResolvedValue(undefined);

    await fieldDeviceRepository.delete({ id: 'device-1', base_version: 7 });

    expect(api).toHaveBeenCalledWith('/facility/field-devices/device-1?base_version=7', {
      method: 'DELETE',
      signal: undefined
    });
  });

  it('sends one versioned command per bulk item', async () => {
    vi.mocked(api).mockResolvedValue({
      results: [],
      total_count: 1,
      success_count: 1,
      failure_count: 0
    });
    const commands = [{ id: 'device-1', base_version: 7 }];

    await fieldDeviceRepository.bulkDelete(commands);

    expect(api).toHaveBeenCalledWith('/facility/field-devices/bulk-delete', {
      method: 'DELETE',
      body: JSON.stringify({ items: commands }),
      signal: undefined
    });
  });
});
