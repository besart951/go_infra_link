<script lang="ts">
  import { goto } from '$app/navigation';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import PaginatedList from '$lib/components/list/PaginatedList.svelte';
  import ControlCabinetForm from '$lib/components/facility/forms/ControlCabinetForm.svelte';
  import HistoryTimelineDialog from '$lib/components/history/HistoryTimelineDialog.svelte';
  import type { ControlCabinet } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
  import PlusIcon from '@lucide/svelte/icons/plus';
  import { useControlCabinetState } from './state/context.svelte.js';

  const t = createTranslator();
  const listState = useControlCabinetState();
  let historyCabinet = $state<ControlCabinet | null>(null);
  let historyOpen = $state(false);

  const columns = $derived.by(() => [
    {
      key: 'cabinet_nr',
      label: listState.isProjectContext
        ? $t('projects.control_cabinets.table.control_cabinet')
        : $t('facility.forms.control_cabinet.number_label')
    },
    {
      key: 'building',
      label: listState.isProjectContext
        ? $t('projects.control_cabinets.table.building')
        : $t('facility.building')
    },
    { key: 'actions', label: '', width: 'w-[100px]' }
  ]);

  const searchPlaceholder = $derived.by(() =>
    listState.isProjectContext
      ? $t('projects.control_cabinets.search_placeholder')
      : $t('facility.search_control_cabinets')
  );

  const emptyMessage = $derived.by(() =>
    listState.isProjectContext
      ? $t('projects.control_cabinets.empty')
      : $t('facility.no_control_cabinets_found')
  );

  const newLabel = $derived.by(() =>
    listState.isProjectContext
      ? $t('projects.control_cabinets.new')
      : $t('facility.new_control_cabinet')
  );

  async function handleView(cabinet: ControlCabinet): Promise<void> {
    await goto(`/facility/control-cabinets/${cabinet.id}`);
  }

  async function handleHistoryRestored(): Promise<void> {
    await listState.reload();
    listState.notifyHistoryRestored();
  }
</script>

<div class="flex flex-col gap-4">
  {#if historyCabinet}
    <HistoryTimelineDialog
      bind:open={historyOpen}
      title={`${$t('history.title')}: ${historyCabinet.control_cabinet_nr ?? historyCabinet.id}`}
      scopeType="control_cabinet"
      scopeId={historyCabinet.id}
      projectId={listState.projectId}
      controlCabinetId={historyCabinet.id}
      onRestored={handleHistoryRestored}
    />
  {/if}

  <div class="flex flex-wrap items-center justify-end gap-2">
    {#if !listState.showForm && listState.canCreateControlCabinet()}
      <Button onclick={() => listState.openCreateForm()}>
        <PlusIcon class="mr-2 size-4" />
        {newLabel}
      </Button>
    {/if}
  </div>

  {#if listState.showForm}
    <ControlCabinetForm
      initialData={listState.editingItem}
      projectId={listState.projectId}
      onSuccess={(cabinet) => void listState.handleFormSuccess(cabinet)}
      onCancel={() => listState.cancelForm()}
    />
  {/if}

  <PaginatedList
    state={listState}
    {columns}
    {searchPlaceholder}
    {emptyMessage}
    onSearch={(text) => void listState.search(text)}
    onPageChange={(page) => void listState.goToPage(page)}
    onReload={() => void listState.reload()}
  >
    {#snippet rowSnippet(cabinet: ControlCabinet)}
      <Table.Cell class="font-medium">
        <Button
          variant="link"
          class="h-auto p-0 font-medium"
          onclick={() => void handleView(cabinet)}
        >
          {cabinet.control_cabinet_nr ?? $t('common.not_available')}
        </Button>
      </Table.Cell>
      <Table.Cell>{listState.getBuildingLabel(cabinet.building_id)}</Table.Cell>
      <Table.Cell class="text-right">
        <DropdownMenu.Root>
          <DropdownMenu.Trigger>
            {#snippet child({ props })}
              <Button variant="ghost" size="icon" {...props}>
                <EllipsisIcon class="size-4" />
              </Button>
            {/snippet}
          </DropdownMenu.Trigger>
          <DropdownMenu.Content align="end" class="w-44">
            <DropdownMenu.Item
              onclick={() =>
                void listState.copyToClipboard(cabinet.control_cabinet_nr ?? cabinet.id)}
            >
              {$t('common.copy')}
            </DropdownMenu.Item>
            {#if listState.canCreateControlCabinet()}
              <DropdownMenu.Item onclick={() => void listState.duplicateControlCabinet(cabinet)}>
                {$t('facility.duplicate')}
              </DropdownMenu.Item>
            {/if}
            <DropdownMenu.Item onclick={() => void handleView(cabinet)}>
              {$t('common.view')}
            </DropdownMenu.Item>
            <DropdownMenu.Item
              onclick={() => {
                historyCabinet = cabinet;
                historyOpen = true;
              }}
            >
              {$t('history.open')}
            </DropdownMenu.Item>
            {#if listState.canUpdateControlCabinet()}
              <DropdownMenu.Item onclick={() => listState.editControlCabinet(cabinet)}>
                {$t('common.edit')}
              </DropdownMenu.Item>
            {/if}
            {#if listState.canDeleteControlCabinet()}
              <DropdownMenu.Separator />
              <DropdownMenu.Item
                variant="destructive"
                onclick={() => void listState.deleteControlCabinet(cabinet)}
              >
                {$t('common.delete')}
              </DropdownMenu.Item>
            {/if}
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      </Table.Cell>
    {/snippet}
  </PaginatedList>
</div>
