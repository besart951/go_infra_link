import type {
    CreateFieldDeviceExportRequest,
    FieldDeviceExportJobResponse
} from '$lib/domain/facility/index.js';
import type { FieldDeviceRepository } from '$lib/domain/ports/facility/fieldDeviceRepository.js';

export class ExportFieldDevicesUseCase {
    constructor(private repository: FieldDeviceRepository) { }

    async createExport(data: CreateFieldDeviceExportRequest, signal?: AbortSignal): Promise<FieldDeviceExportJobResponse> {
        return this.repository.createExport(data, signal);
    }

    async getExportJob(jobId: string, signal?: AbortSignal): Promise<FieldDeviceExportJobResponse> {
        return this.repository.getExportJob(jobId, signal);
    }

    getExportDownloadUrl(jobId: string): string {
        return this.repository.getExportDownloadUrl(jobId);
    }
}
