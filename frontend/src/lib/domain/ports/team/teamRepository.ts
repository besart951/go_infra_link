import type { ListRepository, ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type { Team, CreateTeamRequest, UpdateTeamRequest } from '$lib/domain/team/index.js';

export interface TeamRepository extends ListRepository<Team> {
    list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<Team>>;
    get(id: string, signal?: AbortSignal): Promise<Team>;
    create(data: CreateTeamRequest, signal?: AbortSignal): Promise<Team>;
    update(id: string, data: UpdateTeamRequest, signal?: AbortSignal): Promise<Team>;
    delete(id: string, signal?: AbortSignal): Promise<void>;
}
