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
	FieldDeviceExportJobResponse
} from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';
import {
	getAvailableApparatNumbers as getAvailableApparatNumbersApi,
	getFieldDeviceOptions,
	getFieldDeviceOptionsForProject
} from '$lib/infrastructure/api/facility.adapter.js';

export const fieldDeviceRepository: FieldDeviceRepository = {
	async list(params: ListParams, signal?: AbortSignal): Promise<PaginatedResponse<FieldDevice>> {
		const searchParams = new URLSearchParams();
		searchParams.set('page', String(params.pagination.page));
		searchParams.set('limit', String(params.pagination.pageSize));
		if (params.search.text) searchParams.set('search', params.search.text);

		if (params.filters) {
			Object.entries(params.filters).forEach(([key, value]) => {
				if (value !== undefined && value !== null) searchParams.set(key, value);
			});
		}

		const query = searchParams.toString();
		const response = await api<FieldDeviceListResponse>(
			`/facility/field-devices${query ? `?${query}` : ''}`,
			{ signal }
		);

		return {
			items: response.items,
			metadata: {
				total: response.total,
				page: response.page,
				pageSize: params.pagination.pageSize,
				totalPages: response.total_pages
			}
		};
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
