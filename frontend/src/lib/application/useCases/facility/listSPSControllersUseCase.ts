import type {
    SPSController,
    SPSControllerSystemType,
    SPSControllerSystemTypeListParams,
    SPSControllerSystemTypeListResponse
} from '$lib/domain/facility/index.js';
import type { SPSControllerRepository } from '$lib/domain/ports/facility/spsControllerRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';

export class ListSPSControllersUseCase {
    constructor(private repository: SPSControllerRepository) { }

    async execute(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<SPSController>> {
        return this.repository.list(params, signal);
    }

    async get(id: string, signal?: AbortSignal): Promise<SPSController> {
        const items = await this.repository.getBulk([id], signal);
        return items[0];
    }

    async getBulk(ids: string[], signal?: AbortSignal): Promise<SPSController[]> {
        return this.repository.getBulk(ids, signal);
    }

    async listSystemTypes(params?: SPSControllerSystemTypeListParams, signal?: AbortSignal): Promise<SPSControllerSystemTypeListResponse> {
        return this.repository.listSystemTypes(params, signal);
    }

    async getSystemType(id: string, signal?: AbortSignal): Promise<SPSControllerSystemType> {
        return this.repository.getSystemType(id, signal);
    }
}
