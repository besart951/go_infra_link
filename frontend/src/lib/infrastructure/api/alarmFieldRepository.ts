import type { AlarmFieldRepository } from '$lib/domain/ports/facility/alarmFieldRepository.js';
import type {
	AlarmField,
	CreateAlarmFieldRequest,
	UpdateAlarmFieldRequest
} from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';

interface AlarmFieldListResponse {
	items: AlarmField[];
	total: number;
	page: number;
	total_pages: number;
}

export const alarmFieldRepository: AlarmFieldRepository = {
	async list(params, signal) {
		const page = params.pagination.page;
		const pageSize = params.pagination.pageSize;
		const search = params.search.text;
		const qp = new URLSearchParams({ page: String(page), limit: String(pageSize) });
		if (search) qp.set('search', search);
		const response = await api<AlarmFieldListResponse>(`/facility/alarm-fields?${qp.toString()}`, {
			signal
		});
		return {
			items: response.items,
			metadata: {
				total: response.total,
				page: response.page,
				pageSize,
				totalPages: response.total_pages
			}
		};
	},
	async get(id, signal) {
		return api<AlarmField>(`/facility/alarm-fields/${id}`, { signal });
	},
	async create(data: CreateAlarmFieldRequest, signal?: AbortSignal) {
		return api<AlarmField>('/facility/alarm-fields', {
			method: 'POST',
			body: JSON.stringify(data),
			signal
		});
	},
	async update(id: string, data: UpdateAlarmFieldRequest, signal?: AbortSignal) {
		return api<AlarmField>(`/facility/alarm-fields/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data),
			signal
		});
	},
	async delete(id: string, signal?: AbortSignal) {
		return api<void>(`/facility/alarm-fields/${id}`, { method: 'DELETE', signal });
	}
};
