import type {
	AvailableApparatNumbersResponse,
	CreateFieldDeviceRequest,
	FieldDeviceOptions,
	MultiCreateFieldDeviceResponse,
	SPSControllerSystemType
} from '$lib/domain/facility/index.js';
import type {
	FieldDeviceMultiCreateRepository,
	SPSControllerSystemTypeSearchParams
} from '$lib/domain/ports/facility/fieldDeviceMultiCreateRepository.js';

/**
 * Application use case for the Field Device multi-create flow.
 *
 * Framework-agnostic: safe to use from Svelte components, stores, etc.
 */
export class FieldDeviceMultiCreateUseCase {
	constructor(private repository: FieldDeviceMultiCreateRepository) {}

	getOptions(signal?: AbortSignal): Promise<FieldDeviceOptions> {
		return this.repository.getFieldDeviceOptions(signal);
	}

	searchSpsControllerSystemTypes(
		params: SPSControllerSystemTypeSearchParams,
		signal?: AbortSignal
	): Promise<SPSControllerSystemType[]> {
		return this.repository.listSpsControllerSystemTypes(params, signal);
	}

	getSpsControllerSystemType(id: string, signal?: AbortSignal): Promise<SPSControllerSystemType> {
		return this.repository.getSpsControllerSystemType(id, signal);
	}

	getAvailableApparatNumbers(
		spsControllerSystemTypeId: string,
		apparatId: string,
		systemPartId?: string,
		signal?: AbortSignal
	): Promise<AvailableApparatNumbersResponse> {
		return this.repository.getAvailableApparatNumbers(
			spsControllerSystemTypeId,
			apparatId,
			systemPartId,
			signal
		);
	}

	multiCreate(
		fieldDevices: CreateFieldDeviceRequest[],
		signal?: AbortSignal
	): Promise<MultiCreateFieldDeviceResponse> {
		return this.repository.multiCreateFieldDevices(fieldDevices, signal);
	}
}
