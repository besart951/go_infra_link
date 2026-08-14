import { apparatRepository } from '$lib/infrastructure/api/apparatRepository.js';
import { fieldDeviceRepository } from '$lib/infrastructure/api/fieldDeviceRepository.js';
import { systemPartRepository } from '$lib/infrastructure/api/systemPartRepository.js';
import {
  FacilityReferenceDataCache,
  facilityReferenceDataCache,
  withOptionRelations
} from '$lib/services/facilityReferenceDataCache.js';
import type { Apparat, SystemPart } from '$lib/domain/facility/index.js';
import type { CrudRepository } from '$lib/domain/ports/crudRepository.js';
import type { FieldDeviceRepository } from '$lib/domain/ports/facility/fieldDeviceRepository.js';

export interface FieldDeviceStaticLookupResult {
  apparats?: Apparat[];
  systemParts?: SystemPart[];
  apparatsError?: unknown;
  systemPartsError?: unknown;
}

export class FieldDeviceLookupService {
  private readonly referenceDataCache: FacilityReferenceDataCache;
  private cachedLookups: FieldDeviceStaticLookupResult | null = null;
  private cachedLookupScope: string | undefined;
  private lookupRequest: Promise<FieldDeviceStaticLookupResult> | null = null;
  private lookupRequestScope: string | undefined;

  constructor(
    apparats: CrudRepository<Apparat, unknown, unknown> = apparatRepository,
    systemParts: CrudRepository<SystemPart, unknown, unknown> = systemPartRepository,
    private readonly fieldDevices: Pick<
      FieldDeviceRepository,
      'getOptions' | 'getOptionsForProject'
    > = fieldDeviceRepository,
    referenceDataCache?: FacilityReferenceDataCache
  ) {
    this.referenceDataCache =
      referenceDataCache ??
      (apparats === apparatRepository &&
      systemParts === systemPartRepository &&
      this.fieldDevices === fieldDeviceRepository
        ? facilityReferenceDataCache
        : new FacilityReferenceDataCache({ apparats, systemParts, fieldDevices }));
  }

  async loadStaticLookups(projectId?: string): Promise<FieldDeviceStaticLookupResult> {
    const scope = projectId ?? '';
    if (this.cachedLookups && this.cachedLookupScope === scope) {
      return this.cachedLookups;
    }

    if (this.lookupRequest && this.lookupRequestScope === scope) {
      return this.lookupRequest;
    }

    const request = (async () => {
      try {
        const [referenceData, options] = await Promise.all([
          this.referenceDataCache.load(),
          projectId ? this.fieldDevices.getOptionsForProject(projectId) : Promise.resolve(undefined)
        ]);

        return {
          apparats: options
            ? withOptionRelations(referenceData.apparats, referenceData.systemParts, options)
            : referenceData.apparats,
          systemParts: referenceData.systemParts
        };
      } catch (error) {
        return {
          apparatsError: error,
          systemPartsError: error
        };
      }
    })();

    this.lookupRequest = request;
    this.lookupRequestScope = scope;

    request
      .then((result) => {
        this.cachedLookups = result;
        this.cachedLookupScope = scope;
        this.lookupRequest = null;
        this.lookupRequestScope = undefined;
      })
      .catch(() => {
        this.lookupRequest = null;
        this.lookupRequestScope = undefined;
      });

    return request;
  }

  resetCachedLookups(): void {
    this.cachedLookups = null;
    this.cachedLookupScope = undefined;
    this.lookupRequest = null;
    this.lookupRequestScope = undefined;
  }
}
