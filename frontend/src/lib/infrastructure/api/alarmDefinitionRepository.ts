import type { AlarmDefinitionRepository } from '$lib/domain/ports/facility/alarmDefinitionRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
  AlarmDefinition,
  AlarmDefinitionListResponse,
  CreateAlarmDefinitionRequest,
  UpdateAlarmDefinitionRequest
} from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';
import { buildListUrl, mapPaginatedResponse } from './listHelpers.js';

export const alarmDefinitionRepository: AlarmDefinitionRepository = {
  async list(
    params: ListParams,
    signal?: AbortSignal
  ): Promise<PaginatedResponse<AlarmDefinition>> {
    const response = await api<AlarmDefinitionListResponse>(
      buildListUrl('/facility/alarm-definitions', params),
      { signal }
    );

    return mapPaginatedResponse(response, params);
  },

  async get(id: string, signal?: AbortSignal): Promise<AlarmDefinition> {
    return api<AlarmDefinition>(`/facility/alarm-definitions/${id}`, { signal });
  },

  async create(data: CreateAlarmDefinitionRequest, signal?: AbortSignal): Promise<AlarmDefinition> {
    return api<AlarmDefinition>('/facility/alarm-definitions', {
      method: 'POST',
      body: JSON.stringify(data),
      signal
    });
  },

  async update(
    id: string,
    data: UpdateAlarmDefinitionRequest,
    signal?: AbortSignal
  ): Promise<AlarmDefinition> {
    return api<AlarmDefinition>(`/facility/alarm-definitions/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
      signal
    });
  },

  async delete(id: string, signal?: AbortSignal): Promise<void> {
    return api<void>(`/facility/alarm-definitions/${id}`, {
      method: 'DELETE',
      signal
    });
  }
};
