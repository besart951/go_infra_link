import type { NotificationClassRepository } from '$lib/domain/ports/facility/notificationClassRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
  NotificationClass,
  NotificationClassListResponse,
  CreateNotificationClassRequest,
  UpdateNotificationClassRequest
} from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';
import { buildListUrl, mapPaginatedResponse } from './listHelpers.js';

export const notificationClassRepository: NotificationClassRepository = {
  async list(
    params: ListParams,
    signal?: AbortSignal
  ): Promise<PaginatedResponse<NotificationClass>> {
    const response = await api<NotificationClassListResponse>(
      buildListUrl('/facility/notification-classes', params),
      { signal }
    );

    return mapPaginatedResponse(response, params);
  },

  async get(id: string, signal?: AbortSignal): Promise<NotificationClass> {
    return api<NotificationClass>(`/facility/notification-classes/${id}`, { signal });
  },

  async create(
    data: CreateNotificationClassRequest,
    signal?: AbortSignal
  ): Promise<NotificationClass> {
    return api<NotificationClass>('/facility/notification-classes', {
      method: 'POST',
      body: JSON.stringify(data),
      signal
    });
  },

  async update(
    id: string,
    data: UpdateNotificationClassRequest,
    signal?: AbortSignal
  ): Promise<NotificationClass> {
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
