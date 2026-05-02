import { describe, expect, it } from 'vitest';
import {
  reduceSystemNotificationStreamEvent,
  sortSystemNotifications
} from './systemNotificationCache.js';
import { SYSTEM_NOTIFICATION_STREAM_EVENT, type SystemNotification } from './system.js';

function notification(overrides: Partial<SystemNotification>): SystemNotification {
  return {
    id: 'notification-1',
    recipient_id: 'user-1',
    event_key: 'project.updated',
    title: 'Projekt',
    body: 'Update',
    resource_type: 'project',
    is_important: false,
    created_at: '2026-01-01T10:00:00.000Z',
    updated_at: '2026-01-01T10:00:00.000Z',
    ...overrides
  };
}

describe('system notification cache', () => {
  it('sorts important notifications before newer normal notifications', () => {
    const important = notification({
      id: 'important',
      is_important: true,
      created_at: '2026-01-01T08:00:00.000Z'
    });
    const newer = notification({
      id: 'newer',
      created_at: '2026-01-01T09:00:00.000Z'
    });

    expect(sortSystemNotifications([newer, important]).map((item) => item.id)).toEqual([
      'important',
      'newer'
    ]);
  });

  it('applies created events to preview and asks inbox to reload when active', () => {
    const created = notification({ id: 'created' });

    const result = reduceSystemNotificationStreamEvent(
      {
        previewItems: [],
        inboxItems: [],
        unreadCount: 0,
        inboxUnreadOnly: false,
        inboxLoaded: true
      },
      {
        type: SYSTEM_NOTIFICATION_STREAM_EVENT.Created,
        notification: created,
        unread_count: 4,
        at: '2026-01-01T10:00:00.000Z'
      },
      5
    );

    expect(result.previewItems).toEqual([created]);
    expect(result.unreadCount).toBe(4);
    expect(result.reloadInboxPage).toBe('current');
  });

  it('removes read updates from unread-only inbox', () => {
    const unread = notification({ id: 'n1', read_at: null });
    const read = notification({ id: 'n1', read_at: '2026-01-01T11:00:00.000Z' });

    const result = reduceSystemNotificationStreamEvent(
      {
        previewItems: [unread],
        inboxItems: [unread],
        unreadCount: 1,
        inboxUnreadOnly: true,
        inboxLoaded: true
      },
      {
        type: SYSTEM_NOTIFICATION_STREAM_EVENT.Updated,
        notification: read,
        unread_count: 0,
        at: '2026-01-01T11:00:00.000Z'
      },
      5
    );

    expect(result.previewItems).toEqual([read]);
    expect(result.inboxItems).toEqual([]);
    expect(result.unreadCount).toBe(0);
  });

  it('read-all events keep REST as source of truth by asking first inbox page to reload', () => {
    const result = reduceSystemNotificationStreamEvent(
      {
        previewItems: [notification({ id: 'n1' })],
        inboxItems: [notification({ id: 'n1' })],
        unreadCount: 3,
        inboxUnreadOnly: false,
        inboxLoaded: true
      },
      {
        type: SYSTEM_NOTIFICATION_STREAM_EVENT.ReadAll,
        unread_count: 0,
        at: '2026-01-01T12:00:00.000Z'
      },
      5
    );

    expect(result.unreadCount).toBe(0);
    expect(result.reloadInboxPage).toBe('first');
  });
});
