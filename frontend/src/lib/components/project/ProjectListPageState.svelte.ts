import { goto } from '$app/navigation';
import { createProject } from '$lib/infrastructure/api/project.adapter.js';
import { addToast } from '$lib/components/toast.svelte';
import type { CreateProjectRequest, Project, ProjectStatus } from '$lib/domain/project/index.js';
import { useOptimisticUpdate } from '$lib/hooks/useOptimisticUpdate.svelte.js';
import { t as translate } from '$lib/i18n/index.js';
import { projectListStore } from '$lib/stores/projects/projectListStore.js';

type CreateProjectForm = {
  name: string;
  description: string;
  status: ProjectStatus;
  start_date: string;
  phase_id: string;
};

function todayInputValue(): string {
  const now = new Date();
  const yyyy = now.getFullYear();
  const mm = String(now.getMonth() + 1).padStart(2, '0');
  const dd = String(now.getDate()).padStart(2, '0');
  return `${yyyy}-${mm}-${dd}`;
}

function emptyCreateProjectForm(): CreateProjectForm {
  return {
    name: '',
    description: '',
    status: 'planned',
    start_date: todayInputValue(),
    phase_id: ''
  };
}

export class ProjectListPageState {
  createOpen = $state(false);
  createBusy = $state(false);
  form = $state<CreateProjectForm>(emptyCreateProjectForm());

  private readonly optimisticCreate = useOptimisticUpdate<Project>({
    onSuccess: (project) => {
      void goto(`/projects/${project.id}`);
    },
    onError: (error) => {
      addToast(
        error instanceof Error ? error.message : translate('project.creation_failed'),
        'error'
      );
    }
  });

  initialize(): void {
    projectListStore.load();
  }

  canSubmitCreate(): boolean {
    return (
      this.form.name.trim().length > 0 && this.form.phase_id.trim().length > 0 && !this.createBusy
    );
  }

  async submitCreate(): Promise<void> {
    if (!this.canSubmitCreate()) return;
    this.createBusy = true;

    const payload: CreateProjectRequest = {
      name: this.form.name.trim(),
      description: this.form.description.trim() || undefined,
      status: this.form.status,
      start_date: this.form.start_date
        ? new Date(`${this.form.start_date}T00:00:00Z`).toISOString()
        : undefined,
      phase_id: this.form.phase_id
    };
    const originalStatus = this.form.status;

    try {
      await this.optimisticCreate.execute(
        () => {
          this.createOpen = false;
          this.form = emptyCreateProjectForm();
          addToast(translate('projects.page.creating'), 'info', 2000);
        },
        async () => {
          const project = await createProject(payload);
          addToast(translate('project.project_created'), 'success');
          projectListStore.reload();
          return project;
        },
        () => {
          this.createOpen = true;
          this.form = {
            name: payload.name,
            description: payload.description ?? '',
            status: originalStatus,
            start_date: payload.start_date ? payload.start_date.split('T')[0] : todayInputValue(),
            phase_id: payload.phase_id
          };
        }
      );
    } finally {
      this.createBusy = false;
    }
  }

  handleStatusChange(event: Event): void {
    const value = (event.target as HTMLSelectElement).value;
    projectListStore.setStatus(value === 'all' ? 'all' : (value as ProjectStatus));
  }
}
