/**
 * Auth Store - Global authentication state using Svelte 5 Runes
 *
 * This store manages:
 * - Current authenticated user
 * - User permissions and derived capabilities
 * - Allowed role assignments returned by the backend
 */

import { getCurrentUser, getAllowedRoles, type User, type UserRole } from '$lib/api/users.js';
import { t } from '$lib/i18n/index.js';

interface AuthState {
  user: User | null;
  allowedRoles: import('$lib/api/users.js').AllowedRole[];
  isLoading: boolean;
  error: string | null;
}

// Global auth state using Svelte 5 runes
const authState = $state<AuthState>({
  user: null,
  allowedRoles: [],
  isLoading: false,
  error: null
});

/**
 * Load current user and their permissions
 */
export async function loadAuth(): Promise<void> {
  authState.isLoading = true;
  authState.error = null;

  try {
    const [user, allowedRolesResponse] = await Promise.all([getCurrentUser(), getAllowedRoles()]);

    authState.user = user;
    authState.allowedRoles = allowedRolesResponse.roles;
  } catch (error) {
    authState.error = error instanceof Error ? error.message : t('auth.load_failed');
    authState.user = null;
    authState.allowedRoles = [];
  } finally {
    authState.isLoading = false;
  }
}

/**
 * Clear auth state (on logout)
 */
export function clearAuth(): void {
  authState.user = null;
  authState.allowedRoles = [];
  authState.error = null;
}

const ROLE_LEVELS: Record<UserRole, number> = {
  superadmin: 100,
  admin_fzag: 90,
  fzag: 80,
  admin_planer: 70,
  planer: 60,
  admin_entrepreneur: 50,
  entrepreneur: 40
};

// Legacy compatibility for stale HMR/module consumers. Authorization stays permission-based.
export function getRoleLevel(role: UserRole): number {
  return ROLE_LEVELS[role] ?? 0;
}

export function canManageRole(targetRole: UserRole): boolean {
  if (!authState.user) return false;
  return getRoleLevel(authState.user.role) > getRoleLevel(targetRole);
}

export function hasRole(role: UserRole): boolean {
  return authState.user?.role === role;
}

export function hasMinRole(minRole: UserRole): boolean {
  if (!authState.user) return false;
  return getRoleLevel(authState.user.role) >= getRoleLevel(minRole);
}

export function canAccessUserDirectory(): boolean {
  return Boolean(authState.user?.can_access_user_directory);
}

/**
 * Check if the current user is authenticated
 */
export function isAuthenticated(): boolean {
  return authState.user !== null;
}

/**
 * Get allowed roles for creating new users
 */
export function getAllowedRolesForCreation(): import('$lib/api/users.js').AllowedRole[] {
  return authState.allowedRoles;
}

/**
 * Get current user
 */
export function getCurrentUserState(): User | null {
  return authState.user;
}

/**
 * Check if auth is loading
 */
export function isAuthLoading(): boolean {
  return authState.isLoading;
}

/**
 * Get auth error
 */
export function getAuthError(): string | null {
  return authState.error;
}

// Export reactive getters using Svelte 5 runes
export const auth = {
  get user() {
    return authState.user;
  },
  get allowedRoles() {
    return authState.allowedRoles;
  },
  get isLoading() {
    return authState.isLoading;
  },
  get error() {
    return authState.error;
  },
  get isAuthenticated() {
    return authState.user !== null;
  },
  get canAccessUserDirectory() {
    return Boolean(authState.user?.can_access_user_directory);
  }
};
