/**
 * Project API repository
 * Infrastructure layer - implements ProjectRepository port via HTTP
 */
import { toProjectCapabilities } from '$lib/domain/project/capabilities.js';
import type {
  ProjectRepository,
  PaginationParams
} from '$lib/domain/ports/project/projectRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
  Project,
  ProjectListResponse,
  CreateProjectRequest,
  UpdateProjectRequest,
  ProjectUserListResponse,
  ProjectObjectDataListResponse,
  ProjectControlCabinetListResponse,
  ProjectSPSControllerListResponse,
  ProjectFieldDeviceListResponse,
  ProjectFieldDeviceMultiCreateResponse,
  ProjectCapabilities
} from '$lib/domain/project/index.js';
import type {
  ControlCabinet,
  MultiCreateFieldDeviceRequest,
  MultiCreateFieldDeviceResponse,
  ObjectDataListParams,
  SPSController,
  FacilityJob
} from '$lib/domain/facility/index.js';
import { toFacilityJob } from '$lib/domain/facility/facility-job.js';
import { api } from '$lib/api/client.js';
import { apiClient } from '$lib/api/generated/client.js';
import { versionedDeletePath, versionedProjectLinkDeletePath } from './versionedMutation.js';

function buildQuery(params: Record<string, string | number | boolean | undefined>): string {
  const searchParams = new URLSearchParams();
  for (const [key, value] of Object.entries(params)) {
    if (value !== undefined && value !== '') {
      searchParams.set(key, String(value));
    }
  }
  const query = searchParams.toString();
  return query ? `?${query}` : '';
}

export const projectRepository: ProjectRepository = {
  // ──────────────────────────────────────────────────────────────────────
  // CRUD
  // ──────────────────────────────────────────────────────────────────────

  async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<Project>> {
    const searchParams = new URLSearchParams();
    searchParams.set('page', String(params.pagination.page));
    searchParams.set('limit', String(params.pagination.pageSize));
    if (params.search.text) searchParams.set('search', params.search.text);

    if (params.filters) {
      Object.entries(params.filters).forEach(([key, value]) => {
        if (value !== undefined && value !== null) searchParams.set(key, value);
      });
    }

    const query = searchParams.toString();
    const response = await api<ProjectListResponse>(`/projects${query ? `?${query}` : ''}`, {
      signal
    });

    return {
      items: response.items,
      metadata: {
        total: response.total,
        page: response.page,
        pageSize: params.pagination.pageSize,
        totalPages: response.total_pages
      }
    };
  },

  async get(id: string, signal?: AbortSignal): Promise<Project> {
    return api<Project>(`/projects/${id}`, { signal });
  },

  async getCapabilities(id: string, signal?: AbortSignal): Promise<ProjectCapabilities> {
    const { data } = await apiClient.GET('/api/v1/projects/{id}/capabilities', {
      params: { path: { id } },
      signal
    });
    if (!data) {
      throw new Error('Project capabilities response is empty');
    }
    return toProjectCapabilities(data.permissions);
  },

  async create(data: CreateProjectRequest, signal?: AbortSignal): Promise<Project> {
    return api<Project>('/projects', {
      method: 'POST',
      body: JSON.stringify(data),
      signal
    });
  },

  async update(id: string, data: UpdateProjectRequest, signal?: AbortSignal): Promise<Project> {
    return api<Project>(`/projects/${id}`, {
      method: 'PATCH',
      body: JSON.stringify(data),
      signal
    });
  },

  async delete(command, signal?: AbortSignal): Promise<void> {
    await api<void>(versionedDeletePath('/projects', command), { method: 'DELETE', signal });
  },

  // ──────────────────────────────────────────────────────────────────────
  // Users
  // ──────────────────────────────────────────────────────────────────────

  async listUsers(projectId: string, signal?: AbortSignal): Promise<ProjectUserListResponse> {
    return api<ProjectUserListResponse>(`/projects/${projectId}/users`, { signal });
  },

  async addUser(projectId: string, userId: string, signal?: AbortSignal): Promise<void> {
    await apiClient.POST('/api/v1/projects/{id}/users', {
      params: { path: { id: projectId } },
      body: { user_id: userId },
      signal
    });
  },

  async removeUser(projectId: string, userId: string, signal?: AbortSignal): Promise<void> {
    await apiClient.DELETE('/api/v1/projects/{id}/users/{userId}', {
      params: { path: { id: projectId, userId } },
      signal
    });
  },

  // ──────────────────────────────────────────────────────────────────────
  // Object Data
  // ──────────────────────────────────────────────────────────────────────

  async listObjectData(
    projectId: string,
    params?: ObjectDataListParams,
    signal?: AbortSignal
  ): Promise<ProjectObjectDataListResponse> {
    const query = buildQuery({
      page: params?.page,
      limit: params?.limit,
      search: params?.search,
      apparat_id: params?.apparat_id,
      system_part_id: params?.system_part_id
    });
    return api<ProjectObjectDataListResponse>(`/projects/${projectId}/object-data${query}`, {
      signal
    });
  },

  async addObjectData(
    projectId: string,
    objectDataId: string,
    signal?: AbortSignal
  ): Promise<void> {
    await apiClient.POST('/api/v1/projects/{id}/object-data', {
      params: { path: { id: projectId } },
      body: { object_data_id: objectDataId },
      signal
    });
  },

  async removeObjectData(
    projectId: string,
    objectDataId: string,
    signal?: AbortSignal
  ): Promise<void> {
    await apiClient.DELETE('/api/v1/projects/{id}/object-data/{objectDataId}', {
      params: { path: { id: projectId, objectDataId } },
      signal
    });
  },

  // ──────────────────────────────────────────────────────────────────────
  // Control Cabinets
  // ──────────────────────────────────────────────────────────────────────

  async listControlCabinets(
    projectId: string,
    params?: PaginationParams,
    signal?: AbortSignal
  ): Promise<ProjectControlCabinetListResponse> {
    const query = buildQuery({ page: params?.page, limit: params?.limit });
    return api<ProjectControlCabinetListResponse>(
      `/projects/${projectId}/control-cabinets${query}`,
      { signal }
    );
  },

  async addControlCabinet(
    projectId: string,
    controlCabinetId: string,
    signal?: AbortSignal
  ): Promise<void> {
    await apiClient.POST('/api/v1/projects/{id}/control-cabinets', {
      params: { path: { id: projectId } },
      body: { control_cabinet_id: controlCabinetId },
      signal
    });
  },

  async removeControlCabinet(command, signal?: AbortSignal): Promise<void> {
    await api<void>(versionedProjectLinkDeletePath('control-cabinets', command), {
      method: 'DELETE',
      signal
    });
  },

  async copyControlCabinet(
    projectId: string,
    controlCabinetId: string,
    operationId: string,
    signal?: AbortSignal
  ): Promise<FacilityJob> {
    const { data } = await apiClient.POST(
      '/api/v1/projects/{id}/control-cabinets/{controlCabinetId}/copy',
      {
        params: {
          path: { id: projectId, controlCabinetId },
          header: { 'Idempotency-Key': operationId }
        },
        signal
      }
    );
    if (!data) {
      throw new Error('Copy job response is empty');
    }
    return toFacilityJob(data);
  },

  // ──────────────────────────────────────────────────────────────────────
  // SPS Controllers
  // ──────────────────────────────────────────────────────────────────────

  async listSPSControllers(
    projectId: string,
    params?: PaginationParams,
    signal?: AbortSignal
  ): Promise<ProjectSPSControllerListResponse> {
    const query = buildQuery({ page: params?.page, limit: params?.limit });
    return api<ProjectSPSControllerListResponse>(`/projects/${projectId}/sps-controllers${query}`, {
      signal
    });
  },

  async addSPSController(
    projectId: string,
    spsControllerId: string,
    signal?: AbortSignal
  ): Promise<void> {
    await apiClient.POST('/api/v1/projects/{id}/sps-controllers', {
      params: { path: { id: projectId } },
      body: { sps_controller_id: spsControllerId },
      signal
    });
  },

  async removeSPSController(command, signal?: AbortSignal): Promise<void> {
    await api<void>(versionedProjectLinkDeletePath('sps-controllers', command), {
      method: 'DELETE',
      signal
    });
  },

  async copySPSController(
    projectId: string,
    spsControllerId: string,
    operationId: string,
    signal?: AbortSignal
  ): Promise<FacilityJob> {
    const { data } = await apiClient.POST(
      '/api/v1/projects/{id}/sps-controllers/{spsControllerId}/copy',
      {
        params: {
          path: { id: projectId, spsControllerId },
          header: { 'Idempotency-Key': operationId }
        },
        signal
      }
    );
    if (!data) {
      throw new Error('Copy job response is empty');
    }
    return toFacilityJob(data);
  },

  // ──────────────────────────────────────────────────────────────────────
  // Field Devices
  // ──────────────────────────────────────────────────────────────────────

  async listFieldDevices(
    projectId: string,
    params?: PaginationParams,
    signal?: AbortSignal
  ): Promise<ProjectFieldDeviceListResponse> {
    const query = buildQuery({ page: params?.page, limit: params?.limit });
    return api<ProjectFieldDeviceListResponse>(`/projects/${projectId}/field-devices${query}`, {
      signal
    });
  },

  async addFieldDevice(
    projectId: string,
    fieldDeviceId: string,
    signal?: AbortSignal
  ): Promise<void> {
    await apiClient.POST('/api/v1/projects/{id}/field-devices', {
      params: { path: { id: projectId } },
      body: { field_device_id: fieldDeviceId },
      signal
    });
  },

  async addFieldDevices(
    projectId: string,
    fieldDeviceIds: string[],
    signal?: AbortSignal
  ): Promise<ProjectFieldDeviceMultiCreateResponse> {
    return api<ProjectFieldDeviceMultiCreateResponse>(
      `/projects/${projectId}/field-devices/multi-create`,
      {
        method: 'POST',
        body: JSON.stringify({ field_device_ids: fieldDeviceIds }),
        signal
      }
    );
  },

  async createFieldDevices(
    projectId: string,
    data: MultiCreateFieldDeviceRequest,
    signal?: AbortSignal
  ): Promise<MultiCreateFieldDeviceResponse> {
    return api<MultiCreateFieldDeviceResponse>(
      `/projects/${projectId}/field-devices/multi-create`,
      {
        method: 'POST',
        body: JSON.stringify(data),
        signal
      }
    );
  },

  async removeFieldDevice(command, signal?: AbortSignal): Promise<void> {
    await api<void>(versionedProjectLinkDeletePath('field-devices', command), {
      method: 'DELETE',
      signal
    });
  }
};
