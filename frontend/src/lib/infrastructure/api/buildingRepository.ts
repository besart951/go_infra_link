import type { BuildingRepository } from '$lib/domain/ports/facility/buildingRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
  Building,
  BuildingListResponse,
  BuildingBulkResponse,
  CreateBuildingRequest,
  UpdateBuildingRequest
} from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';
import { buildListUrl, mapPaginatedResponse } from './listHelpers.js';

export const buildingRepository: BuildingRepository = {
  async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<Building>> {
    const response = await api<BuildingListResponse>(buildListUrl('/facility/buildings', params), {
      signal
    });

    return mapPaginatedResponse(response, params);
  },

  async get(id: string, signal?: AbortSignal): Promise<Building> {
    return api<Building>(`/facility/buildings/${id}`, { signal });
  },

  async getBulk(ids: string[], signal?: AbortSignal): Promise<Building[]> {
    const response = await api<BuildingBulkResponse>('/facility/buildings/bulk', {
      method: 'POST',
      body: JSON.stringify({ ids }),
      signal
    });
    return response.items;
  },

  async create(data: CreateBuildingRequest, signal?: AbortSignal): Promise<Building> {
    return api<Building>('/facility/buildings', {
      method: 'POST',
      body: JSON.stringify(data),
      signal
    });
  },

  async update(id: string, data: UpdateBuildingRequest, signal?: AbortSignal): Promise<Building> {
    return api<Building>(`/facility/buildings/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
      signal
    });
  },

  async delete(id: string, signal?: AbortSignal): Promise<void> {
    return api<void>(`/facility/buildings/${id}`, {
      method: 'DELETE',
      signal
    });
  },

  async validate(
    data: { id?: string; iws_code: string; building_group: number },
    signal?: AbortSignal
  ): Promise<void> {
    return api<void>('/facility/buildings/validate', {
      method: 'POST',
      body: JSON.stringify(data),
      signal
    });
  }
};
