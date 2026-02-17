import type { ObjectDataRepository } from '$lib/domain/ports/facility/objectDataRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type { ObjectData, ObjectDataListResponse, CreateObjectDataRequest, UpdateObjectDataRequest } from '$lib/domain/facility/index.js';
import type { BacnetObject } from '$lib/domain/facility/bacnet-object.js';
import { api } from '$lib/api/client.js';

export const objectDataRepository: ObjectDataRepository = {
    async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<ObjectData>> {
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
        const response = await api<ObjectDataListResponse>(`/facility/object-data${query ? `?${query}` : ''}`, { signal });

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

    async get(id: string, signal?: AbortSignal): Promise<ObjectData> {
        return api<ObjectData>(`/facility/object-data/${id}`, { signal });
    },

    async create(data: CreateObjectDataRequest, signal?: AbortSignal): Promise<ObjectData> {
        return api<ObjectData>('/facility/object-data', {
            method: 'POST',
            body: JSON.stringify(data),
            signal
        });
    },

    async update(id: string, data: UpdateObjectDataRequest, signal?: AbortSignal): Promise<ObjectData> {
        return api<ObjectData>(`/facility/object-data/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
            signal
        });
    },

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return api<void>(`/facility/object-data/${id}`, {
            method: 'DELETE',
            signal
        });
    },

    async getBacnetObjects(id: string, signal?: AbortSignal): Promise<BacnetObject[]> {
        return api<BacnetObject[]>(`/facility/object-data/${id}/bacnet-objects`, { signal });
    }
};
