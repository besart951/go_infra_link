import type { ListRepository, ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type { StateText, CreateStateTextRequest, UpdateStateTextRequest } from '$lib/domain/facility/index.js';

export interface StateTextRepository extends ListRepository<StateText> {
    list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<StateText>>;
    get(id: string, signal?: AbortSignal): Promise<StateText>;
    create(data: CreateStateTextRequest, signal?: AbortSignal): Promise<StateText>;
    update(id: string, data: UpdateStateTextRequest, signal?: AbortSignal): Promise<StateText>;
    delete(id: string, signal?: AbortSignal): Promise<void>;
}
