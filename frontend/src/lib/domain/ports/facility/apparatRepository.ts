import type { ListRepository, ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type { Apparat, CreateApparatRequest, UpdateApparatRequest } from '$lib/domain/facility/index.js';

export interface ApparatRepository extends ListRepository<Apparat> {
    list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<Apparat>>;
    get(id: string, signal?: AbortSignal): Promise<Apparat>;
    getBulk(ids: string[], signal?: AbortSignal): Promise<Apparat[]>;
    create(data: CreateApparatRequest, signal?: AbortSignal): Promise<Apparat>;
    update(id: string, data: UpdateApparatRequest, signal?: AbortSignal): Promise<Apparat>;
    delete(id: string, signal?: AbortSignal): Promise<void>;
}
