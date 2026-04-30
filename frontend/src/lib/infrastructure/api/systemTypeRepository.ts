import type { SystemTypeRepository } from '$lib/domain/ports/facility/systemTypeRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
  SystemType,
  SystemTypeListResponse,
  CreateSystemTypeRequest,
  UpdateSystemTypeRequest
} from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';
import { buildListUrl, mapPaginatedResponse } from './listHelpers.js';

export const systemTypeRepository: SystemTypeRepository = {
  async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<SystemType>> {
    const response = await api<SystemTypeListResponse>(
      buildListUrl('/facility/system-types', params),
      { signal }
    );

    return mapPaginatedResponse(response, params);
  },

  async get(id: string, signal?: AbortSignal): Promise<SystemType> {
    return api<SystemType>(`/facility/system-types/${id}`, { signal });
  },

  async create(data: CreateSystemTypeRequest, signal?: AbortSignal): Promise<SystemType> {
    return api<SystemType>('/facility/system-types', {
      method: 'POST',
      body: JSON.stringify(data),
      signal
    });
  },

  async update(
    id: string,
    data: UpdateSystemTypeRequest,
    signal?: AbortSignal
  ): Promise<SystemType> {
    return api<SystemType>(`/facility/system-types/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
      signal
    });
  },

  async delete(id: string, signal?: AbortSignal): Promise<void> {
    return api<void>(`/facility/system-types/${id}`, {
      method: 'DELETE',
      signal
    });
  }
};
