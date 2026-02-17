import type { NotificationClass, CreateNotificationClassRequest, UpdateNotificationClassRequest } from '$lib/domain/facility/index.js';
import type { NotificationClassRepository } from '$lib/domain/ports/facility/notificationClassRepository.js';

export class ManageNotificationClassUseCase {
    constructor(private repository: NotificationClassRepository) { }

    async create(data: CreateNotificationClassRequest, signal?: AbortSignal): Promise<NotificationClass> {
        return this.repository.create(data, signal);
    }

    async update(id: string, data: UpdateNotificationClassRequest, signal?: AbortSignal): Promise<NotificationClass> {
        return this.repository.update(id, data, signal);
    }

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return this.repository.delete(id, signal);
    }
}
