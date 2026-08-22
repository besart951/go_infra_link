import type { ControlCabinetRepository } from '$lib/domain/ports/facility/controlCabinetRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
  ControlCabinet,
  ControlCabinetListResponse,
  ControlCabinetBulkResponse,
  CreateControlCabinetRequest,
  UpdateControlCabinetRequest,
  ControlCabinetDeleteImpact,
  FacilityJob
} from '$lib/domain/facility/index.js';
import { toFacilityJob } from '$lib/domain/facility/facility-job.js';
import { api } from '$lib/api/client.js';
import { apiClient } from '$lib/api/generated/client.js';
import { buildListUrl, mapPaginatedResponse } from './listHelpers.js';
import { versionedDeletePath } from './versionedMutation.js';

export const controlCabinetRepository: ControlCabinetRepository = {
  async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<ControlCabinet>> {
    const response = await api<ControlCabinetListResponse>(
      buildListUrl('/facility/control-cabinets', params),
      { signal }
    );

    return mapPaginatedResponse(response, params);
  },

  async get(id: string, signal?: AbortSignal): Promise<ControlCabinet> {
    return api<ControlCabinet>(`/facility/control-cabinets/${id}`, { signal });
  },

  async getBulk(ids: string[], signal?: AbortSignal): Promise<ControlCabinet[]> {
    const response = await api<ControlCabinetBulkResponse>('/facility/control-cabinets/bulk', {
      method: 'POST',
      body: JSON.stringify({ ids }),
      signal
    });
    return response.items;
  },

  async copy(id: string, operationId: string, signal?: AbortSignal): Promise<FacilityJob> {
    const { data } = await apiClient.POST('/api/v1/facility/control-cabinets/{id}/copy', {
      params: {
        path: { id },
        header: { 'Idempotency-Key': operationId }
      },
      signal
    });
    if (!data) {
      throw new Error('Copy job response is empty');
    }
    return toFacilityJob(data);
  },

  async create(data: CreateControlCabinetRequest, signal?: AbortSignal): Promise<ControlCabinet> {
    return api<ControlCabinet>('/facility/control-cabinets', {
      method: 'POST',
      body: JSON.stringify(data),
      signal
    });
  },

  async update(
    id: string,
    data: UpdateControlCabinetRequest,
    signal?: AbortSignal
  ): Promise<ControlCabinet> {
    return api<ControlCabinet>(`/facility/control-cabinets/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
      signal
    });
  },

  async delete(command, signal?: AbortSignal): Promise<void> {
    return api<void>(versionedDeletePath('/facility/control-cabinets', command), {
      method: 'DELETE',
      signal
    });
  },

  async validate(
    data: { id?: string; building_id: string; control_cabinet_nr?: string },
    signal?: AbortSignal
  ): Promise<void> {
    return api<void>('/facility/control-cabinets/validate', {
      method: 'POST',
      body: JSON.stringify(data),
      signal
    });
  },

  async getDeleteImpact(id: string, signal?: AbortSignal): Promise<ControlCabinetDeleteImpact> {
    return api<ControlCabinetDeleteImpact>(`/facility/control-cabinets/${id}/delete-impact`, {
      signal
    });
  }
};
