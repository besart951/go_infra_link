import type {
    SPSController,
    CreateSPSControllerRequest,
    UpdateSPSControllerRequest,
    SPSControllerSystemType,
    SPSControllerSystemTypeListParams,
    SPSControllerSystemTypeListResponse
} from '$lib/domain/facility/index.js';
import type { SPSControllerRepository } from '$lib/domain/ports/facility/spsControllerRepository.js';
import { ListEntityUseCase } from '$lib/application/useCases/listEntityUseCase.js';

export class ListSPSControllersUseCase extends ListEntityUseCase<SPSController, CreateSPSControllerRequest, UpdateSPSControllerRequest> {
    constructor(private repo: SPSControllerRepository) {
        super(repo);
    }

    async listSystemTypes(params?: SPSControllerSystemTypeListParams, signal?: AbortSignal): Promise<SPSControllerSystemTypeListResponse> {
        return this.repo.listSystemTypes(params, signal);
    }

    async getSystemType(id: string, signal?: AbortSignal): Promise<SPSControllerSystemType> {
        return this.repo.getSystemType(id, signal);
    }
}
