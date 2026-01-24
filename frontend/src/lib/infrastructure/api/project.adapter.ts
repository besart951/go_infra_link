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
	UpdateProjectRequest
} from '$lib/domain/project/index.js';

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

// Re-export types for convenience
export type {
	Project,
	ProjectListParams,
	ProjectListResponse,
	CreateProjectRequest,
	UpdateProjectRequest
};
