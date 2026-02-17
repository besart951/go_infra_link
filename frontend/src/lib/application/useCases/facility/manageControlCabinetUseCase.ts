import type { ControlCabinet, CreateControlCabinetRequest, UpdateControlCabinetRequest, ControlCabinetDeleteImpact } from '$lib/domain/facility/index.js';
import type { ControlCabinetRepository } from '$lib/domain/ports/facility/controlCabinetRepository.js';

export class ManageControlCabinetUseCase {
    constructor(private repository: ControlCabinetRepository) { }

    async create(data: CreateControlCabinetRequest, signal?: AbortSignal): Promise<ControlCabinet> {
        return this.repository.create(data, signal);
    }

    async update(id: string, data: UpdateControlCabinetRequest, signal?: AbortSignal): Promise<ControlCabinet> {
        return this.repository.update(id, data, signal);
    }

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return this.repository.delete(id, signal);
    }

    async validate(data: { id?: string; building_id: string; control_cabinet_nr?: string }, signal?: AbortSignal): Promise<void> {
        return this.repository.validate(data, signal);
    }

    async getDeleteImpact(id: string, signal?: AbortSignal): Promise<ControlCabinetDeleteImpact> {
        return this.repository.getDeleteImpact(id, signal);
    }
}
