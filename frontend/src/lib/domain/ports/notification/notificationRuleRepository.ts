import type {
  NotificationRule,
  NotificationRuleList,
  UpsertNotificationRuleRequest
} from '$lib/domain/notification/index.js';

export interface NotificationRuleRepository {
  list(signal?: AbortSignal): Promise<NotificationRuleList>;
  create(data: UpsertNotificationRuleRequest, signal?: AbortSignal): Promise<NotificationRule>;
  update(
    id: string,
    data: UpsertNotificationRuleRequest,
    signal?: AbortSignal
  ): Promise<NotificationRule>;
  delete(id: string, signal?: AbortSignal): Promise<void>;
}
