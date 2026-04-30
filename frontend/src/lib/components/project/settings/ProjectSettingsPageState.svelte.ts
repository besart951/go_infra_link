import { addToast } from '$lib/components/toast.svelte';
import { objectDataRepository } from '$lib/infrastructure/api/objectDataRepository.js';
import {
  addProjectObjectData,
  addProjectUser,
  getProject,
  listProjectObjectData,
  listProjectUsers,
  removeProjectObjectData,
  removeProjectUser,
  updateProject
} from '$lib/infrastructure/api/project.adapter.js';
import { listUsers } from '$lib/infrastructure/api/user.adapter.js';
import type { ObjectData } from '$lib/domain/facility/index.js';
import type { Project, ProjectStatus, UpdateProjectRequest } from '$lib/domain/project/index.js';
import type { User } from '$lib/domain/user/index.js';
import { t as translate } from '$lib/i18n/index.js';
import { confirm } from '$lib/stores/confirm-dialog.js';
import { canPerform } from '$lib/utils/permissions.js';

export type ProjectSettingsTab = 'settings' | 'users' | 'object-data';
export type ObjectDataStatusFilter = 'all' | 'active' | 'inactive';

const emptyForm = () => ({
  name: '',
  description: '',
  status: 'planned' as ProjectStatus,
  start_date: '',
  phase_id: ''
});

export class ProjectSettingsPageState {
  project = $state<Project | null>(null);
  loading = $state(true);
  error = $state<string | null>(null);
  saving = $state(false);
  activeTab = $state<ProjectSettingsTab>('settings');
  form = $state(emptyForm());

  projectUsers = $state<User[]>([]);
  availableUsers = $state<User[]>([]);
  usersLoading = $state(false);
  usersLoaded = $state(false);
  userSearch = $state('');

  projectObjectData = $state<ObjectData[]>([]);
  objectDataLoading = $state(false);
  objectDataLoaded = $state(false);
  objectDataSearch = $state('');
  objectDataStatusFilter = $state<ObjectDataStatusFilter>('all');
  editingObjectData: ObjectData | undefined = $state(undefined);
  showObjectDataForm = $state(false);

  readonly statusOptions: Array<{ value: ProjectStatus; label: string }> = [
    { value: 'planned', label: 'projects.settings.status.planned' },
    { value: 'ongoing', label: 'projects.settings.status.ongoing' },
    { value: 'completed', label: 'projects.settings.status.completed' }
  ];

  private userSearchTimeout: ReturnType<typeof setTimeout> | null = null;
  private objectDataSearchTimeout: ReturnType<typeof setTimeout> | null = null;

  constructor(private readonly resolveProjectId: () => string) {}

  get projectId(): string {
    return this.resolveProjectId();
  }

  get canUpdateProject(): boolean {
    return canPerform('update', 'project');
  }

  get canUpdateObjectData(): boolean {
    return canPerform('update', 'objectdata');
  }

  getFilteredObjectData(): ObjectData[] {
    if (this.objectDataStatusFilter === 'all') {
      return this.projectObjectData;
    }
    const isActive = this.objectDataStatusFilter === 'active';
    return this.projectObjectData.filter((item) => this.isObjectDataActive(item) === isActive);
  }

  formatDate(value?: string): string {
    if (!value) return '-';
    try {
      return new Date(value).toLocaleDateString();
    } catch {
      return value;
    }
  }

  hydrateForm(project: Project): void {
    this.form = {
      name: project.name ?? '',
      description: project.description ?? '',
      status: project.status,
      start_date: project.start_date ? project.start_date.slice(0, 10) : '',
      phase_id: project.phase_id ?? ''
    };
  }

  resetForm(): void {
    if (this.project) {
      this.hydrateForm(this.project);
    }
  }

  async saveSettings(): Promise<void> {
    if (!this.projectId || !this.project) return;
    if (!this.canUpdateProject) return;
    if (!this.form.phase_id.trim()) {
      addToast(translate('projects.settings.phase_required'), 'error');
      return;
    }

    this.saving = true;
    try {
      const payload: UpdateProjectRequest = {
        id: this.projectId,
        name: this.form.name.trim(),
        description: this.form.description.trim(),
        status: this.form.status,
        start_date: this.form.start_date || null,
        phase_id: this.form.phase_id
      };
      this.project = await updateProject(this.projectId, payload);
      this.hydrateForm(this.project);
      addToast(translate('projects.settings.updated'), 'success');
    } catch (error) {
      addToast(
        error instanceof Error ? error.message : translate('projects.settings.update_failed'),
        'error'
      );
    } finally {
      this.saving = false;
    }
  }

  async load(): Promise<void> {
    if (!this.projectId) {
      this.error = translate('projects.errors.missing_id');
      this.loading = false;
      return;
    }

    this.loading = true;
    this.error = null;
    this.usersLoaded = false;
    this.objectDataLoaded = false;

    try {
      this.project = await getProject(this.projectId);
      this.hydrateForm(this.project);
    } catch (error) {
      const message =
        error instanceof Error ? error.message : translate('projects.errors.load_failed');
      this.error = message;
      addToast(message, 'error');
    } finally {
      this.loading = false;
    }
  }

  async loadUsers(): Promise<void> {
    if (!this.projectId) return;
    this.usersLoading = true;
    try {
      const [projectUsersRes, usersRes] = await Promise.all([
        listProjectUsers(this.projectId),
        listUsers({ page: 1, limit: 100, search: this.userSearch.trim() || undefined })
      ]);
      this.projectUsers = projectUsersRes.items;
      this.availableUsers = usersRes.items;
      this.usersLoaded = true;
    } catch (error) {
      addToast(
        error instanceof Error ? error.message : translate('projects.users.load_failed'),
        'error'
      );
    } finally {
      this.usersLoading = false;
    }
  }

  isUserInProject(userId: string): boolean {
    return this.projectUsers.some((member) => member.id === userId);
  }

  async toggleUser(user: User): Promise<void> {
    if (!this.projectId) return;
    if (!this.canUpdateProject) return;

    if (this.isUserInProject(user.id)) {
      const ok = await confirm({
        title: translate('projects.users.remove_title'),
        message: translate('projects.users.remove_message'),
        confirmText: translate('projects.users.remove_confirm'),
        cancelText: translate('common.cancel'),
        variant: 'destructive'
      });
      if (!ok) return;

      try {
        await removeProjectUser(this.projectId, user.id);
        addToast(translate('projects.users.removed'), 'success');
        await this.loadUsers();
      } catch (error) {
        addToast(
          error instanceof Error ? error.message : translate('projects.users.remove_failed'),
          'error'
        );
      }
      return;
    }

    try {
      await addProjectUser(this.projectId, user.id);
      addToast(translate('projects.users.added'), 'success');
      await this.loadUsers();
    } catch (error) {
      addToast(
        error instanceof Error ? error.message : translate('projects.users.add_failed'),
        'error'
      );
    }
  }

  async loadObjectData(): Promise<void> {
    if (!this.projectId) return;
    this.objectDataLoading = true;
    try {
      const projectRes = await listProjectObjectData(this.projectId, {
        page: 1,
        limit: 100,
        search: this.objectDataSearch.trim() || undefined
      });
      this.projectObjectData = projectRes.items ?? [];
    } catch (error) {
      addToast(
        error instanceof Error ? error.message : translate('projects.object_data.load_failed'),
        'error'
      );
    } finally {
      this.objectDataLoading = false;
      this.objectDataLoaded = true;
    }
  }

  async toggleObjectData(objectData: ObjectData): Promise<void> {
    if (!this.projectId) return;
    if (!this.canUpdateProject) return;
    const isAssigned = objectData.project_id === this.projectId;
    const isActive = isAssigned && objectData.is_active;

    if (isActive) {
      const ok = await confirm({
        title: translate('projects.object_data.deactivate_title'),
        message: translate('projects.object_data.deactivate_message'),
        confirmText: translate('projects.object_data.deactivate_confirm'),
        cancelText: translate('common.cancel'),
        variant: 'destructive'
      });
      if (!ok) return;

      try {
        await removeProjectObjectData(this.projectId, objectData.id);
        addToast(translate('projects.object_data.deactivated'), 'success');
        await this.loadObjectData();
      } catch (error) {
        addToast(
          error instanceof Error
            ? error.message
            : translate('projects.object_data.deactivate_failed'),
          'error'
        );
      }
      return;
    }

    try {
      await addProjectObjectData(this.projectId, objectData.id);
      addToast(translate('projects.object_data.activated'), 'success');
      await this.loadObjectData();
    } catch (error) {
      addToast(
        error instanceof Error ? error.message : translate('projects.object_data.activate_failed'),
        'error'
      );
    }
  }

  isObjectDataActive(item: ObjectData): boolean {
    return item.project_id === this.projectId && item.is_active;
  }

  async editObjectData(item: ObjectData): Promise<void> {
    if (!this.canUpdateObjectData) return;
    try {
      this.editingObjectData = await objectDataRepository.get(item.id);
    } catch {
      this.editingObjectData = item;
    }
    this.showObjectDataForm = true;
  }

  handleObjectDataSuccess(): void {
    this.showObjectDataForm = false;
    this.editingObjectData = undefined;
    void this.loadObjectData();
  }

  handleObjectDataCancel(): void {
    this.showObjectDataForm = false;
    this.editingObjectData = undefined;
  }

  handleUserSearchInput(): void {
    if (this.userSearchTimeout) {
      clearTimeout(this.userSearchTimeout);
    }
    this.userSearchTimeout = setTimeout(() => {
      if (this.activeTab === 'users' && !this.usersLoading) {
        void this.loadUsers();
      }
    }, 300);
  }

  handleObjectDataSearchInput(): void {
    if (this.objectDataSearchTimeout) {
      clearTimeout(this.objectDataSearchTimeout);
    }
    this.objectDataSearchTimeout = setTimeout(() => {
      if (this.activeTab === 'object-data' && !this.objectDataLoading) {
        void this.loadObjectData();
      }
    }, 300);
  }

  ensureActiveTabLoaded(): void {
    if (this.activeTab === 'users' && !this.usersLoaded && !this.usersLoading) {
      void this.loadUsers();
    }

    if (this.activeTab === 'object-data' && !this.objectDataLoaded && !this.objectDataLoading) {
      void this.loadObjectData();
    }
  }
}
