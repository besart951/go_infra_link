<script lang="ts">
  import { onMount } from 'svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import PaginatedList from '$lib/components/list/PaginatedList.svelte';
  import ProjectCreateDialog from '$lib/components/project/ProjectCreateDialog.svelte';
  import ProjectListToolbar from '$lib/components/project/ProjectListToolbar.svelte';
  import { ProjectListPageState } from '$lib/components/project/ProjectListPageState.svelte.js';
  import ProjectStatusBadge from '$lib/components/project/ProjectStatusBadge.svelte';
  import { projectListStore } from '$lib/stores/projects/projectListStore.js';
  import type { Project, ProjectStatus } from '$lib/domain/project/index.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();
  const state = new ProjectListPageState();

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
  <EntityListHeader
    title={$t('navigation.projects')}
    description={$t('pages.projects_desc')}
    backHref="/projects"
    backLabel={$t('common.back')}
    createLabel={$t('common.create')}
    canCreate={canPerform('create', 'project')}
    createActive={state.createOpen}
    onCreateClick={() => state.toggleCreateDialog()}
  />

  <ProjectListToolbar
    statusLabel={$t('common.status')}
    statusValue={$projectListStore.status}
    options={statusOptions}
    phaseLabel={$t('messages.phase')}
    phaseValue={$projectListStore.phaseId}
    allPhasesLabel={$t('messages.all_phases')}
    phaseSearchPlaceholder={$t('phases.page.search_placeholder')}
    phaseEmptyText={$t('messages.no_phases_found')}
    disabled={$projectListStore.loading}
    onStatusChange={(status) => state.setStatusFilter(status)}
    onPhaseChange={(phaseId) => state.setPhaseFilter(phaseId)}
  />

  <PaginatedList
    state={$projectListStore}
    columns={[
      { key: 'name', label: $t('common.name') },
      { key: 'status', label: $t('common.status') },
      { key: 'start_date', label: $t('messages.start_date') },
      { key: 'phase', label: $t('messages.phase') },
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
        <ProjectStatusBadge status={project.status} label={$t(`messages.${project.status}`)} />
      </Table.Cell>
      <Table.Cell>
        {project.start_date ? new Date(project.start_date).toLocaleDateString() : '-'}
      </Table.Cell>
      <Table.Cell>
        {project.phase?.name?.trim() || $t('common.not_available')}
      </Table.Cell>
      <Table.Cell>
        <Button variant="ghost" size="sm" href="/projects/{project.id}"
          >{$t('messages.view')}</Button
        >
      </Table.Cell>
    {/snippet}
  </PaginatedList>
</div>

<ProjectCreateDialog
  {state}
  statusOptions={createStatusOptions}
  labels={{
    title: $t('common.create'),
    description: $t('pages.projects_desc'),
    name: $t('common.name'),
    status: $t('common.status'),
    startDate: $t('messages.start_date'),
    phase: $t('messages.phase'),
    descriptionField: $t('common.description'),
    cancel: $t('common.cancel'),
    create: $t('common.create'),
    namePlaceholder: $t('messages.project_name_placeholder'),
    descriptionPlaceholder: $t('messages.project_description_placeholder'),
    errorTitle: $t('messages.error')
  }}
/>
