import type { SPSControllerSystemTypeRepository } from '$lib/domain/ports/facility/spsControllerSystemTypeRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type { SPSControllerSystemType, SPSControllerSystemTypeListResponse } from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';

export const spsControllerSystemTypeRepository: SPSControllerSystemTypeRepository = {
    async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<SPSControllerSystemType>> {
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
        const response = await api<SPSControllerSystemTypeListResponse>(
            `/facility/sps-controller-system-types${query ? `?${query}` : ''}`,
            { signal }
        );

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

    async get(id: string, signal?: AbortSignal): Promise<SPSControllerSystemType> {
        return api<SPSControllerSystemType>(`/facility/sps-controller-system-types/${id}`, { signal });
    }
};
