import type { FieldDeviceMultiCreateRepository } from '$lib/domain/ports/facility/fieldDeviceMultiCreateRepository.js';
import {
	getAvailableApparatNumbers,
	getFieldDeviceOptions,
	getFieldDeviceOptionsForProject,
	getSPSControllerSystemType,
	listSPSControllerSystemTypes,
	multiCreateFieldDevices
} from '$lib/infrastructure/api/facility.adapter.js';

export const facilityFieldDeviceMultiCreateRepository: FieldDeviceMultiCreateRepository = {
	async getFieldDeviceOptions(signal) {
		return getFieldDeviceOptions({ signal });
	},

	async getFieldDeviceOptionsForProject(projectId, signal) {
		return getFieldDeviceOptionsForProject(projectId, { signal });
	},

	async listSpsControllerSystemTypes(params, signal) {
		const res = await listSPSControllerSystemTypes(
			{ page: 1, limit: params.limit ?? 50, search: params.search ?? '' },
			{ signal }
		);
		return res.items;
	},

	async getSpsControllerSystemType(id, signal) {
		return getSPSControllerSystemType(id, { signal });
	},

	async getAvailableApparatNumbers(spsControllerSystemTypeId, apparatId, systemPartId, signal) {
		return getAvailableApparatNumbers(spsControllerSystemTypeId, apparatId, systemPartId, {
			signal
		});
	},

	async multiCreateFieldDevices(fieldDevices, signal) {
		return multiCreateFieldDevices({ field_devices: fieldDevices }, { signal });
	}
};
