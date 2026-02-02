import type { FieldDeviceOptions } from '$lib/domain/facility/index.js';
import type { FieldDeviceOptionsRepository } from '$lib/domain/ports/facility/fieldDeviceOptionsRepository.js';

export class GetFieldDeviceOptionsUseCase {
	constructor(private repository: FieldDeviceOptionsRepository) {}

	execute(signal?: AbortSignal): Promise<FieldDeviceOptions> {
		return this.repository.getOptions(signal);
	}
}
