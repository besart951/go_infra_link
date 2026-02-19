import type { AlarmTypeRepository } from '$lib/domain/ports/facility/alarmTypeRepository.js';
import type { AlarmType, AlarmTypeListResponse } from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';

export const alarmTypeRepository: AlarmTypeRepository = {
    async list(params = {}, signal?: AbortSignal) {
        const searchParams = new URLSearchParams();
        if (params.page) searchParams.set('page', String(params.page));
        if (params.pageSize) searchParams.set('limit', String(params.pageSize));
        if (params.search) searchParams.set('search', params.search);
        const query = searchParams.toString();
        const response = await api<AlarmTypeListResponse>(
            `/facility/alarm-types${query ? `?${query}` : ''}`,
            { signal }
        );
        return {
            items: response.items,
            total: response.total,
            page: response.page,
            totalPages: response.total_pages
        };
    },

    async get(id: string, signal?: AbortSignal): Promise<AlarmType> {
        return api<AlarmType>(`/facility/alarm-types/${id}`, { signal });
    },

    async getWithFields(id: string, signal?: AbortSignal): Promise<AlarmType> {
        return api<AlarmType>(`/facility/alarm-types/${id}/fields`, { signal });
    }
};
