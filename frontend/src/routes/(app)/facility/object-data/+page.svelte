<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
  import { Plus } from '@lucide/svelte';
  import PaginatedList from '$lib/components/list/PaginatedList.svelte';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import { addToast } from '$lib/components/toast.svelte';
  import { confirm } from '$lib/stores/confirm-dialog.js';
  import { objectDataStore } from '$lib/stores/list/entityStores.js';
  import type { ObjectData } from '$lib/domain/facility/index.js';
  import { ManageObjectDataUseCase } from '$lib/application/useCases/facility/manageObjectDataUseCase.js';
  import { objectDataRepository } from '$lib/infrastructure/api/objectDataRepository.js';
  import { CrudPageActions } from '$lib/components/facility/shared/crudPageActions.svelte.js';
  import { canPerform } from '$lib/utils/permissions.js';
  const manageObjectData = new ManageObjectDataUseCase(objectDataRepository);
  import ObjectDataForm from '$lib/components/facility/forms/ObjectDataForm.svelte';
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();

  const actions = new CrudPageActions<ObjectData>({
    reload: () => objectDataStore.reload(),
    deleteItem: (item) => manageObjectData.delete(item.id),
    confirmDelete: confirm,
    addToast,
    getDeleteTitle: () => $t('facility.delete_object_data_confirm').replace('{desc}', ''),
    getDeleteMessage: (item) =>
      $t('facility.delete_object_data_confirm').replace('{desc}', item.description || ''),
    getDeleteConfirmText: () => $t('common.delete'),
    getDeleteCancelText: () => $t('common.cancel'),
    getDeleteSuccessMessage: () => $t('facility.object_data_deleted'),
    getDeleteFailureMessage: () => $t('facility.delete_object_data_failed')
  });

  async function handleEdit(item: ObjectData) {
    try {
      actions.edit(await manageObjectData.get(item.id));
    } catch (error) {
      console.error(error);
      actions.edit(item);
    }
  }

  onMount(() => {
    objectDataStore.load();
  });
</script>

<svelte:head>
  <title>{$t('facility.object_data')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">{$t('facility.object_data_title')}</h1>
      <p class="text-sm text-muted-foreground">
        {$t('facility.object_data_desc')}
      </p>
    </div>
    {#if !actions.showForm && canPerform('create', 'objectdata')}
      <Button onclick={() => actions.create()}>
        <Plus class="mr-2 size-4" />
        {$t('facility.new_object_data')}
      </Button>
    {/if}
  </div>

  {#if actions.showForm}
    <ObjectDataForm
      initialData={actions.editingItem}
      onSuccess={() => actions.success()}
      onCancel={() => actions.cancel()}
    />
  {/if}

  <PaginatedList
    state={$objectDataStore}
    columns={[
      { key: 'description', label: $t('common.description') },
      { key: 'version', label: $t('facility.version') },
      { key: 'is_active', label: $t('common.status') },
      { key: 'actions', label: '', width: 'w-[100px]' }
    ]}
    searchPlaceholder={$t('facility.search_object_data')}
    emptyMessage={$t('facility.no_object_data_found')}
    onSearch={(text) => objectDataStore.search(text)}
    onPageChange={(page) => objectDataStore.goToPage(page)}
    onReload={() => objectDataStore.reload()}
  >
    {#snippet rowSnippet(item: ObjectData)}
      <Table.Cell class="font-medium">{item.description}</Table.Cell>
      <Table.Cell>
        <code class="rounded bg-muted px-1.5 py-0.5 text-sm">{item.version}</code>
      </Table.Cell>
      <Table.Cell>
        <span
          class="inline-flex items-center rounded-full px-2 py-1 text-xs font-medium {item.is_active
            ? 'bg-green-50 text-green-700'
            : 'bg-gray-50 text-gray-700'}"
        >
          {item.is_active ? $t('common.active') : $t('common.inactive')}
        </span>
      </Table.Cell>
      <Table.Cell class="text-right">
        <DropdownMenu.Root>
          <DropdownMenu.Trigger>
            {#snippet child({ props })}
              <Button variant="ghost" size="icon" {...props}>
                <EllipsisIcon class="size-4" />
              </Button>
            {/snippet}
          </DropdownMenu.Trigger>
          <DropdownMenu.Content align="end" class="w-40">
            <DropdownMenu.Item onclick={() => actions.copy(item.description ?? item.id)}>
              {$t('facility.copy')}
            </DropdownMenu.Item>
            <DropdownMenu.Item onclick={() => goto(`/facility/object-data/${item.id}`)}>
              {$t('facility.view')}
            </DropdownMenu.Item>
            {#if canPerform('update', 'objectdata')}
              <DropdownMenu.Item onclick={() => handleEdit(item)}
                >{$t('common.edit')}</DropdownMenu.Item
              >
            {/if}
            {#if canPerform('delete', 'objectdata')}
              <DropdownMenu.Separator />
              <DropdownMenu.Item variant="destructive" onclick={() => actions.delete(item)}>
                {$t('common.delete')}
              </DropdownMenu.Item>
            {/if}
          </DropdownMenu.Content>
        </DropdownMenu.Root>
      </Table.Cell>
    {/snippet}
  </PaginatedList>
</div>
