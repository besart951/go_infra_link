import {
  FacilityReferenceDataCache,
  type FacilityChangeEvent
} from './facilityReferenceDataCache.js';
import type { Apparat, FieldDeviceOptions, SystemPart } from '$lib/domain/facility/index.js';

const systemPart: SystemPart = {
  id: 'system-part-air',
  version: 1,
  short_name: 'Abl',
  name: 'Abluft',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z'
};

const apparat: Apparat = {
  id: 'apparat-damper',
  version: 1,
  short_name: 'Abk',
  name: 'Abschaltung',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z'
};

function repository<T>(resolveItems: () => T[]) {
  return {
    list: vi.fn().mockImplementation(() =>
      Promise.resolve({
        items: resolveItems(),
        metadata: {
          total: resolveItems().length,
          page: 1,
          pageSize: 1000,
          totalPages: 1
        }
      })
    ),
    get: vi.fn(),
    getBulk: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn()
  };
}

function options(): FieldDeviceOptions {
  return {
    apparats: [apparat],
    system_parts: [systemPart],
    object_datas: [],
    apparat_to_system_part: { [apparat.id]: [systemPart.id] },
    object_data_to_apparat: {}
  };
}

describe('FacilityReferenceDataCache', () => {
  it('uses its reference data cache for 30 minutes before refreshing it', async () => {
    let now = 0;
    const apparats = repository(() => [apparat]);
    const systemParts = repository(() => [systemPart]);
    const fieldDevices = { getOptions: vi.fn().mockResolvedValue(options()) };
    const cache = new FacilityReferenceDataCache(
      { apparats, systemParts, fieldDevices },
      { createStream: () => ({ connect: vi.fn(), disconnect: vi.fn() }), now: () => now }
    );

    await cache.load();
    now = 30 * 60 * 1000 - 1;
    await cache.load();
    expect(apparats.list).toHaveBeenCalledOnce();

    now += 1;
    await cache.load();
    expect(apparats.list).toHaveBeenCalledTimes(2);
  });

  it('clears reference data when the authenticated application session ends', async () => {
    const apparats = repository(() => [apparat]);
    const systemParts = repository(() => [systemPart]);
    const fieldDevices = { getOptions: vi.fn().mockResolvedValue(options()) };
    const cache = new FacilityReferenceDataCache(
      { apparats, systemParts, fieldDevices },
      { createStream: () => ({ connect: vi.fn(), disconnect: vi.fn() }) }
    );

    await cache.load();
    cache.stop({ immediate: true });
    await cache.load();

    expect(apparats.list).toHaveBeenCalledTimes(2);
  });

  it('loads the reference data once and starts its realtime stream without blocking callers', async () => {
    const apparats = repository(() => [apparat]);
    const systemParts = repository(() => [systemPart]);
    const fieldDevices = { getOptions: vi.fn().mockResolvedValue(options()) };
    const stream = { connect: vi.fn(), disconnect: vi.fn() };
    const cache = new FacilityReferenceDataCache(
      { apparats, systemParts, fieldDevices },
      { createStream: () => stream }
    );

    cache.start();
    const result = await cache.load();

    expect(stream.connect).toHaveBeenCalledOnce();
    expect(apparats.list).toHaveBeenCalledOnce();
    expect(systemParts.list).toHaveBeenCalledOnce();
    expect(result).toEqual({
      apparats: [{ ...apparat, system_parts: [systemPart] }],
      systemParts: [systemPart]
    });

    cache.stop({ immediate: true });
    expect(stream.disconnect).toHaveBeenCalledOnce();
  });

  it('keeps the shared stream and cache during a transient authenticated-layout handover', async () => {
    vi.useFakeTimers();
    try {
      const apparats = repository(() => [apparat]);
      const systemParts = repository(() => [systemPart]);
      const fieldDevices = { getOptions: vi.fn().mockResolvedValue(options()) };
      const stream = { connect: vi.fn(), disconnect: vi.fn() };
      const cache = new FacilityReferenceDataCache(
        { apparats, systemParts, fieldDevices },
        { createStream: () => stream, stopGraceMs: 100 }
      );

      cache.start();
      await cache.load();
      cache.stop();
      cache.start();
      await vi.advanceTimersByTimeAsync(100);
      await cache.load();

      expect(stream.disconnect).not.toHaveBeenCalled();
      expect(apparats.list).toHaveBeenCalledOnce();
    } finally {
      vi.useRealTimers();
    }
  });

  it('keeps the realtime stream open for copy progress without fetching unauthorized reference data', () => {
    const apparats = repository(() => [apparat]);
    const systemParts = repository(() => [systemPart]);
    const fieldDevices = { getOptions: vi.fn().mockResolvedValue(options()) };
    const stream = { connect: vi.fn(), disconnect: vi.fn() };
    const cache = new FacilityReferenceDataCache(
      { apparats, systemParts, fieldDevices },
      { createStream: () => stream }
    );

    cache.start({ refreshReferenceData: false });

    expect(stream.connect).toHaveBeenCalledOnce();
    expect(apparats.list).not.toHaveBeenCalled();
    expect(systemParts.list).not.toHaveBeenCalled();
    expect(fieldDevices.getOptions).not.toHaveBeenCalled();
  });

  it('refreshes the cache and notifies subscribers when the realtime stream signals a change', async () => {
    let apparatItems = [apparat];
    let systemPartItems = [systemPart];
    const apparats = repository(() => apparatItems);
    const systemParts = repository(() => systemPartItems);
    const fieldDevices = { getOptions: vi.fn().mockResolvedValue(options()) };
    let notifyChange: (() => void) | undefined;
    const cache = new FacilityReferenceDataCache(
      { apparats, systemParts, fieldDevices },
      {
        createStream: (onChange) => {
          notifyChange = onChange;
          return { connect: vi.fn(), disconnect: vi.fn() };
        }
      }
    );
    const updates: Apparat[][] = [];
    cache.subscribe((data) => updates.push(data.apparats));

    cache.start();
    await cache.load();
    const updatedApparat = { ...apparat, name: 'Updated apparatus' };
    const updatedSystemPart = { ...systemPart, name: 'Updated system part' };
    apparatItems = [updatedApparat];
    systemPartItems = [updatedSystemPart];

    notifyChange?.();

    await vi.waitFor(() => {
      expect(updates).toHaveLength(2);
    });
    expect(updates[1]).toEqual([{ ...updatedApparat, system_parts: [updatedSystemPart] }]);
  });

  it('forwards resource-scoped facility changes from the shared realtime stream', () => {
    const apparats = repository(() => [apparat]);
    const systemParts = repository(() => [systemPart]);
    const fieldDevices = { getOptions: vi.fn().mockResolvedValue(options()) };
    let notifyFacilityChange: ((event: FacilityChangeEvent) => void) | undefined;
    const cache = new FacilityReferenceDataCache(
      { apparats, systemParts, fieldDevices },
      {
        createStream: (_onChange, _onJobProgress, _onOpen, onFacilityChange) => {
          notifyFacilityChange = onFacilityChange;
          return { connect: vi.fn(), disconnect: vi.fn() };
        }
      }
    );
    const events: FacilityChangeEvent[] = [];
    cache.subscribeFacilityChanges((event) => events.push(event));

    const change: FacilityChangeEvent = {
      type: 'facility.changed',
      resource: 'field_devices',
      action: 'bulk_updated',
      ids: ['3fa85f64-5717-4562-b3fc-2c963f66afa6'],
      at: '2026-08-15T00:00:00Z'
    };
    notifyFacilityChange?.(change);

    expect(events).toEqual([change]);
  });

  it('identifies events from the current user without suppressing other users', () => {
    const apparats = repository(() => [apparat]);
    const systemParts = repository(() => [systemPart]);
    const fieldDevices = { getOptions: vi.fn().mockResolvedValue(options()) };
    const cache = new FacilityReferenceDataCache(
      { apparats, systemParts, fieldDevices },
      { createStream: () => ({ connect: vi.fn(), disconnect: vi.fn() }) }
    );
    const ownChange: FacilityChangeEvent = {
      type: 'facility.changed',
      resource: 'sps_controllers',
      action: 'updated',
      ids: ['3fa85f64-5717-4562-b3fc-2c963f66afa6'],
      actor_id: 'f47ac10b-58cc-4372-a567-0e02b2c3d479',
      at: '2026-08-16T00:00:00Z'
    };

    cache.start({ currentUserId: ownChange.actor_id });

    expect(cache.isChangeFromCurrentUser(ownChange)).toBe(true);
    expect(
      cache.isChangeFromCurrentUser({
        ...ownChange,
        actor_id: '550e8400-e29b-41d4-a716-446655440000'
      })
    ).toBe(false);

    cache.stop({ immediate: true });
    expect(cache.isChangeFromCurrentUser(ownChange)).toBe(false);
  });
});
