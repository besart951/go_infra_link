import { render, screen, waitFor } from '@testing-library/svelte';
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
    deleteUserMock: vi.fn()
  };
});

const defaultDirectoryResponse = {
  items: [],
  total: 0,
  page: 1,
  total_pages: 1,
  teams: [],
  capabilities: { can_create_user: false }
};

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
  getAllowedRoles: vi.fn().mockResolvedValue({ roles: [] }),
  getCurrentUser: vi.fn(),
  listUsers: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, total_pages: 1 }),
  listUserDirectory: state.listUserDirectoryMock,
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

  it('loads data from /users/directory and keeps create CTA hidden when capability is false', async () => {
    state.listUserDirectoryMock.mockResolvedValue({
      ...defaultDirectoryResponse,
      capabilities: { can_create_user: false }
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
      capabilities: { can_create_user: true }
    });

    render(UsersPage);

    await waitFor(() => {
      expect(state.listUserDirectoryMock).toHaveBeenCalled();
    });
    expect(screen.getByRole('button', { name: 'common.create_user' })).toBeInTheDocument();
  });
});
