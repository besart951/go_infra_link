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
  import { apparatsStore } from '$lib/stores/list/entityStores.js';
  import type { Apparat } from '$lib/domain/facility/index.js';
  import ApparatForm from '$lib/components/facility/forms/ApparatForm.svelte';
  import { ManageEntityUseCase } from '$lib/application/useCases/manageEntityUseCase.js';
  import { apparatRepository } from '$lib/infrastructure/api/apparatRepository.js';
  import { CrudPageActions } from '$lib/components/facility/shared/crudPageActions.svelte.js';
  import { canPerform } from '$lib/utils/permissions.js';
  const manageApparat = new ManageEntityUseCase(apparatRepository);
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();

  const actions = new CrudPageActions<Apparat>({
    reload: () => apparatsStore.reload(),
    deleteItem: (item) => manageApparat.delete(item.id),
    confirmDelete: confirm,
    addToast,
    getDeleteTitle: () => $t('common.delete'),
    getDeleteMessage: (item) =>
      $t('facility.delete_apparat_confirm').replace('{name}', item.short_name ?? item.name),
    getDeleteConfirmText: () => $t('common.delete'),
    getDeleteCancelText: () => $t('common.cancel'),
    getDeleteSuccessMessage: () => $t('facility.apparat_deleted'),
    getDeleteFailureMessage: () => $t('facility.delete_apparat_failed')
  });

  onMount(() => {
    apparatsStore.load();
  });
</script>

<svelte:head>
  <title>{$t('facility.apparats_title')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">{$t('facility.apparats_title')}</h1>
      <p class="text-sm text-muted-foreground">{$t('facility.apparats_desc')}</p>
    </div>
    {#if !actions.showForm && canPerform('create', 'apparat')}
      <Button onclick={() => actions.create()}>
        <Plus class="mr-2 size-4" />
        {$t('facility.new_apparat')}
      </Button>
    {/if}
  </div>

  {#if actions.showForm}
    <ApparatForm
      initialData={actions.editingItem}
      onSuccess={() => actions.success()}
      onCancel={() => actions.cancel()}
    />
  {/if}

  <PaginatedList
    state={$apparatsStore}
    columns={[
      { key: 'short_name', label: $t('facility.short_name') },
      { key: 'name', label: $t('common.name') },
      { key: 'description', label: $t('common.description') },
      { key: 'actions', label: '', width: 'w-[100px]' }
    ]}
    searchPlaceholder={$t('facility.search_apparats')}
    emptyMessage={$t('facility.no_apparats_found')}
    onSearch={(text) => apparatsStore.search(text)}
    onPageChange={(page) => apparatsStore.goToPage(page)}
    onReload={() => apparatsStore.reload()}
  >
    {#snippet rowSnippet(item: Apparat)}
      <Table.Cell class="font-medium">{item.short_name}</Table.Cell>
      <Table.Cell>{item.name}</Table.Cell>
      <Table.Cell>{item.description ?? $t('common.not_available')}</Table.Cell>
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
            <DropdownMenu.Item onclick={() => actions.copy(item.short_name ?? item.id)}>
              {$t('facility.copy')}
            </DropdownMenu.Item>
            <DropdownMenu.Item onclick={() => goto(`/facility/apparats/${item.id}`)}>
              {$t('facility.view')}
            </DropdownMenu.Item>
            {#if canPerform('update', 'apparat')}
              <DropdownMenu.Item onclick={() => actions.edit(item)}
                >{$t('common.edit')}</DropdownMenu.Item
              >
            {/if}
            {#if canPerform('delete', 'apparat')}
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
