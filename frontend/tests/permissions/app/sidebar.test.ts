import { render, screen } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { buildAdminUser, buildUser, permission } from '../../helpers/permissions.js';
import type { Team } from '../../../src/lib/domain/team/index.js';
import type { Project } from '../../../src/lib/domain/project/index.js';

const DEFAULT_TIMESTAMP = '2026-01-01T00:00:00.000Z';

function buildTeam(overrides: Partial<Team> = {}): Team {
  return {
    id: 'team-1',
    name: 'Ops',
    description: '',
    created_at: DEFAULT_TIMESTAMP,
    updated_at: DEFAULT_TIMESTAMP,
    ...overrides
  };
}

function buildProject(overrides: Partial<Project> = {}): Project {
  return {
    id: 'project-1',
    name: 'Alpha',
    description: '',
    status: 'planned',
    phase_id: 'phase-1',
    creator_id: 'user-1',
    created_at: DEFAULT_TIMESTAMP,
    updated_at: DEFAULT_TIMESTAMP,
    ...overrides
  };
}

const state = vi.hoisted(() => {
  const grantedPermissions = new Set<string>();

  return {
    gotoMock: vi.fn(),
    pageStore: {
      subscribe(run: (value: unknown) => void) {
        run({ url: new URL('http://localhost/users'), params: {}, data: {} });
        return () => {};
      }
    },
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

vi.mock('$app/stores', () => ({
  page: state.pageStore
}));

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

vi.mock('$lib/components/ui/sidebar/index.js', async () => {
  const { default: SlotContainer } = await import('../../setup/stubs/SlotContainer.svelte');
  return {
    Root: SlotContainer,
    Header: SlotContainer,
    Content: SlotContainer,
    Footer: SlotContainer,
    Rail: SlotContainer
  };
});

vi.mock('$lib/components/sidebar/index.js', async () => {
  const { default: NavMain } = await import('../../setup/stubs/NavMainStub.svelte');
  const { default: NavProjects } = await import('../../setup/stubs/NavProjectsStub.svelte');
  const { default: NavUser } = await import('../../setup/stubs/NavUserStub.svelte');
  const { default: TeamSwitcher } = await import('../../setup/stubs/TeamSwitcherStub.svelte');
  return {
    NavMain,
    NavProjects,
    NavUser,
    TeamSwitcher
  };
});

import AppSidebar from '../../../src/lib/components/app-sidebar.svelte';

describe('permission-aware sidebar navigation', () => {
  beforeEach(() => {
    state.gotoMock.mockReset();
    state.resetPermissions();
  });

  it('hides protected navigation links when the user has no matching permissions', () => {
    render(AppSidebar, {
      user: buildUser(),
      teams: [],
      projects: []
    });

    expect(screen.queryByTestId('nav-link:/users')).not.toBeInTheDocument();
    expect(screen.queryByTestId('nav-link:/teams')).not.toBeInTheDocument();
    expect(screen.queryByTestId('nav-link:/projects')).not.toBeInTheDocument();
    expect(screen.queryByTestId('nav-link:/facility/buildings')).not.toBeInTheDocument();
    expect(screen.queryByTestId('nav-link:/admin/notifications/smtp')).not.toBeInTheDocument();
  });

  it('shows navigation entries that match the granted permission set', () => {
    state.setPermissions([
      permission('user'),
      permission('team'),
      permission('role'),
      permission('project'),
      permission('phase'),
      permission('building')
    ]);

    render(AppSidebar, {
      user: buildUser(),
      teams: [buildTeam()],
      projects: [buildProject()]
    });

    expect(screen.getByTestId('nav-link:/users')).toBeInTheDocument();
    expect(screen.getByTestId('nav-link:/teams')).toBeInTheDocument();
    expect(screen.getByTestId('nav-link:/users/roles')).toBeInTheDocument();
    expect(screen.getByTestId('nav-link:/projects')).toBeInTheDocument();
    expect(screen.getByTestId('nav-link:/projects/phases')).toBeInTheDocument();
    expect(screen.getByTestId('nav-link:/facility/buildings')).toBeInTheDocument();
    expect(screen.getByTestId('project-link:project-1')).toBeInTheDocument();
  });

  it('keeps the admin notifications link restricted to superadmin users', () => {
    render(AppSidebar, {
      user: buildAdminUser(),
      teams: [],
      projects: []
    });

    expect(screen.getByTestId('nav-link:/admin/notifications/smtp')).toBeInTheDocument();
  });
});
