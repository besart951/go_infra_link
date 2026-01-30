/**
 * Project API adapter
 * Infrastructure layer - implements project data operations via HTTP
 */
import { api } from '$lib/api/client.js';
import type {
	Project,
	ProjectListParams,
	ProjectListResponse,
	CreateProjectRequest,
	UpdateProjectRequest,
	ProjectUserListResponse,
	ProjectObjectDataListResponse,
	ProjectControlCabinetListResponse,
	ProjectSPSControllerListResponse,
	ProjectFieldDeviceListResponse
} from '$lib/domain/project/index.js';
import type { ObjectDataListParams } from '$lib/domain/facility/index.js';

/**
 * List projects with optional filters
 */
export async function listProjects(
	params?: ProjectListParams,
	options?: RequestInit
): Promise<ProjectListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	if (params?.search) searchParams.set('search', params.search);
	if (params?.status) searchParams.set('status', params.status);

	const query = searchParams.toString();
	const endpoint = `/projects${query ? `?${query}` : ''}`;

	return api<ProjectListResponse>(endpoint, options);
}

/**
 * Get a single project by ID
 */
export async function getProject(id: string, options?: RequestInit): Promise<Project> {
	return api<Project>(`/projects/${id}`, options);
}

/**
 * Create a new project
 */
export async function createProject(
	data: CreateProjectRequest,
	options?: RequestInit
): Promise<Project> {
	return api<Project>('/projects', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
}

/**
 * Update an existing project
 */
export async function updateProject(
	id: string,
	data: UpdateProjectRequest,
	options?: RequestInit
): Promise<Project> {
	return api<Project>(`/projects/${id}`, {
		...options,
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

/**
 * Delete a project
 */
export async function deleteProject(id: string, options?: RequestInit): Promise<void> {
	return api<void>(`/projects/${id}`, {
		...options,
		method: 'DELETE'
	});
}

/**
 * List users in a project
 */
export async function listProjectUsers(
	projectId: string,
	options?: RequestInit
): Promise<ProjectUserListResponse> {
	return api<ProjectUserListResponse>(`/projects/${projectId}/users`, options);
}

/**
 * Add a user to a project
 */
export async function addProjectUser(
	projectId: string,
	userId: string,
	options?: RequestInit
): Promise<void> {
	return api<void>(`/projects/${projectId}/users`, {
		...options,
		method: 'POST',
		body: JSON.stringify({ user_id: userId })
	});
}

/**
 * Remove a user from a project
 */
export async function removeProjectUser(
	projectId: string,
	userId: string,
	options?: RequestInit
): Promise<void> {
	return api<void>(`/projects/${projectId}/users/${userId}`, {
		...options,
		method: 'DELETE'
	});
}

/**
 * List project object data
 */
export async function listProjectObjectData(
	projectId: string,
	params?: ObjectDataListParams,
	options?: RequestInit
): Promise<ProjectObjectDataListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	if (params?.search) searchParams.set('search', params.search);
	if (params?.apparat_id) searchParams.set('apparat_id', params.apparat_id);
	if (params?.system_part_id) searchParams.set('system_part_id', params.system_part_id);

	const query = searchParams.toString();
	return api<ProjectObjectDataListResponse>(
		`/projects/${projectId}/object-data${query ? `?${query}` : ''}`,
		options
	);
}

/**
 * Attach object data to a project
 */
export async function addProjectObjectData(
	projectId: string,
	objectDataId: string,
	options?: RequestInit
): Promise<void> {
	return api<void>(`/projects/${projectId}/object-data`, {
		...options,
		method: 'POST',
		body: JSON.stringify({ object_data_id: objectDataId })
	});
}

/**
 * Detach object data from a project
 */
export async function removeProjectObjectData(
	projectId: string,
	objectDataId: string,
	options?: RequestInit
): Promise<void> {
	return api<void>(`/projects/${projectId}/object-data/${objectDataId}`, {
		...options,
		method: 'DELETE'
	});
}

// ============================================================================
// PROJECT CONTROL CABINETS
// ==========================================================================

export async function listProjectControlCabinets(
	projectId: string,
	params?: { page?: number; limit?: number },
	options?: RequestInit
): Promise<ProjectControlCabinetListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	const query = searchParams.toString();
	return api<ProjectControlCabinetListResponse>(
		`/projects/${projectId}/control-cabinets${query ? `?${query}` : ''}`,
		options
	);
}

export async function addProjectControlCabinet(
	projectId: string,
	controlCabinetId: string,
	options?: RequestInit
): Promise<void> {
	return api<void>(`/projects/${projectId}/control-cabinets`, {
		...options,
		method: 'POST',
		body: JSON.stringify({ control_cabinet_id: controlCabinetId })
	});
}

export async function removeProjectControlCabinet(
	projectId: string,
	linkId: string,
	options?: RequestInit
): Promise<void> {
	return api<void>(`/projects/${projectId}/control-cabinets/${linkId}`, {
		...options,
		method: 'DELETE'
	});
}

// ============================================================================
// PROJECT SPS CONTROLLERS
// ==========================================================================

export async function listProjectSPSControllers(
	projectId: string,
	params?: { page?: number; limit?: number },
	options?: RequestInit
): Promise<ProjectSPSControllerListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	const query = searchParams.toString();
	return api<ProjectSPSControllerListResponse>(
		`/projects/${projectId}/sps-controllers${query ? `?${query}` : ''}`,
		options
	);
}

export async function addProjectSPSController(
	projectId: string,
	spsControllerId: string,
	options?: RequestInit
): Promise<void> {
	return api<void>(`/projects/${projectId}/sps-controllers`, {
		...options,
		method: 'POST',
		body: JSON.stringify({ sps_controller_id: spsControllerId })
	});
}

export async function removeProjectSPSController(
	projectId: string,
	linkId: string,
	options?: RequestInit
): Promise<void> {
	return api<void>(`/projects/${projectId}/sps-controllers/${linkId}`, {
		...options,
		method: 'DELETE'
	});
}

// ============================================================================
// PROJECT FIELD DEVICES
// ==========================================================================

export async function listProjectFieldDevices(
	projectId: string,
	params?: { page?: number; limit?: number },
	options?: RequestInit
): Promise<ProjectFieldDeviceListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	const query = searchParams.toString();
	return api<ProjectFieldDeviceListResponse>(
		`/projects/${projectId}/field-devices${query ? `?${query}` : ''}`,
		options
	);
}

export async function addProjectFieldDevice(
	projectId: string,
	fieldDeviceId: string,
	options?: RequestInit
): Promise<void> {
	return api<void>(`/projects/${projectId}/field-devices`, {
		...options,
		method: 'POST',
		body: JSON.stringify({ field_device_id: fieldDeviceId })
	});
}

export async function removeProjectFieldDevice(
	projectId: string,
	linkId: string,
	options?: RequestInit
): Promise<void> {
	return api<void>(`/projects/${projectId}/field-devices/${linkId}`, {
		...options,
		method: 'DELETE'
	});
}

// Re-export types for convenience
export type {
	Project,
	ProjectListParams,
	ProjectListResponse,
	CreateProjectRequest,
	UpdateProjectRequest,
	ProjectUserListResponse,
	ProjectObjectDataListResponse,
	ProjectControlCabinetListResponse,
	ProjectSPSControllerListResponse,
	ProjectFieldDeviceListResponse
};
