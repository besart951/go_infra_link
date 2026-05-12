<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Dialog from '$lib/components/ui/dialog/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { Textarea } from '$lib/components/ui/textarea/index.js';
  import type { ProjectStatus } from '$lib/domain/project/index.js';
  import ProjectPhaseSelect from './ProjectPhaseSelect.svelte';
  import type { CreateProjectField, ProjectListPageState } from './ProjectListPageState.svelte.js';

  type StatusOption = {
    value: ProjectStatus;
    label: string;
  };

  type ProjectCreateDialogLabels = {
    title: string;
    description: string;
    name: string;
    status: string;
    startDate: string;
    phase: string;
    descriptionField: string;
    cancel: string;
    create: string;
    namePlaceholder: string;
    descriptionPlaceholder: string;
    errorTitle: string;
  };

  type Props = {
    state: ProjectListPageState;
    statusOptions: readonly StatusOption[];
    labels: ProjectCreateDialogLabels;
  };

  let { state, statusOptions, labels }: Props = $props();

  function errorId(field: CreateProjectField): string {
    return `project_${field}_create_error`;
  }

  function fieldError(field: CreateProjectField): string | undefined {
    return state.getCreateFieldError(field);
  }

  function hasFieldError(field: CreateProjectField): boolean {
    return Boolean(fieldError(field));
  }

  function describedBy(field: CreateProjectField): string | undefined {
    return hasFieldError(field) ? errorId(field) : undefined;
  }

  function clearFieldError(field: CreateProjectField): void {
    state.clearCreateFieldError(field);
  }

  function handleSubmit(event: SubmitEvent): void {
    event.preventDefault();
    void state.submitCreate();
  }
</script>

<Dialog.Root bind:open={state.createOpen}>
  <Dialog.Content class="sm:max-w-2xl">
    <Dialog.Header>
      <Dialog.Title>{labels.title}</Dialog.Title>
      <Dialog.Description>{labels.description}</Dialog.Description>
    </Dialog.Header>

    <form class="space-y-6" onsubmit={handleSubmit}>
      {#if state.createError}
        <div
          role="alert"
          class="rounded-xl border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive"
        >
          <p class="font-medium">{labels.errorTitle}</p>
          <p>{state.createError}</p>
        </div>
      {/if}

      <div class="grid gap-4 md:grid-cols-2">
        <div class="flex flex-col gap-2">
          <Label for="project_name_create">{labels.name}</Label>
          <Input
            id="project_name_create"
            placeholder={labels.namePlaceholder}
            bind:value={state.form.name}
            disabled={state.createBusy}
            aria-invalid={hasFieldError('name')}
            aria-describedby={describedBy('name')}
            oninput={() => clearFieldError('name')}
          />
          {#if fieldError('name')}
            <p id={errorId('name')} class="text-sm text-destructive">{fieldError('name')}</p>
          {/if}
        </div>

        <div class="flex flex-col gap-2">
          <Label for="project_status_create">{labels.status}</Label>
          <select
            id="project_status_create"
            class="h-9 rounded-md border border-input bg-background px-3 text-sm font-medium shadow-xs transition-[color,box-shadow] outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50"
            bind:value={state.form.status}
            disabled={state.createBusy}
            aria-invalid={hasFieldError('status')}
            aria-describedby={describedBy('status')}
            onchange={() => clearFieldError('status')}
          >
            {#each statusOptions as option (option.value)}
              <option value={option.value}>{option.label}</option>
            {/each}
          </select>
          {#if fieldError('status')}
            <p id={errorId('status')} class="text-sm text-destructive">{fieldError('status')}</p>
          {/if}
        </div>

        <div class="flex flex-col gap-2">
          <Label for="project_start_create">{labels.startDate}</Label>
          <Input
            id="project_start_create"
            type="date"
            bind:value={state.form.start_date}
            disabled={state.createBusy}
            aria-invalid={hasFieldError('start_date')}
            aria-describedby={describedBy('start_date')}
            oninput={() => clearFieldError('start_date')}
          />
          {#if fieldError('start_date')}
            <p id={errorId('start_date')} class="text-sm text-destructive">
              {fieldError('start_date')}
            </p>
          {/if}
        </div>

        <div class="flex flex-col gap-2">
          <Label for="project_phase_create">{labels.phase}</Label>
          <ProjectPhaseSelect
            id="project_phase_create"
            bind:value={state.form.phase_id}
            width="w-full"
            disabled={state.createBusy}
          />
          {#if fieldError('phase_id')}
            <p id={errorId('phase_id')} class="text-sm text-destructive">
              {fieldError('phase_id')}
            </p>
          {/if}
        </div>

        <div class="flex flex-col gap-2 md:col-span-2">
          <Label for="project_desc_create">{labels.descriptionField}</Label>
          <Textarea
            id="project_desc_create"
            placeholder={labels.descriptionPlaceholder}
            rows={3}
            bind:value={state.form.description}
            disabled={state.createBusy}
            aria-invalid={hasFieldError('description')}
            aria-describedby={describedBy('description')}
            oninput={() => clearFieldError('description')}
          />
          {#if fieldError('description')}
            <p id={errorId('description')} class="text-sm text-destructive">
              {fieldError('description')}
            </p>
          {/if}
        </div>
      </div>

      <Dialog.Footer>
        <Button
          type="button"
          variant="outline"
          onclick={() => state.setCreateOpen(false)}
          disabled={state.createBusy}
        >
          {labels.cancel}
        </Button>
        <Button type="submit" disabled={!state.canSubmitCreate}>
          {labels.create}
        </Button>
      </Dialog.Footer>
    </form>
  </Dialog.Content>
</Dialog.Root>
