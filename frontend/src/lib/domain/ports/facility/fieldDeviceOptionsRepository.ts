import type { FieldDeviceOptions } from '$lib/domain/facility/index.js';

/**
 * Port for retrieving FieldDevice options metadata.
 */
export interface FieldDeviceOptionsRepository {
	getOptions(signal?: AbortSignal): Promise<FieldDeviceOptions>;
	getOptionsForProject(projectId: string, signal?: AbortSignal): Promise<FieldDeviceOptions>;
}
