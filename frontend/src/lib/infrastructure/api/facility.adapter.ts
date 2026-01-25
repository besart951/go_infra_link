/**
 * Facility API adapters
 * Infrastructure layer - implements facility data operations via HTTP
 */
import { api, type ApiOptions } from '$lib/api/client.js';
import type {
	Building,
	BuildingListParams,
	BuildingListResponse,
	CreateBuildingRequest,
	UpdateBuildingRequest,
	ControlCabinet,
	ControlCabinetListParams,
	ControlCabinetListResponse,
	CreateControlCabinetRequest,
	UpdateControlCabinetRequest,
	SPSController,
	SPSControllerListParams,
	SPSControllerListResponse,
	CreateSPSControllerRequest,
	UpdateSPSControllerRequest,
	FieldDevice,
	FieldDeviceListParams,
	FieldDeviceListResponse,
	CreateFieldDeviceRequest,
	UpdateFieldDeviceRequest
} from '$lib/domain/facility/index.js';

// ============================================================================
// BUILDINGS
// ============================================================================

export async function listBuildings(
	params?: BuildingListParams,
	options?: ApiOptions
): Promise<BuildingListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	if (params?.search) searchParams.set('search', params.search);

	const query = searchParams.toString();
	return api<BuildingListResponse>(`/facility/buildings${query ? `?${query}` : ''}`, options);
}

export async function getBuilding(id: string, options?: ApiOptions): Promise<Building> {
	return api<Building>(`/facility/buildings/${id}`, options);
}

export async function createBuilding(
	data: CreateBuildingRequest,
	options?: ApiOptions
): Promise<Building> {
	return api<Building>('/facility/buildings', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
}

export async function updateBuilding(
	id: string,
	data: UpdateBuildingRequest,
	options?: ApiOptions
): Promise<Building> {
	return api<Building>(`/facility/buildings/${id}`, {
		...options,
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

export async function deleteBuilding(id: string, options?: ApiOptions): Promise<void> {
	return api<void>(`/facility/buildings/${id}`, { ...options, method: 'DELETE' });
}

// ============================================================================
// CONTROL CABINETS
// ============================================================================

export async function listControlCabinets(
	params?: ControlCabinetListParams,
	options?: RequestInit
): Promise<ControlCabinetListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	if (params?.search) searchParams.set('search', params.search);
	if (params?.building_id) searchParams.set('building_id', params.building_id);

	const query = searchParams.toString();
	return api<ControlCabinetListResponse>(
		`/facility/control-cabinets${query ? `?${query}` : ''}`,
		options
	);
}

export async function getControlCabinet(
	id: string,
	options?: RequestInit
): Promise<ControlCabinet> {
	return api<ControlCabinet>(`/facility/control-cabinets/${id}`, options);
}

export async function createControlCabinet(
	data: CreateControlCabinetRequest,
	options?: RequestInit
): Promise<ControlCabinet> {
	return api<ControlCabinet>('/facility/control-cabinets', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
}

export async function updateControlCabinet(
	id: string,
	data: UpdateControlCabinetRequest,
	options?: RequestInit
): Promise<ControlCabinet> {
	return api<ControlCabinet>(`/facility/control-cabinets/${id}`, {
		...options,
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

export async function deleteControlCabinet(id: string, options?: RequestInit): Promise<void> {
	return api<void>(`/facility/control-cabinets/${id}`, { ...options, method: 'DELETE' });
}

// ============================================================================
// SPS CONTROLLERS
// ============================================================================

export async function listSPSControllers(
	params?: SPSControllerListParams,
	options?: RequestInit
): Promise<SPSControllerListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	if (params?.search) searchParams.set('search', params.search);
	if (params?.control_cabinet_id) searchParams.set('control_cabinet_id', params.control_cabinet_id);

	const query = searchParams.toString();
	return api<SPSControllerListResponse>(
		`/facility/sps-controllers${query ? `?${query}` : ''}`,
		options
	);
}

export async function getSPSController(id: string, options?: RequestInit): Promise<SPSController> {
	return api<SPSController>(`/facility/sps-controllers/${id}`, options);
}

export async function createSPSController(
	data: CreateSPSControllerRequest,
	options?: RequestInit
): Promise<SPSController> {
	return api<SPSController>('/facility/sps-controllers', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
}

export async function updateSPSController(
	id: string,
	data: UpdateSPSControllerRequest,
	options?: RequestInit
): Promise<SPSController> {
	return api<SPSController>(`/facility/sps-controllers/${id}`, {
		...options,
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

export async function deleteSPSController(id: string, options?: RequestInit): Promise<void> {
	return api<void>(`/facility/sps-controllers/${id}`, { ...options, method: 'DELETE' });
}

// ============================================================================
// FIELD DEVICES
// ============================================================================

export async function listFieldDevices(
	params?: FieldDeviceListParams,
	options?: RequestInit
): Promise<FieldDeviceListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	if (params?.search) searchParams.set('search', params.search);
	if (params?.sps_controller_system_type_id) {
		searchParams.set('sps_controller_system_type_id', params.sps_controller_system_type_id);
	}

	const query = searchParams.toString();
	return api<FieldDeviceListResponse>(
		`/facility/field-devices${query ? `?${query}` : ''}`,
		options
	);
}

export async function getFieldDevice(id: string, options?: RequestInit): Promise<FieldDevice> {
	return api<FieldDevice>(`/facility/field-devices/${id}`, options);
}

export async function createFieldDevice(
	data: CreateFieldDeviceRequest,
	options?: RequestInit
): Promise<FieldDevice> {
	return api<FieldDevice>('/facility/field-devices', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
}

export async function updateFieldDevice(
	id: string,
	data: UpdateFieldDeviceRequest,
	options?: RequestInit
): Promise<FieldDevice> {
	return api<FieldDevice>(`/facility/field-devices/${id}`, {
		...options,
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

export async function deleteFieldDevice(id: string, options?: RequestInit): Promise<void> {
	return api<void>(`/facility/field-devices/${id}`, { ...options, method: 'DELETE' });
}

// Re-export all types
export type {
	Building,
	BuildingListParams,
	BuildingListResponse,
	CreateBuildingRequest,
	UpdateBuildingRequest,
	ControlCabinet,
	ControlCabinetListParams,
	ControlCabinetListResponse,
	CreateControlCabinetRequest,
	UpdateControlCabinetRequest,
	SPSController,
	SPSControllerListParams,
	SPSControllerListResponse,
	CreateSPSControllerRequest,
	UpdateSPSControllerRequest,
	FieldDevice,
	FieldDeviceListParams,
	FieldDeviceListResponse,
	CreateFieldDeviceRequest,
	UpdateFieldDeviceRequest
};
