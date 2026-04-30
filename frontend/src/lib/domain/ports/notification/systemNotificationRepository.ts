import type {
  ListSystemNotificationsParams,
  SystemNotification,
  SystemNotificationList
} from '$lib/domain/notification/index.js';

export interface SystemNotificationRepository {
  list(
    params?: ListSystemNotificationsParams,
    signal?: AbortSignal
  ): Promise<SystemNotificationList>;
  markRead(id: string, signal?: AbortSignal): Promise<SystemNotification>;
  markAllRead(signal?: AbortSignal): Promise<void>;
}
