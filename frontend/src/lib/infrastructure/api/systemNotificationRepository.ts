import { api } from '$lib/api/client.js';
import type {
  ListSystemNotificationsParams,
  SystemNotification,
  SystemNotificationList
} from '$lib/domain/notification/index.js';
import type { SystemNotificationRepository } from '$lib/domain/ports/notification/systemNotificationRepository.js';

function buildQuery(params?: ListSystemNotificationsParams): string {
  const query = new URLSearchParams();
  if (params?.page) query.set('page', String(params.page));
  if (params?.limit) query.set('limit', String(params.limit));
  if (params?.unread_only) query.set('unread_only', 'true');
  const text = query.toString();
  return text ? `?${text}` : '';
}

export const systemNotificationRepository: SystemNotificationRepository = {
  list(
    params?: ListSystemNotificationsParams,
    signal?: AbortSignal
  ): Promise<SystemNotificationList> {
    return api<SystemNotificationList>(`/account/notifications${buildQuery(params)}`, { signal });
  },

  markRead(id: string, signal?: AbortSignal): Promise<SystemNotification> {
    return api<SystemNotification>(`/account/notifications/${id}/read`, {
      method: 'POST',
      signal
    });
  },

  markAllRead(signal?: AbortSignal): Promise<void> {
    return api<void>('/account/notifications/read-all', {
      method: 'POST',
      signal
    });
  }
};
