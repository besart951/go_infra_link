import { api } from '$lib/api/client.js';
import type {
  HistoryListResponse,
  HistoryTimelineParams,
  RestoreResult
} from '$lib/domain/history.js';

function buildTimelineUrl(path: string, params: HistoryTimelineParams = {}): string {
  const query = new URLSearchParams();
  if (params.scopeType) query.set('scope_type', params.scopeType);
  if (params.scopeId) query.set('scope_id', params.scopeId);
  if (params.entityTable) query.set('entity_table', params.entityTable);
  if (params.entityId) query.set('entity_id', params.entityId);
  if (params.page) query.set('page', String(params.page));
  if (params.limit) query.set('limit', String(params.limit));
  const suffix = query.toString();
  return suffix ? `${path}?${suffix}` : path;
}

export const historyRepository = {
  listTimeline(params: HistoryTimelineParams, signal?: AbortSignal): Promise<HistoryListResponse> {
    return api<HistoryListResponse>(buildTimelineUrl('/history/timeline', params), { signal });
  },

  listProjectTimeline(
    projectId: string,
    params: HistoryTimelineParams = {},
    signal?: AbortSignal
  ): Promise<HistoryListResponse> {
    return api<HistoryListResponse>(
      buildTimelineUrl(`/projects/${projectId}/history/timeline`, params),
      {
        signal
      }
    );
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
