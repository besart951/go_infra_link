import { afterEach, describe, expect, it, vi } from 'vitest';
import type { ActivityDataSource } from './contract.js';
import { clearActivityCacheForTests } from './activityCache.js';
import { ActivityTimelineController } from './activityTimelineController.svelte.js';

function response(id: string) {
  return {
    items: [
      {
        id,
        occurred_at: '2026-08-15T10:00:00Z',
        action: 'update' as const,
        entity_table: 'field_devices',
        entity_id: 'field-device-1'
      }
    ],
    next_cursor: 'next-page'
  };
}

afterEach(() => clearActivityCacheForTests());

describe('ActivityTimelineController', () => {
  it('uses the cached query result before starting another request', async () => {
    const source: ActivityDataSource = {
      cacheNamespace: 'test:global',
      list: vi.fn().mockResolvedValue(response('event-1'))
    };
    const controller = new ActivityTimelineController(source);

    await controller.load({ limit: 25 });
    await controller.load({ limit: 25 });

    expect(source.list).toHaveBeenCalledOnce();
    expect(controller.events.map((event) => event.id)).toEqual(['event-1']);
  });

  it('ignores a late response from a superseded filter request', async () => {
    let resolveFirst: ((value: ReturnType<typeof response>) => void) | undefined;
    let resolveSecond: ((value: ReturnType<typeof response>) => void) | undefined;
    const source: ActivityDataSource = {
      cacheNamespace: 'test:race',
      list: vi
        .fn()
        .mockImplementationOnce(
          () => new Promise<ReturnType<typeof response>>((resolve) => (resolveFirst = resolve))
        )
        .mockImplementationOnce(
          () => new Promise<ReturnType<typeof response>>((resolve) => (resolveSecond = resolve))
        )
    };
    const controller = new ActivityTimelineController(source);

    const first = controller.load({ entityTable: 'field_devices' });
    const second = controller.load({ entityTable: 'apparats' });
    resolveSecond?.(response('new-event'));
    await second;
    resolveFirst?.(response('old-event'));
    await first;

    expect(controller.events.map((event) => event.id)).toEqual(['new-event']);
  });

  it('uses the server cursor and appends the next page without duplicates', async () => {
    const first = response('event-1');
    const second = { ...response('event-2'), next_cursor: undefined, previous_cursor: 'previous' };
    const source: ActivityDataSource = {
      cacheNamespace: 'test:cursor',
      list: vi.fn().mockResolvedValueOnce(first).mockResolvedValueOnce(second)
    };
    const controller = new ActivityTimelineController(source);

    await controller.load({ limit: 25 });
    await controller.load({ limit: 25 }, { append: true });

    expect(source.list).toHaveBeenLastCalledWith(
      { limit: 25, cursor: 'next-page' },
      expect.any(AbortSignal)
    );
    expect(controller.events.map((event) => event.id)).toEqual(['event-1', 'event-2']);
    expect(controller.hasMore).toBe(false);
    expect(controller.hasPrevious).toBe(true);
  });
});
