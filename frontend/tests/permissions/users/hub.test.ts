import { render, screen } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { buildUser, permission } from '../../helpers/permissions.js';

function createUser(overrides: Parameters<typeof buildUser>[0] = {}) {
  return buildUser(overrides);
}

function renderUsersHub() {
  return render(UsersHubPage, {
    data: {
      backendAvailable: true,
      user: state.user,
      teams: [],
      projects: []
    }
  });
}

const state = vi.hoisted(() => {
  const grantedPermissions = new Set<string>();

  return {
    user: null as ReturnType<typeof createUser> | null,
    setPermissions(permissions: string[]) {
      grantedPermissions.clear();
      for (const granted of permissions) {
        grantedPermissions.add(granted);
      }
    },
    resetPermissions() {
      grantedPermissions.clear();
    },
    canPerform(action: string, resource: string) {
      return grantedPermissions.has(`${resource}.${action}`);
    }
  };
});

vi.mock('$lib/i18n/translator.js', () => ({
  createTranslator: () => ({
    subscribe(run: (value: (key: string) => string) => void) {
      run((key: string) => key);
      return () => {};
    }
  })
}));

vi.mock('$lib/stores/auth.svelte.js', () => ({
  auth: {
    get user() {
      return state.user;
    }
  }
}));

vi.mock('$lib/utils/permissions.js', () => ({
  canPerform: (action: string, resource: string) => state.canPerform(action, resource)
}));

import UsersHubPage from '../../../src/routes/(app)/users/+page.svelte';

describe('/users hub permission surface', () => {
  beforeEach(() => {
    state.user = createUser();
    state.resetPermissions();
  });

  it('hides the roles card when the user only has directory access', () => {
    state.user = createUser({ can_access_user_directory: true });

    renderUsersHub();

    expect(screen.getByText('hub.users.directory_title')).toBeInTheDocument();
    expect(screen.queryByText('navigation.roles_permissions')).not.toBeInTheDocument();
  });

  it('shows the roles card for FZAG admins without relying on role.read', () => {
    state.user = createUser({ role: 'admin_fzag' });

    renderUsersHub();

    expect(screen.getByText('navigation.roles_permissions')).toBeInTheDocument();
  });

  it('hides the roles card from non-admin roles even when role.read is granted', () => {
    state.user = createUser({ role: 'fzag' });
    state.setPermissions([permission('role')]);

    renderUsersHub();

    expect(screen.queryByText('navigation.roles_permissions')).not.toBeInTheDocument();
  });
});
