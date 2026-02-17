import type { ProjectRepository } from '$lib/domain/ports/project/projectRepository.js';
import type {
    Project,
    CreateProjectRequest,
    UpdateProjectRequest,
    ProjectUserListResponse,
    ProjectObjectDataListResponse,
    ProjectControlCabinetListResponse,
    ProjectSPSControllerListResponse,
    ProjectFieldDeviceListResponse
} from '$lib/domain/project/index.js';
import type { ObjectDataListParams } from '$lib/domain/facility/index.js';
import type { PaginationParams } from '$lib/domain/ports/project/projectRepository.js';

export class ManageProjectUseCase {
    constructor(private repository: ProjectRepository) { }

    async get(id: string, signal?: AbortSignal): Promise<Project> {
        return this.repository.get(id, signal);
    }

    async create(data: CreateProjectRequest, signal?: AbortSignal): Promise<Project> {
        return this.repository.create(data, signal);
    }

    async update(
        id: string,
        data: UpdateProjectRequest,
        signal?: AbortSignal
    ): Promise<Project> {
        return this.repository.update(id, data, signal);
    }

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return this.repository.delete(id, signal);
    }

    // ── User links ──────────────────────────────────────────────────────

    async listUsers(
        projectId: string,
        signal?: AbortSignal
    ): Promise<ProjectUserListResponse> {
        return this.repository.listUsers(projectId, signal);
    }

    async addUser(
        projectId: string,
        userId: string,
        signal?: AbortSignal
    ): Promise<void> {
        return this.repository.addUser(projectId, userId, signal);
    }

    async removeUser(
        projectId: string,
        userId: string,
        signal?: AbortSignal
    ): Promise<void> {
        return this.repository.removeUser(projectId, userId, signal);
    }

    // ── Object Data links ───────────────────────────────────────────────

    async listObjectData(
        projectId: string,
        params?: ObjectDataListParams,
        signal?: AbortSignal
    ): Promise<ProjectObjectDataListResponse> {
        return this.repository.listObjectData(projectId, params, signal);
    }

    async addObjectData(
        projectId: string,
        objectDataId: string,
        signal?: AbortSignal
    ): Promise<void> {
        return this.repository.addObjectData(projectId, objectDataId, signal);
    }

    async removeObjectData(
        projectId: string,
        objectDataId: string,
        signal?: AbortSignal
    ): Promise<void> {
        return this.repository.removeObjectData(projectId, objectDataId, signal);
    }

    // ── Control Cabinet links ───────────────────────────────────────────

    async listControlCabinets(
        projectId: string,
        params?: PaginationParams,
        signal?: AbortSignal
    ): Promise<ProjectControlCabinetListResponse> {
        return this.repository.listControlCabinets(projectId, params, signal);
    }

    async addControlCabinet(
        projectId: string,
        controlCabinetId: string,
        signal?: AbortSignal
    ): Promise<void> {
        return this.repository.addControlCabinet(projectId, controlCabinetId, signal);
    }

    async removeControlCabinet(
        projectId: string,
        linkId: string,
        signal?: AbortSignal
    ): Promise<void> {
        return this.repository.removeControlCabinet(projectId, linkId, signal);
    }

    // ── SPS Controller links ────────────────────────────────────────────

    async listSPSControllers(
        projectId: string,
        params?: PaginationParams,
        signal?: AbortSignal
    ): Promise<ProjectSPSControllerListResponse> {
        return this.repository.listSPSControllers(projectId, params, signal);
    }

    async addSPSController(
        projectId: string,
        spsControllerId: string,
        signal?: AbortSignal
    ): Promise<void> {
        return this.repository.addSPSController(projectId, spsControllerId, signal);
    }

    async removeSPSController(
        projectId: string,
        linkId: string,
        signal?: AbortSignal
    ): Promise<void> {
        return this.repository.removeSPSController(projectId, linkId, signal);
    }

    // ── Field Device links ──────────────────────────────────────────────

    async listFieldDevices(
        projectId: string,
        params?: PaginationParams,
        signal?: AbortSignal
    ): Promise<ProjectFieldDeviceListResponse> {
        return this.repository.listFieldDevices(projectId, params, signal);
    }

    async addFieldDevice(
        projectId: string,
        fieldDeviceId: string,
        signal?: AbortSignal
    ): Promise<void> {
        return this.repository.addFieldDevice(projectId, fieldDeviceId, signal);
    }

    async removeFieldDevice(
        projectId: string,
        linkId: string,
        signal?: AbortSignal
    ): Promise<void> {
        return this.repository.removeFieldDevice(projectId, linkId, signal);
    }
}
