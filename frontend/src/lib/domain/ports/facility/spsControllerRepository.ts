import type { ListRepository, ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
    SPSController,
    CreateSPSControllerRequest,
    UpdateSPSControllerRequest,
    NextGADeviceResponse,
    SPSControllerSystemType,
    SPSControllerSystemTypeListParams,
    SPSControllerSystemTypeListResponse
} from '$lib/domain/facility/index.js';

export interface SPSControllerRepository extends ListRepository<SPSController> {
    list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<SPSController>>;
    get(id: string, signal?: AbortSignal): Promise<SPSController>;
    getBulk(ids: string[], signal?: AbortSignal): Promise<SPSController[]>;
    create(data: CreateSPSControllerRequest, signal?: AbortSignal): Promise<SPSController>;
    update(id: string, data: UpdateSPSControllerRequest, signal?: AbortSignal): Promise<SPSController>;
    delete(id: string, signal?: AbortSignal): Promise<void>;
    validate(data: { id?: string; control_cabinet_id: string; ga_device?: string; device_name: string; ip_address?: string; subnet?: string; gateway?: string; vlan?: string }, signal?: AbortSignal): Promise<void>;
    getNextGADevice(controlCabinetId: string, spsControllerId?: string, signal?: AbortSignal): Promise<NextGADeviceResponse>;

    // SPS Controller System Types (Sub-resource)
    listSystemTypes(params?: SPSControllerSystemTypeListParams, signal?: AbortSignal): Promise<SPSControllerSystemTypeListResponse>;
    getSystemType(id: string, signal?: AbortSignal): Promise<SPSControllerSystemType>;
}
