import type { components } from '$lib/api/generated/schema.js';
import type {
  FacilityChangeEvent,
  FacilityRealtimeResource
} from './facilityReferenceDataCache.js';

export type FacilityDetailKind =
  | 'buildings'
  | 'control-cabinets'
  | 'sps-controllers'
  | 'sps-controller-system-types'
  | 'field-devices';

export type FacilityDetail =
  | components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.BuildingDetailResponse']
  | components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.ControlCabinetDetailResponse']
  | components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerDetailResponse']
  | components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.SPSControllerSystemTypeDetailResponse']
  | components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_facility.FieldDeviceDetailResponse'];

export interface FacilityDetailScope {
  projectId?: string;
  relationPage?: number;
}

interface CacheEntry {
  detail: FacilityDetail;
  expiresAt: number;
}

export const FACILITY_DETAIL_CACHE_TTL_MS = 30 * 60 * 1000;

const resourceToKinds: Partial<Record<FacilityRealtimeResource, FacilityDetailKind[]>> = {
  all: [
    'buildings',
    'control-cabinets',
    'sps-controllers',
    'sps-controller-system-types',
    'field-devices'
  ],
  buildings: [
    'buildings',
    'control-cabinets',
    'sps-controllers',
    'sps-controller-system-types',
    'field-devices'
  ],
  control_cabinets: [
    'buildings',
    'control-cabinets',
    'sps-controllers',
    'sps-controller-system-types',
    'field-devices'
  ],
  sps_controllers: [
    'control-cabinets',
    'sps-controllers',
    'sps-controller-system-types',
    'field-devices'
  ],
  sps_controller_system_types: ['sps-controllers', 'sps-controller-system-types', 'field-devices'],
  field_devices: ['sps-controller-system-types', 'field-devices'],
  apparats: ['field-devices'],
  system_parts: ['field-devices'],
  bacnet_objects: ['field-devices']
};

/**
 * Browser-session cache for permission-filtered hierarchy details. The active
 * user ID is part of its lifecycle so data from a prior account is never
 * reused after logout/login. Realtime only invalidates dependent entries; a
 * clean detail page decides when to revalidate its visible state.
 */
export class FacilityDetailCache {
  private readonly entries = new Map<string, CacheEntry>();
  private activeUserId: string | null = null;

  start(userId?: string): void {
    const nextUserId = userId ?? null;
    if (nextUserId !== this.activeUserId) {
      this.clear();
      this.activeUserId = nextUserId;
    }
  }

  stop(): void {
    this.activeUserId = null;
    this.clear();
  }

  get(
    kind: FacilityDetailKind,
    id: string,
    scope: FacilityDetailScope = {}
  ): FacilityDetail | null {
    const entry = this.entries.get(this.key(kind, id, scope));
    if (!entry) return null;
    if (entry.expiresAt <= Date.now()) {
      this.entries.delete(this.key(kind, id, scope));
      return null;
    }
    return structuredClone(entry.detail);
  }

  set(
    kind: FacilityDetailKind,
    id: string,
    detail: FacilityDetail,
    scope: FacilityDetailScope = {}
  ): void {
    this.entries.set(this.key(kind, id, scope), {
      detail: structuredClone(detail),
      expiresAt: Date.now() + FACILITY_DETAIL_CACHE_TTL_MS
    });
  }

  invalidate(kind: FacilityDetailKind, id: string, scope: FacilityDetailScope = {}): void {
    this.entries.delete(this.key(kind, id, scope));
  }

  invalidateProject(projectId: string): void {
    for (const key of this.entries.keys()) {
      if (key.includes(`:project:${projectId}:`)) this.entries.delete(key);
    }
  }

  invalidateForFacilityChange(event: FacilityChangeEvent): void {
    const affectedKinds = resourceToKinds[event.resource] ?? [];
    if (affectedKinds.length === 0) return;

    for (const key of this.entries.keys()) {
      if (affectedKinds.some((kind) => key.includes(`:${kind}:`))) {
        this.entries.delete(key);
      }
    }
  }

  clear(): void {
    this.entries.clear();
  }

  private key(kind: FacilityDetailKind, id: string, scope: FacilityDetailScope): string {
    const user = this.activeUserId ?? 'anonymous';
    const context = scope.projectId ? `project:${scope.projectId}` : 'global';
    const page = scope.relationPage ?? 1;
    return `${user}:${context}:${kind}:${id}:relation-page:${page}`;
  }
}

export const facilityDetailCache = new FacilityDetailCache();
