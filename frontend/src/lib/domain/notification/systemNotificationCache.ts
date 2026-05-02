import type { SystemNotification, SystemNotificationStreamEvent } from './system.js';
import { SYSTEM_NOTIFICATION_STREAM_EVENT } from './system.js';

export interface SystemNotificationCache {
  previewItems: SystemNotification[];
  inboxItems: SystemNotification[];
  unreadCount: number;
  inboxUnreadOnly: boolean;
  inboxLoaded: boolean;
}

export interface SystemNotificationCacheReduction extends SystemNotificationCache {
  reloadInboxPage?: 'current' | 'first';
}

export function reduceSystemNotificationStreamEvent(
  cache: SystemNotificationCache,
  event: SystemNotificationStreamEvent,
  previewLimit: number
): SystemNotificationCacheReduction {
  const unreadCount =
    typeof event.unread_count === 'number' && event.unread_count >= 0
      ? event.unread_count
      : cache.unreadCount;

  switch (event.type) {
    case SYSTEM_NOTIFICATION_STREAM_EVENT.Created:
      if (!event.notification) {
        return { ...cache, unreadCount };
      }
      return {
        ...cache,
        unreadCount,
        previewItems: limitSystemNotificationPreview(
          upsertSystemNotification(cache.previewItems, event.notification),
          previewLimit
        ),
        reloadInboxPage: cache.inboxLoaded ? 'current' : undefined
      };
    case SYSTEM_NOTIFICATION_STREAM_EVENT.Updated:
      if (!event.notification) {
        return { ...cache, unreadCount };
      }
      return {
        ...cache,
        unreadCount,
        previewItems: limitSystemNotificationPreview(
          upsertSystemNotification(cache.previewItems, event.notification),
          previewLimit
        ),
        inboxItems: applyInboxSystemNotificationUpdate(
          cache.inboxItems,
          event.notification,
          cache.inboxUnreadOnly
        )
      };
    case SYSTEM_NOTIFICATION_STREAM_EVENT.Deleted:
      if (!event.notification_id) {
        return { ...cache, unreadCount };
      }
      return {
        ...cache,
        unreadCount,
        previewItems: cache.previewItems.filter((item) => item.id !== event.notification_id),
        inboxItems: cache.inboxItems.filter((item) => item.id !== event.notification_id)
      };
    case SYSTEM_NOTIFICATION_STREAM_EVENT.ReadAll:
      return {
        ...cache,
        unreadCount,
        reloadInboxPage: 'first'
      };
    default:
      return { ...cache, unreadCount };
  }
}

export function upsertSystemNotification(
  items: SystemNotification[],
  notification: SystemNotification
): SystemNotification[] {
  const next = items.filter((item) => item.id !== notification.id);
  next.push(notification);
  return sortSystemNotifications(next);
}

export function applyInboxSystemNotificationUpdate(
  items: SystemNotification[],
  notification: SystemNotification,
  unreadOnly: boolean
): SystemNotification[] {
  if (unreadOnly && notification.read_at) {
    return items.filter((item) => item.id !== notification.id);
  }
  return upsertSystemNotification(items, notification);
}

export function limitSystemNotificationPreview(
  items: SystemNotification[],
  limit: number
): SystemNotification[] {
  return sortSystemNotifications(items).slice(0, limit);
}

export function sortSystemNotifications(items: SystemNotification[]): SystemNotification[] {
  return [...items].sort((a, b) => {
    if (a.is_important !== b.is_important) {
      return a.is_important ? -1 : 1;
    }
    return Date.parse(b.created_at) - Date.parse(a.created_at);
  });
}
