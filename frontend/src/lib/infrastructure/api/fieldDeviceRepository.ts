import type { FieldDeviceRepository } from '$lib/domain/ports/facility/fieldDeviceRepository.js';
import type { ListParams, PaginatedResponse } from '$lib/domain/ports/listRepository.js';
import type {
  FieldDevice,
  FieldDeviceListResponse,
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
  BacnetObject
} from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';
import { buildListUrl, mapPaginatedResponse } from './listHelpers.js';
import {
  getAvailableApparatNumbers as getAvailableApparatNumbersApi,
  getFieldDeviceOptions,
  getFieldDeviceOptionsForProject
} from './fieldDeviceOptionsEndpoint.js';

export const fieldDeviceRepository: FieldDeviceRepository = {
  async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<FieldDevice>> {
    const response = await api<FieldDeviceListResponse>(
      buildListUrl('/facility/field-devices', params),
      { signal }
    );

    return mapPaginatedResponse(response, params);
  },

  async get(id: string, signal?: AbortSignal): Promise<FieldDevice> {
    return api<FieldDevice>(`/facility/field-devices/${id}`, { signal });
  },

  async create(data: CreateFieldDeviceRequest, signal?: AbortSignal): Promise<FieldDevice> {
    return api<FieldDevice>('/facility/field-devices', {
      method: 'POST',
      body: JSON.stringify(data),
      signal
    });
  },

  async update(
    id: string,
    data: UpdateFieldDeviceRequest,
    signal?: AbortSignal
  ): Promise<FieldDevice> {
    return api<FieldDevice>(`/facility/field-devices/${id}`, {
      method: 'PUT',
      body: JSON.stringify(data),
      signal
    });
  },

  async delete(id: string, signal?: AbortSignal): Promise<void> {
    return api<void>(`/facility/field-devices/${id}`, {
      method: 'DELETE',
      signal
    });
  },

  async multiCreate(
    data: MultiCreateFieldDeviceRequest,
    signal?: AbortSignal
  ): Promise<MultiCreateFieldDeviceResponse> {
    return api<MultiCreateFieldDeviceResponse>('/facility/field-devices/multi-create', {
      method: 'POST',
      body: JSON.stringify(data),
      signal
    });
  },

  async bulkUpdate(
    data: BulkUpdateFieldDeviceRequest,
    signal?: AbortSignal
  ): Promise<BulkUpdateFieldDeviceResponse> {
    return api<BulkUpdateFieldDeviceResponse>('/facility/field-devices/bulk-update', {
      method: 'PATCH',
      body: JSON.stringify(data),
      signal
    });
  },

  async bulkDelete(ids: string[], signal?: AbortSignal): Promise<BulkDeleteFieldDeviceResponse> {
    return api<BulkDeleteFieldDeviceResponse>('/facility/field-devices/bulk-delete', {
      method: 'DELETE',
      body: JSON.stringify({ ids }),
      signal
    });
  },

  async getOptions(signal?: AbortSignal): Promise<FieldDeviceOptions> {
    return getFieldDeviceOptions({ signal });
  },

  async getOptionsForProject(projectId: string, signal?: AbortSignal): Promise<FieldDeviceOptions> {
    return getFieldDeviceOptionsForProject(projectId, { signal });
  },

  async listBacnetObjects(fieldDeviceId: string, signal?: AbortSignal): Promise<BacnetObject[]> {
    return api<BacnetObject[]>(`/facility/field-devices/${fieldDeviceId}/bacnet-objects`, {
      signal
    });
  },

  async getAvailableApparatNumbers(
    spsControllerSystemTypeId: string,
    apparatId: string,
    systemPartId?: string,
    signal?: AbortSignal
  ): Promise<AvailableApparatNumbersResponse> {
    return getAvailableApparatNumbersApi(spsControllerSystemTypeId, apparatId, systemPartId, {
      signal
    });
  },

  async createExport(
    data: CreateFieldDeviceExportRequest,
    signal?: AbortSignal
  ): Promise<FieldDeviceExportJobResponse> {
    return api<FieldDeviceExportJobResponse>('/facility/exports/field-devices', {
      method: 'POST',
      body: JSON.stringify(data),
      signal
    });
  },

  async getExportJob(jobId: string, signal?: AbortSignal): Promise<FieldDeviceExportJobResponse> {
    return api<FieldDeviceExportJobResponse>(`/facility/exports/jobs/${jobId}`, { signal });
  },

  getExportDownloadUrl(jobId: string): string {
    return `/api/v1/facility/exports/jobs/${jobId}/download`;
  }
};
