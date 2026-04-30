import { ManageObjectDataUseCase } from '$lib/application/useCases/facility/manageObjectDataUseCase.js';
import { objectDataRepository } from '$lib/infrastructure/api/objectDataRepository.js';
import type { ObjectData } from '$lib/domain/facility/index.js';

export class MultiCreateObjectDataPreviewService {
  private readonly manageObjectDataUseCase = new ManageObjectDataUseCase(objectDataRepository);

  fetch(objectDataId: string, signal?: AbortSignal): Promise<ObjectData> {
    return this.manageObjectDataUseCase.get(objectDataId, signal);
  }
}
