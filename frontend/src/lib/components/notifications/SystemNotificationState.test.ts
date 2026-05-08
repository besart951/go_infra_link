import { describe, expect, it } from 'vitest';
import { parseSystemNotificationStreamEvent } from './SystemNotificationState.svelte.js';

const notificationId = '11111111-1111-4111-8111-111111111111';
const userId = '22222222-2222-4222-8222-222222222222';

function notification() {
  return {
    id: notificationId,
    recipient_id: userId,
    actor_id: null,
    event_key: 'notification.test',
    title: 'Test',
    body: 'Body',
    resource_type: 'project',
    resource_id: null,
    metadata: { project_id: notificationId },
    read_at: null,
    is_important: false,
    created_at: '2026-01-01T10:00:00Z',
    updated_at: '2026-01-01T10:00:00Z'
  };
}

describe('system notification realtime message validation', () => {
  it('accepts created notification events', () => {
    const parsed = parseSystemNotificationStreamEvent({
      type: 'notification.created',
      notification: notification(),
      unread_count: 1,
      at: '2026-01-01T10:00:01Z'
    });

    expect(parsed.type).toBe('notification.created');
  });

  it('accepts deleted notification events with a notification id', () => {
    const parsed = parseSystemNotificationStreamEvent({
      type: 'notification.deleted',
      notification_id: notificationId,
      unread_count: 0,
      at: '2026-01-01T10:00:01Z'
    });

    expect(parsed.notification_id).toBe(notificationId);
  });

  it('rejects invalid event types and malformed IDs', () => {
    expect(() =>
      parseSystemNotificationStreamEvent({
        type: 'notification.created',
        notification: { ...notification(), id: 'not-a-uuid' },
        unread_count: 1,
        at: '2026-01-01T10:00:01Z'
      })
    ).toThrow();

    expect(() =>
      parseSystemNotificationStreamEvent({
        type: 'notification.admin',
        unread_count: 1,
        at: '2026-01-01T10:00:01Z'
      })
    ).toThrow();
  });
});
