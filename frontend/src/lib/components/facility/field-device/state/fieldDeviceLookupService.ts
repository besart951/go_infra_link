import { ListEntityUseCase } from '$lib/application/useCases/listEntityUseCase.js';
import { apparatRepository } from '$lib/infrastructure/api/apparatRepository.js';
import { systemPartRepository } from '$lib/infrastructure/api/systemPartRepository.js';
import type { Apparat, SystemPart } from '$lib/domain/facility/index.js';
import type { CrudRepository } from '$lib/domain/ports/crudRepository.js';

export interface FieldDeviceStaticLookupResult {
  apparats?: Apparat[];
  systemParts?: SystemPart[];
  apparatsError?: unknown;
  systemPartsError?: unknown;
}

export class FieldDeviceLookupService {
  private static readonly instances = new Set<FieldDeviceLookupService>();

  private readonly listApparatsUseCase: ListEntityUseCase<Apparat>;
  private readonly listSystemPartsUseCase: ListEntityUseCase<SystemPart>;
  private cachedLookups: FieldDeviceStaticLookupResult | null = null;
  private lookupRequest: Promise<FieldDeviceStaticLookupResult> | null = null;

  constructor(
    apparats: CrudRepository<Apparat, unknown, unknown> = apparatRepository,
    systemParts: CrudRepository<SystemPart, unknown, unknown> = systemPartRepository
  ) {
    this.listApparatsUseCase = new ListEntityUseCase(apparats);
    this.listSystemPartsUseCase = new ListEntityUseCase(systemParts);
    FieldDeviceLookupService.instances.add(this);
  }

  async loadStaticLookups(): Promise<FieldDeviceStaticLookupResult> {
    if (this.cachedLookups) {
      return this.cachedLookups;
    }

    if (this.lookupRequest) {
      return this.lookupRequest;
    }

    const request = (async () => {
      const [apparatsResult, systemPartsResult] = await Promise.allSettled([
        this.listApparatsUseCase.execute({
          pagination: { page: 1, pageSize: 1000 },
          search: { text: '' }
        }),
        this.listSystemPartsUseCase.execute({
          pagination: { page: 1, pageSize: 1000 },
          search: { text: '' }
        })
      ]);

      return {
        apparats: apparatsResult.status === 'fulfilled' ? apparatsResult.value.items : undefined,
        systemParts:
          systemPartsResult.status === 'fulfilled' ? systemPartsResult.value.items : undefined,
        apparatsError: apparatsResult.status === 'rejected' ? apparatsResult.reason : undefined,
        systemPartsError:
          systemPartsResult.status === 'rejected' ? systemPartsResult.reason : undefined
      };
    })();

    this.lookupRequest = request;

    request
      .then((result) => {
        this.cachedLookups = result;
        this.lookupRequest = null;
      })
      .catch(() => {
        this.lookupRequest = null;
      });

    return request;
  }

  resetCachedLookups(): void {
    this.cachedLookups = null;
    this.lookupRequest = null;
  }

  static resetAllCachedLookups(): void {
    for (const service of FieldDeviceLookupService.instances) {
      service.resetCachedLookups();
    }
    FieldDeviceLookupService.instances.clear();
  }
}
