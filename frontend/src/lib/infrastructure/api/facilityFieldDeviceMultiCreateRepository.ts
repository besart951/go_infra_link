import type { FieldDeviceMultiCreateRepository } from '$lib/domain/ports/facility/fieldDeviceMultiCreateRepository.js';
import type {
	FieldDeviceOptions,
	AvailableApparatNumbersResponse,
	MultiCreateFieldDeviceResponse
} from '$lib/domain/facility/field-device.js';
import type { SPSControllerSystemType } from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';
import { spsControllerSystemTypeRepository } from '$lib/infrastructure/api/spsControllerSystemTypeRepository.js';

export const facilityFieldDeviceMultiCreateRepository: FieldDeviceMultiCreateRepository = {
	async getFieldDeviceOptions(signal) {
		return api<FieldDeviceOptions>('/facility/field-devices/options', { signal });
	},

	async getFieldDeviceOptionsForProject(projectId, signal) {
		return api<FieldDeviceOptions>(`/projects/${projectId}/field-device-options`, { signal });
	},

	async listSpsControllerSystemTypes(params, signal) {
		const res = await spsControllerSystemTypeRepository.list({
			pagination: { page: 1, pageSize: params.limit ?? 50 },
			search: { text: params.search ?? '' }
		}, signal);
		return res.items;
	},

	async getSpsControllerSystemType(id, signal) {
		return spsControllerSystemTypeRepository.get(id, signal);
	},

	async getAvailableApparatNumbers(spsControllerSystemTypeId, apparatId, systemPartId, signal) {
		const searchParams = new URLSearchParams();
		searchParams.set('sps_controller_system_type_id', spsControllerSystemTypeId);
		searchParams.set('apparat_id', apparatId);
		if (systemPartId) {
			searchParams.set('system_part_id', systemPartId);
		}
		return api<AvailableApparatNumbersResponse>(
			`/facility/field-devices/available-apparat-nr?${searchParams.toString()}`,
			{ signal }
		);
	},

	async multiCreateFieldDevices(fieldDevices, signal) {
		return api<MultiCreateFieldDeviceResponse>('/facility/field-devices/multi-create', {
			method: 'POST',
			body: JSON.stringify({ field_devices: fieldDevices }),
			signal
		});
	}
};
