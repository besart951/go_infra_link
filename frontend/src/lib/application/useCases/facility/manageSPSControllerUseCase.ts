import type {
    SPSController,
    CreateSPSControllerRequest,
    UpdateSPSControllerRequest,
    NextGADeviceResponse
} from '$lib/domain/facility/index.js';
import type { SPSControllerRepository } from '$lib/domain/ports/facility/spsControllerRepository.js';

export class ManageSPSControllerUseCase {
    constructor(private repository: SPSControllerRepository) { }

    async create(data: CreateSPSControllerRequest, signal?: AbortSignal): Promise<SPSController> {
        return this.repository.create(data, signal);
    }

    async update(id: string, data: UpdateSPSControllerRequest, signal?: AbortSignal): Promise<SPSController> {
        return this.repository.update(id, data, signal);
    }

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return this.repository.delete(id, signal);
    }

    async validate(data: { id?: string; control_cabinet_id: string; ga_device?: string; device_name: string; ip_address?: string; subnet?: string; gateway?: string; vlan?: string }, signal?: AbortSignal): Promise<void> {
        return this.repository.validate(data, signal);
    }

    async getNextGADevice(controlCabinetId: string, spsControllerId?: string, signal?: AbortSignal): Promise<NextGADeviceResponse> {
        return this.repository.getNextGADevice(controlCabinetId, spsControllerId, signal);
    }
}
