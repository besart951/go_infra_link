import { ListEntityUseCase } from '$lib/application/useCases/listEntityUseCase.js';
import { fetchAllPages } from '$lib/components/facility/shared/paginatedListFetcher.js';
import type { Apparat, FieldDeviceOptions, SystemPart } from '$lib/domain/facility/index.js';
import type { CopyJobKind, CopyJobStatus } from '$lib/domain/facility/copy-job.js';
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

export interface FacilityCopyJobProgressEvent {
  type: 'facility.copy_job.progress';
  job_id: string;
  kind: CopyJobKind;
  status: CopyJobStatus;
  progress: number;
  stage: string;
  error?: string;
  updated_at: string;
}

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
  createStream?: (
    onChange: () => void,
    onCopyJobProgress: (event: FacilityCopyJobProgressEvent) => void,
    onOpen: () => void
  ) => FacilityReferenceDataStream;
  ttlMs?: number;
  now?: () => number;
}

interface FacilityReferenceDataRequest {
  id: number;
  force: boolean;
  promise: Promise<FacilityReferenceData>;
}

export const FACILITY_REFERENCE_DATA_CACHE_TTL_MS = 30 * 60 * 1000;

const facilityReferenceDataChangedEventSchema = z.object({
  type: z.literal('facility_reference_data.changed'),
  resources: z.array(z.enum(['apparats', 'system_parts'])).min(1),
  at: z.string()
});

const facilityCopyJobProgressEventSchema = z.object({
  type: z.literal('facility.copy_job.progress'),
  job_id: z.string().uuid(),
  kind: z.enum(['control_cabinet', 'sps_controller', 'sps_controller_system_type']),
  status: z.enum(['queued', 'running', 'completed', 'failed']),
  progress: z.number().int().min(0).max(100),
  stage: z.string(),
  error: z.string().optional(),
  updated_at: z.string()
});

const facilityRealtimeEventSchema = z.discriminatedUnion('type', [
  facilityReferenceDataChangedEventSchema,
  facilityCopyJobProgressEventSchema
]);

export class FacilityReferenceDataCache {
  private readonly listApparatsUseCase: ListEntityUseCase<Apparat>;
  private readonly listSystemPartsUseCase: ListEntityUseCase<SystemPart>;
  private readonly fieldDevices: Pick<FieldDeviceRepository, 'getOptions'>;
  private readonly stream: FacilityReferenceDataStream;
  private readonly listeners = new Set<(data: FacilityReferenceData) => void>();
  private readonly copyJobListeners = new Set<(event: FacilityCopyJobProgressEvent) => void>();
  private readonly realtimeOpenListeners = new Set<() => void>();
  private readonly ttlMs: number;
  private readonly now: () => number;
  private data: FacilityReferenceData | null = null;
  private expiresAt = 0;
  private request: FacilityReferenceDataRequest | null = null;
  private cacheGeneration = 0;
  private requestID = 0;
  private refreshReferenceData = false;

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
    this.ttlMs = options.ttlMs ?? FACILITY_REFERENCE_DATA_CACHE_TTL_MS;
    this.now = options.now ?? Date.now;
    this.stream =
      options.createStream?.(
        () => this.handleReferenceDataChange(),
        (event) => this.notifyCopyJobProgress(event),
        () => this.notifyRealtimeOpen()
      ) ??
      createFacilityReferenceDataStream(
        () => this.handleReferenceDataChange(),
        (event) => this.notifyCopyJobProgress(event),
        () => this.notifyRealtimeOpen()
      );
  }

  /**
   * Opens the single facility realtime stream for every signed-in user. Users
   * without both reference-data permissions still receive their own copy-job
   * progress, while reference data is fetched only when it is authorized.
   */
  start(options: { refreshReferenceData?: boolean } = {}): void {
    this.refreshReferenceData = options.refreshReferenceData ?? true;
    this.stream.connect();
    if (this.refreshReferenceData) this.loadInBackground();
  }

  stop(): void {
    this.refreshReferenceData = false;
    this.cacheGeneration += 1;
    this.data = null;
    this.expiresAt = 0;
    this.request = null;
    this.stream.disconnect();
  }

  async load(): Promise<FacilityReferenceData> {
    if (this.data && this.expiresAt > this.now()) return this.data;
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

  subscribeCopyJobProgress(listener: (event: FacilityCopyJobProgressEvent) => void): () => void {
    this.copyJobListeners.add(listener);
    return () => {
      this.copyJobListeners.delete(listener);
    };
  }

  subscribeRealtimeOpen(listener: () => void): () => void {
    this.realtimeOpenListeners.add(listener);
    return () => {
      this.realtimeOpenListeners.delete(listener);
    };
  }

  private loadInBackground(): void {
    void this.load().catch(() => undefined);
  }

  private refreshInBackground(): void {
    void this.refresh().catch(() => undefined);
  }

  private handleReferenceDataChange(): void {
    if (this.refreshReferenceData) this.refreshInBackground();
  }

  private fetch(force: boolean): Promise<FacilityReferenceData> {
    if (!force && this.data && this.expiresAt > this.now()) return Promise.resolve(this.data);
    if (this.request && (!force || this.request.force)) return this.request.promise;

    const request = this.createRequest(force);
    this.request = request;
    return request.promise;
  }

  private createRequest(force: boolean): FacilityReferenceDataRequest {
    const requestID = ++this.requestID;
    const generation = this.cacheGeneration;
    const promise = this.fetchReferenceData()
      .then((data) => {
        if (generation !== this.cacheGeneration || this.request?.id !== requestID) {
          return data;
        }
        this.data = data;
        this.expiresAt = this.now() + this.ttlMs;
        this.notify(data);
        return data;
      })
      .finally(() => {
        if (this.request?.id === requestID) {
          this.request = null;
        }
      });

    return { id: requestID, force, promise };
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

  private notifyCopyJobProgress(event: FacilityCopyJobProgressEvent): void {
    for (const listener of this.copyJobListeners) {
      listener(event);
    }
  }

  private notifyRealtimeOpen(): void {
    for (const listener of this.realtimeOpenListeners) {
      listener();
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

function createFacilityReferenceDataStream(
  onChange: () => void,
  onCopyJobProgress: (event: FacilityCopyJobProgressEvent) => void,
  onOpen: () => void
): FacilityReferenceDataStream {
  return new RealtimeJsonStream({
    url: () => buildSameOriginWebSocketUrl('/api/v1/facility/reference-data/stream'),
    parseMessage: (message) => facilityRealtimeEventSchema.parse(message),
    onMessage: (event) => {
      if (event.type === 'facility_reference_data.changed') {
        onChange();
        return;
      }
      onCopyJobProgress(event);
    },
    onOpen: ({ wasReconnect }) => {
      if (wasReconnect) onChange();
      onOpen();
    },
    onInvalidMessage: (raw, error) => {
      console.warn('Ignored invalid facility reference data event', { raw, error });
    }
  });
}

export const facilityReferenceDataCache = new FacilityReferenceDataCache();
