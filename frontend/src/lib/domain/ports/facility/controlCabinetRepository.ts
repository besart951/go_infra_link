import type { ListRepository, ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
    ControlCabinet,
    CreateControlCabinetRequest,
    UpdateControlCabinetRequest,
    ControlCabinetDeleteImpact
} from '$lib/domain/facility/index.js';

export interface ControlCabinetRepository extends ListRepository<ControlCabinet> {
    list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<ControlCabinet>>;
    get(id: string, signal?: AbortSignal): Promise<ControlCabinet>;
    getBulk(ids: string[], signal?: AbortSignal): Promise<ControlCabinet[]>;
    create(data: CreateControlCabinetRequest, signal?: AbortSignal): Promise<ControlCabinet>;
    update(id: string, data: UpdateControlCabinetRequest, signal?: AbortSignal): Promise<ControlCabinet>;
    delete(id: string, signal?: AbortSignal): Promise<void>;
    validate(data: { id?: string; building_id: string; control_cabinet_nr?: string }, signal?: AbortSignal): Promise<void>;
    getDeleteImpact(id: string, signal?: AbortSignal): Promise<ControlCabinetDeleteImpact>;
}
