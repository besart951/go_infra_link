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
  import { stateTextsStore } from '$lib/stores/list/entityStores.js';
  import type { StateText } from '$lib/domain/facility/index.js';
  import StateTextForm from '$lib/components/facility/forms/StateTextForm.svelte';
  import { createStateTextActions } from '$lib/components/facility/shared/facilityCrudPageActions.svelte.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();

  const actions = createStateTextActions();

  onMount(() => {
    stateTextsStore.load();
  });
</script>

<svelte:head>
  <title>{$t('facility.state_texts')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">{$t('facility.state_texts_title')}</h1>
      <p class="text-sm text-muted-foreground">{$t('facility.state_texts_desc')}</p>
    </div>
    {#if !actions.showForm && canPerform('create', 'statetext')}
      <Button onclick={() => actions.create()}>
        <Plus class="mr-2 size-4" />
        {$t('facility.new_state_text')}
      </Button>
    {/if}
  </div>

  {#if actions.showForm}
    <StateTextForm
      initialData={actions.editingItem}
      onSuccess={() => actions.success()}
      onCancel={() => actions.cancel()}
    />
  {/if}

  <PaginatedList
    state={$stateTextsStore}
    columns={[
      { key: 'ref_number', label: $t('facility.ref_number') },
      { key: 'state_text1', label: $t('facility.state_text1') },
      { key: 'actions', label: '', width: 'w-[100px]' }
    ]}
    searchPlaceholder={$t('facility.search_state_texts')}
    emptyMessage={$t('facility.no_state_texts_found')}
    onSearch={(text) => stateTextsStore.search(text)}
    onPageChange={(page) => stateTextsStore.goToPage(page)}
    onReload={() => stateTextsStore.reload()}
  >
    {#snippet rowSnippet(item: StateText)}
      <Table.Cell class="font-medium">{item.ref_number}</Table.Cell>
      <Table.Cell>{item.state_text1 ?? $t('common.not_available')}</Table.Cell>
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
            <DropdownMenu.Item onclick={() => actions.copy(String(item.ref_number ?? item.id))}>
              {$t('facility.copy')}
            </DropdownMenu.Item>
            <DropdownMenu.Item onclick={() => goto(`/facility/state-texts/${item.id}`)}>
              {$t('facility.view')}
            </DropdownMenu.Item>
            {#if canPerform('update', 'statetext')}
              <DropdownMenu.Item onclick={() => actions.edit(item)}
                >{$t('common.edit')}</DropdownMenu.Item
              >
            {/if}
            {#if canPerform('delete', 'statetext')}
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
