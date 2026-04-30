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
  import { systemTypesStore } from '$lib/stores/list/entityStores.js';
  import type { SystemType } from '$lib/domain/facility/index.js';
  import SystemTypeForm from '$lib/components/facility/forms/SystemTypeForm.svelte';
  import { createSystemTypeActions } from '$lib/components/facility/shared/facilityCrudPageActions.svelte.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();

  const actions = createSystemTypeActions();

  function formatNumber(value: number) {
    return String(value).padStart(4, '0');
  }

  onMount(() => {
    systemTypesStore.load();
  });
</script>

<svelte:head>
  <title>{$t('facility.system_types_title')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">{$t('facility.system_types_title')}</h1>
      <p class="text-sm text-muted-foreground">{$t('facility.system_types_desc')}</p>
    </div>
    {#if !actions.showForm && canPerform('create', 'systemtype')}
      <Button onclick={() => actions.create()}>
        <Plus class="mr-2 size-4" />
        {$t('facility.new_system_type')}
      </Button>
    {/if}
  </div>

  {#if actions.showForm}
    <SystemTypeForm
      initialData={actions.editingItem}
      onSuccess={() => actions.success()}
      onCancel={() => actions.cancel()}
    />
  {/if}

  <PaginatedList
    state={$systemTypesStore}
    columns={[
      { key: 'name', label: $t('common.name') },
      { key: 'number_min', label: $t('facility.forms.system_type.min_label') },
      { key: 'number_max', label: $t('facility.forms.system_type.max_label') },
      { key: 'actions', label: '', width: 'w-[100px]' }
    ]}
    searchPlaceholder={$t('facility.search_system_types')}
    emptyMessage={$t('facility.no_system_types_found')}
    onSearch={(text) => systemTypesStore.search(text)}
    onPageChange={(page) => systemTypesStore.goToPage(page)}
    onReload={() => systemTypesStore.reload()}
  >
    {#snippet rowSnippet(item: SystemType)}
      <Table.Cell class="font-medium">{item.name}</Table.Cell>
      <Table.Cell>{formatNumber(item.number_min)}</Table.Cell>
      <Table.Cell>{formatNumber(item.number_max)}</Table.Cell>
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
            <DropdownMenu.Item onclick={() => actions.copy(item.name ?? item.id)}>
              {$t('facility.copy')}
            </DropdownMenu.Item>
            <DropdownMenu.Item onclick={() => goto(`/facility/system-types/${item.id}`)}>
              {$t('facility.view')}
            </DropdownMenu.Item>
            {#if canPerform('update', 'systemtype')}
              <DropdownMenu.Item onclick={() => actions.edit(item)}
                >{$t('common.edit')}</DropdownMenu.Item
              >
            {/if}
            {#if canPerform('delete', 'systemtype')}
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
