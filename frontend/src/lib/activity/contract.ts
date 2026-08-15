import type {
  ChangeEvent,
  HistoryAction,
  HistoryListResponse,
  HistoryTimelineParams
} from '$lib/domain/history.js';

export type ActivityAction = HistoryAction | 'relation_changed';

export interface ActivityChange {
  field: string;
  before: unknown;
  after: unknown;
}

export interface ActivityEntity {
  table: string;
  id: string;
}

export interface ActivityScope {
  scopeType?: string;
  scopeId?: string;
  entity?: ActivityEntity;
  fields?: string[];
}

export interface ActivityItem {
  id: string;
  event: ChangeEvent;
  action: ActivityAction;
  occurredAt: string;
  actorName?: string;
  entity: ActivityEntity;
  changes: ActivityChange[];
  scopes: Array<{ type: string; id: string; label?: string }>;
  summary?: string;
  batchId?: string;
}

export interface ActivityDataSource {
  cacheNamespace: string;
  list(params: HistoryTimelineParams, signal?: AbortSignal): Promise<HistoryListResponse>;
}
