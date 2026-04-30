<script lang="ts">
  import { onMount } from 'svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import { ArrowLeft, Plus, Pencil, Trash2, Eye } from '@lucide/svelte';
  import PaginatedList from '$lib/components/list/PaginatedList.svelte';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { phaseListStore } from '$lib/stores/phases/phaseListStore.js';
  import type { Phase } from '$lib/domain/phase/index.js';
  import PhaseForm from '$lib/components/project/PhaseForm.svelte';
  import { PhaseListPageState } from '$lib/components/project/PhaseListPageState.svelte.js';
  import { canPerform } from '$lib/utils/permissions.js';

  const t = createTranslator();
  const state = new PhaseListPageState();

  onMount(() => {
    state.initialize();
  });
</script>

<ConfirmDialog />

<svelte:head>
  <title>{$t('phases.page.title')}</title>
</svelte:head>

<div class="flex flex-col gap-6">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">{$t('phases.page.heading')}</h1>
      <p class="text-sm text-muted-foreground">{$t('phases.page.description')}</p>
    </div>
    <div class="flex flex-col gap-2 sm:flex-row">
      <Button variant="outline" href="/projects">
        <ArrowLeft class="size-4" />
        {$t('hub.back_to_overview')}
      </Button>
      {#if !state.showForm && canPerform('create', 'phase')}
        <Button onclick={() => state.handleCreate()}>
          <Plus class="mr-2 size-4" />
          {$t('phases.page.new')}
        </Button>
      {/if}
    </div>
  </div>

  {#if state.showForm}
    <PhaseForm
      initialData={state.editingPhase}
      onSuccess={() => state.handleSuccess()}
      onCancel={() => state.handleCancel()}
    />
  {/if}

  <PaginatedList
    state={$phaseListStore}
    columns={[
      { key: 'name', label: $t('common.name') },
      { key: 'created', label: $t('common.created') },
      { key: 'actions', label: $t('common.actions'), width: 'w-[140px]' }
    ]}
    searchPlaceholder={$t('phases.page.search_placeholder')}
    emptyMessage={$t('phases.page.empty')}
    onSearch={(text) => phaseListStore.search(text)}
    onPageChange={(page) => phaseListStore.goToPage(page)}
    onReload={() => phaseListStore.reload()}
  >
    {#snippet rowSnippet(phase: Phase)}
      <Table.Cell class="font-medium">
        <a href="/projects/phases/{phase.id}" class="hover:underline">
          {phase.name}
        </a>
      </Table.Cell>
      <Table.Cell>
        {phase.created_at
          ? new Date(phase.created_at).toLocaleDateString()
          : $t('common.not_available')}
      </Table.Cell>
      <Table.Cell>
        <div class="flex items-center gap-2">
          {#if canPerform('update', 'phase')}
            <Button variant="ghost" size="icon" onclick={() => state.handleEdit(phase)}>
              <Pencil class="size-4" />
            </Button>
          {/if}
          <Button variant="ghost" size="icon" href="/projects/phases/{phase.id}">
            <Eye class="size-4" />
          </Button>
          {#if canPerform('delete', 'phase')}
            <Button
              variant="ghost"
              size="icon"
              disabled={state.deleting}
              onclick={() => state.handleDelete(phase)}
            >
              <Trash2 class="size-4 text-destructive" />
            </Button>
          {/if}
        </div>
      </Table.Cell>
    {/snippet}
  </PaginatedList>
</div>
