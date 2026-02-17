/**
 * Project API repository
 * Infrastructure layer - implements ProjectRepository port via HTTP
 */
import type { ProjectRepository, PaginationParams } from '$lib/domain/ports/project/projectRepository.js';
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
    ProjectFieldDeviceListResponse
} from '$lib/domain/project/index.js';
import type { ObjectDataListParams } from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';

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
        const response = await api<ProjectListResponse>(`/projects${query ? `?${query}` : ''}`, { signal });

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

    async create(data: CreateProjectRequest, signal?: AbortSignal): Promise<Project> {
        return api<Project>('/projects', {
            method: 'POST',
            body: JSON.stringify(data),
            signal
        });
    },

    async update(id: string, data: UpdateProjectRequest, signal?: AbortSignal): Promise<Project> {
        return api<Project>(`/projects/${id}`, {
            method: 'PUT',
            body: JSON.stringify(data),
            signal
        });
    },

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return api<void>(`/projects/${id}`, { method: 'DELETE', signal });
    },

    // ──────────────────────────────────────────────────────────────────────
    // Users
    // ──────────────────────────────────────────────────────────────────────

    async listUsers(projectId: string, signal?: AbortSignal): Promise<ProjectUserListResponse> {
        return api<ProjectUserListResponse>(`/projects/${projectId}/users`, { signal });
    },

    async addUser(projectId: string, userId: string, signal?: AbortSignal): Promise<void> {
        return api<void>(`/projects/${projectId}/users`, {
            method: 'POST',
            body: JSON.stringify({ user_id: userId }),
            signal
        });
    },

    async removeUser(projectId: string, userId: string, signal?: AbortSignal): Promise<void> {
        return api<void>(`/projects/${projectId}/users/${userId}`, {
            method: 'DELETE',
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
        return api<ProjectObjectDataListResponse>(
            `/projects/${projectId}/object-data${query}`,
            { signal }
        );
    },

    async addObjectData(
        projectId: string,
        objectDataId: string,
        signal?: AbortSignal
    ): Promise<void> {
        return api<void>(`/projects/${projectId}/object-data`, {
            method: 'POST',
            body: JSON.stringify({ object_data_id: objectDataId }),
            signal
        });
    },

    async removeObjectData(
        projectId: string,
        objectDataId: string,
        signal?: AbortSignal
    ): Promise<void> {
        return api<void>(`/projects/${projectId}/object-data/${objectDataId}`, {
            method: 'DELETE',
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
        return api<void>(`/projects/${projectId}/control-cabinets`, {
            method: 'POST',
            body: JSON.stringify({ control_cabinet_id: controlCabinetId }),
            signal
        });
    },

    async removeControlCabinet(
        projectId: string,
        linkId: string,
        signal?: AbortSignal
    ): Promise<void> {
        return api<void>(`/projects/${projectId}/control-cabinets/${linkId}`, {
            method: 'DELETE',
            signal
        });
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
        return api<ProjectSPSControllerListResponse>(
            `/projects/${projectId}/sps-controllers${query}`,
            { signal }
        );
    },

    async addSPSController(
        projectId: string,
        spsControllerId: string,
        signal?: AbortSignal
    ): Promise<void> {
        return api<void>(`/projects/${projectId}/sps-controllers`, {
            method: 'POST',
            body: JSON.stringify({ sps_controller_id: spsControllerId }),
            signal
        });
    },

    async removeSPSController(
        projectId: string,
        linkId: string,
        signal?: AbortSignal
    ): Promise<void> {
        return api<void>(`/projects/${projectId}/sps-controllers/${linkId}`, {
            method: 'DELETE',
            signal
        });
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
        return api<ProjectFieldDeviceListResponse>(
            `/projects/${projectId}/field-devices${query}`,
            { signal }
        );
    },

    async addFieldDevice(
        projectId: string,
        fieldDeviceId: string,
        signal?: AbortSignal
    ): Promise<void> {
        return api<void>(`/projects/${projectId}/field-devices`, {
            method: 'POST',
            body: JSON.stringify({ field_device_id: fieldDeviceId }),
            signal
        });
    },

    async removeFieldDevice(
        projectId: string,
        linkId: string,
        signal?: AbortSignal
    ): Promise<void> {
        return api<void>(`/projects/${projectId}/field-devices/${linkId}`, {
            method: 'DELETE',
            signal
        });
    }
};
