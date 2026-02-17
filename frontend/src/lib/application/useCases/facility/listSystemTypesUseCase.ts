import type { SystemType } from '$lib/domain/facility/index.js';
import type { SystemTypeRepository } from '$lib/domain/ports/facility/systemTypeRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';

export class ListSystemTypesUseCase {
    constructor(private repository: SystemTypeRepository) { }

    async execute(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<SystemType>> {
        return this.repository.list(params, signal);
    }

    async get(id: string, signal?: AbortSignal): Promise<SystemType> {
        return this.repository.get(id, signal);
    }
}
