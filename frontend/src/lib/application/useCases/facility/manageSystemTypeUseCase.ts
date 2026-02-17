import type { SystemType, CreateSystemTypeRequest, UpdateSystemTypeRequest } from '$lib/domain/facility/index.js';
import type { SystemTypeRepository } from '$lib/domain/ports/facility/systemTypeRepository.js';

export class ManageSystemTypeUseCase {
    constructor(private repository: SystemTypeRepository) { }

    async create(data: CreateSystemTypeRequest, signal?: AbortSignal): Promise<SystemType> {
        return this.repository.create(data, signal);
    }

    async update(id: string, data: UpdateSystemTypeRequest, signal?: AbortSignal): Promise<SystemType> {
        return this.repository.update(id, data, signal);
    }

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return this.repository.delete(id, signal);
    }
}
