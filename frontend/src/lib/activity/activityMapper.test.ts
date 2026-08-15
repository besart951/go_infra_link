import { describe, expect, it } from 'vitest';
import type { ChangeEvent } from '$lib/domain/history.js';
import { toActivityItem } from './activityMapper.js';

function event(overrides: Partial<ChangeEvent> = {}): ChangeEvent {
  return {
    id: 'event-1',
    occurred_at: '2026-08-15T10:00:00Z',
    action: 'update',
    entity_table: 'field_devices',
    entity_id: 'field-device-1',
    diff_json: {
      description: { before: 'Alt', after: 'Neu' },
      updated_at: { before: '2026-08-15T09:00:00Z', after: '2026-08-15T10:00:00Z' }
    },
    ...overrides
  };
}

describe('toActivityItem', () => {
  it('keeps the actual audit diff while hiding mechanical timestamps', () => {
    const item = toActivityItem(event());

    expect(item.changes).toEqual([{ field: 'description', before: 'Alt', after: 'Neu' }]);
    expect(item.action).toBe('update');
  });

  it('marks relation changes without replacing the authoritative audit event', () => {
    const source = event({
      diff_json: {
        apparat_id: { before: 'apparat-a', after: 'apparat-b' }
      }
    });

    const item = toActivityItem(source);

    expect(item.action).toBe('relation_changed');
    expect(item.event).toBe(source);
    expect(item.changes).toEqual([
      { field: 'apparat_id', before: 'apparat-a', after: 'apparat-b' }
    ]);
  });

  it('preserves readable scope labels for every rendering level', () => {
    const item = toActivityItem(
      event({
        scopes: [
          { scope_type: 'project', scope_id: 'project-1', label: 'Umbau Süd' },
          { scope_type: 'field_device', scope_id: 'field-device-1', label: 'FD-18' }
        ]
      })
    );

    expect(item.scopes).toEqual([
      { type: 'project', id: 'project-1', label: 'Umbau Süd' },
      { type: 'field_device', id: 'field-device-1', label: 'FD-18' }
    ]);
  });
});
