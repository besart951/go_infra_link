import { FieldDeviceLookupService } from './fieldDeviceLookupService.js';
import { FacilityReferenceDataCache } from '$lib/services/facilityReferenceDataCache.js';
import type { Apparat, FieldDeviceOptions, SystemPart } from '$lib/domain/facility/index.js';

const systemPartAir: SystemPart = {
  id: 'system-part-air',
  version: 1,
  short_name: 'Abl',
  name: 'Abluft',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z'
};

const systemPartHeat: SystemPart = {
  id: 'system-part-heat',
  version: 1,
  short_name: 'Hzg',
  name: 'Heizung',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z'
};

const apparatDamper: Apparat = {
  id: 'apparat-damper',
  version: 1,
  short_name: 'Abk',
  name: 'Abschaltung',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z'
};

const apparatPump: Apparat = {
  id: 'apparat-pump',
  version: 1,
  short_name: 'Pmp',
  name: 'Pumpe',
  created_at: '2026-01-01T00:00:00Z',
  updated_at: '2026-01-01T00:00:00Z'
};

function repository<T>(items: T[]) {
  return {
    list: vi.fn().mockResolvedValue({
      items,
      metadata: {
        total: items.length,
        page: 1,
        pageSize: 1000,
        totalPages: 1
      }
    }),
    get: vi.fn(),
    getBulk: vi.fn(),
    create: vi.fn(),
    update: vi.fn(),
    delete: vi.fn()
  };
}

function options(overrides: Partial<FieldDeviceOptions> = {}): FieldDeviceOptions {
  return {
    apparats: [apparatDamper],
    system_parts: [systemPartAir],
    object_datas: [],
    apparat_to_system_part: {
      [apparatDamper.id]: [systemPartAir.id]
    },
    object_data_to_apparat: {},
    ...overrides
  };
}

describe('FieldDeviceLookupService', () => {
  it('falls back to field device options when paginated lookup lists are empty', async () => {
    const apparats = repository<Apparat>([]);
    const systemParts = repository<SystemPart>([]);
    const fieldDevices = {
      getOptions: vi.fn().mockResolvedValue(options()),
      getOptionsForProject: vi.fn()
    };
    const service = new FieldDeviceLookupService(
      apparats,
      systemParts,
      fieldDevices,
      new FacilityReferenceDataCache({ apparats, systemParts, fieldDevices })
    );

    const result = await service.loadStaticLookups();

    expect(result.apparats).toEqual([
      {
        ...apparatDamper,
        system_parts: [systemPartAir]
      }
    ]);
    expect(result.systemParts).toEqual([systemPartAir]);
  });

  it('keeps all paginated apparats and enriches relations from field device options', async () => {
    const apparats = repository([apparatDamper, apparatPump]);
    const systemParts = repository([systemPartAir, systemPartHeat]);
    const fieldDevices = {
      getOptions: vi.fn().mockResolvedValue(
        options({
          apparats: [apparatDamper, apparatPump],
          system_parts: [systemPartAir, systemPartHeat],
          apparat_to_system_part: {
            [apparatDamper.id]: [systemPartAir.id],
            [apparatPump.id]: [systemPartHeat.id]
          }
        })
      ),
      getOptionsForProject: vi.fn()
    };
    const service = new FieldDeviceLookupService(
      apparats,
      systemParts,
      fieldDevices,
      new FacilityReferenceDataCache({ apparats, systemParts, fieldDevices })
    );

    const result = await service.loadStaticLookups();

    expect(result.apparats).toEqual([
      {
        ...apparatDamper,
        system_parts: [systemPartAir]
      },
      {
        ...apparatPump,
        system_parts: [systemPartHeat]
      }
    ]);
    expect(result.systemParts).toEqual([systemPartAir, systemPartHeat]);
  });
});
