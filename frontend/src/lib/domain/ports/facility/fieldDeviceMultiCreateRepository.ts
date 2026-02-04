import type {
	AvailableApparatNumbersResponse,
	CreateFieldDeviceRequest,
	FieldDeviceOptions,
	MultiCreateFieldDeviceResponse,
	SPSControllerSystemType
} from '$lib/domain/facility/index.js';

export interface SPSControllerSystemTypeSearchParams {
	search?: string;
	limit?: number;
}

/**
 * Port for the Field Device multi-create flow.
 *
 * UI should depend on this interface (domain port), not on infrastructure adapters.
 */
export interface FieldDeviceMultiCreateRepository {
	getFieldDeviceOptions(signal?: AbortSignal): Promise<FieldDeviceOptions>;
	getFieldDeviceOptionsForProject(
		projectId: string,
		signal?: AbortSignal
	): Promise<FieldDeviceOptions>;

	listSpsControllerSystemTypes(
		params: SPSControllerSystemTypeSearchParams,
		signal?: AbortSignal
	): Promise<SPSControllerSystemType[]>;

	getSpsControllerSystemType(id: string, signal?: AbortSignal): Promise<SPSControllerSystemType>;

	getAvailableApparatNumbers(
		spsControllerSystemTypeId: string,
		apparatId: string,
		systemPartId?: string,
		signal?: AbortSignal
	): Promise<AvailableApparatNumbersResponse>;

	multiCreateFieldDevices(
		fieldDevices: CreateFieldDeviceRequest[],
		signal?: AbortSignal
	): Promise<MultiCreateFieldDeviceResponse>;
}
