import type { TeamRepository } from '$lib/domain/ports/team/teamRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
    Team,
    TeamListResponse,
    CreateTeamRequest,
    UpdateTeamRequest
} from '$lib/domain/team/index.js';
import { api } from '$lib/api/client.js';

export const teamRepository: TeamRepository = {
    async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<Team>> {
        const searchParams = new URLSearchParams();
        searchParams.set('page', String(params.pagination.page));
        searchParams.set('limit', String(params.pagination.pageSize));
        if (params.search.text) searchParams.set('search', params.search.text);

        if (params.filters) {
            Object.entries(params.filters).forEach(([key, value]) => {
                if (value !== undefined && value !== null) searchParams.set(key, value);
            });
        }

        const query = searchParams.toString();
        const response = await api<TeamListResponse>(
            `/teams${query ? `?${query}` : ''}`,
            { signal }
        );

        return {
            items: response.items,
            metadata: {
                total: response.total,
                page: response.page,
                pageSize: params.pagination.pageSize,
                totalPages: response.total_pages
            }
        };
    },

    async get(id: string, signal?: AbortSignal): Promise<Team> {
        return api<Team>(`/teams/${id}`, { signal });
    },

    async create(data: CreateTeamRequest, signal?: AbortSignal): Promise<Team> {
        return api<Team>('/teams', {
            method: 'POST',
            body: JSON.stringify(data),
            signal
        });
    },

    async update(id: string, data: UpdateTeamRequest, signal?: AbortSignal): Promise<Team> {
        return api<Team>(`/teams/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
            signal
        });
    },

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return api<void>(`/teams/${id}`, {
            method: 'DELETE',
            signal
        });
    }
};
