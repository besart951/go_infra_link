import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const state = vi.hoisted(() => {
  return {
    gotoMock: vi.fn(),
    canAccessUserDirectory: true,
    addToastMock: vi.fn(),
    confirmMock: vi.fn().mockResolvedValue(true),
    listUserDirectoryMock: vi.fn(),
    setUserRoleMock: vi.fn(),
    disableUserMock: vi.fn(),
    enableUserMock: vi.fn(),
    restoreUserMock: vi.fn(),
    deleteUserMock: vi.fn()
  };
});

const defaultDirectoryResponse = {
  items: [],
  total: 0,
  page: 1,
  total_pages: 1,
  teams: [],
  roles: [],
  capabilities: { can_create_user: false, can_read_deleted: false }
};

function directoryUser(overrides: Record<string, unknown> = {}) {
  return {
    id: '00000000-0000-0000-0000-000000000001',
    first_name: 'Ada',
    last_name: 'Lovelace',
    email: 'ada@example.com',
    is_active: false,
    role: 'planer',
    role_display_name: 'Planer',
    created_at: '2026-05-08T10:00:00.000Z',
    updated_at: '2026-05-08T10:00:00.000Z',
    last_login_at: null,
    disabled_at: null,
    locked_until: null,
    failed_login_attempts: 0,
    teams: [],
    capabilities: {
      can_update: false,
      can_delete: false,
      can_disable: false,
      can_enable: false,
      can_restore: false,
      can_change_role: false
    },
    ...overrides
  };
}

vi.mock('$app/navigation', () => ({
  goto: state.gotoMock
}));

vi.mock('$lib/i18n/translator', () => ({
  createTranslator: () => ({
    subscribe(run: (value: (key: string) => string) => void) {
      run((key: string) => key);
      return () => {};
    }
  })
}));

vi.mock('$lib/i18n/index.js', () => ({
  t: (key: string) => key
}));

vi.mock('$lib/api/client.js', () => ({
  getErrorMessage: (error: unknown) => (error instanceof Error ? error.message : 'unknown')
}));

vi.mock('$lib/api/users.js', () => ({
  createUser: vi.fn(),
  deleteUser: state.deleteUserMock,
  disableUser: state.disableUserMock,
  enableUser: state.enableUserMock,
  restoreUser: state.restoreUserMock,
  getAllowedRoles: vi.fn().mockResolvedValue({ roles: [] }),
  getCurrentUser: vi.fn(),
  getUserRegistration: vi.fn(),
  inviteUser: vi.fn(),
  listUsers: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, total_pages: 1 }),
  listUserDirectory: state.listUserDirectoryMock,
  resendUserRegistration: vi.fn(),
  setUserRole: state.setUserRoleMock,
  updateCurrentUser: vi.fn(),
  updateCurrentUserPassword: vi.fn()
}));

vi.mock('$lib/stores/auth.svelte.js', () => ({
  getAllowedRolesForCreation: () => [],
  auth: {
    get canAccessUserDirectory() {
      return state.canAccessUserDirectory;
    }
  }
}));

vi.mock('$lib/components/toast.svelte', () => ({
  addToast: state.addToastMock
}));

vi.mock('$lib/stores/confirm-dialog.js', () => ({
  confirm: state.confirmMock
}));

vi.mock('$lib/components/confirm-dialog.svelte', async () => {
  const { default: SlotContainer } = await import('../../setup/stubs/SlotContainer.svelte');
  return { default: SlotContainer };
});

vi.mock('$lib/components/role-badge.svelte', async () => {
  const { default: SlotContainer } = await import('../../setup/stubs/SlotContainer.svelte');
  return { default: SlotContainer };
});

vi.mock('$lib/components/user-avatar.svelte', async () => {
  const { default: SlotContainer } = await import('../../setup/stubs/SlotContainer.svelte');
  return { default: SlotContainer };
});

vi.mock('$lib/components/user-management-form.svelte', async () => {
  const { default: SlotContainer } = await import('../../setup/stubs/SlotContainer.svelte');
  return { default: SlotContainer };
});

vi.mock('$lib/components/ui/dialog/index.js', async () => {
  const { default: SlotContainer } = await import('../../setup/stubs/SlotContainer.svelte');
  return {
    Root: SlotContainer,
    Content: SlotContainer,
    Header: SlotContainer,
    Title: SlotContainer,
    Description: SlotContainer
  };
});

import UsersPage from '../../../src/routes/(app)/users/directory/+page.svelte';

describe('/users/directory permission surface', () => {
  beforeEach(() => {
    state.gotoMock.mockReset();
    state.canAccessUserDirectory = true;
    state.addToastMock.mockReset();
    state.confirmMock.mockReset();
    state.confirmMock.mockResolvedValue(true);
    state.listUserDirectoryMock.mockReset();
    state.listUserDirectoryMock.mockResolvedValue(defaultDirectoryResponse);
    state.setUserRoleMock.mockReset();
    state.disableUserMock.mockReset();
    state.enableUserMock.mockReset();
    state.restoreUserMock.mockReset();
    state.deleteUserMock.mockReset();
  });

  it('redirects to / when auth says user cannot access directory', async () => {
    state.canAccessUserDirectory = false;

    render(UsersPage);

    await waitFor(() => {
      expect(state.gotoMock).toHaveBeenCalledWith('/');
    });
    expect(state.listUserDirectoryMock).not.toHaveBeenCalled();
  });

  it('uses layout user data on hard reload before the auth store hydrates', async () => {
    state.canAccessUserDirectory = false;

    render(UsersPage, {
      data: {
        user: {
          can_access_user_directory: true
        }
      }
    });

    await waitFor(() => {
      expect(state.listUserDirectoryMock).toHaveBeenCalled();
    });
    expect(state.gotoMock).not.toHaveBeenCalled();
  });

  it('loads data from /users/directory and keeps create CTA hidden when capability is false', async () => {
    state.listUserDirectoryMock.mockResolvedValue({
      ...defaultDirectoryResponse,
      capabilities: { can_create_user: false, can_read_deleted: false }
    });

    render(UsersPage);

    await waitFor(() => {
      expect(state.listUserDirectoryMock).toHaveBeenCalled();
    });
    expect(screen.getByText('pages.user_management')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'common.create_user' })).not.toBeInTheDocument();
  });

  it('shows create CTA when directory page capability allows user creation', async () => {
    state.listUserDirectoryMock.mockResolvedValue({
      ...defaultDirectoryResponse,
      capabilities: { can_create_user: true, can_read_deleted: false }
    });

    render(UsersPage);

    await waitFor(() => {
      expect(state.listUserDirectoryMock).toHaveBeenCalled();
    });
    expect(screen.getByRole('button', { name: 'common.create_user' })).toBeInTheDocument();
  });

  it('keeps the row action menu visible when invitation resend is cooling down', async () => {
    state.listUserDirectoryMock.mockResolvedValue({
      ...defaultDirectoryResponse,
      items: [
        directoryUser({
          registration_process: {
            status: 'pending',
            email_status: 'sent',
            steps: [
              { key: 'created', label: 'Angelegt', status: 'completed' },
              { key: 'email_sent', label: 'E-Mail versendet', status: 'completed' },
              { key: 'registered', label: 'Registriert', status: 'current' },
              { key: 'first_login', label: 'Erste Anmeldung', status: 'pending' }
            ],
            can_resend: false,
            resend_available_at: new Date(Date.now() + 60_000).toISOString(),
            send_count: 1
          }
        })
      ],
      total: 1,
      capabilities: { can_create_user: true, can_read_deleted: false }
    });

    render(UsersPage);

    await waitFor(() => {
      expect(state.listUserDirectoryMock).toHaveBeenCalled();
    });
    await fireEvent.click(screen.getByRole('button', { name: 'common.actions' }));

    expect(await screen.findByText('user.resend_invitation')).toBeInTheDocument();
  });

  it('hides deleted-user toggle without read-deleted capability', async () => {
    render(UsersPage);

    await waitFor(() => {
      expect(state.listUserDirectoryMock).toHaveBeenCalled();
    });

    expect(screen.queryByText('common.show_deleted_users')).not.toBeInTheDocument();
  });

  it('shows deleted-user toggle when read-deleted capability is present', async () => {
    state.listUserDirectoryMock.mockResolvedValue({
      ...defaultDirectoryResponse,
      capabilities: { can_create_user: false, can_read_deleted: true }
    });

    render(UsersPage);

    await waitFor(() => {
      expect(state.listUserDirectoryMock).toHaveBeenCalled();
    });

    expect(screen.getByText('common.show_deleted_users')).toBeInTheDocument();
  });

  it('applies team and role filters from the toolbar', async () => {
    state.listUserDirectoryMock.mockResolvedValue({
      ...defaultDirectoryResponse,
      teams: [{ id: 'team-1', name: 'Team Alpha', count: 1 }],
      roles: [{ role: 'planer', display_name: 'Planer', count: 1 }]
    });

    render(UsersPage);

    await waitFor(() => {
      expect(state.listUserDirectoryMock).toHaveBeenCalled();
    });

    await fireEvent.change(screen.getByLabelText('common.team'), {
      target: { value: 'team-1' }
    });
    await waitFor(() => {
      expect(state.listUserDirectoryMock).toHaveBeenLastCalledWith(
        expect.objectContaining({ team_id: 'team-1', role: undefined })
      );
    });

    await fireEvent.change(screen.getByLabelText('common.role'), {
      target: { value: 'planer' }
    });
    await waitFor(() => {
      expect(state.listUserDirectoryMock).toHaveBeenLastCalledWith(
        expect.objectContaining({ team_id: 'team-1', role: 'planer' })
      );
    });
  });

  it('shows restore action from can_restore instead of can_enable', async () => {
    state.listUserDirectoryMock.mockResolvedValue({
      ...defaultDirectoryResponse,
      items: [
        directoryUser({
          is_deleted: true,
          capabilities: {
            can_update: false,
            can_delete: false,
            can_disable: false,
            can_enable: false,
            can_restore: true,
            can_change_role: false
          }
        })
      ],
      total: 1,
      capabilities: { can_create_user: false, can_read_deleted: true }
    });

    render(UsersPage);

    await waitFor(() => {
      expect(state.listUserDirectoryMock).toHaveBeenCalled();
    });
    await fireEvent.click(screen.getByRole('button', { name: 'common.actions' }));

    expect(await screen.findByText('actions.restore_user')).toBeInTheDocument();
  });
});
