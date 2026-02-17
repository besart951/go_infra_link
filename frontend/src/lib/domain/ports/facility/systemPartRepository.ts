import type { ListRepository, ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type { SystemPart, CreateSystemPartRequest, UpdateSystemPartRequest } from '$lib/domain/facility/index.js';

export interface SystemPartRepository extends ListRepository<SystemPart> {
    list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<SystemPart>>;
    get(id: string, signal?: AbortSignal): Promise<SystemPart>;
    create(data: CreateSystemPartRequest, signal?: AbortSignal): Promise<SystemPart>;
    update(id: string, data: UpdateSystemPartRequest, signal?: AbortSignal): Promise<SystemPart>;
    delete(id: string, signal?: AbortSignal): Promise<void>;
}
