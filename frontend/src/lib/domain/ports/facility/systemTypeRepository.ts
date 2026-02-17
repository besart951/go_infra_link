import type { ListRepository, ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type { SystemType, CreateSystemTypeRequest, UpdateSystemTypeRequest } from '$lib/domain/facility/index.js';

export interface SystemTypeRepository extends ListRepository<SystemType> {
    list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<SystemType>>;
    get(id: string, signal?: AbortSignal): Promise<SystemType>;
    create(data: CreateSystemTypeRequest, signal?: AbortSignal): Promise<SystemType>;
    update(id: string, data: UpdateSystemTypeRequest, signal?: AbortSignal): Promise<SystemType>;
    delete(id: string, signal?: AbortSignal): Promise<void>;
}
