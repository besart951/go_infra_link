import type { SPSControllerRepository } from '$lib/domain/ports/facility/spsControllerRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
    SPSController,
    SPSControllerListResponse,
    SPSControllerBulkResponse,
    CreateSPSControllerRequest,
    UpdateSPSControllerRequest,
    NextGADeviceResponse,
    SPSControllerSystemType,
    SPSControllerSystemTypeListParams,
    SPSControllerSystemTypeListResponse
} from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';

export const spsControllerRepository: SPSControllerRepository = {
    async list(
        params: ListParams,
        signal?: AbortSignal
    ): Promise<PaginatedResponse<SPSController>> {
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
        const response = await api<SPSControllerListResponse>(
            `/facility/sps-controllers${query ? `?${query}` : ''}`,
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

    async get(id: string, signal?: AbortSignal): Promise<SPSController> {
        return api<SPSController>(`/facility/sps-controllers/${id}`, { signal });
    },

    async getBulk(ids: string[], signal?: AbortSignal): Promise<SPSController[]> {
        const response = await api<SPSControllerBulkResponse>('/facility/sps-controllers/bulk', {
            method: 'POST',
            body: JSON.stringify({ ids }),
            signal
        });
        return response.items;
    },

    async create(
        data: CreateSPSControllerRequest,
        signal?: AbortSignal
    ): Promise<SPSController> {
        return api<SPSController>('/facility/sps-controllers', {
            method: 'POST',
            body: JSON.stringify(data),
            signal
        });
    },

    async update(
        id: string,
        data: UpdateSPSControllerRequest,
        signal?: AbortSignal
    ): Promise<SPSController> {
        return api<SPSController>(`/facility/sps-controllers/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
            signal
        });
    },

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return api<void>(`/facility/sps-controllers/${id}`, {
            method: 'DELETE',
            signal
        });
    },

    async validate(
        data: {
            id?: string;
            control_cabinet_id: string;
            ga_device?: string;
            device_name: string;
            ip_address?: string;
            subnet?: string;
            gateway?: string;
            vlan?: string;
        },
        signal?: AbortSignal
    ): Promise<void> {
        return api<void>('/facility/sps-controllers/validate', {
            method: 'POST',
            body: JSON.stringify(data),
            signal
        });
    },

    async getNextGADevice(
        controlCabinetId: string,
        spsControllerId?: string,
        signal?: AbortSignal
    ): Promise<NextGADeviceResponse> {
        const searchParams = new URLSearchParams();
        searchParams.set('control_cabinet_id', controlCabinetId);
        if (spsControllerId) {
            searchParams.set('sps_controller_id', spsControllerId);
        }
        return api<NextGADeviceResponse>(
            `/facility/sps-controllers/next-ga-device?${searchParams.toString()}`,
            { signal }
        );
    },

    async listSystemTypes(
        params?: SPSControllerSystemTypeListParams,
        signal?: AbortSignal
    ): Promise<SPSControllerSystemTypeListResponse> {
        const searchParams = new URLSearchParams();
        if (params?.page) searchParams.set('page', String(params.page));
        if (params?.limit) searchParams.set('limit', String(params.limit));
        if (params?.search) searchParams.set('search', params.search);
        if (params?.sps_controller_id)
            searchParams.set('sps_controller_id', params.sps_controller_id);

        const query = searchParams.toString();
        return api<SPSControllerSystemTypeListResponse>(
            `/facility/sps-controller-system-types${query ? `?${query}` : ''}`,
            { signal }
        );
    },

    async getSystemType(id: string, signal?: AbortSignal): Promise<SPSControllerSystemType> {
        return api<SPSControllerSystemType>(`/facility/sps-controller-system-types/${id}`, {
            signal
        });
    }
};
