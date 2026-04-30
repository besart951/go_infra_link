import type { ApparatRepository } from '$lib/domain/ports/facility/apparatRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
  Apparat,
  ApparatListResponse,
  ApparatBulkResponse,
  CreateApparatRequest,
  UpdateApparatRequest
} from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';
import { buildListUrl, mapPaginatedResponse } from './listHelpers.js';

export const apparatRepository: ApparatRepository = {
  async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<Apparat>> {
    const response = await api<ApparatListResponse>(buildListUrl('/facility/apparats', params), {
      signal
    });

    return mapPaginatedResponse(response, params);
  },

  async get(id: string, signal?: AbortSignal): Promise<Apparat> {
    return api<Apparat>(`/facility/apparats/${id}`, { signal });
  },

  async getBulk(ids: string[], signal?: AbortSignal): Promise<Apparat[]> {
    const response = await api<ApparatBulkResponse>('/facility/apparats/bulk', {
      method: 'POST',
      body: JSON.stringify({ ids }),
      signal
    });
    return response.items;
  },

  async create(
    data: import('$lib/domain/facility/index.js').CreateApparatRequest,
    signal?: AbortSignal
  ): Promise<Apparat> {
    return api<Apparat>('/facility/apparats', {
      method: 'POST',
      body: JSON.stringify(data),
      signal
    });
  },

  async update(
    id: string,
    data: import('$lib/domain/facility/index.js').UpdateApparatRequest,
    signal?: AbortSignal
  ): Promise<Apparat> {
    return api<Apparat>(`/facility/apparats/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
      signal
    });
  },

  async delete(id: string, signal?: AbortSignal): Promise<void> {
    return api<void>(`/facility/apparats/${id}`, {
      method: 'DELETE',
      signal
    });
  }
};
