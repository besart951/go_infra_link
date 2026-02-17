import type { ListRepository, ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
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

export interface PaginationParams {
    page?: number;
    limit?: number;
}

export interface ProjectRepository extends ListRepository<Project> {
    list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<Project>>;
    get(id: string, signal?: AbortSignal): Promise<Project>;
    create(data: CreateProjectRequest, signal?: AbortSignal): Promise<Project>;
    update(id: string, data: UpdateProjectRequest, signal?: AbortSignal): Promise<Project>;
    delete(id: string, signal?: AbortSignal): Promise<void>;

    // Project user links
    listUsers(projectId: string, signal?: AbortSignal): Promise<ProjectUserListResponse>;
    addUser(projectId: string, userId: string, signal?: AbortSignal): Promise<void>;
    removeUser(projectId: string, userId: string, signal?: AbortSignal): Promise<void>;

    // Project object data links
    listObjectData(
        projectId: string,
        params?: ObjectDataListParams,
        signal?: AbortSignal
    ): Promise<ProjectObjectDataListResponse>;
    addObjectData(
        projectId: string,
        objectDataId: string,
        signal?: AbortSignal
    ): Promise<void>;
    removeObjectData(
        projectId: string,
        objectDataId: string,
        signal?: AbortSignal
    ): Promise<void>;

    // Project control cabinet links
    listControlCabinets(
        projectId: string,
        params?: PaginationParams,
        signal?: AbortSignal
    ): Promise<ProjectControlCabinetListResponse>;
    addControlCabinet(
        projectId: string,
        controlCabinetId: string,
        signal?: AbortSignal
    ): Promise<void>;
    removeControlCabinet(
        projectId: string,
        linkId: string,
        signal?: AbortSignal
    ): Promise<void>;

    // Project SPS controller links
    listSPSControllers(
        projectId: string,
        params?: PaginationParams,
        signal?: AbortSignal
    ): Promise<ProjectSPSControllerListResponse>;
    addSPSController(
        projectId: string,
        spsControllerId: string,
        signal?: AbortSignal
    ): Promise<void>;
    removeSPSController(
        projectId: string,
        linkId: string,
        signal?: AbortSignal
    ): Promise<void>;

    // Project field device links
    listFieldDevices(
        projectId: string,
        params?: PaginationParams,
        signal?: AbortSignal
    ): Promise<ProjectFieldDeviceListResponse>;
    addFieldDevice(
        projectId: string,
        fieldDeviceId: string,
        signal?: AbortSignal
    ): Promise<void>;
    removeFieldDevice(
        projectId: string,
        linkId: string,
        signal?: AbortSignal
    ): Promise<void>;
}
