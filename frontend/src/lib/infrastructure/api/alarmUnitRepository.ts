import type { AlarmUnitRepository } from '$lib/domain/ports/facility/alarmUnitRepository.js';
import type { CreateUnitRequest, Unit, UpdateUnitRequest } from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';
import {
  buildListUrl,
  mapPaginatedResponse,
  type ApiPaginatedListResponse
} from './listHelpers.js';

export const alarmUnitRepository: AlarmUnitRepository = {
  async list(params, signal) {
    const response = await api<ApiPaginatedListResponse<Unit>>(
      buildListUrl('/facility/alarm-units', params),
      {
        signal
      }
    );
    return mapPaginatedResponse<Unit>(response, params);
  },
  async get(id, signal) {
    return api<Unit>(`/facility/alarm-units/${id}`, { signal });
  },
  async create(data: CreateUnitRequest, signal?: AbortSignal) {
    return api<Unit>('/facility/alarm-units', {
      method: 'POST',
      body: JSON.stringify(data),
      signal
    });
  },
  async update(id: string, data: UpdateUnitRequest, signal?: AbortSignal) {
    return api<Unit>(`/facility/alarm-units/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
      signal
    });
  },
  async delete(id: string, signal?: AbortSignal) {
    return api<void>(`/facility/alarm-units/${id}`, { method: 'DELETE', signal });
  }
};
