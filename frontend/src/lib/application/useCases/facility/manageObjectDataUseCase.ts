import type { ObjectData, CreateObjectDataRequest, UpdateObjectDataRequest } from '$lib/domain/facility/index.js';
import type { ObjectDataRepository } from '$lib/domain/ports/facility/objectDataRepository.js';
import type { BacnetObject } from '$lib/domain/facility/bacnet-object.js';
import { ManageEntityUseCase } from '$lib/application/useCases/manageEntityUseCase.js';

export class ManageObjectDataUseCase extends ManageEntityUseCase<ObjectData, CreateObjectDataRequest, UpdateObjectDataRequest> {
    constructor(private repo: ObjectDataRepository) {
        super(repo);
    }

    async getBacnetObjects(id: string, signal?: AbortSignal): Promise<BacnetObject[]> {
        return this.repo.getBacnetObjects(id, signal);
    }
}
