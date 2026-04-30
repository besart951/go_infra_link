import type { SystemPartRepository } from '$lib/domain/ports/facility/systemPartRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
  SystemPart,
  CreateSystemPartRequest,
  UpdateSystemPartRequest
} from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';
import { buildListUrl, mapPaginatedResponse } from './listHelpers.js';

export const systemPartRepository: SystemPartRepository = {
  async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<SystemPart>> {
    const response = await api<any>(buildListUrl('/facility/system-parts', params), {
      signal
    });

    return mapPaginatedResponse(response, params);
  },

  async get(id: string, signal?: AbortSignal): Promise<SystemPart> {
    return api<SystemPart>(`/facility/system-parts/${id}`, { signal });
  },

  async create(data: CreateSystemPartRequest, signal?: AbortSignal): Promise<SystemPart> {
    return api<SystemPart>('/facility/system-parts', {
      method: 'POST',
      body: JSON.stringify(data),
      signal
    });
  },

  async update(
    id: string,
    data: UpdateSystemPartRequest,
    signal?: AbortSignal
  ): Promise<SystemPart> {
    return api<SystemPart>(`/facility/system-parts/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
      signal
    });
  },

  async delete(id: string, signal?: AbortSignal): Promise<void> {
    return api<void>(`/facility/system-parts/${id}`, {
      method: 'DELETE',
      signal
    });
  }
};
