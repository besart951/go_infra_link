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

type ApiError = { error: string; message?: string };

const API_BASE = '/api/v1';

function getCookie(name: string): string | undefined {
	if (typeof document === 'undefined') return undefined;
	const m = document.cookie.match(
		new RegExp(`(?:^|; )${name.replace(/[-[\]{}()*+?.,\\^$|#\s]/g, '\\$&')}=([^;]*)`)
	);
	return m ? decodeURIComponent(m[1]) : undefined;
}

async function fetchAPI<T>(endpoint: string, options?: RequestInit): Promise<T> {
	const csrf = getCookie('csrf_token');
	const headers: Record<string, string> = {
		'Content-Type': 'application/json',
		...(options?.headers as Record<string, string> | undefined)
	};
	if (csrf) headers['X-CSRF-Token'] = csrf;

	const response = await fetch(`${API_BASE}${endpoint}`, {
		...options,
		credentials: 'include',
		headers
	});

	if (!response.ok) {
		const error: ApiError = await response.json().catch(() => ({
			error: 'unknown_error',
			message: 'An unknown error occurred'
		}));
		throw new Error(error.message || error.error);
	}

	return response.json();
}

export async function listTeams(
	params: { page?: number; limit?: number; search?: string } = {}
): Promise<TeamListResponse> {
	const sp = new URLSearchParams();
	if (params.page) sp.set('page', String(params.page));
	if (params.limit) sp.set('limit', String(params.limit));
	if (params.search) sp.set('search', params.search);
	const q = sp.toString();
	return fetchAPI<TeamListResponse>(q ? `/teams?${q}` : '/teams');
}

export async function listTeamMembers(
	teamId: string,
	params: { page?: number; limit?: number } = {}
): Promise<TeamMemberListResponse> {
	const sp = new URLSearchParams();
	if (params.page) sp.set('page', String(params.page));
	if (params.limit) sp.set('limit', String(params.limit));
	const q = sp.toString();
	return fetchAPI<TeamMemberListResponse>(
		q ? `/teams/${teamId}/members?${q}` : `/teams/${teamId}/members`
	);
}
