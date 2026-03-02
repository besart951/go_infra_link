import type { FieldDeviceMultiCreateRepository } from '$lib/domain/ports/facility/fieldDeviceMultiCreateRepository.js';
import type {
	FieldDeviceOptions,
	AvailableApparatNumbersResponse,
	MultiCreateFieldDeviceResponse
} from '$lib/domain/facility/field-device.js';
import type { SPSControllerSystemType } from '$lib/domain/facility/index.js';
import { api } from '$lib/api/client.js';
import {
	getAvailableApparatNumbers as getAvailableApparatNumbersApi,
	getFieldDeviceOptions,
	getFieldDeviceOptionsForProject
} from '$lib/infrastructure/api/facility.adapter.js';
import { spsControllerSystemTypeRepository } from '$lib/infrastructure/api/spsControllerSystemTypeRepository.js';

export const facilityFieldDeviceMultiCreateRepository: FieldDeviceMultiCreateRepository = {
	async getFieldDeviceOptions(signal) {
		return getFieldDeviceOptions({ signal });
	},

	async getFieldDeviceOptionsForProject(projectId, signal) {
		return getFieldDeviceOptionsForProject(projectId, { signal });
	},

	async listSpsControllerSystemTypes(params, signal) {
		const res = await spsControllerSystemTypeRepository.list(
			{
				pagination: { page: 1, pageSize: params.limit ?? 50 },
				search: { text: params.search ?? '' }
			},
			signal
		);
		return res.items;
	},

	async getSpsControllerSystemType(id, signal) {
		return spsControllerSystemTypeRepository.get(id, signal);
	},

	async getAvailableApparatNumbers(spsControllerSystemTypeId, apparatId, systemPartId, signal) {
		return getAvailableApparatNumbersApi(spsControllerSystemTypeId, apparatId, systemPartId, {
			signal
		});
	},

	async multiCreateFieldDevices(fieldDevices, signal) {
		return api<MultiCreateFieldDeviceResponse>('/facility/field-devices/multi-create', {
			method: 'POST',
			body: JSON.stringify({ field_devices: fieldDevices }),
			signal
		});
	}
};
