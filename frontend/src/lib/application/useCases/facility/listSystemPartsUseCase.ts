import type { SystemPartListParams, SystemPartListResponse } from '$lib/domain/facility/system.js';
import type { SystemPartRepository } from '$lib/domain/ports/facility/systemPartRepository.js';

export class ListSystemPartsUseCase {
    constructor(private repository: SystemPartRepository) { }

    async execute(params?: SystemPartListParams, signal?: AbortSignal): Promise<SystemPartListResponse> {
        return this.repository.list(params, signal);
    }
}
