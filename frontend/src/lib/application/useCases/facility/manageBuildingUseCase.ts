import type { Building, CreateBuildingRequest, UpdateBuildingRequest } from '$lib/domain/facility/index.js';
import type { BuildingRepository } from '$lib/domain/ports/facility/buildingRepository.js';

export class ManageBuildingUseCase {
    constructor(private repository: BuildingRepository) { }

    async create(data: CreateBuildingRequest, signal?: AbortSignal): Promise<Building> {
        return this.repository.create(data, signal);
    }

    async update(id: string, data: UpdateBuildingRequest, signal?: AbortSignal): Promise<Building> {
        return this.repository.update(id, data, signal);
    }

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return this.repository.delete(id, signal);
    }

    async validate(data: { id?: string; iws_code: string; building_group: number }, signal?: AbortSignal): Promise<void> {
        return this.repository.validate(data, signal);
    }
}
