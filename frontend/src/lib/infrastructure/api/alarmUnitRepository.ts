import type { AlarmUnitRepository } from '$lib/domain/ports/facility/alarmUnitRepository.js';
import type { CreateUnitRequest, Unit, UpdateUnitRequest } from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';

interface UnitListResponse {
	items: Unit[];
	total: number;
	page: number;
	total_pages: number;
}

export const alarmUnitRepository: AlarmUnitRepository = {
	async list(params, signal) {
		const page = params.pagination.page;
		const pageSize = params.pagination.pageSize;
		const search = params.search.text;
		const qp = new URLSearchParams({ page: String(page), limit: String(pageSize) });
		if (search) qp.set('search', search);
		const response = await api<UnitListResponse>(`/facility/alarm-units?${qp.toString()}`, {
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
		return api<Unit>(`/facility/alarm-units/${id}`, { signal });
	},
	async create(data: CreateUnitRequest, signal?: AbortSignal) {
		return api<Unit>('/facility/alarm-units', {
			method: 'POST',
			body: JSON.stringify(data),
			signal
		});
	},
	async update(id: string, data: UpdateUnitRequest, signal?: AbortSignal) {
		return api<Unit>(`/facility/alarm-units/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data),
			signal
		});
	},
	async delete(id: string, signal?: AbortSignal) {
		return api<void>(`/facility/alarm-units/${id}`, { method: 'DELETE', signal });
	}
};
