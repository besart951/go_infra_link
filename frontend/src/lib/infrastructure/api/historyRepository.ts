import { api } from '$lib/api/client.js';
import type {
  ChangeEvent,
  ChangeEventScope,
  HistoryListResponse,
  HistoryTimelineParams,
  RestoreResult
} from '$lib/domain/history.js';

export const historyRepository = {
  async listTimeline(
    params: HistoryTimelineParams,
    signal?: AbortSignal
  ): Promise<HistoryListResponse> {
    return api<HistoryListResponseWire>(timelineURL('/history/timeline', params), { signal }).then(
      toHistoryListResponse
    );
  },

  async listProjectTimeline(
    projectId: string,
    params: HistoryTimelineParams = {},
    signal?: AbortSignal
  ): Promise<HistoryListResponse> {
    return api<HistoryListResponseWire>(
      timelineURL(`/projects/${projectId}/history/timeline`, params),
      { signal }
    ).then(toHistoryListResponse);
  },

  restoreEvent(
    eventId: string,
    mode: 'before' | 'after' = 'after',
    signal?: AbortSignal
  ): Promise<RestoreResult> {
    return api<RestoreResult>(`/history/events/${eventId}/restore`, {
      method: 'POST',
      body: JSON.stringify({ mode }),
      signal
    });
  },

  restoreControlCabinet(
    controlCabinetId: string,
    eventId: string,
    signal?: AbortSignal
  ): Promise<RestoreResult> {
    return api<RestoreResult>(`/history/control-cabinets/${controlCabinetId}/restore`, {
      method: 'POST',
      body: JSON.stringify({ event_id: eventId }),
      signal
    });
  },

  restoreProjectControlCabinet(
    projectId: string,
    controlCabinetId: string,
    eventId: string,
    signal?: AbortSignal
  ): Promise<RestoreResult> {
    return api<RestoreResult>(
      `/projects/${projectId}/history/control-cabinets/${controlCabinetId}/restore`,
      {
        method: 'POST',
        body: JSON.stringify({ event_id: eventId }),
        signal
      }
    );
  }
};

interface HistoryListResponseWire {
  items?: TimelineEvent[];
  next_cursor?: string;
  previous_cursor?: string;
}

type TimelineEvent = Partial<ChangeEvent>;

function timelineURL(path: string, params: HistoryTimelineParams): string {
  const query = new URLSearchParams();
  setTimelineScalarParams(query, params);
  for (const action of params.actions ?? []) query.append('action', action);
  for (const field of params.fields ?? []) query.append('field', field);
  const encoded = query.toString();
  return `${path}${encoded ? `?${encoded}` : ''}`;
}

function setTimelineScalarParams(query: URLSearchParams, params: HistoryTimelineParams): void {
  const values: Array<[string, string | number | undefined]> = [
    ['scope_type', params.scopeType], ['scope_id', params.scopeId],
    ['entity_table', params.entityTable], ['entity_id', params.entityId],
    ['actor_id', params.actorId], ['occurred_from', params.occurredFrom],
    ['occurred_to', params.occurredTo], ['cursor', params.cursor], ['limit', params.limit]
  ];
  for (const [key, value] of values) {
    if (value !== undefined && value !== '') query.set(key, String(value));
  }
}

function toHistoryListResponse(response: HistoryListResponseWire | undefined): HistoryListResponse {
  if (!response) throw new Error('Die Timeline-Antwort ist leer.');
  return {
    items: (response.items ?? []).map(toChangeEvent),
    ...(response.next_cursor ? { next_cursor: response.next_cursor } : {}),
    ...(response.previous_cursor ? { previous_cursor: response.previous_cursor } : {})
  };
}

function toChangeEvent(event: TimelineEvent): ChangeEvent {
  if (!event.id || !event.occurred_at || !event.action || !event.entity_table || !event.entity_id) {
    throw new Error('Die Timeline-Antwort enthält ein unvollständiges Ereignis.');
  }
  return {
    id: event.id,
    occurred_at: event.occurred_at,
    actor_id: event.actor_id,
    actor_name: event.actor_name,
    action: event.action,
    entity_table: event.entity_table,
    entity_id: event.entity_id,
    batch_id: event.batch_id,
    summary: event.summary,
    scopes: (event.scopes ?? []).map(toChangeEventScope),
    before_json: toJSONValue(event.before_json),
    after_json: toJSONValue(event.after_json),
    diff_json: toJSONValue(event.diff_json) as ChangeEvent['diff_json'],
    metadata_json: toJSONValue(event.metadata_json)
  };
}

function toChangeEventScope(scope: {
  scope_type?: string;
  scope_id?: string;
  label?: string;
}): ChangeEventScope {
  if (!scope.scope_type || !scope.scope_id) {
    throw new Error('Die Timeline-Antwort enthält einen unvollständigen Kontext.');
  }
  return { scope_type: scope.scope_type, scope_id: scope.scope_id, label: scope.label };
}

function toJSONValue(value: Record<string, unknown> | null | undefined): Record<string, unknown> | null {
  return value ? (value as unknown as Record<string, unknown>) : null;
}
