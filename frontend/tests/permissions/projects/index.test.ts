import { render, screen } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { permission } from '../../helpers/permissions.js';

const state = vi.hoisted(() => {
  const grantedPermissions = new Set<string>();

  const projectListValue = {
    items: [],
    total: 0,
    page: 1,
    total_pages: 1,
    loading: false,
    error: null,
    status: 'all'
  };

  return {
    gotoMock: vi.fn(),
    addToastMock: vi.fn(),
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
    projectListStore: {
      subscribe(run: (value: unknown) => void) {
        run(projectListValue);
        return () => {};
      },
      load: vi.fn(),
      reload: vi.fn(),
      search: vi.fn(),
      goToPage: vi.fn(),
      setStatus: vi.fn()
    },
    optimisticExecuteMock: vi.fn()
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

vi.mock('$lib/components/toast.svelte', () => ({
  addToast: state.addToastMock
}));

vi.mock('$lib/stores/projects/projectListStore.js', () => ({
  projectListStore: state.projectListStore
}));

vi.mock('$lib/infrastructure/api/project.adapter.js', () => ({
  createProject: vi.fn()
}));

vi.mock('$lib/hooks/useOptimisticUpdate.svelte.js', () => ({
  useOptimisticUpdate: () => ({
    execute: state.optimisticExecuteMock
  })
}));

vi.mock('$lib/components/list/PaginatedList.svelte', async () => {
  const { default: EmptyList } = await import('../../setup/stubs/EmptyList.svelte');
  return { default: EmptyList };
});

vi.mock('$lib/components/project/ProjectPhaseSelect.svelte', async () => {
  const { default: SlotContainer } = await import('../../setup/stubs/SlotContainer.svelte');
  return { default: SlotContainer };
});

import ProjectsPage from '../../../src/routes/(app)/projects/+page.svelte';

describe('/projects permission surface', () => {
  beforeEach(() => {
    state.gotoMock.mockReset();
    state.addToastMock.mockReset();
    state.optimisticExecuteMock.mockReset();
    state.projectListStore.load.mockClear();
    state.resetPermissions();
  });

  it('renders the scoped project list without requiring project.listAll', () => {
    render(ProjectsPage);

    expect(screen.getByText('navigation.projects')).toBeInTheDocument();
    expect(screen.queryByText('common.create')).not.toBeInTheDocument();
  });

  it('still renders for a logged-in user with an unrelated permission set', () => {
    state.setPermissions([permission('team')]);

    render(ProjectsPage);

    expect(screen.getByText('navigation.projects')).toBeInTheDocument();
    expect(screen.queryByText('common.create')).not.toBeInTheDocument();
  });

  it('shows the create CTA when project.create is granted', () => {
    state.setPermissions([permission('project', 'create')]);

    render(ProjectsPage);

    expect(screen.getByText('common.create')).toBeInTheDocument();
  });
});
