import type {
    FieldDevice,
    CreateFieldDeviceRequest,
    UpdateFieldDeviceRequest,
    MultiCreateFieldDeviceRequest,
    MultiCreateFieldDeviceResponse,
    BulkUpdateFieldDeviceRequest,
    BulkUpdateFieldDeviceResponse,
    BulkDeleteFieldDeviceResponse,
    AvailableApparatNumbersResponse
} from '$lib/domain/facility/index.js';
import type { FieldDeviceRepository } from '$lib/domain/ports/facility/fieldDeviceRepository.js';

export class ManageFieldDeviceUseCase {
    constructor(private repository: FieldDeviceRepository) { }

    async create(data: CreateFieldDeviceRequest, signal?: AbortSignal): Promise<FieldDevice> {
        return this.repository.create(data, signal);
    }

    async update(id: string, data: UpdateFieldDeviceRequest, signal?: AbortSignal): Promise<FieldDevice> {
        return this.repository.update(id, data, signal);
    }

    async delete(id: string, signal?: AbortSignal): Promise<void> {
        return this.repository.delete(id, signal);
    }

    async multiCreate(data: MultiCreateFieldDeviceRequest, signal?: AbortSignal): Promise<MultiCreateFieldDeviceResponse> {
        return this.repository.multiCreate(data, signal);
    }

    async bulkUpdate(data: BulkUpdateFieldDeviceRequest, signal?: AbortSignal): Promise<BulkUpdateFieldDeviceResponse> {
        return this.repository.bulkUpdate(data, signal);
    }

    async bulkDelete(ids: string[], signal?: AbortSignal): Promise<BulkDeleteFieldDeviceResponse> {
        return this.repository.bulkDelete(ids, signal);
    }

    async getAvailableApparatNumbers(
        spsControllerSystemTypeId: string,
        apparatId: string,
        systemPartId?: string,
        signal?: AbortSignal
    ): Promise<AvailableApparatNumbersResponse> {
        return this.repository.getAvailableApparatNumbers(
            spsControllerSystemTypeId,
            apparatId,
            systemPartId,
            signal
        );
    }
}
