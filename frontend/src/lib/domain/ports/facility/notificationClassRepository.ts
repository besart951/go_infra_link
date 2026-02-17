import type { CrudRepository } from '$lib/domain/ports/crudRepository.js';
import type { NotificationClass, CreateNotificationClassRequest, UpdateNotificationClassRequest } from '$lib/domain/facility/index.js';

export interface NotificationClassRepository extends CrudRepository<NotificationClass, CreateNotificationClassRequest, UpdateNotificationClassRequest> {
}
