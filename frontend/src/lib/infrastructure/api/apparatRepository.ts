import type { ApparatRepository } from '$lib/domain/ports/facility/apparatRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type { Apparat, ApparatListResponse, ApparatBulkResponse, CreateApparatRequest, UpdateApparatRequest } from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';

export const apparatRepository: ApparatRepository = {
    async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<Apparat>> {
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
        const response = await api<ApparatListResponse>(`/facility/apparats${query ? `?${query}` : ''}`, { signal });

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

    async get(id: string, signal?: AbortSignal): Promise<Apparat> {
        return api<Apparat>(`/facility/apparats/${id}`, { signal });
    },

    async getBulk(ids: string[], signal?: AbortSignal): Promise<Apparat[]> {
        const response = await api<ApparatBulkResponse>('/facility/apparats/bulk', {
            method: 'POST',
            body: JSON.stringify({ ids }),
            signal
        });
        return response.items;
    },

    async create(data: import('$lib/domain/facility/index.js').CreateApparatRequest, signal?: AbortSignal): Promise<Apparat> {
        return api<Apparat>('/facility/apparats', {
            method: 'POST',
            body: JSON.stringify(data),
            signal
        });
    },

    async update(id: string, data: import('$lib/domain/facility/index.js').UpdateApparatRequest, signal?: AbortSignal): Promise<Apparat> {
        return api<Apparat>(`/facility/apparats/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
            signal
        });
    },

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return api<void>(`/facility/apparats/${id}`, {
            method: 'DELETE',
            signal
        });
    }
};
