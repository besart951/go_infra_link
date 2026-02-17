import type { Apparat } from '$lib/domain/facility/index.js';
import type { ApparatRepository } from '$lib/domain/ports/facility/apparatRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';

export class ListApparatsUseCase {
    constructor(private repository: ApparatRepository) { }

    async execute(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<Apparat>> {
        return this.repository.list(params, signal);
    }


    async get(id: string, signal?: AbortSignal): Promise<Apparat> {
        const items = await this.repository.getBulk([id], signal);
        return items[0];
    }

    async getBulk(ids: string[], signal?: AbortSignal): Promise<Apparat[]> {
        return this.repository.getBulk(ids, signal);
    }
}
