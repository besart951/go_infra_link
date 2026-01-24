import { api } from './client.js';

export interface Team {
	id: string;
	name: string;
	description?: string | null;
	created_at: string;
	updated_at: string;
}

export interface TeamListResponse {
	items: Team[];
	total: number;
	page: number;
	total_pages: number;
}

export interface TeamMember {
	team_id: string;
	user_id: string;
	role: string;
	joined_at: string;
}

export interface TeamMemberListResponse {
	items: TeamMember[];
	total: number;
	page: number;
	total_pages: number;
}

export interface CreateTeamRequest {
	name: string;
	description?: string | null;
}

export interface AddTeamMemberRequest {
	user_id: string;
	role: 'member' | 'manager' | 'owner';
}

/**
 * List all teams
 * CSRF token is automatically included by the api() client
 */
export async function listTeams(
	params: { page?: number; limit?: number; search?: string } = {}
): Promise<TeamListResponse> {
	const sp = new URLSearchParams();
	if (params.page) sp.set('page', String(params.page));
	if (params.limit) sp.set('limit', String(params.limit));
	if (params.search) sp.set('search', params.search);
	const q = sp.toString();
	return api<TeamListResponse>(q ? `/teams?${q}` : '/teams');
}

/**
 * Get a single team by ID
 */
export async function getTeam(teamId: string): Promise<Team> {
	return api<Team>(`/teams/${teamId}`);
}

/**
 * Create a new team
 * CSRF token is automatically included
 */
export async function createTeam(req: CreateTeamRequest): Promise<Team> {
	return api<Team>('/teams', {
		method: 'POST',
		body: JSON.stringify({
			name: req.name,
			description: req.description ?? null
		})
	});
}

/**
 * Update a team
 * CSRF token is automatically included
 */
export async function updateTeam(teamId: string, req: Partial<CreateTeamRequest>): Promise<Team> {
	return api<Team>(`/teams/${teamId}`, {
		method: 'PATCH',
		body: JSON.stringify(req)
	});
}

/**
 * Delete a team
 * CSRF token is automatically included
 */
export async function deleteTeam(teamId: string): Promise<void> {
	return api<void>(`/teams/${teamId}`, { method: 'DELETE' });
}

/**
 * List team members
 */
export async function listTeamMembers(
	teamId: string,
	params: { page?: number; limit?: number } = {}
): Promise<TeamMemberListResponse> {
	const sp = new URLSearchParams();
	if (params.page) sp.set('page', String(params.page));
	if (params.limit) sp.set('limit', String(params.limit));
	const q = sp.toString();
	return api<TeamMemberListResponse>(
		q ? `/teams/${teamId}/members?${q}` : `/teams/${teamId}/members`
	);
}

/**
 * Add a member to a team
 * CSRF token is automatically included
 */
export async function addTeamMember(teamId: string, req: AddTeamMemberRequest): Promise<void> {
	return api<void>(`/teams/${teamId}/members`, {
		method: 'POST',
		body: JSON.stringify(req)
	});
}

/**
 * Remove a member from a team
 * CSRF token is automatically included
 */
export async function removeTeamMember(teamId: string, userId: string): Promise<void> {
	return api<void>(`/teams/${teamId}/members/${userId}`, {
		method: 'DELETE'
	});
}
