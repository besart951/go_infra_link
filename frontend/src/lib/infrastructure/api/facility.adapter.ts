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
	UpdateFieldDeviceRequest,
	SystemType,
	SystemTypeListParams,
	SystemTypeListResponse,
	CreateSystemTypeRequest,
	UpdateSystemTypeRequest,
	SystemPart,
	SystemPartListParams,
	SystemPartListResponse,
	CreateSystemPartRequest,
	UpdateSystemPartRequest,
	Apparat,
	ApparatListParams,
	ApparatListResponse,
	CreateApparatRequest,
	UpdateApparatRequest,
	Specification,
	SpecificationListParams,
	SpecificationListResponse,
	CreateSpecificationRequest,
	UpdateSpecificationRequest,
	StateText,
	StateTextListParams,
	StateTextListResponse,
	CreateStateTextRequest,
	UpdateStateTextRequest,
	NotificationClass,
	NotificationClassListParams,
	NotificationClassListResponse,
	CreateNotificationClassRequest,
	UpdateNotificationClassRequest,
	AlarmDefinition,
	AlarmDefinitionListParams,
	AlarmDefinitionListResponse,
	CreateAlarmDefinitionRequest,
	UpdateAlarmDefinitionRequest,
	ObjectData,
	ObjectDataListParams,
	ObjectDataListResponse,
	CreateObjectDataRequest,
	UpdateObjectDataRequest,
	SPSControllerSystemType,
	SPSControllerSystemTypeListParams,
	SPSControllerSystemTypeListResponse
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
	options?: ApiOptions
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
	options?: ApiOptions
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
	options?: ApiOptions
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

// ============================================================================
// SYSTEM TYPES
// ============================================================================

export async function listSystemTypes(
	params?: SystemTypeListParams,
	options?: ApiOptions
): Promise<SystemTypeListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	if (params?.search) searchParams.set('search', params.search);

	const query = searchParams.toString();
	return api<SystemTypeListResponse>(`/facility/system-types${query ? `?${query}` : ''}`, options);
}

export async function createSystemType(
	data: CreateSystemTypeRequest,
	options?: ApiOptions
): Promise<SystemType> {
	return api<SystemType>('/facility/system-types', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
}

export async function updateSystemType(
	id: string,
	data: UpdateSystemTypeRequest,
	options?: ApiOptions
): Promise<SystemType> {
	return api<SystemType>(`/facility/system-types/${id}`, {
		...options,
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

export async function deleteSystemType(id: string, options?: ApiOptions): Promise<void> {
	return api<void>(`/facility/system-types/${id}`, { ...options, method: 'DELETE' });
}

export async function getSystemType(id: string, options?: ApiOptions): Promise<SystemType> {
	return api<SystemType>(`/facility/system-types/${id}`, options);
}

// ============================================================================
// SYSTEM PARTS
// ============================================================================

export async function listSystemParts(
	params?: SystemPartListParams,
	options?: ApiOptions
): Promise<SystemPartListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	if (params?.search) searchParams.set('search', params.search);

	const query = searchParams.toString();
	return api<SystemPartListResponse>(`/facility/system-parts${query ? `?${query}` : ''}`, options);
}

export async function createSystemPart(
	data: CreateSystemPartRequest,
	options?: ApiOptions
): Promise<SystemPart> {
	return api<SystemPart>('/facility/system-parts', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
}

export async function updateSystemPart(
	id: string,
	data: UpdateSystemPartRequest,
	options?: ApiOptions
): Promise<SystemPart> {
	return api<SystemPart>(`/facility/system-parts/${id}`, {
		...options,
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

export async function deleteSystemPart(id: string, options?: ApiOptions): Promise<void> {
	return api<void>(`/facility/system-parts/${id}`, { ...options, method: 'DELETE' });
}

export async function getSystemPart(id: string, options?: ApiOptions): Promise<SystemPart> {
	return api<SystemPart>(`/facility/system-parts/${id}`, options);
}

// ============================================================================
// APPARATS
// ============================================================================

export async function listApparats(
	params?: ApparatListParams,
	options?: ApiOptions
): Promise<ApparatListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	if (params?.search) searchParams.set('search', params.search);

	const query = searchParams.toString();
	return api<ApparatListResponse>(`/facility/apparats${query ? `?${query}` : ''}`, options);
}

export async function createApparat(
	data: CreateApparatRequest,
	options?: ApiOptions
): Promise<Apparat> {
	return api<Apparat>('/facility/apparats', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
}

export async function updateApparat(
	id: string,
	data: UpdateApparatRequest,
	options?: ApiOptions
): Promise<Apparat> {
	return api<Apparat>(`/facility/apparats/${id}`, {
		...options,
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

export async function deleteApparat(id: string, options?: ApiOptions): Promise<void> {
	return api<void>(`/facility/apparats/${id}`, { ...options, method: 'DELETE' });
}

export async function getApparat(id: string, options?: ApiOptions): Promise<Apparat> {
	return api<Apparat>(`/facility/apparats/${id}`, options);
}

// ============================================================================
// SPECIFICATIONS
// ============================================================================

export async function listSpecifications(
	params?: SpecificationListParams,
	options?: ApiOptions
): Promise<SpecificationListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	if (params?.search) searchParams.set('search', params.search);

	const query = searchParams.toString();
	return api<SpecificationListResponse>(
		`/facility/specifications${query ? `?${query}` : ''}`,
		options
	);
}

export async function createSpecification(
	data: CreateSpecificationRequest,
	options?: ApiOptions
): Promise<Specification> {
	return api<Specification>('/facility/specifications', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
}

export async function updateSpecification(
	id: string,
	data: UpdateSpecificationRequest,
	options?: ApiOptions
): Promise<Specification> {
	return api<Specification>(`/facility/specifications/${id}`, {
		...options,
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

export async function deleteSpecification(id: string, options?: ApiOptions): Promise<void> {
	return api<void>(`/facility/specifications/${id}`, { ...options, method: 'DELETE' });
}

export async function getSpecification(id: string, options?: ApiOptions): Promise<Specification> {
	return api<Specification>(`/facility/specifications/${id}`, options);
}

// ============================================================================
// STATE TEXTS
// ============================================================================

export async function listStateTexts(
	params?: StateTextListParams,
	options?: ApiOptions
): Promise<StateTextListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	if (params?.search) searchParams.set('search', params.search);

	const query = searchParams.toString();
	return api<StateTextListResponse>(`/facility/state-texts${query ? `?${query}` : ''}`, options);
}

export async function createStateText(
	data: CreateStateTextRequest,
	options?: ApiOptions
): Promise<StateText> {
	return api<StateText>('/facility/state-texts', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
}

export async function updateStateText(
	id: string,
	data: UpdateStateTextRequest,
	options?: ApiOptions
): Promise<StateText> {
	return api<StateText>(`/facility/state-texts/${id}`, {
		...options,
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

export async function deleteStateText(id: string, options?: ApiOptions): Promise<void> {
	return api<void>(`/facility/state-texts/${id}`, { ...options, method: 'DELETE' });
}

export async function getStateText(id: string, options?: ApiOptions): Promise<StateText> {
	return api<StateText>(`/facility/state-texts/${id}`, options);
}

// ============================================================================
// NOTIFICATION CLASSES
// ============================================================================

export async function listNotificationClasses(
	params?: NotificationClassListParams,
	options?: ApiOptions
): Promise<NotificationClassListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	if (params?.search) searchParams.set('search', params.search);

	const query = searchParams.toString();
	return api<NotificationClassListResponse>(
		`/facility/notification-classes${query ? `?${query}` : ''}`,
		options
	);
}

export async function createNotificationClass(
	data: CreateNotificationClassRequest,
	options?: ApiOptions
): Promise<NotificationClass> {
	return api<NotificationClass>('/facility/notification-classes', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
}

export async function updateNotificationClass(
	id: string,
	data: UpdateNotificationClassRequest,
	options?: ApiOptions
): Promise<NotificationClass> {
	return api<NotificationClass>(`/facility/notification-classes/${id}`, {
		...options,
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

export async function deleteNotificationClass(id: string, options?: ApiOptions): Promise<void> {
	return api<void>(`/facility/notification-classes/${id}`, { ...options, method: 'DELETE' });
}

export async function getNotificationClass(
	id: string,
	options?: ApiOptions
): Promise<NotificationClass> {
	return api<NotificationClass>(`/facility/notification-classes/${id}`, options);
}

// ============================================================================
// ALARM DEFINITIONS
// ============================================================================

export async function listAlarmDefinitions(
	params?: AlarmDefinitionListParams,
	options?: ApiOptions
): Promise<AlarmDefinitionListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	if (params?.search) searchParams.set('search', params.search);

	const query = searchParams.toString();
	return api<AlarmDefinitionListResponse>(
		`/facility/alarm-definitions${query ? `?${query}` : ''}`,
		options
	);
}

export async function createAlarmDefinition(
	data: CreateAlarmDefinitionRequest,
	options?: ApiOptions
): Promise<AlarmDefinition> {
	return api<AlarmDefinition>('/facility/alarm-definitions', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
}

export async function updateAlarmDefinition(
	id: string,
	data: UpdateAlarmDefinitionRequest,
	options?: ApiOptions
): Promise<AlarmDefinition> {
	return api<AlarmDefinition>(`/facility/alarm-definitions/${id}`, {
		...options,
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

export async function deleteAlarmDefinition(id: string, options?: ApiOptions): Promise<void> {
	return api<void>(`/facility/alarm-definitions/${id}`, { ...options, method: 'DELETE' });
}

export async function getAlarmDefinition(
	id: string,
	options?: ApiOptions
): Promise<AlarmDefinition> {
	return api<AlarmDefinition>(`/facility/alarm-definitions/${id}`, options);
}

// ============================================================================
// OBJECT DATA
// ============================================================================

export async function listObjectData(
	params?: ObjectDataListParams,
	options?: ApiOptions
): Promise<ObjectDataListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	if (params?.search) searchParams.set('search', params.search);

	const query = searchParams.toString();
	return api<ObjectDataListResponse>(`/facility/object-data${query ? `?${query}` : ''}`, options);
}

export async function createObjectData(
	data: CreateObjectDataRequest,
	options?: ApiOptions
): Promise<ObjectData> {
	return api<ObjectData>('/facility/object-data', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
}

export async function updateObjectData(
	id: string,
	data: UpdateObjectDataRequest,
	options?: ApiOptions
): Promise<ObjectData> {
	return api<ObjectData>(`/facility/object-data/${id}`, {
		...options,
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

export async function deleteObjectData(id: string, options?: ApiOptions): Promise<void> {
	return api<void>(`/facility/object-data/${id}`, { ...options, method: 'DELETE' });
}

export async function getObjectData(id: string, options?: ApiOptions): Promise<ObjectData> {
	return api<ObjectData>(`/facility/object-data/${id}`, options);
}

// ============================================================================
// SPS CONTROLLER SYSTEM TYPES
// ============================================================================

export async function listSPSControllerSystemTypes(
	params?: SPSControllerSystemTypeListParams,
	options?: ApiOptions
): Promise<SPSControllerSystemTypeListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	if (params?.search) searchParams.set('search', params.search);

	const query = searchParams.toString();
	return api<SPSControllerSystemTypeListResponse>(
		`/facility/sps-controller-system-types${query ? `?${query}` : ''}`,
		options
	);
}

export async function getSPSControllerSystemType(
	id: string,
	options?: ApiOptions
): Promise<SPSControllerSystemType> {
	return api<SPSControllerSystemType>(`/facility/sps-controller-system-types/${id}`, options);
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
	UpdateFieldDeviceRequest,
	SystemType,
	SystemTypeListParams,
	SystemTypeListResponse,
	SystemPart,
	SystemPartListParams,
	SystemPartListResponse,
	Apparat,
	ApparatListParams,
	ApparatListResponse,
	Specification,
	SpecificationListParams,
	SpecificationListResponse,
	StateText,
	StateTextListParams,
	StateTextListResponse,
	NotificationClass,
	NotificationClassListParams,
	NotificationClassListResponse,
	AlarmDefinition,
	AlarmDefinitionListParams,
	AlarmDefinitionListResponse,
	ObjectData,
	ObjectDataListParams,
	ObjectDataListResponse,
	SPSControllerSystemType,
	SPSControllerSystemTypeListParams,
	SPSControllerSystemTypeListResponse
};
