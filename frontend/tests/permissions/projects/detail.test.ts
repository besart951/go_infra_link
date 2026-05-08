import { render, screen, waitFor } from '@testing-library/svelte';
import { beforeEach, describe, expect, it, vi } from 'vitest';

import { permission } from '../../helpers/permissions.js';

const state = vi.hoisted(() => {
  const grantedPermissions = new Set<string>();

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
    pageStore: {
      subscribe(run: (value: unknown) => void) {
        run({
          url: new URL('http://localhost/projects/project-1'),
          params: { id: 'project-1' },
          data: {
            user: {
              id: 'user-1',
              first_name: 'Test',
              last_name: 'User',
              email: 'test@example.com',
              is_active: true,
              role: 'planer',
              permissions: [],
              can_access_user_directory: false,
              created_at: '2026-01-01T00:00:00.000Z',
              updated_at: '2026-01-01T00:00:00.000Z',
              failed_login_attempts: 0
            }
          }
        });
        return () => {};
      }
    },
    projectDetailService: {
      getProject: vi.fn().mockResolvedValue({
        id: 'project-1',
        name: 'Alpha',
        description: '',
        status: 'planned',
        phase_id: 'phase-1',
        creator_id: 'user-1',
        created_at: '2026-01-01T00:00:00.000Z',
        updated_at: '2026-01-01T00:00:00.000Z'
      }),
      listUsers: vi.fn().mockResolvedValue({ items: [] })
    }
  };
});

vi.mock('$app/stores', () => ({
  page: state.pageStore
}));

vi.mock('$lib/i18n/translator.js', () => ({
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

vi.mock('$lib/components/project/ProjectDetailService.js', () => ({
  projectDetailService: state.projectDetailService
}));

vi.mock('$lib/services/projectCollaboration.svelte.js', () => ({
  ProjectCollaborationState: class {
    socketStatus = 'connected';
    onlineUsers = [];
    connect = vi.fn();
    disconnect = vi.fn();
    buildFieldDeviceEditorsByDevice = vi.fn(() => new Map());
    publishFieldDeviceDraftState = vi.fn();
    publishFieldDeviceDelta = vi.fn();
  }
}));

vi.mock('$lib/components/history/HistoryTimelineDialog.svelte', async () => {
  const { default: SlotContainer } = await import('../../setup/stubs/SlotContainer.svelte');
  return { default: SlotContainer };
});

vi.mock('$lib/components/confirm-dialog.svelte', async () => {
  const { default: SlotContainer } = await import('../../setup/stubs/SlotContainer.svelte');
  return { default: SlotContainer };
});

vi.mock('$lib/components/user-avatar.svelte', async () => {
  const { default: SlotContainer } = await import('../../setup/stubs/SlotContainer.svelte');
  return { default: SlotContainer };
});

vi.mock('$lib/components/facility/control-cabinets/ControlCabinetListView.svelte', async () => {
  const { default: SlotContainer } = await import('../../setup/stubs/SlotContainer.svelte');
  return { default: SlotContainer };
});

vi.mock('$lib/components/facility/sps-controllers/SPSControllerListView.svelte', async () => {
  const { default: SlotContainer } = await import('../../setup/stubs/SlotContainer.svelte');
  return { default: SlotContainer };
});

vi.mock('$lib/components/facility/field-device/FieldDeviceListView.svelte', async () => {
  const { default: SlotContainer } = await import('../../setup/stubs/SlotContainer.svelte');
  return { default: SlotContainer };
});

import ProjectDetailPage from '../../../src/routes/(app)/projects/[id]/+page.svelte';

describe('/projects/:id permission surface', () => {
  beforeEach(() => {
    state.resetPermissions();
    state.projectDetailService.getProject.mockClear();
    state.projectDetailService.listUsers.mockClear();
  });

  it('hides the settings link when project.update would be forbidden', () => {
    state.setPermissions([permission('project', 'listAll')]);

    const { container } = render(ProjectDetailPage);

    expect(container.querySelector('a[href="/projects/project-1/settings"]')).toBeNull();
  });

  it('does not open project subareas that would be forbidden for a project.listAll-only user', () => {
    state.setPermissions([permission('project', 'listAll')]);

    render(ProjectDetailPage);

    expect(state.projectDetailService.listUsers).not.toHaveBeenCalled();
    expect(screen.queryByLabelText('history.open')).not.toBeInTheDocument();
    expect(screen.queryByText('projects.control_cabinets.title')).not.toBeInTheDocument();
    expect(screen.queryByText('projects.sps_controllers.title')).not.toBeInTheDocument();
    expect(screen.queryByText('projects.field_devices.title')).not.toBeInTheDocument();
  });

  it('shows the settings link when project.update is granted', () => {
    state.setPermissions([permission('project', 'listAll'), permission('project', 'update')]);

    const { container } = render(ProjectDetailPage);

    expect(container.querySelector('a[href="/projects/project-1/settings"]')).not.toBeNull();
    expect(state.projectDetailService.listUsers).toHaveBeenCalledWith('project-1');
  });

  it('shows only the project facility tabs backed by granted read permissions', async () => {
    state.setPermissions([
      permission('project', 'listAll'),
      permission('project.controlcabinet'),
      permission('project.fielddevice')
    ]);

    render(ProjectDetailPage);

    await waitFor(() =>
      expect(screen.getByText('projects.control_cabinets.title')).toBeInTheDocument()
    );
    expect(screen.queryByText('projects.sps_controllers.title')).not.toBeInTheDocument();
    expect(screen.getByText('projects.field_devices.title')).toBeInTheDocument();
  });

  it('shows the project history action only when timeline.read is granted', () => {
    state.setPermissions([permission('project', 'listAll'), permission('timeline')]);

    render(ProjectDetailPage);

    expect(screen.getByLabelText('history.open')).toBeInTheDocument();
  });
});
