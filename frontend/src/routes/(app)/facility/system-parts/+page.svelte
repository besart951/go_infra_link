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
  import { systemPartsStore } from '$lib/stores/list/entityStores.js';
  import type { SystemPart } from '$lib/domain/facility/index.js';
  import SystemPartForm from '$lib/components/facility/forms/SystemPartForm.svelte';
  import { createSystemPartActions } from '$lib/components/facility/shared/facilityCrudPageActions.svelte.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();

  const actions = createSystemPartActions();

  onMount(() => {
    systemPartsStore.load();
  });
</script>

<svelte:head>
  <title>{$t('facility.system_parts_title')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">{$t('facility.system_parts_title')}</h1>
      <p class="text-sm text-muted-foreground">{$t('facility.system_parts_desc')}</p>
    </div>
    {#if !actions.showForm && canPerform('create', 'systempart')}
      <Button onclick={() => actions.create()}>
        <Plus class="mr-2 size-4" />
        {$t('facility.new_system_part')}
      </Button>
    {/if}
  </div>

  {#if actions.showForm}
    <SystemPartForm
      initialData={actions.editingItem}
      onSuccess={() => actions.success()}
      onCancel={() => actions.cancel()}
    />
  {/if}

  <PaginatedList
    state={$systemPartsStore}
    columns={[
      { key: 'short_name', label: $t('facility.short_name') },
      { key: 'name', label: $t('common.name') },
      { key: 'description', label: $t('common.description') },
      { key: 'actions', label: '', width: 'w-[100px]' }
    ]}
    searchPlaceholder={$t('facility.search_system_parts')}
    emptyMessage={$t('facility.no_system_parts_found')}
    onSearch={(text) => systemPartsStore.search(text)}
    onPageChange={(page) => systemPartsStore.goToPage(page)}
    onReload={() => systemPartsStore.reload()}
  >
    {#snippet rowSnippet(item: SystemPart)}
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
            <DropdownMenu.Item onclick={() => goto(`/facility/system-parts/${item.id}`)}>
              {$t('facility.view')}
            </DropdownMenu.Item>
            {#if canPerform('update', 'systempart')}
              <DropdownMenu.Item onclick={() => actions.edit(item)}
                >{$t('common.edit')}</DropdownMenu.Item
              >
            {/if}
            {#if canPerform('delete', 'systempart')}
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
