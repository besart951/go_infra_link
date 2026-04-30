import { api } from '$lib/api/client.js';
import type {
  NotificationRule,
  NotificationRuleList,
  UpsertNotificationRuleRequest
} from '$lib/domain/notification/index.js';
import type { NotificationRuleRepository } from '$lib/domain/ports/notification/notificationRuleRepository.js';

export const notificationRuleRepository: NotificationRuleRepository = {
  list(signal?: AbortSignal): Promise<NotificationRuleList> {
    return api<NotificationRuleList>('/admin/notifications/rules', { signal });
  },

  create(data: UpsertNotificationRuleRequest, signal?: AbortSignal): Promise<NotificationRule> {
    return api<NotificationRule>('/admin/notifications/rules', {
      method: 'POST',
      body: JSON.stringify(data),
      signal
    });
  },

  update(
    id: string,
    data: UpsertNotificationRuleRequest,
    signal?: AbortSignal
  ): Promise<NotificationRule> {
    return api<NotificationRule>(`/admin/notifications/rules/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
      signal
    });
  },

  delete(id: string, signal?: AbortSignal): Promise<void> {
    return api<void>(`/admin/notifications/rules/${id}`, {
      method: 'DELETE',
      signal
    });
  }
};
