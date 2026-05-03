import { t as translate } from '$lib/i18n/index.js';
import type { ChangeEvent, HistoryAction } from '$lib/domain/history.js';

const HIDDEN_DIFF_FIELDS = new Set(['created_at', 'updated_at']);

export function historyActionLabel(action: HistoryAction): string {
  return translate(`history.actions.${action}`);
}

export function historyActionVariant(
  action: HistoryAction
): 'default' | 'secondary' | 'destructive' | 'outline' | 'success' | 'warning' {
  if (action === 'delete') return 'destructive';
  if (action === 'create') return 'success';
  if (action === 'restore') return 'warning';
  return 'secondary';
}

export function historyTableLabel(table: string): string {
  const key = `history.tables.${table}`;
  const label = translate(key);
  return label === key ? readableFieldFallback(table) : label;
}

export function historyFieldLabel(field: string): string {
  if (field === '__record__') return translate('history.record');
  const key = `history.fields.${field}`;
  const label = translate(key);
  if (label !== key) return label;

  const normalizedField = field.split('.').at(-1) ?? field;
  const normalizedKey = `history.fields.${normalizedField}`;
  const normalizedLabel = translate(normalizedKey);
  if (normalizedLabel !== normalizedKey) return normalizedLabel;

  return readableFieldFallback(normalizedField);
}

export function formatHistoryDate(value: string): string {
  return new Intl.DateTimeFormat('de-CH', {
    dateStyle: 'medium',
    timeStyle: 'short'
  }).format(new Date(value));
}

export function formatHistoryValue(value: unknown): string {
  if (value === null || value === undefined || value === '') return '∅';
  if (typeof value === 'object') return JSON.stringify(value);
  return String(value);
}

export function historyActorLabel(event: ChangeEvent): string {
  if (event.actor_name) return `${translate('history.actor')}: ${event.actor_name}`;
  if (event.actor_id) return `${translate('history.actor')}: ${event.actor_id}`;
  return translate('history.system');
}

export function historyVisibleDiffEntries(
  event: ChangeEvent
): Array<[string, { before: unknown; after: unknown }]> {
  return Object.entries(event.diff_json ?? {}).filter(([field]) => !HIDDEN_DIFF_FIELDS.has(field));
}

export function historyPrimaryField(event: ChangeEvent): string {
  return historyVisibleDiffEntries(event)[0]?.[0] ?? '__record__';
}

function readableFieldFallback(field: string): string {
  return field
    .replace(/_id$/u, '')
    .replace(/_/gu, ' ')
    .replace(/\bid\b/giu, 'ID')
    .replace(/\bip\b/giu, 'IP')
    .replace(/\bsps\b/giu, 'SPS')
    .replace(/\bbacnet\b/giu, 'BACnet')
    .replace(/\b\p{L}/gu, (char) => char.toLocaleUpperCase('de-CH'));
}
