import type { ListRepository, ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type { NotificationClass, CreateNotificationClassRequest, UpdateNotificationClassRequest } from '$lib/domain/facility/index.js';

export interface NotificationClassRepository extends ListRepository<NotificationClass> {
    list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<NotificationClass>>;
    get(id: string, signal?: AbortSignal): Promise<NotificationClass>;
    create(data: CreateNotificationClassRequest, signal?: AbortSignal): Promise<NotificationClass>;
    update(id: string, data: UpdateNotificationClassRequest, signal?: AbortSignal): Promise<NotificationClass>;
    delete(id: string, signal?: AbortSignal): Promise<void>;
}
