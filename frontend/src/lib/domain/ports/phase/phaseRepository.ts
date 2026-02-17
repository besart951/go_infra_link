import type { ListRepository, ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type { Phase, CreatePhaseRequest, UpdatePhaseRequest } from '$lib/domain/phase/index.js';

export interface PhaseRepository extends ListRepository<Phase> {
    list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<Phase>>;
    get(id: string, signal?: AbortSignal): Promise<Phase>;
    create(data: CreatePhaseRequest, signal?: AbortSignal): Promise<Phase>;
    update(id: string, data: UpdatePhaseRequest, signal?: AbortSignal): Promise<Phase>;
    delete(id: string, signal?: AbortSignal): Promise<void>;
}
