import type { AlarmTypeRepository } from '$lib/domain/ports/facility/alarmTypeRepository.js';
import type {
	AlarmType,
	AlarmTypeField,
	AlarmTypeListResponse,
	CreateAlarmTypeFieldRequest,
	CreateAlarmTypeRequest,
	UpdateAlarmTypeFieldRequest,
	UpdateAlarmTypeRequest
} from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';

export const alarmTypeRepository: AlarmTypeRepository = {
	async list(params = {}, signal?: AbortSignal) {
		const searchParams = new URLSearchParams();
		if (params.page) searchParams.set('page', String(params.page));
		if (params.pageSize) searchParams.set('limit', String(params.pageSize));
		if (params.search) searchParams.set('search', params.search);
		const query = searchParams.toString();
		const response = await api<AlarmTypeListResponse>(
			`/facility/alarm-types${query ? `?${query}` : ''}`,
			{ signal }
		);
		return {
			items: response.items,
			total: response.total,
			page: response.page,
			totalPages: response.total_pages
		};
	},

	async get(id: string, signal?: AbortSignal): Promise<AlarmType> {
		return api<AlarmType>(`/facility/alarm-types/${id}`, { signal });
	},

	async getWithFields(id: string, signal?: AbortSignal): Promise<AlarmType> {
		return api<AlarmType>(`/facility/alarm-types/${id}/fields`, { signal });
	},

	async create(data: CreateAlarmTypeRequest, signal?: AbortSignal): Promise<AlarmType> {
		return api<AlarmType>('/facility/alarm-types', {
			method: 'POST',
			body: JSON.stringify(data),
			signal
		});
	},

	async update(id: string, data: UpdateAlarmTypeRequest, signal?: AbortSignal): Promise<AlarmType> {
		return api<AlarmType>(`/facility/alarm-types/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data),
			signal
		});
	},

	async delete(id: string, signal?: AbortSignal): Promise<void> {
		return api<void>(`/facility/alarm-types/${id}`, {
			method: 'DELETE',
			signal
		});
	},

	async createField(
		alarmTypeId: string,
		data: CreateAlarmTypeFieldRequest,
		signal?: AbortSignal
	): Promise<AlarmTypeField> {
		return api<AlarmTypeField>(`/facility/alarm-types/${alarmTypeId}/fields`, {
			method: 'POST',
			body: JSON.stringify(data),
			signal
		});
	},

	async updateField(
		id: string,
		data: UpdateAlarmTypeFieldRequest,
		signal?: AbortSignal
	): Promise<AlarmTypeField> {
		return api<AlarmTypeField>(`/facility/alarm-type-fields/${id}`, {
			method: 'PUT',
			body: JSON.stringify(data),
			signal
		});
	},

	async deleteField(id: string, signal?: AbortSignal): Promise<void> {
		return api<void>(`/facility/alarm-type-fields/${id}`, {
			method: 'DELETE',
			signal
		});
	}
};
