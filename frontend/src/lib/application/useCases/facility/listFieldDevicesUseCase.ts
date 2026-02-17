import type {
    FieldDevice,
    CreateFieldDeviceRequest,
    UpdateFieldDeviceRequest,
    FieldDeviceOptions
} from '$lib/domain/facility/index.js';
import type { FieldDeviceRepository } from '$lib/domain/ports/facility/fieldDeviceRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';

export class ListFieldDevicesUseCase {
    constructor(private repository: FieldDeviceRepository) { }

    async execute(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<FieldDevice>> {
        return this.repository.list(params, signal);
    }

    async get(id: string, signal?: AbortSignal): Promise<FieldDevice> {
        return this.repository.get(id, signal);
    }

    async getOptions(signal?: AbortSignal): Promise<FieldDeviceOptions> {
        return this.repository.getOptions(signal);
    }

    async getOptionsForProject(projectId: string, signal?: AbortSignal): Promise<FieldDeviceOptions> {
        return this.repository.getOptionsForProject(projectId, signal);
    }
}
