import type { SystemTypeRepository } from '$lib/domain/ports/facility/systemTypeRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type { SystemType, SystemTypeListResponse, CreateSystemTypeRequest, UpdateSystemTypeRequest } from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';

export const systemTypeRepository: SystemTypeRepository = {
    async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<SystemType>> {
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
        const response = await api<SystemTypeListResponse>(`/facility/system-types${query ? `?${query}` : ''}`, { signal });

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

    async get(id: string, signal?: AbortSignal): Promise<SystemType> {
        return api<SystemType>(`/facility/system-types/${id}`, { signal });
    },

    async create(data: CreateSystemTypeRequest, signal?: AbortSignal): Promise<SystemType> {
        return api<SystemType>('/facility/system-types', {
            method: 'POST',
            body: JSON.stringify(data),
            signal
        });
    },

    async update(id: string, data: UpdateSystemTypeRequest, signal?: AbortSignal): Promise<SystemType> {
        return api<SystemType>(`/facility/system-types/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
            signal
        });
    },

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return api<void>(`/facility/system-types/${id}`, {
            method: 'DELETE',
            signal
        });
    }
};
