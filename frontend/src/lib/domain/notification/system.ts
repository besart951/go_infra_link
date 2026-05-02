export interface SystemNotification {
  id: string;
  recipient_id: string;
  actor_id?: string | null;
  event_key: string;
  title: string;
  body: string;
  resource_type: string;
  resource_id?: string | null;
  metadata?: Record<string, string>;
  read_at?: string | null;
  is_important: boolean;
  created_at: string;
  updated_at: string;
}

export interface SystemNotificationList {
  items: SystemNotification[];
  total: number;
  page: number;
  total_pages: number;
  unread_count: number;
}

export interface ListSystemNotificationsParams {
  page?: number;
  limit?: number;
  unread_only?: boolean;
}

export const SYSTEM_NOTIFICATION_STREAM_EVENT = {
  Created: 'notification.created',
  Updated: 'notification.updated',
  Deleted: 'notification.deleted',
  ReadAll: 'notification.read_all'
} as const;

export type SystemNotificationStreamEventType =
  (typeof SYSTEM_NOTIFICATION_STREAM_EVENT)[keyof typeof SYSTEM_NOTIFICATION_STREAM_EVENT];

export interface SystemNotificationStreamEvent {
  type: SystemNotificationStreamEventType;
  notification?: SystemNotification;
  notification_id?: string;
  unread_count: number;
  at: string;
}
