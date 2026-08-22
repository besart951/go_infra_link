import { ListEntityUseCase } from '$lib/application/useCases/listEntityUseCase.js';
import { fetchAllPages } from '$lib/components/facility/shared/paginatedListFetcher.js';
import type { Apparat, FieldDeviceOptions, SystemPart } from '$lib/domain/facility/index.js';
import type { FacilityJobKind, FacilityJobStatus } from '$lib/domain/facility/facility-job.js';
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

export interface FacilityJobProgressEvent {
  type: 'facility.job.progress';
  job_id: string;
  kind: FacilityJobKind;
  status: FacilityJobStatus;
  progress: number;
  stage: string;
  job_type?: 'copy' | 'export' | 'bulk' | 'delete' | 'restore';
  class?: 'mutation' | 'export';
  processed?: number;
  total?: number;
  success_count?: number;
  failure_count?: number;
  error?: string;
  updated_at: string;
}

export type FacilityRealtimeResource =
  | 'all'
  | 'buildings'
  | 'system_types'
  | 'system_parts'
  | 'apparats'
  | 'control_cabinets'
  | 'sps_controllers'
  | 'sps_controller_system_types'
  | 'field_devices'
  | 'bacnet_objects'
  | 'object_data'
  | 'state_texts'
  | 'notification_classes'
  | 'alarm_definitions'
  | 'alarm_types'
  | 'alarm_type_fields'
  | 'alarm_fields'
  | 'units';

export interface FacilityChangeEvent {
  type: 'facility.changed';
  resource: FacilityRealtimeResource;
  action:
    | 'created'
    | 'updated'
    | 'deleted'
    | 'copied'
    | 'bulk_created'
    | 'bulk_updated'
    | 'bulk_deleted'
    | 'reconnected';
  ids: string[];
  actor_id?: string;
  at: string;
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
    onJobProgress: (event: FacilityJobProgressEvent) => void,
    onOpen: () => void,
    onFacilityChange: (event: FacilityChangeEvent) => void
  ) => FacilityReferenceDataStream;
  ttlMs?: number;
  stopGraceMs?: number;
  now?: () => number;
}

export interface FacilityReferenceDataStartOptions {
  refreshReferenceData?: boolean;
  currentUserId?: string;
}

interface FacilityReferenceDataRequest {
  id: number;
  force: boolean;
  promise: Promise<FacilityReferenceData>;
}

export const FACILITY_REFERENCE_DATA_CACHE_TTL_MS = 30 * 60 * 1000;
export const FACILITY_REFERENCE_DATA_STOP_GRACE_MS = 1500;

const facilityReferenceDataChangedEventSchema = z.object({
  type: z.literal('facility_reference_data.changed'),
  resources: z.array(z.enum(['apparats', 'system_parts'])).min(1),
  at: z.string()
});

const facilityJobProgressEventSchema = z.object({
  type: z.literal('facility.job.progress'),
  job_id: z.string().uuid(),
  kind: z.enum([
    'control_cabinet',
    'sps_controller',
    'sps_controller_system_type',
    'field_device',
    'object_data'
  ]),
  job_type: z.enum(['copy', 'export', 'bulk', 'delete', 'restore']).optional(),
  class: z.enum(['mutation', 'export']).optional(),
  status: z.enum(['queued', 'running', 'completed', 'failed']),
  progress: z.number().int().min(0).max(100),
  stage: z.string(),
  processed: z.number().int().nonnegative().optional(),
  total: z.number().int().nonnegative().optional(),
  success_count: z.number().int().nonnegative().optional(),
  failure_count: z.number().int().nonnegative().optional(),
  error: z.string().optional(),
  updated_at: z.string()
});

const facilityChangeEventSchema = z.object({
  type: z.literal('facility.changed'),
  resource: z.enum([
    'buildings',
    'system_types',
    'system_parts',
    'apparats',
    'control_cabinets',
    'sps_controllers',
    'sps_controller_system_types',
    'field_devices',
    'bacnet_objects',
    'object_data',
    'state_texts',
    'notification_classes',
    'alarm_definitions',
    'alarm_types',
    'alarm_type_fields',
    'alarm_fields',
    'units'
  ]),
  action: z.enum([
    'created',
    'updated',
    'deleted',
    'copied',
    'bulk_created',
    'bulk_updated',
    'bulk_deleted'
  ]),
  ids: z.array(z.string().uuid()),
  actor_id: z.string().uuid().optional(),
  at: z.string()
});

const facilityRealtimeEventSchema = z.union([
  facilityReferenceDataChangedEventSchema,
  facilityJobProgressEventSchema,
  facilityChangeEventSchema
]);

export class FacilityReferenceDataCache {
  private readonly listApparatsUseCase: ListEntityUseCase<Apparat>;
  private readonly listSystemPartsUseCase: ListEntityUseCase<SystemPart>;
  private readonly fieldDevices: Pick<FieldDeviceRepository, 'getOptions'>;
  private readonly stream: FacilityReferenceDataStream;
  private readonly listeners = new Set<(data: FacilityReferenceData) => void>();
  private readonly jobListeners = new Set<(event: FacilityJobProgressEvent) => void>();
  private readonly facilityChangeListeners = new Set<(event: FacilityChangeEvent) => void>();
  private readonly realtimeOpenListeners = new Set<() => void>();
  private readonly ttlMs: number;
  private readonly stopGraceMs: number;
  private readonly now: () => number;
  private data: FacilityReferenceData | null = null;
  private expiresAt = 0;
  private request: FacilityReferenceDataRequest | null = null;
  private cacheGeneration = 0;
  private requestID = 0;
  private refreshReferenceData = false;
  private currentUserId: string | null = null;
  private stopTimer: ReturnType<typeof setTimeout> | null = null;

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
    this.stopGraceMs = options.stopGraceMs ?? FACILITY_REFERENCE_DATA_STOP_GRACE_MS;
    this.now = options.now ?? Date.now;
    this.stream =
      options.createStream?.(
        () => this.handleReferenceDataChange(),
        (event) => this.notifyJobProgress(event),
        () => this.notifyRealtimeOpen(),
        (event) => this.notifyFacilityChange(event)
      ) ??
      createFacilityReferenceDataStream(
        () => this.handleReferenceDataChange(),
        (event) => this.notifyJobProgress(event),
        () => this.notifyRealtimeOpen(),
        (event) => this.notifyFacilityChange(event)
      );
  }

  /**
   * Opens the single facility realtime stream for every signed-in user. Users
   * without both reference-data permissions still receive their own job
   * progress, while reference data is fetched only when it is authorized.
   */
  start(options: FacilityReferenceDataStartOptions = {}): void {
    this.cancelScheduledStop();
    this.refreshReferenceData = options.refreshReferenceData ?? true;
    this.currentUserId = options.currentUserId ?? null;
    this.stream.connect();
    if (this.refreshReferenceData) this.loadInBackground();
  }

  stop(options: { immediate?: boolean } = {}): void {
    if (options.immediate) {
      this.cancelScheduledStop();
      this.finishStop();
      return;
    }
    if (this.stopTimer) return;

    // SvelteKit can briefly unmount and remount the authenticated layout while
    // resolving route data. Keep the shared stream alive across that handover
    // so a single browser never opens a second facility connection.
    this.stopTimer = setTimeout(() => {
      this.stopTimer = null;
      this.finishStop();
    }, this.stopGraceMs);
  }

  private finishStop(): void {
    this.refreshReferenceData = false;
    this.currentUserId = null;
    this.cacheGeneration += 1;
    this.data = null;
    this.expiresAt = 0;
    this.request = null;
    this.stream.disconnect();
  }

  private cancelScheduledStop(): void {
    if (!this.stopTimer) return;
    clearTimeout(this.stopTimer);
    this.stopTimer = null;
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

  subscribeJobProgress(listener: (event: FacilityJobProgressEvent) => void): () => void {
    this.jobListeners.add(listener);
    return () => {
      this.jobListeners.delete(listener);
    };
  }

  subscribeFacilityChanges(listener: (event: FacilityChangeEvent) => void): () => void {
    this.facilityChangeListeners.add(listener);
    return () => {
      this.facilityChangeListeners.delete(listener);
    };
  }

  /**
   * Returns whether an event was caused by the user of this browser session.
   * Callers use this only while guarding an unsaved form: the list/detail can
   * still refresh after the user's own completed mutation, but it must not
   * present that mutation as a change made by someone else.
   */
  isChangeFromCurrentUser(event: FacilityChangeEvent): boolean {
    return Boolean(this.currentUserId && event.actor_id === this.currentUserId);
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

  private notifyJobProgress(event: FacilityJobProgressEvent): void {
    for (const listener of this.jobListeners) {
      listener(event);
    }
  }

  private notifyRealtimeOpen(): void {
    for (const listener of this.realtimeOpenListeners) {
      listener();
    }
  }

  private notifyFacilityChange(event: FacilityChangeEvent): void {
    for (const listener of this.facilityChangeListeners) {
      listener(event);
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
  onJobProgress: (event: FacilityJobProgressEvent) => void,
  onOpen: () => void,
  onFacilityChange: (event: FacilityChangeEvent) => void
): FacilityReferenceDataStream {
  return new RealtimeJsonStream({
    url: () => buildSameOriginWebSocketUrl('/api/v1/facility/reference-data/stream'),
    parseMessage: (message) => facilityRealtimeEventSchema.parse(message),
    onMessage: (event) => {
      if (event.type === 'facility_reference_data.changed') {
        onChange();
        return;
      }
      if (event.type === 'facility.job.progress') {
        onJobProgress(event);
        return;
      }
      if (event.type === 'facility.changed') {
        onFacilityChange(event);
      }
    },
    onOpen: ({ wasReconnect }) => {
      if (wasReconnect) {
        onChange();
        onFacilityChange({
          type: 'facility.changed',
          resource: 'all',
          action: 'reconnected',
          ids: [],
          at: new Date().toISOString()
        });
      }
      onOpen();
    },
    onInvalidMessage: (raw, error) => {
      console.warn('Ignored invalid facility reference data event', { raw, error });
    }
  });
}

const facilityReferenceDataCacheGlobalKey = '__goInfraLinkFacilityReferenceDataCache__';

type FacilityReferenceDataCacheGlobal = typeof globalThis & {
  [facilityReferenceDataCacheGlobalKey]?: FacilityReferenceDataCache;
};

function getFacilityReferenceDataCache(): FacilityReferenceDataCache {
  const globalState = globalThis as FacilityReferenceDataCacheGlobal;
  return (globalState[facilityReferenceDataCacheGlobalKey] ??= new FacilityReferenceDataCache());
}

// Lazy project chunks can evaluate their own copy of this module. Persist the
// cache on the browser global so every chunk still uses one facility stream.
export const facilityReferenceDataCache = getFacilityReferenceDataCache();
