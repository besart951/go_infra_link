import type { ListRepository, ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
    FieldDevice,
    CreateFieldDeviceRequest,
    UpdateFieldDeviceRequest,
    MultiCreateFieldDeviceRequest,
    MultiCreateFieldDeviceResponse,
    BulkUpdateFieldDeviceRequest,
    BulkUpdateFieldDeviceResponse,
    BulkDeleteFieldDeviceResponse,
    FieldDeviceOptions,
    AvailableApparatNumbersResponse,
    CreateFieldDeviceExportRequest,
    FieldDeviceExportJobResponse
} from '$lib/domain/facility/index.js';

export interface FieldDeviceRepository extends ListRepository<FieldDevice> {
    // Standard CRUD
    list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<FieldDevice>>;
    get(id: string, signal?: AbortSignal): Promise<FieldDevice>;
    create(data: CreateFieldDeviceRequest, signal?: AbortSignal): Promise<FieldDevice>;
    update(id: string, data: UpdateFieldDeviceRequest, signal?: AbortSignal): Promise<FieldDevice>;
    delete(id: string, signal?: AbortSignal): Promise<void>;

    // Bulk / Multi Operations
    multiCreate(data: MultiCreateFieldDeviceRequest, signal?: AbortSignal): Promise<MultiCreateFieldDeviceResponse>;
    bulkUpdate(data: BulkUpdateFieldDeviceRequest, signal?: AbortSignal): Promise<BulkUpdateFieldDeviceResponse>;
    bulkDelete(ids: string[], signal?: AbortSignal): Promise<BulkDeleteFieldDeviceResponse>;

    // Options / Helpers
    getOptions(signal?: AbortSignal): Promise<FieldDeviceOptions>;
    getOptionsForProject(projectId: string, signal?: AbortSignal): Promise<FieldDeviceOptions>;
    getAvailableApparatNumbers(
        spsControllerSystemTypeId: string,
        apparatId: string,
        systemPartId?: string,
        signal?: AbortSignal
    ): Promise<AvailableApparatNumbersResponse>;

    // Exports
    createExport(data: CreateFieldDeviceExportRequest, signal?: AbortSignal): Promise<FieldDeviceExportJobResponse>;
    getExportJob(jobId: string, signal?: AbortSignal): Promise<FieldDeviceExportJobResponse>;
    getExportDownloadUrl(jobId: string): string;
}
