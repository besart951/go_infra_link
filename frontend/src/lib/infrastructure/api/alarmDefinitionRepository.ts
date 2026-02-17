import type { AlarmDefinitionRepository } from '$lib/domain/ports/facility/alarmDefinitionRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type { AlarmDefinition, AlarmDefinitionListResponse, CreateAlarmDefinitionRequest, UpdateAlarmDefinitionRequest } from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';

export const alarmDefinitionRepository: AlarmDefinitionRepository = {
    async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<AlarmDefinition>> {
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
        const response = await api<AlarmDefinitionListResponse>(
            `/facility/alarm-definitions${query ? `?${query}` : ''}`,
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

    async get(id: string, signal?: AbortSignal): Promise<AlarmDefinition> {
        return api<AlarmDefinition>(`/facility/alarm-definitions/${id}`, { signal });
    },

    async create(data: CreateAlarmDefinitionRequest, signal?: AbortSignal): Promise<AlarmDefinition> {
        return api<AlarmDefinition>('/facility/alarm-definitions', {
            method: 'POST',
            body: JSON.stringify(data),
            signal
        });
    },

    async update(id: string, data: UpdateAlarmDefinitionRequest, signal?: AbortSignal): Promise<AlarmDefinition> {
        return api<AlarmDefinition>(`/facility/alarm-definitions/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
            signal
        });
    },

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return api<void>(`/facility/alarm-definitions/${id}`, {
            method: 'DELETE',
            signal
        });
    }
};
