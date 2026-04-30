/**
 * Auth Store - Global authentication state using Svelte 5 Runes
 *
 * This store manages:
 * - Current authenticated user
 * - User permissions and derived capabilities
 * - Allowed role assignments returned by the backend
 */

import { ApiException } from '$lib/api/client.js';
import { t } from '$lib/i18n/index.js';
import {
  userRepository,
  type AllowedRole,
  type User
} from '$lib/infrastructure/api/userRepository.js';

interface AuthState {
  user: User | null;
  allowedRoles: AllowedRole[];
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
    const user = await userRepository.getCurrent();
    authState.user = user;

    try {
      const allowedRolesResponse = await userRepository.getAllowedRoles();
      authState.allowedRoles = allowedRolesResponse.roles;
    } catch (error) {
      authState.error = error instanceof Error ? error.message : t('auth.load_failed');
      if (authState.allowedRoles.length === 0) {
        authState.allowedRoles = [];
      }
    }
  } catch (error) {
    authState.error = error instanceof Error ? error.message : t('auth.load_failed');
    if (!(error instanceof ApiException && error.status === 0 && authState.user)) {
      authState.user = null;
      authState.allowedRoles = [];
    }
  } finally {
    authState.isLoading = false;
  }
}

/**
 * Get allowed roles for creating new users
 */
export function getAllowedRolesForCreation(): AllowedRole[] {
  return authState.allowedRoles;
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
