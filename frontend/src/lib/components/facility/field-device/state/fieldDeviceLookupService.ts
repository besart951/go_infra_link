import { ListEntityUseCase } from '$lib/application/useCases/listEntityUseCase.js';
import { apparatRepository } from '$lib/infrastructure/api/apparatRepository.js';
import { fieldDeviceRepository } from '$lib/infrastructure/api/fieldDeviceRepository.js';
import { systemPartRepository } from '$lib/infrastructure/api/systemPartRepository.js';
import { fetchAllPages } from '$lib/components/facility/shared/paginatedListFetcher.js';
import type { Apparat, FieldDeviceOptions, SystemPart } from '$lib/domain/facility/index.js';
import type { CrudRepository } from '$lib/domain/ports/crudRepository.js';
import type { FieldDeviceRepository } from '$lib/domain/ports/facility/fieldDeviceRepository.js';

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
  private cachedLookupScope: string | undefined;
  private lookupRequest: Promise<FieldDeviceStaticLookupResult> | null = null;
  private lookupRequestScope: string | undefined;

  constructor(
    apparats: CrudRepository<Apparat, unknown, unknown> = apparatRepository,
    systemParts: CrudRepository<SystemPart, unknown, unknown> = systemPartRepository,
    private readonly fieldDevices: Pick<
      FieldDeviceRepository,
      'getOptions' | 'getOptionsForProject'
    > = fieldDeviceRepository
  ) {
    this.listApparatsUseCase = new ListEntityUseCase(apparats);
    this.listSystemPartsUseCase = new ListEntityUseCase(systemParts);
    FieldDeviceLookupService.instances.add(this);
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
      const [optionsResult, apparatsResult, systemPartsResult] = await Promise.allSettled([
        projectId
          ? this.fieldDevices.getOptionsForProject(projectId)
          : this.fieldDevices.getOptions(),
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
      const listedApparats =
        apparatsResult.status === 'fulfilled' && apparatsResult.value.length > 0
          ? apparatsResult.value
          : undefined;
      const listedSystemParts =
        systemPartsResult.status === 'fulfilled' && systemPartsResult.value.length > 0
          ? systemPartsResult.value
          : undefined;
      const systemParts = listedSystemParts ?? options?.system_parts;
      const apparats = withOptionRelations(
        listedApparats ?? options?.apparats,
        systemParts,
        options
      );

      return {
        apparats,
        systemParts,
        apparatsError: apparatsResult.status === 'rejected' ? apparatsResult.reason : undefined,
        systemPartsError:
          systemPartsResult.status === 'rejected' ? systemPartsResult.reason : undefined
      };
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

  static resetAllCachedLookups(): void {
    for (const service of FieldDeviceLookupService.instances) {
      service.resetCachedLookups();
    }
    FieldDeviceLookupService.instances.clear();
  }
}

function withOptionRelations(
  apparats: Apparat[] | undefined,
  systemParts: SystemPart[] | undefined,
  options: FieldDeviceOptions | undefined
): Apparat[] | undefined {
  if (!apparats) return undefined;
  if (!systemParts || !options?.apparat_to_system_part) return apparats;

  const systemPartsById = new Map(systemParts.map((systemPart) => [systemPart.id, systemPart]));

  return apparats.map((apparat) => {
    const relationIds = options.apparat_to_system_part[apparat.id];
    if (!relationIds) return apparat;

    return {
      ...apparat,
      system_parts: relationIds
        .map((systemPartId) => systemPartsById.get(systemPartId))
        .filter((systemPart): systemPart is SystemPart => Boolean(systemPart))
    };
  });
}
