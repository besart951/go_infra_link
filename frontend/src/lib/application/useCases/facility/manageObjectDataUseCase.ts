import type { ObjectData, CreateObjectDataRequest, UpdateObjectDataRequest } from '$lib/domain/facility/index.js';
import type { ObjectDataRepository } from '$lib/domain/ports/facility/objectDataRepository.js';
import type { BacnetObject } from '$lib/domain/facility/bacnet-object.js';

export class ManageObjectDataUseCase {
    constructor(private repository: ObjectDataRepository) { }

    async create(data: CreateObjectDataRequest, signal?: AbortSignal): Promise<ObjectData> {
        return this.repository.create(data, signal);
    }

    async update(id: string, data: UpdateObjectDataRequest, signal?: AbortSignal): Promise<ObjectData> {
        return this.repository.update(id, data, signal);
    }

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return this.repository.delete(id, signal);
    }

    async get(id: string, signal?: AbortSignal): Promise<ObjectData> {
        return this.repository.get(id, signal);
    }

    async getBacnetObjects(id: string, signal?: AbortSignal): Promise<BacnetObject[]> {
        return this.repository.getBacnetObjects(id, signal);
    }
}
