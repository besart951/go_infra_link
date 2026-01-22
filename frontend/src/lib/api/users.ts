export interface User {
	id: string;
	first_name: string;
	last_name: string;
	email: string;
	is_active: boolean;
	role: 'user' | 'admin' | 'superadmin';
	created_at: string;
	updated_at: string;
	last_login_at?: string | null;
	disabled_at?: string | null;
	locked_until?: string | null;
	failed_login_attempts: number;
}

export interface PaginatedUserResponse {
	items: User[];
	total: number;
	page: number;
	total_pages: number;
}

export interface ListUsersParams {
	page?: number;
	limit?: number;
	search?: string;
	order_by?: string;
	order?: 'asc' | 'desc';
}

export interface ApiError {
	error: string;
	message?: string;
}

const API_BASE = '/api/v1';

function getCookie(name: string): string | undefined {
	if (typeof document === 'undefined') return undefined;
	const m = document.cookie.match(
		new RegExp(
			`(?:^|; )${name.replace(/[-[\]{}()*+?.,\\^$|#\s]/g, '\\$&')}=([^;]*)`
		)
	);
	return m ? decodeURIComponent(m[1]) : undefined;
}

async function fetchAPI<T>(endpoint: string, options?: RequestInit): Promise<T> {
	const csrf = getCookie('csrf_token');
	const headers: Record<string, string> = {
		'Content-Type': 'application/json',
		...((options?.headers as Record<string, string> | undefined) ?? {})
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

export async function listUsers(
	params: ListUsersParams = {},
	options?: RequestInit
): Promise<PaginatedUserResponse> {
	const searchParams = new URLSearchParams();
	if (params.page) searchParams.set('page', params.page.toString());
	if (params.limit) searchParams.set('limit', params.limit.toString());
	if (params.search) searchParams.set('search', params.search);
	if (params.order_by) searchParams.set('order_by', params.order_by);
	if (params.order) searchParams.set('order', params.order);

	const query = searchParams.toString();
	const endpoint = query ? `/users?${query}` : '/users';

	return fetchAPI<PaginatedUserResponse>(endpoint, options);
}

export async function setUserRole(
	userId: string,
	role: 'user' | 'admin' | 'superadmin'
): Promise<void> {
	await fetchAPI(`/admin/users/${userId}/role`, {
		method: 'POST',
		body: JSON.stringify({ role })
	});
}

export async function disableUser(userId: string): Promise<void> {
	await fetchAPI(`/admin/users/${userId}/disable`, {
		method: 'POST'
	});
}

export async function enableUser(userId: string): Promise<void> {
	await fetchAPI(`/admin/users/${userId}/enable`, {
		method: 'POST'
	});
}

export async function deleteUser(userId: string): Promise<void> {
	await fetchAPI(`/users/${userId}`, {
		method: 'DELETE'
	});
}
