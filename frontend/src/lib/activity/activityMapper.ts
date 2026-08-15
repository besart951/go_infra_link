import type { ChangeEvent } from '$lib/domain/history.js';
import type { ActivityAction, ActivityItem } from './contract.js';

const HIDDEN_FIELDS = new Set(['created_at', 'updated_at']);

export function toActivityItem(event: ChangeEvent): ActivityItem {
  const changes = Object.entries(event.diff_json ?? {})
    .filter(([field]) => !HIDDEN_FIELDS.has(field))
    .map(([field, value]) => ({ field, before: value.before, after: value.after }));

  return {
    id: event.id,
    event,
    action: activityAction(event, changes),
    occurredAt: event.occurred_at,
    actorName: event.actor_name,
    entity: { table: event.entity_table, id: event.entity_id },
    changes,
    scopes: (event.scopes ?? []).map((scope) => ({
      type: scope.scope_type,
      id: scope.scope_id,
      label: scope.label
    })),
    summary: event.summary,
    batchId: event.batch_id
  };
}

export function toActivityItems(events: ChangeEvent[]): ActivityItem[] {
  return events.map(toActivityItem);
}

function activityAction(event: ChangeEvent, changes: ActivityItem['changes']): ActivityAction {
  if (event.action !== 'update') return event.action;

  return changes.some(isRelationChange) ? 'relation_changed' : event.action;
}

function isRelationChange(change: ActivityItem['changes'][number]): boolean {
  return (
    change.field.endsWith('_id') &&
    change.before !== undefined &&
    change.after !== undefined &&
    change.before !== change.after
  );
}
