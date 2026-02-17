import type { Building } from '$lib/domain/facility/index.js';
import type { BuildingRepository } from '$lib/domain/ports/facility/buildingRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';

export class ListBuildingsUseCase {
    constructor(private repository: BuildingRepository) { }

    async execute(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<Building>> {
        return this.repository.list(params, signal);
    }

    async get(id: string, signal?: AbortSignal): Promise<Building> {
        const items = await this.repository.getBulk([id], signal);
        return items[0];
    }

    async getBulk(ids: string[], signal?: AbortSignal): Promise<Building[]> {
        return this.repository.getBulk(ids, signal);
    }
}
