import { api } from '$lib/api/client.js';
import { apiClient } from '$lib/api/generated/client.js';
import type {
  ChangeEvent,
  ChangeEventScope,
  HistoryListResponse,
  HistoryTimelineParams,
  RestoreResult
} from '$lib/domain/history.js';
import type { components } from '$lib/api/generated/schema.js';

export const historyRepository = {
  async listTimeline(
    params: HistoryTimelineParams,
    signal?: AbortSignal
  ): Promise<HistoryListResponse> {
    const { data } = await apiClient.GET('/api/v1/history/timeline', {
      params: { query: toTimelineQuery(params) },
      signal
    });
    return toHistoryListResponse(data);
  },

  async listProjectTimeline(
    projectId: string,
    params: HistoryTimelineParams = {},
    signal?: AbortSignal
  ): Promise<HistoryListResponse> {
    const { data } = await apiClient.GET('/api/v1/projects/{id}/history/timeline', {
      params: { path: { id: projectId }, query: toTimelineQuery(params) },
      signal
    });
    return toHistoryListResponse(data);
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

type TimelineResponse =
  components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_history.TimelineResponse'];
type TimelineEvent =
  components['schemas']['github_com_besart951_go_infra_link_backend_internal_handler_dto_history.ChangeEventResponse'];

function toTimelineQuery(params: HistoryTimelineParams) {
  return {
    scope_type: params.scopeType,
    scope_id: params.scopeId,
    entity_table: params.entityTable,
    entity_id: params.entityId,
    actor_id: params.actorId,
    occurred_from: params.occurredFrom,
    occurred_to: params.occurredTo,
    action: params.actions,
    field: params.fields,
    page: params.page,
    limit: params.limit
  };
}

function toHistoryListResponse(response: TimelineResponse | undefined): HistoryListResponse {
  if (!response) throw new Error('Die Timeline-Antwort ist leer.');
  return {
    items: (response.items ?? []).map(toChangeEvent),
    total: response.total ?? 0,
    page: response.page ?? 1,
    total_pages: response.total_pages ?? 1
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

function toJSONValue(value: Record<string, never> | undefined): Record<string, unknown> | null {
  return value ? (value as unknown as Record<string, unknown>) : null;
}
