import type { ControlCabinetRepository } from '$lib/domain/ports/facility/controlCabinetRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
    ControlCabinet,
    ControlCabinetListResponse,
    ControlCabinetBulkResponse,
    CreateControlCabinetRequest,
    UpdateControlCabinetRequest,
    ControlCabinetDeleteImpact
} from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';

export const controlCabinetRepository: ControlCabinetRepository = {
    async list(
        params: ListParams,
        signal?: AbortSignal
    ): Promise<PaginatedResponse<ControlCabinet>> {
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
        const response = await api<ControlCabinetListResponse>(
            `/facility/control-cabinets${query ? `?${query}` : ''}`,
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

    async get(id: string, signal?: AbortSignal): Promise<ControlCabinet> {
        return api<ControlCabinet>(`/facility/control-cabinets/${id}`, { signal });
    },

    async getBulk(ids: string[], signal?: AbortSignal): Promise<ControlCabinet[]> {
        const response = await api<ControlCabinetBulkResponse>('/facility/control-cabinets/bulk', {
            method: 'POST',
            body: JSON.stringify({ ids }),
            signal
        });
        return response.items;
    },

    async create(
        data: CreateControlCabinetRequest,
        signal?: AbortSignal
    ): Promise<ControlCabinet> {
        return api<ControlCabinet>('/facility/control-cabinets', {
            method: 'POST',
            body: JSON.stringify(data),
            signal
        });
    },

    async update(
        id: string,
        data: UpdateControlCabinetRequest,
        signal?: AbortSignal
    ): Promise<ControlCabinet> {
        return api<ControlCabinet>(`/facility/control-cabinets/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
            signal
        });
    },

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return api<void>(`/facility/control-cabinets/${id}`, {
            method: 'DELETE',
            signal
        });
    },

    async validate(
        data: { id?: string; building_id: string; control_cabinet_nr?: string },
        signal?: AbortSignal
    ): Promise<void> {
        return api<void>('/facility/control-cabinets/validate', {
            method: 'POST',
            body: JSON.stringify(data),
            signal
        });
    },

    async getDeleteImpact(
        id: string,
        signal?: AbortSignal
    ): Promise<ControlCabinetDeleteImpact> {
        return api<ControlCabinetDeleteImpact>(
            `/facility/control-cabinets/${id}/delete-impact`,
            { signal }
        );
    }
};
