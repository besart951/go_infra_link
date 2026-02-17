import type {
    SPSController,
    CreateSPSControllerRequest,
    UpdateSPSControllerRequest,
    NextGADeviceResponse
} from '$lib/domain/facility/index.js';
import type { SPSControllerRepository } from '$lib/domain/ports/facility/spsControllerRepository.js';
import { ManageEntityUseCase } from '$lib/application/useCases/manageEntityUseCase.js';

export class ManageSPSControllerUseCase extends ManageEntityUseCase<SPSController, CreateSPSControllerRequest, UpdateSPSControllerRequest> {
    constructor(private repo: SPSControllerRepository) {
        super(repo);
    }

    async validate(data: { id?: string; control_cabinet_id: string; ga_device?: string; device_name: string; ip_address?: string; subnet?: string; gateway?: string; vlan?: string }, signal?: AbortSignal): Promise<void> {
        return this.repo.validate(data, signal);
    }

    async getNextGADevice(controlCabinetId: string, spsControllerId?: string, signal?: AbortSignal): Promise<NextGADeviceResponse> {
        return this.repo.getNextGADevice(controlCabinetId, spsControllerId, signal);
    }
}
