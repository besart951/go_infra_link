import type { SystemPartRepository } from '$lib/domain/ports/facility/systemPartRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
    SystemPart,
    CreateSystemPartRequest,
    UpdateSystemPartRequest
} from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';

export const systemPartRepository: SystemPartRepository = {
    async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<SystemPart>> {
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
        const response = await api<any>(`/facility/system-parts${query ? `?${query}` : ''}`, { signal });

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

    async get(id: string, signal?: AbortSignal): Promise<SystemPart> {
        return api<SystemPart>(`/facility/system-parts/${id}`, { signal });
    },

    async create(data: CreateSystemPartRequest, signal?: AbortSignal): Promise<SystemPart> {
        return api<SystemPart>('/facility/system-parts', {
            method: 'POST',
            body: JSON.stringify(data),
            signal
        });
    },

    async update(id: string, data: UpdateSystemPartRequest, signal?: AbortSignal): Promise<SystemPart> {
        return api<SystemPart>(`/facility/system-parts/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
            signal
        });
    },

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return api<void>(`/facility/system-parts/${id}`, {
            method: 'DELETE',
            signal
        });
    }
};
