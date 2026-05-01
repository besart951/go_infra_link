import { render, screen } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { permission } from '../../helpers/permissions.js';

const state = vi.hoisted(() => {
  const grantedPermissions = new Set<string>();

  const teamsStoreValue = {
    items: [],
    total: 0,
    page: 1,
    total_pages: 1,
    loading: false,
    error: null
  };

  return {
    gotoMock: vi.fn(),
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
    teamsStore: {
      subscribe(run: (value: unknown) => void) {
        run(teamsStoreValue);
        return () => {};
      },
      load: vi.fn(),
      reload: vi.fn(),
      search: vi.fn(),
      goToPage: vi.fn()
    }
  };
});

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

vi.mock('$lib/utils/permissions.js', () => ({
  canPerform: (action: string, resource: string) => state.canPerform(action, resource)
}));

vi.mock('$lib/api/teams.js', () => ({
  addTeamMember: vi.fn(),
  createTeam: vi.fn(),
  deleteTeam: vi.fn(),
  getTeam: vi.fn(),
  listTeams: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, total_pages: 1 }),
  listTeamMembers: vi.fn().mockResolvedValue({ items: [], total: 0, page: 1, total_pages: 1 }),
  removeTeamMember: vi.fn(),
  updateTeam: vi.fn()
}));

vi.mock('$lib/components/toast.svelte', () => ({
  addToast: state.addToastMock
}));

vi.mock('$lib/stores/confirm-dialog.js', () => ({
  confirm: state.confirmMock
}));

vi.mock('$lib/stores/list/entityStores.js', () => ({
  teamsStore: state.teamsStore
}));

vi.mock('$lib/components/list/PaginatedList.svelte', async () => {
  const { default: EmptyList } = await import('../../setup/stubs/EmptyList.svelte');
  return { default: EmptyList };
});

vi.mock('$lib/components/confirm-dialog.svelte', async () => {
  const { default: SlotContainer } = await import('../../setup/stubs/SlotContainer.svelte');
  return { default: SlotContainer };
});

import TeamsPage from '../../../src/routes/(app)/teams/+page.svelte';

describe('/teams permission surface', () => {
  beforeEach(() => {
    state.gotoMock.mockReset();
    state.resetPermissions();
    state.addToastMock.mockReset();
    state.confirmMock.mockReset();
    state.confirmMock.mockResolvedValue(true);
    state.teamsStore.load.mockClear();
  });

  it('still renders for a logged-in user without team.read, exposing the current route leak', () => {
    render(TeamsPage);

    expect(screen.getByText('navigation.teams')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'pages.create_team' })).not.toBeInTheDocument();
  });

  it('still renders for a logged-in user with an unrelated permission set', () => {
    state.setPermissions([permission('user')]);

    render(TeamsPage);

    expect(screen.getByText('navigation.teams')).toBeInTheDocument();
    expect(screen.queryByRole('button', { name: 'pages.create_team' })).not.toBeInTheDocument();
  });

  it('shows the create CTA when team.create is granted', () => {
    state.setPermissions([permission('team'), permission('team', 'create')]);

    render(TeamsPage);

    expect(screen.getByRole('button', { name: 'pages.create_team' })).toBeInTheDocument();
  });
});
