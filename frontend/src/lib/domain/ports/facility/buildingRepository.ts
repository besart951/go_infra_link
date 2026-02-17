import type { CrudRepository } from '$lib/domain/ports/crudRepository.js';
import type { Building, CreateBuildingRequest, UpdateBuildingRequest } from '$lib/domain/facility/index.js';

export interface BuildingRepository extends CrudRepository<Building, CreateBuildingRequest, UpdateBuildingRequest> {
    getBulk(ids: string[], signal?: AbortSignal): Promise<Building[]>;
    validate(data: { id?: string; iws_code: string; building_group: number }, signal?: AbortSignal): Promise<void>;
}
