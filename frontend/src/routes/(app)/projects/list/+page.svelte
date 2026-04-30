<script lang="ts">
  import { onMount } from 'svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Textarea } from '$lib/components/ui/textarea/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import { ArrowLeft, Plus } from '@lucide/svelte';
  import PaginatedList from '$lib/components/list/PaginatedList.svelte';
  import ProjectPhaseSelect from '$lib/components/project/ProjectPhaseSelect.svelte';
  import { ProjectListPageState } from '$lib/components/project/ProjectListPageState.svelte.js';
  import { projectListStore } from '$lib/stores/projects/projectListStore.js';
  import type { Project, ProjectStatus } from '$lib/domain/project/index.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();
  const state = new ProjectListPageState();

  function getStatusClass(status: string): string {
    switch (status) {
      case 'completed':
        return 'bg-green-100 text-green-800 dark:bg-green-900 dark:text-green-200';
      case 'ongoing':
        return 'bg-blue-100 text-blue-800 dark:bg-blue-900 dark:text-blue-200';
      default:
        return 'bg-gray-100 text-gray-800 dark:bg-gray-800 dark:text-gray-200';
    }
  }

  const statusOptions: Array<{ value: ProjectStatus | 'all'; label: string }> = [
    { value: 'all', label: $t('messages.all_statuses') },
    { value: 'planned', label: $t('messages.planned') },
    { value: 'ongoing', label: $t('messages.ongoing') },
    { value: 'completed', label: $t('messages.completed') }
  ];

  const createStatusOptions: Array<{ value: ProjectStatus; label: string }> = [
    { value: 'planned', label: $t('messages.planned') },
    { value: 'ongoing', label: $t('messages.ongoing') },
    { value: 'completed', label: $t('messages.completed') }
  ];

  onMount(() => {
    state.initialize();
  });
</script>

<svelte:head>
  <title>{$t('navigation.projects')} | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">{$t('navigation.projects')}</h1>
      <p class="text-sm text-muted-foreground">
        {$t('pages.projects_desc')}
      </p>
    </div>
    <div class="flex flex-col gap-2 sm:flex-row">
      <Button variant="outline" href="/projects">
        <ArrowLeft class="size-4" />
        {$t('hub.back_to_overview')}
      </Button>
      {#if canPerform('create', 'project')}
        <Button onclick={() => (state.createOpen = !state.createOpen)}>
          <Plus class="mr-2 size-4" />
          {$t('common.create')}
        </Button>
      {/if}
    </div>
  </div>

  {#if state.createOpen}
    <div class="rounded-lg border bg-background p-4">
      <div class="grid gap-4 md:grid-cols-2">
        <div class="flex flex-col gap-2">
          <label class="text-sm font-medium" for="project_name_create">{$t('common.name')}</label>
          <Input
            id="project_name_create"
            placeholder={$t('messages.project_name_placeholder')}
            bind:value={state.form.name}
            disabled={state.createBusy}
          />
        </div>

        <div class="flex flex-col gap-2">
          <label class="text-sm font-medium" for="project_status_create"
            >{$t('common.status')}</label
          >
          <select
            id="project_status_create"
            class="h-9 rounded-md border border-input bg-background px-3 text-sm font-medium shadow-xs"
            bind:value={state.form.status}
            disabled={state.createBusy}
          >
            {#each createStatusOptions as opt}
              <option value={opt.value}>{opt.label}</option>
            {/each}
          </select>
        </div>

        <div class="flex flex-col gap-2">
          <label class="text-sm font-medium" for="project_start_create"
            >{$t('messages.start_date')}</label
          >
          <Input
            id="project_start_create"
            type="date"
            bind:value={state.form.start_date}
            disabled={state.createBusy}
          />
        </div>

        <div class="flex flex-col gap-2">
          <label class="text-sm font-medium" for="project_phase_create"
            >{$t('messages.phase')}</label
          >
          <ProjectPhaseSelect
            id="project_phase_create"
            bind:value={state.form.phase_id}
            width="w-full"
            disabled={state.createBusy}
          />
        </div>

        <div class="flex flex-col gap-2 md:col-span-2">
          <label class="text-sm font-medium" for="project_desc_create"
            >{$t('common.description')}</label
          >
          <Textarea
            id="project_desc_create"
            placeholder={$t('messages.project_description_placeholder')}
            rows={3}
            bind:value={state.form.description}
            disabled={state.createBusy}
          />
        </div>
      </div>

      <div class="mt-4 flex items-center justify-end gap-2">
        <Button
          variant="outline"
          onclick={() => (state.createOpen = false)}
          disabled={state.createBusy}>{$t('common.cancel')}</Button
        >
        <Button onclick={() => state.submitCreate()} disabled={!state.canSubmitCreate()}
          >{$t('common.create')}</Button
        >
      </div>
    </div>
  {/if}

  <div class="flex flex-wrap items-center gap-3">
    <label class="text-sm font-medium" for="project_status_filter">{$t('common.status')}</label>
    <select
      id="project_status_filter"
      class="h-9 rounded-md border border-input bg-background px-3 text-sm font-medium shadow-xs"
      value={$projectListStore.status}
      onchange={(event) => state.handleStatusChange(event)}
    >
      {#each statusOptions as opt}
        <option value={opt.value}>{opt.label}</option>
      {/each}
    </select>
  </div>

  <PaginatedList
    state={$projectListStore}
    columns={[
      { key: 'name', label: $t('common.name') },
      { key: 'status', label: $t('common.status') },
      { key: 'start_date', label: $t('messages.start_date') },
      { key: 'created', label: $t('messages.created') },
      { key: 'actions', label: $t('messages.actions'), width: 'w-[100px]' }
    ]}
    searchPlaceholder={$t('messages.search_projects')}
    emptyMessage={$t('messages.no_projects_found')}
    onSearch={(text) => projectListStore.search(text)}
    onPageChange={(page) => projectListStore.goToPage(page)}
    onReload={() => projectListStore.reload()}
  >
    {#snippet rowSnippet(project: Project)}
      <Table.Cell class="font-medium">
        <a href="/projects/{project.id}" class="hover:underline">
          {project.name}
        </a>
      </Table.Cell>
      <Table.Cell>
        <span
          class="inline-flex items-center rounded-full px-2.5 py-0.5 text-xs font-medium {getStatusClass(
            project.status
          )}"
        >
          {project.status}
        </span>
      </Table.Cell>
      <Table.Cell>
        {project.start_date ? new Date(project.start_date).toLocaleDateString() : '-'}
      </Table.Cell>
      <Table.Cell>
        {new Date(project.created_at).toLocaleDateString()}
      </Table.Cell>
      <Table.Cell>
        <Button variant="ghost" size="sm" href="/projects/{project.id}"
          >{$t('messages.view')}</Button
        >
      </Table.Cell>
    {/snippet}
  </PaginatedList>
</div>
