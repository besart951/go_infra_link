<script lang="ts">
  import { goto } from '$app/navigation';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import PaginatedList from '$lib/components/list/PaginatedList.svelte';
  import ControlCabinetForm from '$lib/components/facility/forms/ControlCabinetForm.svelte';
  import type { ControlCabinet } from '$lib/domain/facility/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
  import PlusIcon from '@lucide/svelte/icons/plus';

  type ListState = {
    items: ControlCabinet[];
    total: number;
    page: number;
    pageSize: number;
    totalPages: number;
    searchText: string;
    loading: boolean;
    error: string | null;
  };

  type Props = {
    state: ListState;
    showForm: boolean;
    editingItem?: ControlCabinet;
    projectId?: string;
    searchPlaceholder: string;
    emptyMessage: string;
    cabinetColumnLabel: string;
    buildingColumnLabel: string;
    newLabel: string;
    canCreate: boolean;
    canDuplicate: boolean;
    canUpdate: boolean;
    canDelete: boolean;
    canView?: boolean;
    getBuildingLabel: (buildingId: string) => string;
    onCreate: () => void;
    onSearch: (text: string) => void;
    onPageChange: (page: number) => void;
    onReload: () => void;
    onFormSuccess: (cabinet: ControlCabinet) => void;
    onFormCancel: () => void;
    onEdit: (cabinet: ControlCabinet) => void;
    onDelete: (cabinet: ControlCabinet) => void | Promise<void>;
    onDuplicate: (cabinet: ControlCabinet) => void | Promise<void>;
    onCopy: (value: string) => void | Promise<void>;
    onView?: (cabinet: ControlCabinet) => void | Promise<void>;
  };

  let {
    state,
    showForm,
    editingItem = undefined,
    projectId = undefined,
    searchPlaceholder,
    emptyMessage,
    cabinetColumnLabel,
    buildingColumnLabel,
    newLabel,
    canCreate,
    canDuplicate,
    canUpdate,
    canDelete,
    canView = true,
    getBuildingLabel,
    onCreate,
    onSearch,
    onPageChange,
    onReload,
    onFormSuccess,
    onFormCancel,
    onEdit,
    onDelete,
    onDuplicate,
    onCopy,
    onView
  }: Props = $props();

  const t = createTranslator();

  async function handleView(cabinet: ControlCabinet) {
    if (onView) {
      await onView(cabinet);
      return;
    }
    await goto(`/facility/control-cabinets/${cabinet.id}`);
  }
</script>

<div class="flex flex-col gap-4">
  <div class="flex flex-wrap items-center justify-end gap-2">
    <Button variant="outline" onclick={onReload} disabled={state.loading}>{$t('common.refresh')}</Button>
    {#if !showForm && canCreate}
      <Button onclick={onCreate}>
        <PlusIcon class="mr-2 size-4" />
        {newLabel}
      </Button>
    {/if}
  </div>

  {#if showForm}
    <ControlCabinetForm
      initialData={editingItem}
      projectId={projectId}
      onSuccess={onFormSuccess}
      onCancel={onFormCancel}
    />
  {/if}

  <PaginatedList
    {state}
    columns={[
      { key: 'cabinet_nr', label: cabinetColumnLabel },
      { key: 'building', label: buildingColumnLabel },
      { key: 'actions', label: '', width: 'w-[100px]' }
    ]}
    {searchPlaceholder}
    {emptyMessage}
    {onSearch}
    {onPageChange}
    {onReload}
  >
    {#snippet rowSnippet(cabinet: ControlCabinet)}
      <Table.Cell class="font-medium">
        <button class="hover:underline" type="button" onclick={() => handleView(cabinet)}>
          {cabinet.control_cabinet_nr ?? $t('common.not_available')}
        </button>
      </Table.Cell>
      <Table.Cell>{getBuildingLabel(cabinet.building_id)}</Table.Cell>
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
            <DropdownMenu.Item onclick={() => onCopy(cabinet.control_cabinet_nr ?? cabinet.id)}>
              {$t('common.copy')}
            </DropdownMenu.Item>
            {#if canDuplicate}
              <DropdownMenu.Item onclick={() => onDuplicate(cabinet)}>
                {$t('facility.duplicate')}
              </DropdownMenu.Item>
            {/if}
            {#if canView}
              <DropdownMenu.Item onclick={() => handleView(cabinet)}>
                {$t('common.view')}
              </DropdownMenu.Item>
            {/if}
            {#if canUpdate}
              <DropdownMenu.Item onclick={() => onEdit(cabinet)}>
                {$t('common.edit')}
              </DropdownMenu.Item>
            {/if}
            {#if canDelete}
              <DropdownMenu.Separator />
              <DropdownMenu.Item variant="destructive" onclick={() => onDelete(cabinet)}>
                {$t('common.delete')}
              </DropdownMenu.Item>
            {/if}
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      </Table.Cell>
    {/snippet}
  </PaginatedList>
</div>

