/**
 * Domain entity representing a User
 */
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
