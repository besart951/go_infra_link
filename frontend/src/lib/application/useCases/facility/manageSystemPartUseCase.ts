import type { SystemPart, CreateSystemPartRequest, UpdateSystemPartRequest } from '$lib/domain/facility/index.js';
import type { SystemPartRepository } from '$lib/domain/ports/facility/systemPartRepository.js';

export class ManageSystemPartUseCase {
    constructor(private repository: SystemPartRepository) { }

    async create(data: CreateSystemPartRequest, signal?: AbortSignal): Promise<SystemPart> {
        return this.repository.create(data, signal);
    }

    async update(id: string, data: UpdateSystemPartRequest, signal?: AbortSignal): Promise<SystemPart> {
        return this.repository.update(id, data, signal);
    }

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return this.repository.delete(id, signal);
    }
}
