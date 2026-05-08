import { fireEvent, render, screen, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

const state = vi.hoisted(() => ({
  createTeamMock: vi.fn(),
  gotoMock: vi.fn(),
  addToastMock: vi.fn(),
  teamsStoreReloadMock: vi.fn()
}));

vi.mock('$app/navigation', () => ({
  goto: state.gotoMock
}));

vi.mock('$lib/i18n/translator.js', () => ({
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

vi.mock('$lib/api/teams.js', () => ({
  addTeamMember: vi.fn(),
  createTeam: state.createTeamMock,
  deleteTeam: vi.fn(),
  getTeam: vi.fn(),
  listTeams: vi.fn(),
  listTeamMembers: vi.fn(),
  removeTeamMember: vi.fn(),
  updateTeam: vi.fn()
}));

vi.mock('$lib/components/toast.svelte', () => ({
  addToast: state.addToastMock
}));

vi.mock('$lib/stores/list/entityStores.js', () => ({
  teamsStore: {
    reload: state.teamsStoreReloadMock
  }
}));

import TeamNewPage from '../../../src/routes/(app)/teams/new/+page.svelte';

describe('/teams/new', () => {
  beforeEach(() => {
    state.createTeamMock.mockReset();
    state.gotoMock.mockReset();
    state.addToastMock.mockReset();
    state.teamsStoreReloadMock.mockReset();
  });

  it('creates a team and navigates to the member management page', async () => {
    state.createTeamMock.mockResolvedValue({
      id: 'team-1',
      name: 'Automation',
      description: null,
      created_at: '2026-01-01T00:00:00.000Z',
      updated_at: '2026-01-01T00:00:00.000Z'
    });

    render(TeamNewPage);

    await fireEvent.input(screen.getByLabelText('common.name'), {
      target: { value: 'Automation' }
    });
    await fireEvent.input(screen.getByLabelText('messages.team_description'), {
      target: { value: 'Controls work' }
    });
    await fireEvent.click(screen.getByRole('button', { name: 'common.create' }));

    await waitFor(() => {
      expect(state.createTeamMock).toHaveBeenCalledWith({
        name: 'Automation',
        description: 'Controls work'
      });
    });
    expect(state.teamsStoreReloadMock).toHaveBeenCalled();
    expect(state.gotoMock).toHaveBeenCalledWith('/teams/team-1');
  });
});
