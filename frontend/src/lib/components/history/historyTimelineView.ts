import type { ChangeEvent, ChangeEventScope } from '$lib/domain/history.js';

const HIDDEN_DIFF_FIELDS = new Set(['created_at', 'updated_at']);
const FIELD_DEVICE_OWN_TABLES = new Set(['field_devices', 'specifications']);

const HIERARCHY_SCOPE_ORDER = [
  'building',
  'control_cabinet',
  'sps_controller',
  'sps_controller_system_type',
  'field_device',
  'bacnet_object'
] as const;

type HierarchyScopeType = (typeof HIERARCHY_SCOPE_ORDER)[number];

const HIERARCHY_SCOPE_SET = new Set<string>(HIERARCHY_SCOPE_ORDER);

const ENTITY_SCOPE_BY_TABLE: Record<string, HierarchyScopeType | undefined> = {
  buildings: 'building',
  control_cabinets: 'control_cabinet',
  project_control_cabinets: 'control_cabinet',
  sps_controllers: 'sps_controller',
  project_sps_controllers: 'sps_controller',
  sps_controller_system_types: 'sps_controller_system_type',
  field_devices: 'field_device',
  specifications: 'field_device',
  project_field_devices: 'field_device',
  bacnet_objects: 'bacnet_object',
  bacnet_object_alarm_values: 'bacnet_object'
};

export interface HistoryTimelineContext {
  scopeType?: string;
  scopeId?: string;
  controlCabinetId?: string;
}

export interface HistoryTimelineRow {
  event: ChangeEvent;
  field: string;
  before: unknown;
  after: unknown;
  moreFields: number;
  summaryOnly: boolean;
}

export interface HistoryTimelineGroup {
  key: string;
  kind: HierarchyScopeType | 'other';
  label?: string;
  rows: HistoryTimelineRow[];
  children: HistoryTimelineGroup[];
}

export interface HistoryTimelineView {
  directRows: HistoryTimelineRow[];
  childGroups: HistoryTimelineGroup[];
  flatRows: HistoryTimelineRow[];
  isHierarchicalView: boolean;
}

export function buildHistoryTimelineView(
  events: ChangeEvent[],
  context: HistoryTimelineContext
): HistoryTimelineView {
  const contextScopeType = normalizeScopeType(context.scopeType);
  const isHierarchicalView = Boolean(contextScopeType && context.scopeId);
  const rows = events.filter((event) => shouldShowEvent(event, contextScopeType)).map(toRow);

  if (!isHierarchicalView || !contextScopeType) {
    return {
      directRows: [],
      childGroups: [],
      flatRows: rows,
      isHierarchicalView: false
    };
  }

  const directRows: HistoryTimelineRow[] = [];
  const childGroups: HistoryTimelineGroup[] = [];

  for (const row of rows) {
    if (isDirectContextEvent(row.event, contextScopeType, context.scopeId)) {
      directRows.push(row);
      continue;
    }

    const path = groupPath(row.event, contextScopeType);
    if (path.length === 0) {
      directRows.push(row);
      continue;
    }

    appendToGroupTree(childGroups, path, row);
  }

  return {
    directRows,
    childGroups,
    flatRows: [],
    isHierarchicalView: true
  };
}

function shouldShowEvent(
  event: ChangeEvent,
  contextScopeType: HierarchyScopeType | undefined
): boolean {
  if (!contextScopeType) return true;

  const eventScopeType = eventScopeForTable(event.entity_table);
  if (!eventScopeType) return true;

  return hierarchyIndex(eventScopeType) >= hierarchyIndex(contextScopeType);
}

function isDirectContextEvent(
  event: ChangeEvent,
  contextScopeType: HierarchyScopeType,
  contextScopeId?: string
): boolean {
  if (!contextScopeId) return false;

  if (FIELD_DEVICE_OWN_TABLES.has(event.entity_table) && contextScopeType === 'field_device') {
    return hasScope(event, 'field_device', contextScopeId);
  }

  return eventScopeForTable(event.entity_table) === contextScopeType && event.entity_id === contextScopeId;
}

function groupPath(
  event: ChangeEvent,
  contextScopeType: HierarchyScopeType
): Omit<HistoryTimelineGroup, 'rows' | 'children'>[] {
  const contextIndex = hierarchyIndex(contextScopeType);
  const eventScopeType = eventScopeForTable(event.entity_table);
  const eventIndex = eventScopeType ? hierarchyIndex(eventScopeType) : HIERARCHY_SCOPE_ORDER.length;
  const path: Omit<HistoryTimelineGroup, 'rows' | 'children'>[] = [];

  for (let index = contextIndex + 1; index < HIERARCHY_SCOPE_ORDER.length; index += 1) {
    const scopeType = HIERARCHY_SCOPE_ORDER[index];
    if (index > eventIndex) break;

    const scope = findScope(event, scopeType);
    if (!scope) continue;

    path.push({
      key: `${scopeType}:${scope.scope_id}`,
      kind: scopeType,
      label: scope.label ?? shortId(scope.scope_id)
    });
  }

  if (path.length === 0 && eventScopeType && eventScopeType !== contextScopeType) {
    path.push({
      key: `${eventScopeType}:${event.entity_id}`,
      kind: eventScopeType,
      label: shortId(event.entity_id)
    });
  }

  return path;
}

function appendToGroupTree(
  groups: HistoryTimelineGroup[],
  path: Omit<HistoryTimelineGroup, 'rows' | 'children'>[],
  row: HistoryTimelineRow
): void {
  let currentGroups = groups;
  let currentGroup: HistoryTimelineGroup | undefined;

  for (const segment of path) {
    currentGroup = currentGroups.find((group) => group.key === segment.key);
    if (!currentGroup) {
      currentGroup = { ...segment, rows: [], children: [] };
      currentGroups.push(currentGroup);
    }
    currentGroups = currentGroup.children;
  }

  if (currentGroup) {
    currentGroup.rows.push(row);
  }
}

function eventScopeForTable(table: string): HierarchyScopeType | undefined {
  return ENTITY_SCOPE_BY_TABLE[table];
}

function normalizeScopeType(scopeType?: string): HierarchyScopeType | undefined {
  if (!scopeType || !HIERARCHY_SCOPE_SET.has(scopeType)) return undefined;
  return scopeType as HierarchyScopeType;
}

function hierarchyIndex(scopeType: HierarchyScopeType): number {
  return HIERARCHY_SCOPE_ORDER.indexOf(scopeType);
}

function findScope(event: ChangeEvent, scopeType: HierarchyScopeType): ChangeEventScope | undefined {
  return event.scopes?.find((scope) => scope.scope_type === scopeType);
}

function hasScope(event: ChangeEvent, scopeType: HierarchyScopeType, scopeId: string): boolean {
  return (
    event.scopes?.some((scope) => scope.scope_type === scopeType && scope.scope_id === scopeId) === true
  );
}

function toRow(event: ChangeEvent): HistoryTimelineRow {
  const entries = visibleDiffEntries(event);
  const [field, diff] = entries[0] ?? ['__record__', { before: null, after: null }];

  return {
    event,
    field,
    before: diff.before,
    after: diff.after,
    moreFields: Math.max(entries.length - 1, 0),
    summaryOnly: entries.length === 0
  };
}

function visibleDiffEntries(
  event: ChangeEvent
): Array<[string, { before: unknown; after: unknown }]> {
  return Object.entries(event.diff_json ?? {}).filter(([field]) => !HIDDEN_DIFF_FIELDS.has(field));
}

function shortId(id: string): string {
  return id.slice(0, 8);
}
