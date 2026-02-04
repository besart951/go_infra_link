import type { FieldDeviceOptionsRepository } from '$lib/domain/ports/facility/fieldDeviceOptionsRepository.js';
import {
	getFieldDeviceOptions,
	getFieldDeviceOptionsForProject
} from '$lib/infrastructure/api/facility.adapter.js';

export const facilityFieldDeviceOptionsRepository: FieldDeviceOptionsRepository = {
	async getOptions(signal) {
		return getFieldDeviceOptions({ signal });
	},
	async getOptionsForProject(projectId, signal) {
		return getFieldDeviceOptionsForProject(projectId, { signal });
	}
};
