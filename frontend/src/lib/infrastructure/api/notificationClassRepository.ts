import type { NotificationClassRepository } from '$lib/domain/ports/facility/notificationClassRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type { NotificationClass, NotificationClassListResponse, CreateNotificationClassRequest, UpdateNotificationClassRequest } from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';

export const notificationClassRepository: NotificationClassRepository = {
    async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<NotificationClass>> {
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
        const response = await api<NotificationClassListResponse>(
            `/facility/notification-classes${query ? `?${query}` : ''}`,
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

    async get(id: string, signal?: AbortSignal): Promise<NotificationClass> {
        return api<NotificationClass>(`/facility/notification-classes/${id}`, { signal });
    },

    async create(data: CreateNotificationClassRequest, signal?: AbortSignal): Promise<NotificationClass> {
        return api<NotificationClass>('/facility/notification-classes', {
            method: 'POST',
            body: JSON.stringify(data),
            signal
        });
    },

    async update(id: string, data: UpdateNotificationClassRequest, signal?: AbortSignal): Promise<NotificationClass> {
        return api<NotificationClass>(`/facility/notification-classes/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
            signal
        });
    },

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return api<void>(`/facility/notification-classes/${id}`, {
            method: 'DELETE',
            signal
        });
    }
};
