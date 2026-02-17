import type { FieldDeviceOptionsRepository } from '$lib/domain/ports/facility/fieldDeviceOptionsRepository.js';
import type { FieldDeviceOptions } from '$lib/domain/facility/field-device.js';
import { api } from '$lib/api/client.js';

export const facilityFieldDeviceOptionsRepository: FieldDeviceOptionsRepository = {
	async getOptions(signal) {
		return api<FieldDeviceOptions>('/facility/field-devices/options', { signal });
	},
	async getOptionsForProject(projectId, signal) {
		return api<FieldDeviceOptions>(`/projects/${projectId}/field-device-options`, { signal });
	}
};
