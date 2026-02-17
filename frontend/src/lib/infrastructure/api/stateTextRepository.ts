import type { StateTextRepository } from '$lib/domain/ports/facility/stateTextRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type { StateText, StateTextListResponse, CreateStateTextRequest, UpdateStateTextRequest } from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';

export const stateTextRepository: StateTextRepository = {
    async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<StateText>> {
        const searchParams = new URLSearchParams();
        searchParams.set('page', String(params.pagination.page));
        searchParams.set('limit', String(params.pagination.pageSize));
        if (params.search.text) searchParams.set('search', params.search.text);

        if (params.filters) {
            Object.entries(params.filters).forEach(([key, value]) => {
                if (value !== undefined && value !== null) searchParams.set(key, String(value));
            });
        }

        const query = searchParams.toString();
        const response = await api<StateTextListResponse>(`/facility/state-texts${query ? `?${query}` : ''}`, { signal });

        return {
            items: response.items,
            metadata: {
                total: response.total,
                page: response.page,
                pageSize: response.limit,
                totalPages: response.total_pages
            }
        };
    },

    async get(id: string, signal?: AbortSignal): Promise<StateText> {
        return api<StateText>(`/facility/state-texts/${id}`, { signal });
    },

    async create(data: CreateStateTextRequest, signal?: AbortSignal): Promise<StateText> {
        return api<StateText>('/facility/state-texts', {
            method: 'POST',
            body: JSON.stringify(data),
            signal
        });
    },

    async update(id: string, data: UpdateStateTextRequest, signal?: AbortSignal): Promise<StateText> {
        return api<StateText>(`/facility/state-texts/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
            signal
        });
    },

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return api<void>(`/facility/state-texts/${id}`, {
            method: 'DELETE',
            signal
        });
    }
};
