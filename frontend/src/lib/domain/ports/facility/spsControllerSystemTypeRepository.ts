import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type { SPSControllerSystemType } from '$lib/domain/facility/index.js';

export interface SPSControllerSystemTypeRepository {
    list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<SPSControllerSystemType>>;
    get(id: string, signal?: AbortSignal): Promise<SPSControllerSystemType>;
    copy(id: string, signal?: AbortSignal): Promise<SPSControllerSystemType>;
}
