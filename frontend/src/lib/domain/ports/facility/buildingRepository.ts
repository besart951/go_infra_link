import type { ListRepository, ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type { Building, CreateBuildingRequest, UpdateBuildingRequest } from '$lib/domain/facility/index.js';

export interface BuildingRepository extends ListRepository<Building> {
    list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<Building>>;
    get(id: string, signal?: AbortSignal): Promise<Building>;
    getBulk(ids: string[], signal?: AbortSignal): Promise<Building[]>;
    create(data: CreateBuildingRequest, signal?: AbortSignal): Promise<Building>;
    update(id: string, data: UpdateBuildingRequest, signal?: AbortSignal): Promise<Building>;
    delete(id: string, signal?: AbortSignal): Promise<void>;
    validate(data: { id?: string; iws_code: string; building_group: number }, signal?: AbortSignal): Promise<void>;
}
