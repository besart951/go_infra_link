import { ListEntityUseCase } from '$lib/application/useCases/listEntityUseCase.js';
import { fetchAllPages } from '$lib/components/facility/shared/paginatedListFetcher.js';
import type { Apparat, FieldDeviceOptions, SystemPart } from '$lib/domain/facility/index.js';
import type { CrudRepository } from '$lib/domain/ports/crudRepository.js';
import type { FieldDeviceRepository } from '$lib/domain/ports/facility/fieldDeviceRepository.js';
import { apparatRepository } from '$lib/infrastructure/api/apparatRepository.js';
import { fieldDeviceRepository } from '$lib/infrastructure/api/fieldDeviceRepository.js';
import { systemPartRepository } from '$lib/infrastructure/api/systemPartRepository.js';
import {
  buildSameOriginWebSocketUrl,
  RealtimeJsonStream
} from '$lib/infrastructure/realtime/realtimeJsonStream.js';
import { z } from 'zod';

export interface FacilityReferenceData {
  apparats: Apparat[];
  systemParts: SystemPart[];
}

interface FacilityReferenceDataDependencies {
  apparats: CrudRepository<Apparat, unknown, unknown>;
  systemParts: CrudRepository<SystemPart, unknown, unknown>;
  fieldDevices: Pick<FieldDeviceRepository, 'getOptions'>;
}

export interface FacilityReferenceDataStream {
  connect(): void;
  disconnect(): void;
}

interface FacilityReferenceDataCacheOptions {
  createStream?: (onChange: () => void) => FacilityReferenceDataStream;
}

const facilityReferenceDataEventSchema = z.object({
  type: z.literal('facility_reference_data.changed'),
  resources: z.array(z.enum(['apparats', 'system_parts'])).min(1),
  at: z.string()
});

export class FacilityReferenceDataCache {
  private readonly listApparatsUseCase: ListEntityUseCase<Apparat>;
  private readonly listSystemPartsUseCase: ListEntityUseCase<SystemPart>;
  private readonly fieldDevices: Pick<FieldDeviceRepository, 'getOptions'>;
  private readonly stream: FacilityReferenceDataStream;
  private readonly listeners = new Set<(data: FacilityReferenceData) => void>();
  private data: FacilityReferenceData | null = null;
  private request: Promise<FacilityReferenceData> | null = null;

  constructor(
    dependencies: FacilityReferenceDataDependencies = {
      apparats: apparatRepository,
      systemParts: systemPartRepository,
      fieldDevices: fieldDeviceRepository
    },
    options: FacilityReferenceDataCacheOptions = {}
  ) {
    this.listApparatsUseCase = new ListEntityUseCase(dependencies.apparats);
    this.listSystemPartsUseCase = new ListEntityUseCase(dependencies.systemParts);
    this.fieldDevices = dependencies.fieldDevices;
    this.stream =
      options.createStream?.(() => this.refreshInBackground()) ??
      createFacilityReferenceDataStream(() => this.refreshInBackground());
  }

  start(): void {
    this.stream.connect();
    this.refreshInBackground();
  }

  stop(): void {
    this.stream.disconnect();
  }

  async load(): Promise<FacilityReferenceData> {
    if (this.data) return this.data;
    return this.fetch(false);
  }

  async refresh(): Promise<FacilityReferenceData> {
    return this.fetch(true);
  }

  subscribe(listener: (data: FacilityReferenceData) => void): () => void {
    this.listeners.add(listener);
    if (this.data) listener(this.data);

    return () => {
      this.listeners.delete(listener);
    };
  }

  private refreshInBackground(): void {
    void this.refresh().catch(() => undefined);
  }

  private fetch(force: boolean): Promise<FacilityReferenceData> {
    if (!force && this.data) return Promise.resolve(this.data);
    if (this.request) return this.request;

    const request = this.fetchReferenceData()
      .then((data) => {
        this.data = data;
        this.notify(data);
        return data;
      })
      .finally(() => {
        if (this.request === request) {
          this.request = null;
        }
      });

    this.request = request;
    return request;
  }

  private async fetchReferenceData(): Promise<FacilityReferenceData> {
    const [optionsResult, apparatsResult, systemPartsResult] = await Promise.allSettled([
      this.fieldDevices.getOptions(),
      fetchAllPages((page, pageSize) =>
        this.listApparatsUseCase.execute({
          pagination: { page, pageSize },
          search: { text: '' }
        })
      ),
      fetchAllPages((page, pageSize) =>
        this.listSystemPartsUseCase.execute({
          pagination: { page, pageSize },
          search: { text: '' }
        })
      )
    ]);

    const options = optionsResult.status === 'fulfilled' ? optionsResult.value : undefined;
    const apparats = resolveReferenceItems(apparatsResult, options?.apparats, 'apparats');
    const systemParts = resolveReferenceItems(
      systemPartsResult,
      options?.system_parts,
      'system parts'
    );

    return {
      apparats: withOptionRelations(apparats, systemParts, options),
      systemParts
    };
  }

  private notify(data: FacilityReferenceData): void {
    for (const listener of this.listeners) {
      listener(data);
    }
  }
}

function resolveReferenceItems<T>(
  result: PromiseSettledResult<T[]>,
  fallback: T[] | undefined,
  resource: string
): T[] {
  if (result.status === 'fulfilled' && result.value.length > 0) {
    return result.value;
  }
  if (fallback) return fallback;
  if (result.status === 'fulfilled') return result.value;
  throw new Error(`Failed to load ${resource}`, { cause: result.reason });
}

export function withOptionRelations(
  apparats: Apparat[],
  systemParts: SystemPart[],
  options: FieldDeviceOptions | undefined
): Apparat[] {
  if (!options?.apparat_to_system_part) return apparats;

  const systemPartsById = new Map(systemParts.map((systemPart) => [systemPart.id, systemPart]));

  return apparats.map((apparat) => {
    const relationIDs = options.apparat_to_system_part[apparat.id];
    if (!relationIDs) return apparat;

    return {
      ...apparat,
      system_parts: relationIDs
        .map((systemPartID) => systemPartsById.get(systemPartID))
        .filter((systemPart): systemPart is SystemPart => Boolean(systemPart))
    };
  });
}

function createFacilityReferenceDataStream(onChange: () => void): FacilityReferenceDataStream {
  return new RealtimeJsonStream({
    url: () => buildSameOriginWebSocketUrl('/api/v1/facility/reference-data/stream'),
    parseMessage: (message) => facilityReferenceDataEventSchema.parse(message),
    onMessage: onChange,
    onOpen: onChange,
    onInvalidMessage: (raw, error) => {
      console.warn('Ignored invalid facility reference data event', { raw, error });
    }
  });
}

export const facilityReferenceDataCache = new FacilityReferenceDataCache();
