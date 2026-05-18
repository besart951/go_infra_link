import { goto } from '$app/navigation';
import { ApiException } from '$lib/api/client.js';
import { fieldErrorsFromApiDetails, type FieldErrorMap } from '$lib/api/errorResponse.js';
import { createProject } from '$lib/infrastructure/api/project.adapter.js';
import { addToast } from '$lib/components/toast.svelte';
import type { CreateProjectRequest, Project, ProjectStatus } from '$lib/domain/project/index.js';
import { useOptimisticUpdate } from '$lib/hooks/useOptimisticUpdate.svelte.js';
import { t as translate } from '$lib/i18n/index.js';
import {
  projectListStore,
  type ProjectStatusFilter
} from '$lib/stores/projects/projectListStore.js';

export type CreateProjectForm = {
  name: string;
  description: string;
  status: ProjectStatus;
  start_date: string;
  phase_id: string;
};

export type CreateProjectField = keyof CreateProjectForm;
type CreateProjectErrors = Partial<Record<CreateProjectField, string>>;

const CREATE_PROJECT_FIELDS: readonly CreateProjectField[] = [
  'name',
  'description',
  'status',
  'start_date',
  'phase_id'
];

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
  createErrors = $state<CreateProjectErrors>({});
  createError = $state<string | null>(null);

  canSubmitCreate = $derived.by(() => {
    return (
      this.form.name.trim().length > 0 && this.form.phase_id.trim().length > 0 && !this.createBusy
    );
  });

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

  setCreateOpen(open: boolean): void {
    this.createOpen = open;
    if (!open) {
      this.clearCreateFeedback();
    }
  }

  toggleCreateDialog(): void {
    this.setCreateOpen(!this.createOpen);
  }

  clearCreateFieldError(field: CreateProjectField): void {
    if (!this.createErrors[field]) return;

    const nextErrors: CreateProjectErrors = {};
    for (const createField of CREATE_PROJECT_FIELDS) {
      if (createField !== field && this.createErrors[createField]) {
        nextErrors[createField] = this.createErrors[createField];
      }
    }

    this.createErrors = nextErrors;
    if (Object.keys(nextErrors).length === 0) {
      this.createError = null;
    }
  }

  getCreateFieldError(field: CreateProjectField): string | undefined {
    return this.createErrors[field];
  }

  async submitCreate(): Promise<void> {
    if (!this.canSubmitCreate) return;
    this.createBusy = true;
    this.clearCreateFeedback();

    const payload = this.toCreatePayload();
    const originalStatus = this.form.status;

    try {
      await this.optimisticCreate.execute(
        () => {
          this.setCreateOpen(false);
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
          this.form = this.restoreCreateForm(payload, originalStatus);
        }
      );
    } catch (error) {
      this.createErrors = this.mapCreateFieldErrors(error);
      this.createError =
        Object.keys(this.createErrors).length > 0
          ? translate('errors.validation_error')
          : this.toCreateErrorMessage(error);
    } finally {
      this.createBusy = false;
    }
  }

  setStatusFilter(status: ProjectStatusFilter): void {
    projectListStore.setStatus(status);
  }

  setPhaseFilter(phaseId: string): void {
    projectListStore.setPhase(phaseId);
  }

  private clearCreateFeedback(): void {
    this.createErrors = {};
    this.createError = null;
  }

  private toCreatePayload(): CreateProjectRequest {
    return {
      name: this.form.name.trim(),
      description: this.form.description.trim() || undefined,
      status: this.form.status,
      start_date: this.form.start_date
        ? new Date(`${this.form.start_date}T00:00:00Z`).toISOString()
        : undefined,
      phase_id: this.form.phase_id
    };
  }

  private restoreCreateForm(
    payload: CreateProjectRequest,
    originalStatus: ProjectStatus
  ): CreateProjectForm {
    return {
      name: payload.name,
      description: payload.description ?? '',
      status: originalStatus,
      start_date: payload.start_date ? payload.start_date.split('T')[0] : todayInputValue(),
      phase_id: payload.phase_id
    };
  }

  private mapCreateFieldErrors(error: unknown): CreateProjectErrors {
    if (!(error instanceof ApiException)) {
      return {};
    }

    const apiErrors = fieldErrorsFromApiDetails(error.details);
    const createErrors: CreateProjectErrors = {};

    for (const field of CREATE_PROJECT_FIELDS) {
      const message = this.resolveCreateFieldError(apiErrors, field);
      if (message) {
        createErrors[field] = message;
      }
    }

    return createErrors;
  }

  private resolveCreateFieldError(
    errors: FieldErrorMap,
    field: CreateProjectField
  ): string | undefined {
    const directMessage = errors[field];
    if (directMessage) {
      return directMessage;
    }

    for (const prefix of ['project', 'data', 'body', 'payload', 'request']) {
      const prefixedMessage = errors[`${prefix}.${field}`];
      if (prefixedMessage) {
        return prefixedMessage;
      }
    }

    return undefined;
  }

  private toCreateErrorMessage(error: unknown): string {
    if (error instanceof Error) {
      return error.message;
    }

    return translate('project.creation_failed');
  }
}
