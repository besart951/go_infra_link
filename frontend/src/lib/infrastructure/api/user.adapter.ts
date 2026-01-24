/**
 * User API adapter
 * Infrastructure layer - implements user data operations via HTTP
 */
import { api, ApiException } from '$lib/api/client.js';
import type {
	User,
	UserListParams,
	UserListResponse,
	CreateUserRequest,
	UpdateUserRequest
} from '$lib/domain/user/index.js';

/**
 * List users with optional filters
 */
export async function listUsers(
	params?: UserListParams,
	options?: RequestInit
): Promise<UserListResponse> {
	const searchParams = new URLSearchParams();
	if (params?.page) searchParams.set('page', String(params.page));
	if (params?.limit) searchParams.set('limit', String(params.limit));
	if (params?.search) searchParams.set('search', params.search);
	if (params?.role) searchParams.set('role', params.role);
	if (params?.is_active !== undefined) searchParams.set('is_active', String(params.is_active));

	const query = searchParams.toString();
	const endpoint = `/users${query ? `?${query}` : ''}`;

	return api<UserListResponse>(endpoint, options);
}

/**
 * Get a single user by ID
 */
export async function getUser(id: string, options?: RequestInit): Promise<User> {
	return api<User>(`/users/${id}`, options);
}

/**
 * Get the current authenticated user
 */
export async function getCurrentUser(options?: RequestInit): Promise<User> {
	return api<User>('/auth/me', options);
}

/**
 * Create a new user
 */
export async function createUser(data: CreateUserRequest, options?: RequestInit): Promise<User> {
	return api<User>('/users', {
		...options,
		method: 'POST',
		body: JSON.stringify(data)
	});
}

/**
 * Update an existing user
 */
export async function updateUser(
	id: string,
	data: UpdateUserRequest,
	options?: RequestInit
): Promise<User> {
	return api<User>(`/users/${id}`, {
		...options,
		method: 'PUT',
		body: JSON.stringify(data)
	});
}

/**
 * Delete a user
 */
export async function deleteUser(id: string, options?: RequestInit): Promise<void> {
	return api<void>(`/users/${id}`, {
		...options,
		method: 'DELETE'
	});
}

// Re-export types for convenience
export type { User, UserListParams, UserListResponse, CreateUserRequest, UpdateUserRequest };
