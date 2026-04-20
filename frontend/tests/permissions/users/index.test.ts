import { render, screen } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { permission } from '../../helpers/permissions.js';

const state = vi.hoisted(() => {
  const grantedPermissions = new Set<string>();

  const usersStoreValue = {
    items: [],
    total: 0,
    page: 1,
    total_pages: 1,
    loading: false,
    error: null
  };

  return {
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
    },
    addToastMock: vi.fn(),
    confirmMock: vi.fn().mockResolvedValue(true),
    listTeamsMock: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, total_pages: 1 }),
    listTeamMembersMock: vi
      .fn()
      .mockResolvedValue({ items: [], total: 0, page: 1, total_pages: 1 }),
    usersStore: {
      subscribe(run: (value: unknown) => void) {
        run(usersStoreValue);
        return () => {};
      },
      load: vi.fn(),
      reload: vi.fn(),
      search: vi.fn(),
      goToPage: vi.fn()
    }
  };
});

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

vi.mock('$lib/utils/permissions.js', () => ({
  canPerform: (action: string, resource: string) => state.canPerform(action, resource)
}));

vi.mock('$lib/api/teams.js', () => ({
  listTeams: state.listTeamsMock,
  listTeamMembers: state.listTeamMembersMock
}));

vi.mock('$lib/api/users.js', () => ({
  setUserRole: vi.fn(),
  disableUser: vi.fn(),
  enableUser: vi.fn(),
  deleteUser: vi.fn()
}));

vi.mock('$lib/stores/auth.svelte.js', () => ({
  getAllowedRolesForCreation: () => []
}));

vi.mock('$lib/components/toast.svelte', () => ({
  addToast: state.addToastMock
}));

vi.mock('$lib/stores/confirm-dialog.js', () => ({
  confirm: state.confirmMock
}));

vi.mock('$lib/stores/list/entityStores.js', () => ({
  usersStore: state.usersStore
}));

vi.mock('$lib/components/list/PaginatedList.svelte', async () => {
  const { default: EmptyList } = await import('../../setup/stubs/EmptyList.svelte');
  return { default: EmptyList };
});

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

import UsersPage from '../../../src/routes/(app)/users/+page.svelte';

describe('/users permission surface', () => {
  beforeEach(() => {
    state.resetPermissions();
    state.addToastMock.mockReset();
    state.confirmMock.mockReset();
    state.confirmMock.mockResolvedValue(true);
    state.usersStore.load.mockClear();
  });

  it('still renders for a logged-in user without user.read, exposing the current route leak', () => {
    render(UsersPage);

    expect(screen.getByText('pages.user_management')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'common.create_user' })).not.toBeInTheDocument();
  });

  it('still renders for a logged-in user with an unrelated permission set', () => {
    state.setPermissions([permission('team')]);

    render(UsersPage);

    expect(screen.getByText('pages.user_management')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'common.create_user' })).not.toBeInTheDocument();
  });

  it('keeps the create CTA hidden when the user can only read users', () => {
    state.setPermissions([permission('user')]);

    render(UsersPage);

    expect(screen.getByText('pages.user_management')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'common.create_user' })).not.toBeInTheDocument();
  });

  it('shows the create CTA when user.create is granted', () => {
    state.setPermissions([permission('user'), permission('user', 'create')]);

    render(UsersPage);

    expect(screen.getByRole('button', { name: 'common.create_user' })).toBeInTheDocument();
  });
});
