import type { FieldDeviceOptionsRepository } from '$lib/domain/ports/facility/fieldDeviceOptionsRepository.js';
import { getFieldDeviceOptions } from '$lib/infrastructure/api/facility.adapter.js';

export const facilityFieldDeviceOptionsRepository: FieldDeviceOptionsRepository = {
	async getOptions(signal) {
		return getFieldDeviceOptions({ signal });
	}
};
