/**
 * Team API adapter
 * Infrastructure layer - implements team data operations via HTTP
 */
import { api } from '$lib/api/client.js';
import type {
	Team,
	TeamListParams,
	TeamListResponse,
	CreateTeamRequest,
	UpdateTeamRequest
} from '$lib/domain/team/index.js';

/**
 * List teams with optional filters
 */
export async function listTeams(
	params?: TeamListParams,
	options?: RequestInit
): Promise<TeamListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	if (params?.search) searchParams.set('search', params.search);

	const query = searchParams.toString();
	return api<TeamListResponse>(`/teams${query ? `?${query}` : ''}`, options);
}

/**
 * Get a single team by ID
 */
export async function getTeam(id: string, options?: RequestInit): Promise<Team> {
	return api<Team>(`/teams/${id}`, options);
}

/**
 * Create a new team
 */
export async function createTeam(data: CreateTeamRequest, options?: RequestInit): Promise<Team> {
	return api<Team>('/teams', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
}

/**
 * Update an existing team
 */
export async function updateTeam(
	id: string,
	data: UpdateTeamRequest,
	options?: RequestInit
): Promise<Team> {
	return api<Team>(`/teams/${id}`, {
		...options,
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

/**
 * Delete a team
 */
export async function deleteTeam(id: string, options?: RequestInit): Promise<void> {
	return api<void>(`/teams/${id}`, {
		...options,
		method: 'DELETE'
	});
}

// Re-export types for convenience
export type { Team, TeamListParams, TeamListResponse, CreateTeamRequest, UpdateTeamRequest };
