<script lang="ts">
  import { goto } from '$app/navigation';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import PaginatedList from '$lib/components/list/PaginatedList.svelte';
  import ControlCabinetForm from '$lib/components/facility/forms/ControlCabinetForm.svelte';
  import type { ControlCabinet } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
  import PlusIcon from '@lucide/svelte/icons/plus';
  import { useControlCabinetState } from './state/context.svelte.js';

  const t = createTranslator();
  const state = useControlCabinetState();

  const columns = $derived.by(() => [
    {
      key: 'cabinet_nr',
      label: state.isProjectContext
        ? $t('projects.control_cabinets.table.control_cabinet')
        : $t('facility.forms.control_cabinet.number_label')
    },
    {
      key: 'building',
      label: state.isProjectContext
        ? $t('projects.control_cabinets.table.building')
        : $t('facility.building')
    },
    { key: 'actions', label: '', width: 'w-[100px]' }
  ]);

  const searchPlaceholder = $derived.by(() =>
    state.isProjectContext
      ? $t('projects.control_cabinets.search_placeholder')
      : $t('facility.search_control_cabinets')
  );

  const emptyMessage = $derived.by(() =>
    state.isProjectContext
      ? $t('projects.control_cabinets.empty')
      : $t('facility.no_control_cabinets_found')
  );

  const newLabel = $derived.by(() =>
    state.isProjectContext
      ? $t('projects.control_cabinets.new')
      : $t('facility.new_control_cabinet')
  );

  async function handleView(cabinet: ControlCabinet): Promise<void> {
    await goto(`/facility/control-cabinets/${cabinet.id}`);
  }
</script>

<div class="flex flex-col gap-4">
  <div class="flex flex-wrap items-center justify-end gap-2">
    {#if !state.showForm && canPerform('create', 'controlcabinet')}
      <Button onclick={() => state.openCreateForm()}>
        <PlusIcon class="mr-2 size-4" />
        {newLabel}
      </Button>
    {/if}
  </div>

  {#if state.showForm}
    <ControlCabinetForm
      initialData={state.editingItem}
      projectId={state.projectId}
      onSuccess={(cabinet) => void state.handleFormSuccess(cabinet)}
      onCancel={() => state.cancelForm()}
    />
  {/if}

  <PaginatedList
    {state}
    {columns}
    {searchPlaceholder}
    {emptyMessage}
    onSearch={(text) => void state.search(text)}
    onPageChange={(page) => void state.goToPage(page)}
    onReload={() => void state.reload()}
  >
    {#snippet rowSnippet(cabinet: ControlCabinet)}
      <Table.Cell class="font-medium">
        <button class="hover:underline" type="button" onclick={() => void handleView(cabinet)}>
          {cabinet.control_cabinet_nr ?? $t('common.not_available')}
        </button>
      </Table.Cell>
      <Table.Cell>{state.getBuildingLabel(cabinet.building_id)}</Table.Cell>
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
              onclick={() => void state.copyToClipboard(cabinet.control_cabinet_nr ?? cabinet.id)}
            >
              {$t('common.copy')}
            </DropdownMenu.Item>
            {#if canPerform('create', 'controlcabinet')}
              <DropdownMenu.Item onclick={() => void state.duplicateControlCabinet(cabinet)}>
                {$t('facility.duplicate')}
              </DropdownMenu.Item>
            {/if}
            <DropdownMenu.Item onclick={() => void handleView(cabinet)}>
              {$t('common.view')}
            </DropdownMenu.Item>
            {#if canPerform('update', 'controlcabinet')}
              <DropdownMenu.Item onclick={() => state.editControlCabinet(cabinet)}>
                {$t('common.edit')}
              </DropdownMenu.Item>
            {/if}
            {#if canPerform('delete', 'controlcabinet')}
              <DropdownMenu.Separator />
              <DropdownMenu.Item
                variant="destructive"
                onclick={() => void state.deleteControlCabinet(cabinet)}
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
