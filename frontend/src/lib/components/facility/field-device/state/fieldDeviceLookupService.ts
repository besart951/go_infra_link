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
  private readonly listApparatsUseCase: ListEntityUseCase<Apparat>;
  private readonly listSystemPartsUseCase: ListEntityUseCase<SystemPart>;

  constructor(
    apparats: CrudRepository<Apparat, unknown, unknown> = apparatRepository,
    systemParts: CrudRepository<SystemPart, unknown, unknown> = systemPartRepository
  ) {
    this.listApparatsUseCase = new ListEntityUseCase(apparats);
    this.listSystemPartsUseCase = new ListEntityUseCase(systemParts);
  }

  async loadStaticLookups(): Promise<FieldDeviceStaticLookupResult> {
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
  }
}
