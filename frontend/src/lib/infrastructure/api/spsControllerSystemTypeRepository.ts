import type { SPSControllerSystemTypeRepository } from '$lib/domain/ports/facility/spsControllerSystemTypeRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
  SPSControllerSystemType,
  SPSControllerSystemTypeListResponse
} from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';
import { buildListUrl, mapPaginatedResponse } from './listHelpers.js';

export const spsControllerSystemTypeRepository: SPSControllerSystemTypeRepository = {
  async list(
    params: ListParams,
    signal?: AbortSignal
  ): Promise<PaginatedResponse<SPSControllerSystemType>> {
    const response = await api<SPSControllerSystemTypeListResponse>(
      buildListUrl('/facility/sps-controller-system-types', params),
      { signal }
    );

    return mapPaginatedResponse(response, params);
  },

  async get(id: string, signal?: AbortSignal): Promise<SPSControllerSystemType> {
    return api<SPSControllerSystemType>(`/facility/sps-controller-system-types/${id}`, { signal });
  },

  async copy(id: string, signal?: AbortSignal): Promise<SPSControllerSystemType> {
    return api<SPSControllerSystemType>(`/facility/sps-controller-system-types/${id}/copy`, {
      method: 'POST',
      signal
    });
  },
  async delete(id: string, signal?: AbortSignal): Promise<void> {
    await api<void>(`/facility/sps-controller-system-types/${id}`, {
      method: 'DELETE',
      signal
    });
  }
};
