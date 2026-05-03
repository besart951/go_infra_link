export type HistoryAction = 'create' | 'update' | 'delete' | 'restore';

export interface ChangeEvent {
  id: string;
  occurred_at: string;
  actor_id?: string;
  actor_name?: string;
  action: HistoryAction;
  entity_table: string;
  entity_id: string;
  batch_id?: string;
  summary?: string;
  before_json?: Record<string, unknown> | null;
  after_json?: Record<string, unknown> | null;
  diff_json?: Record<string, { before: unknown; after: unknown }> | null;
  metadata_json?: Record<string, unknown> | null;
  scopes?: ChangeEventScope[];
}

export interface ChangeEventScope {
  scope_type: string;
  scope_id: string;
  label?: string;
}

export interface HistoryListResponse {
  items: ChangeEvent[];
  total: number;
  page: number;
  total_pages: number;
}

export interface RestoreResult {
  restored_count: number;
  deleted_count: number;
  skipped_count: number;
  batch_id: string;
}

export interface HistoryTimelineParams {
  scopeType?: string;
  scopeId?: string;
  entityTable?: string;
  entityId?: string;
  page?: number;
  limit?: number;
}
