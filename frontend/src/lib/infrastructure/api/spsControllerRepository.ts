import type { SPSControllerRepository } from '$lib/domain/ports/facility/spsControllerRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
  SPSController,
  SPSControllerListResponse,
  SPSControllerBulkResponse,
  CreateSPSControllerRequest,
  UpdateSPSControllerRequest,
  NextGADeviceResponse,
  SPSControllerSystemType,
  SPSControllerSystemTypeListParams,
  SPSControllerSystemTypeListResponse
} from '$lib/domain/facility/index.js';
import { createCachedBulkFetchByIds } from '$lib/infrastructure/api/createCachedBulkFetch.js';
import { spsControllerSystemTypeRepository } from '$lib/infrastructure/api/spsControllerSystemTypeRepository.js';
import { api } from '$lib/api/client.js';
import { buildListUrl, mapPaginatedResponse } from './listHelpers.js';

const getBulkCached = createCachedBulkFetchByIds('facility-sps-controllers', (ids, signal) =>
  api<SPSControllerBulkResponse>('/facility/sps-controllers/bulk', {
    method: 'POST',
    body: JSON.stringify({ ids }),
    signal
  }).then((response) => response.items)
);

export const spsControllerRepository: SPSControllerRepository = {
  async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<SPSController>> {
    const response = await api<SPSControllerListResponse>(
      buildListUrl('/facility/sps-controllers', params),
      { signal }
    );

    return mapPaginatedResponse(response, params);
  },

  async get(id: string, signal?: AbortSignal): Promise<SPSController> {
    return api<SPSController>(`/facility/sps-controllers/${id}`, { signal });
  },

  async getBulk(ids: string[], signal?: AbortSignal): Promise<SPSController[]> {
    return getBulkCached(ids, signal);
  },

  async copy(id: string, signal?: AbortSignal): Promise<SPSController> {
    return api<SPSController>(`/facility/sps-controllers/${id}/copy`, {
      method: 'POST',
      signal
    });
  },

  async create(data: CreateSPSControllerRequest, signal?: AbortSignal): Promise<SPSController> {
    return api<SPSController>('/facility/sps-controllers', {
      method: 'POST',
      body: JSON.stringify(data),
      signal
    });
  },

  async update(
    id: string,
    data: UpdateSPSControllerRequest,
    signal?: AbortSignal
  ): Promise<SPSController> {
    return api<SPSController>(`/facility/sps-controllers/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
      signal
    });
  },

  async delete(id: string, signal?: AbortSignal): Promise<void> {
    return api<void>(`/facility/sps-controllers/${id}`, {
      method: 'DELETE',
      signal
    });
  },

  async validate(
    data: {
      id?: string;
      control_cabinet_id: string;
      ga_device?: string;
      device_name: string;
      ip_address?: string;
      subnet?: string;
      gateway?: string;
      vlan?: string;
    },
    signal?: AbortSignal
  ): Promise<void> {
    return api<void>('/facility/sps-controllers/validate', {
      method: 'POST',
      body: JSON.stringify(data),
      signal
    });
  },

  async getNextGADevice(
    controlCabinetId: string,
    spsControllerId?: string,
    signal?: AbortSignal
  ): Promise<NextGADeviceResponse> {
    const searchParams = new URLSearchParams();
    searchParams.set('control_cabinet_id', controlCabinetId);
    if (spsControllerId) {
      searchParams.set('sps_controller_id', spsControllerId);
    }
    return api<NextGADeviceResponse>(
      `/facility/sps-controllers/next-ga-device?${searchParams.toString()}`,
      { signal }
    );
  },

  async listSystemTypes(
    params?: SPSControllerSystemTypeListParams,
    signal?: AbortSignal
  ): Promise<SPSControllerSystemTypeListResponse> {
    const filters: Record<string, string> = {};
    if (params?.project_id) {
      filters.project_id = params.project_id;
    }
    if (params?.sps_controller_id) {
      filters.sps_controller_id = params.sps_controller_id;
    }

    const listResponse = await spsControllerSystemTypeRepository.list(
      {
        pagination: {
          page: params?.page ?? 1,
          pageSize: params?.limit ?? 1000
        },
        search: {
          text: params?.search ?? ''
        },
        ...(Object.keys(filters).length > 0 ? { filters } : {})
      },
      signal
    );

    return {
      items: listResponse.items,
      total: listResponse.metadata.total,
      page: listResponse.metadata.page,
      limit: listResponse.metadata.pageSize,
      total_pages: listResponse.metadata.totalPages
    };
  },

  async getSystemType(id: string, signal?: AbortSignal): Promise<SPSControllerSystemType> {
    return api<SPSControllerSystemType>(`/facility/sps-controller-system-types/${id}`, {
      signal
    });
  }
};
