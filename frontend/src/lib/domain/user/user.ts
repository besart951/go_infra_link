/**
 * User domain types
 * Mirrors backend: internal/domain/user/user.go
 */

export type UserRole = 'user' | 'admin' | 'superadmin';

export interface User {
	id: string;
	first_name: string;
	last_name: string;
	email: string;
	role: UserRole;
	is_active: boolean;
	disabled_at?: string;
	locked_until?: string;
	last_login_at?: string;
	failed_login_attempts: number;
	created_by_id?: string;
	created_at: string;
	updated_at: string;
}

export interface CreateUserRequest {
	first_name: string;
	last_name: string;
	email: string;
	password: string;
	role?: UserRole;
}

export interface UpdateUserRequest {
	first_name?: string;
	last_name?: string;
	email?: string;
	password?: string;
	role?: UserRole;
	is_active?: boolean;
}

export interface UserListParams {
	page?: number;
	limit?: number;
	search?: string;
	role?: UserRole;
	is_active?: boolean;
}

export interface UserListResponse {
	users: User[];
	total: number;
	page: number;
	limit: number;
}
