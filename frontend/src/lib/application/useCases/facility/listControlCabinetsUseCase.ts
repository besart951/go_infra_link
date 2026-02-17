import type { ControlCabinet } from '$lib/domain/facility/index.js';
import type { ControlCabinetRepository } from '$lib/domain/ports/facility/controlCabinetRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';

export class ListControlCabinetsUseCase {
    constructor(private repository: ControlCabinetRepository) { }

    async execute(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<ControlCabinet>> {
        return this.repository.list(params, signal);
    }

    async get(id: string, signal?: AbortSignal): Promise<ControlCabinet> {
        const items = await this.repository.getBulk([id], signal);
        return items[0];
    }

    async getBulk(ids: string[], signal?: AbortSignal): Promise<ControlCabinet[]> {
        return this.repository.getBulk(ids, signal);
    }
}
