import type { CrudRepository } from '$lib/domain/ports/crudRepository.js';
import type {
	SPSController,
	CreateSPSControllerRequest,
	UpdateSPSControllerRequest,
	NextGADeviceResponse,
	SPSControllerSystemType,
	SPSControllerSystemTypeListParams,
	SPSControllerSystemTypeListResponse
} from '$lib/domain/facility/index.js';

export interface SPSControllerRepository extends CrudRepository<
	SPSController,
	CreateSPSControllerRequest,
	UpdateSPSControllerRequest
> {
	getBulk(ids: string[], signal?: AbortSignal): Promise<SPSController[]>;
	copy(id: string, signal?: AbortSignal): Promise<SPSController>;
	validate(
		data: {
			id?: string;
			control_cabinet_id: string;
			ga_device?: string;
			device_name: string;
			ip_address?: string;
			subnet?: string;
			gateway?: string;
			vlan?: string;
		},
		signal?: AbortSignal
	): Promise<void>;
	getNextGADevice(
		controlCabinetId: string,
		spsControllerId?: string,
		signal?: AbortSignal
	): Promise<NextGADeviceResponse>;

	// SPS Controller System Types (Sub-resource)
	listSystemTypes(
		params?: SPSControllerSystemTypeListParams,
		signal?: AbortSignal
	): Promise<SPSControllerSystemTypeListResponse>;
	getSystemType(id: string, signal?: AbortSignal): Promise<SPSControllerSystemType>;
}
