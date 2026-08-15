import { FacilityReferenceDataCache } from './facilityReferenceDataCache.js';
import type { Apparat, FieldDeviceOptions, SystemPart } from '$lib/domain/facility/index.js';

const systemPart: SystemPart = {
  id: 'system-part-air',
  short_name: 'Abl',
  name: 'Abluft',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z'
};

const apparat: Apparat = {
  id: 'apparat-damper',
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
    cache.stop();
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

    cache.stop();
    expect(stream.disconnect).toHaveBeenCalledOnce();
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
});
