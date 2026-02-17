import type { Building, CreateBuildingRequest, UpdateBuildingRequest } from '$lib/domain/facility/index.js';
import type { BuildingRepository } from '$lib/domain/ports/facility/buildingRepository.js';
import { ManageEntityUseCase } from '$lib/application/useCases/manageEntityUseCase.js';

export class ManageBuildingUseCase extends ManageEntityUseCase<Building, CreateBuildingRequest, UpdateBuildingRequest> {
    constructor(private repo: BuildingRepository) {
        super(repo);
    }

    async validate(data: { id?: string; iws_code: string; building_group: number }, signal?: AbortSignal): Promise<void> {
        return this.repo.validate(data, signal);
    }
}
