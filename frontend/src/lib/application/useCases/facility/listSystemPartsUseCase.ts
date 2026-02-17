import type { SystemPart } from '$lib/domain/facility/index.js';
import type { SystemPartRepository } from '$lib/domain/ports/facility/systemPartRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';

export class ListSystemPartsUseCase {
    constructor(private repository: SystemPartRepository) { }

    async execute(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<SystemPart>> {
        return this.repository.list(params, signal);
    }
}
