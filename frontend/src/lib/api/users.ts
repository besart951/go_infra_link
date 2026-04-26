import { api } from './client.js';

export type UserRole =
  | 'superadmin'
  | 'admin_fzag'
  | 'fzag'
  | 'admin_planer'
  | 'planer'
  | 'admin_entrepreneur'
  | 'entrepreneur';

export interface User {
  id: string;
  first_name: string;
  last_name: string;
  email: string;
  is_active: boolean;
  role: UserRole;
  role_display_name: string;
  permissions?: string[];
  can_access_user_directory?: boolean;
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

export interface CreateUserRequest {
  first_name: string;
  last_name: string;
  email: string;
  password: string;
  is_active: boolean;
  role?: UserRole;
}

export interface UpdateUserRequest {
  first_name?: string;
  last_name?: string;
  email?: string;
  password?: string;
  is_active?: boolean;
  role?: UserRole;
}

export interface AllowedRole {
  role: UserRole;
  display_name: string;
}

export interface AllowedRolesResponse {
  roles: AllowedRole[];
}

export interface UserDirectoryTeam {
  id: string;
  name: string;
}

export interface UserDirectoryCapabilities {
  can_update: boolean;
  can_delete: boolean;
  can_disable: boolean;
  can_enable: boolean;
  can_change_role: boolean;
}

export interface UserDirectoryUser extends User {
  teams: UserDirectoryTeam[];
  capabilities: UserDirectoryCapabilities;
}

export interface UserDirectoryTeamFilter {
  id: string;
  name: string;
  count: number;
}

export interface UserDirectoryPageCapabilities {
  can_create_user: boolean;
}

export interface UserDirectoryResponse {
  items: UserDirectoryUser[];
  total: number;
  page: number;
  total_pages: number;
  teams: UserDirectoryTeamFilter[];
  capabilities: UserDirectoryPageCapabilities;
}

/**
 * List all users with pagination and filtering
 * CSRF token is automatically included
 */
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

  return api<PaginatedUserResponse>(endpoint, options);
}

export async function listUserDirectory(
  params: ListUsersParams & { team_id?: string } = {},
  options?: RequestInit
): Promise<UserDirectoryResponse> {
  const searchParams = new URLSearchParams();
  if (params.page) searchParams.set('page', params.page.toString());
  if (params.limit) searchParams.set('limit', params.limit.toString());
  if (params.search) searchParams.set('search', params.search);
  if (params.order_by) searchParams.set('order_by', params.order_by);
  if (params.order) searchParams.set('order', params.order);
  if (params.team_id) searchParams.set('team_id', params.team_id);

  const query = searchParams.toString();
  const endpoint = query ? `/users/directory?${query}` : '/users/directory';

  return api<UserDirectoryResponse>(endpoint, options);
}

/**
 * Get current authenticated user
 */
export async function getCurrentUser(): Promise<User> {
  return api<User>('/auth/me');
}

/**
 * Get allowed roles for the current user
 * CSRF token is automatically included
 */
export async function getAllowedRoles(): Promise<AllowedRolesResponse> {
  return api<AllowedRolesResponse>('/users/allowed-roles');
}

/**
 * Create a new user
 * CSRF token is automatically included
 */
export async function createUser(req: CreateUserRequest): Promise<User> {
  return api<User>('/users', {
    method: 'POST',
    body: JSON.stringify(req)
  });
}

/**
 * Set a user's role (admin only)
 * CSRF token is automatically included
 */
export async function setUserRole(userId: string, role: UserRole): Promise<void> {
  return api<void>(`/admin/users/${userId}/role`, {
    method: 'POST',
    body: JSON.stringify({ role })
  });
}

/**
 * Disable a user (admin only)
 * CSRF token is automatically included
 */
export async function disableUser(userId: string): Promise<void> {
  return api<void>(`/admin/users/${userId}/disable`, {
    method: 'POST'
  });
}

/**
 * Enable a user (admin only)
 * CSRF token is automatically included
 */
export async function enableUser(userId: string): Promise<void> {
  return api<void>(`/admin/users/${userId}/enable`, {
    method: 'POST'
  });
}

/**
 * Delete a user
 * CSRF token is automatically included
 */
export async function deleteUser(userId: string): Promise<void> {
  return api<void>(`/users/${userId}`, {
    method: 'DELETE'
  });
}

/**
 * Update current user profile fields
 */
export async function updateCurrentUser(userId: string, data: UpdateUserRequest): Promise<User> {
  return api<User>(`/users/${userId}`, {
    method: 'PUT',
    body: JSON.stringify(data)
  });
}

/**
 * Update current user password
 */
export async function updateCurrentUserPassword(userId: string, password: string): Promise<User> {
  return api<User>(`/users/${userId}`, {
    method: 'PUT',
    body: JSON.stringify({ password })
  });
}
