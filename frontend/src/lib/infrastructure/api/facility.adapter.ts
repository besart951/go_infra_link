/**
 * Facility API adapters
 * Infrastructure layer - implements facility data operations via HTTP
 */
import { api, type ApiOptions } from '$lib/api/client.js';
import type {
	Building,
	BuildingListParams,
	BuildingListResponse,
	BuildingBulkRequest,
	BuildingBulkResponse,
	CreateBuildingRequest,
	UpdateBuildingRequest,
	ControlCabinet,
	ControlCabinetListParams,
	ControlCabinetListResponse,
	ControlCabinetBulkRequest,
	ControlCabinetBulkResponse,
	ControlCabinetDeleteImpact,
	CreateControlCabinetRequest,
	UpdateControlCabinetRequest,
	SPSController,
	SPSControllerListParams,
	SPSControllerListResponse,
	SPSControllerBulkRequest,
	SPSControllerBulkResponse,
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
	ApparatBulkRequest,
	ApparatBulkResponse,
	CreateApparatRequest,
	UpdateApparatRequest,
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
	SPSControllerSystemTypeListResponse,
	BacnetObject,
	CreateBacnetObjectRequest,
	UpdateBacnetObjectRequest
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

export async function getBuildings(
	ids: BuildingBulkRequest['ids'],
	options?: ApiOptions
): Promise<BuildingBulkResponse> {
	return api<BuildingBulkResponse>('/facility/buildings/bulk', {
		...options,
		method: 'POST',
		body: JSON.stringify({ ids })
	});
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

export async function validateBuilding(
	data: { id?: string; iws_code: string; building_group: number },
	options?: ApiOptions
): Promise<void> {
	return api<void>('/facility/buildings/validate', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
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

export async function getControlCabinets(
	ids: ControlCabinetBulkRequest['ids'],
	options?: ApiOptions
): Promise<ControlCabinetBulkResponse> {
	return api<ControlCabinetBulkResponse>('/facility/control-cabinets/bulk', {
		...options,
		method: 'POST',
		body: JSON.stringify({ ids })
	});
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

export async function validateControlCabinet(
	data: { id?: string; building_id: string; control_cabinet_nr?: string },
	options?: ApiOptions
): Promise<void> {
	return api<void>('/facility/control-cabinets/validate', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
}

export async function getControlCabinetDeleteImpact(
	id: string,
	options?: RequestInit
): Promise<ControlCabinetDeleteImpact> {
	return api<ControlCabinetDeleteImpact>(`/facility/control-cabinets/${id}/delete-impact`, options);
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

export async function getNextSPSControllerGADevice(
	controlCabinetId: string,
	excludeId?: string,
	options?: ApiOptions
): Promise<import('$lib/domain/facility/sps-controller.js').NextGADeviceResponse> {
	const searchParams = new URLSearchParams();
	searchParams.set('control_cabinet_id', controlCabinetId);
	if (excludeId) searchParams.set('exclude_id', excludeId);
	const query = searchParams.toString();
	return api<import('$lib/domain/facility/sps-controller.js').NextGADeviceResponse>(
		`/facility/sps-controllers/next-ga-device?${query}`,
		options
	);
}

export async function getSPSController(id: string, options?: RequestInit): Promise<SPSController> {
	return api<SPSController>(`/facility/sps-controllers/${id}`, options);
}

export async function getSPSControllers(
	ids: SPSControllerBulkRequest['ids'],
	options?: ApiOptions
): Promise<SPSControllerBulkResponse> {
	return api<SPSControllerBulkResponse>('/facility/sps-controllers/bulk', {
		...options,
		method: 'POST',
		body: JSON.stringify({ ids })
	});
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

export async function validateSPSController(
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
	options?: ApiOptions
): Promise<void> {
	return api<void>('/facility/sps-controllers/validate', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
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

export async function getFieldDeviceOptions(
	options?: ApiOptions
): Promise<import('$lib/domain/facility/field-device.js').FieldDeviceOptions> {
	return api<import('$lib/domain/facility/field-device.js').FieldDeviceOptions>(
		'/facility/field-devices/options',
		options
	);
}

export async function getFieldDeviceOptionsForProject(
	projectId: string,
	options?: ApiOptions
): Promise<import('$lib/domain/facility/field-device.js').FieldDeviceOptions> {
	return api<import('$lib/domain/facility/field-device.js').FieldDeviceOptions>(
		`/projects/${projectId}/field-device-options`,
		options
	);
}

export async function getAvailableApparatNumbers(
	spsControllerSystemTypeId: string,
	apparatId: string,
	systemPartId?: string,
	options?: ApiOptions
): Promise<import('$lib/domain/facility/field-device.js').AvailableApparatNumbersResponse> {
	const searchParams = new URLSearchParams();
	searchParams.set('sps_controller_system_type_id', spsControllerSystemTypeId);
	searchParams.set('apparat_id', apparatId);
	if (systemPartId) {
		searchParams.set('system_part_id', systemPartId);
	}

	return api<import('$lib/domain/facility/field-device.js').AvailableApparatNumbersResponse>(
		`/facility/field-devices/available-apparat-nr?${searchParams.toString()}`,
		options
	);
}

export async function multiCreateFieldDevices(
	data: import('$lib/domain/facility/field-device.js').MultiCreateFieldDeviceRequest,
	options?: ApiOptions
): Promise<import('$lib/domain/facility/field-device.js').MultiCreateFieldDeviceResponse> {
	return api<import('$lib/domain/facility/field-device.js').MultiCreateFieldDeviceResponse>(
		'/facility/field-devices/multi-create',
		{
			...options,
			method: 'POST',
			body: JSON.stringify(data)
		}
	);
}

export async function bulkUpdateFieldDevices(
	data: import('$lib/domain/facility/field-device.js').BulkUpdateFieldDeviceRequest,
	options?: ApiOptions
): Promise<import('$lib/domain/facility/field-device.js').BulkUpdateFieldDeviceResponse> {
	return api<import('$lib/domain/facility/field-device.js').BulkUpdateFieldDeviceResponse>(
		'/facility/field-devices/bulk-update',
		{
			...options,
			method: 'PATCH',
			body: JSON.stringify(data)
		}
	);
}

export async function bulkDeleteFieldDevices(
	ids: string[],
	options?: ApiOptions
): Promise<import('$lib/domain/facility/field-device.js').BulkDeleteFieldDeviceResponse> {
	return api<import('$lib/domain/facility/field-device.js').BulkDeleteFieldDeviceResponse>(
		'/facility/field-devices/bulk-delete',
		{
			...options,
			method: 'DELETE',
			body: JSON.stringify({ ids })
		}
	);
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
	if (params?.apparat_id) searchParams.set('apparat_id', params.apparat_id);
	if (params?.object_data_id) searchParams.set('object_data_id', params.object_data_id);

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
	if (params?.object_data_id) searchParams.set('object_data_id', params.object_data_id);
	if (params?.system_part_id) searchParams.set('system_part_id', params.system_part_id);

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

export async function getApparats(
	ids: ApparatBulkRequest['ids'],
	options?: ApiOptions
): Promise<ApparatBulkResponse> {
	return api<ApparatBulkResponse>('/facility/apparats/bulk', {
		...options,
		method: 'POST',
		body: JSON.stringify({ ids })
	});
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
	if (params?.apparat_id) searchParams.set('apparat_id', params.apparat_id);
	if (params?.system_part_id) searchParams.set('system_part_id', params.system_part_id);

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

export async function getObjectDataBacnetObjects(
	id: string,
	options?: ApiOptions
): Promise<BacnetObject[]> {
	return api<BacnetObject[]>(`/facility/object-data/${id}/bacnet-objects`, options);
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
	if (params?.sps_controller_id) {
		searchParams.set('sps_controller_id', params.sps_controller_id);
	}

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

// ============================================================================
// BACNET OBJECTS
// ============================================================================

export async function listBacnetObjects(options?: ApiOptions): Promise<BacnetObject[]> {
	return api<BacnetObject[]>('/facility/bacnet-objects', options);
}

export async function getBacnetObject(id: string, options?: ApiOptions): Promise<BacnetObject> {
	return api<BacnetObject>(`/facility/bacnet-objects/${id}`, options);
}

export async function createBacnetObject(
	data: CreateBacnetObjectRequest,
	options?: ApiOptions
): Promise<BacnetObject> {
	return api<BacnetObject>('/facility/bacnet-objects', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
}

export async function updateBacnetObject(
	id: string,
	data: UpdateBacnetObjectRequest,
	options?: ApiOptions
): Promise<BacnetObject> {
	return api<BacnetObject>(`/facility/bacnet-objects/${id}`, {
		...options,
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

export async function deleteBacnetObject(id: string, options?: ApiOptions): Promise<void> {
	return api<void>(`/facility/bacnet-objects/${id}`, { ...options, method: 'DELETE' });
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
	SPSControllerSystemTypeListResponse,
	BacnetObject,
	CreateBacnetObjectRequest,
	UpdateBacnetObjectRequest
};
