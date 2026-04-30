import type { AlarmFieldRepository } from '$lib/domain/ports/facility/alarmFieldRepository.js';
import type {
  AlarmField,
  CreateAlarmFieldRequest,
  UpdateAlarmFieldRequest
} from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';
import {
  buildListUrl,
  mapPaginatedResponse,
  type ApiPaginatedListResponse
} from './listHelpers.js';

export const alarmFieldRepository: AlarmFieldRepository = {
  async list(params, signal) {
    const response = await api<ApiPaginatedListResponse<AlarmField>>(
      buildListUrl('/facility/alarm-fields', params),
      {
        signal
      }
    );
    return mapPaginatedResponse<AlarmField>(response, params);
  },
  async get(id, signal) {
    return api<AlarmField>(`/facility/alarm-fields/${id}`, { signal });
  },
  async create(data: CreateAlarmFieldRequest, signal?: AbortSignal) {
    return api<AlarmField>('/facility/alarm-fields', {
      method: 'POST',
      body: JSON.stringify(data),
      signal
    });
  },
  async update(id: string, data: UpdateAlarmFieldRequest, signal?: AbortSignal) {
    return api<AlarmField>(`/facility/alarm-fields/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
      signal
    });
  },
  async delete(id: string, signal?: AbortSignal) {
    return api<void>(`/facility/alarm-fields/${id}`, { method: 'DELETE', signal });
  }
};
