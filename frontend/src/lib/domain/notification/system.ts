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
