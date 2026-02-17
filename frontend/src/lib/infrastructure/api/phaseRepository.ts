import type { PhaseRepository } from '$lib/domain/ports/phase/phaseRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
    Phase,
    PhaseListResponse,
    CreatePhaseRequest,
    UpdatePhaseRequest
} from '$lib/domain/phase/index.js';
import { api } from '$lib/api/client.js';

export const phaseRepository: PhaseRepository = {
    async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<Phase>> {
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
        const response = await api<PhaseListResponse>(
            `/phases${query ? `?${query}` : ''}`,
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

    async get(id: string, signal?: AbortSignal): Promise<Phase> {
        return api<Phase>(`/phases/${id}`, { signal });
    },

    async create(data: CreatePhaseRequest, signal?: AbortSignal): Promise<Phase> {
        return api<Phase>('/phases', {
            method: 'POST',
            body: JSON.stringify(data),
            signal
        });
    },

    async update(id: string, data: UpdatePhaseRequest, signal?: AbortSignal): Promise<Phase> {
        return api<Phase>(`/phases/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
            signal
        });
    },

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return api<void>(`/phases/${id}`, {
            method: 'DELETE',
            signal
        });
    }
};
