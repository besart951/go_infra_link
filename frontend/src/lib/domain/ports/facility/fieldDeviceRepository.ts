import type {
  ListRepository,
  ListParams,
  PaginatedResponse
} from '$lib/domain/ports/listRepository.js';
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
  FieldDeviceExportJobResponse,
  BacnetObject,
  FieldDeviceDeleteCommand
} from '$lib/domain/facility/index.js';

export interface FieldDeviceRepository extends ListRepository<FieldDevice> {
  // Standard CRUD
  list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<FieldDevice>>;
  listCursor(
    params: FieldDeviceCursorListParams,
    signal?: AbortSignal
  ): Promise<FieldDeviceCursorPage>;
  get(id: string, signal?: AbortSignal): Promise<FieldDevice>;
  create(data: CreateFieldDeviceRequest, signal?: AbortSignal): Promise<FieldDevice>;
  update(id: string, data: UpdateFieldDeviceRequest, signal?: AbortSignal): Promise<FieldDevice>;
  delete(command: FieldDeviceDeleteCommand, signal?: AbortSignal): Promise<void>;

  // Bulk / Multi Operations
  multiCreate(
    data: MultiCreateFieldDeviceRequest,
    signal?: AbortSignal
  ): Promise<MultiCreateFieldDeviceResponse>;
  bulkUpdate(
    data: BulkUpdateFieldDeviceRequest,
    signal?: AbortSignal
  ): Promise<BulkUpdateFieldDeviceResponse>;
  bulkDelete(
    commands: FieldDeviceDeleteCommand[],
    signal?: AbortSignal
  ): Promise<BulkDeleteFieldDeviceResponse>;

  // Options / Helpers
  getOptions(signal?: AbortSignal): Promise<FieldDeviceOptions>;
  getOptionsForProject(projectId: string, signal?: AbortSignal): Promise<FieldDeviceOptions>;
  listBacnetObjects(fieldDeviceId: string, signal?: AbortSignal): Promise<BacnetObject[]>;
  getAvailableApparatNumbers(
    spsControllerSystemTypeId: string,
    apparatId: string,
    systemPartId?: string,
    signal?: AbortSignal
  ): Promise<AvailableApparatNumbersResponse>;

  // Exports
  createExport(
    data: CreateFieldDeviceExportRequest,
    signal?: AbortSignal
  ): Promise<FieldDeviceExportJobResponse>;
  getExportJob(jobId: string, signal?: AbortSignal): Promise<FieldDeviceExportJobResponse>;
  getExportDownloadUrl(jobId: string): string;
}

export interface FieldDeviceCursorListParams {
  limit: number;
  cursor?: string;
  search?: string;
  orderBy?: string;
  order?: 'asc' | 'desc';
  filters?: Record<string, string>;
}

export interface FieldDeviceCursorPage {
  items: FieldDevice[];
  nextCursor?: string;
  previousCursor?: string;
}
