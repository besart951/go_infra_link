<script lang="ts">
  import { goto } from '$app/navigation';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
  import { apparatsStore } from '$lib/stores/list/entityStores.js';
  import type { Apparat } from '$lib/domain/facility/index.js';
  import ApparatForm from '$lib/components/facility/forms/ApparatForm.svelte';
  import { createApparatActions } from '$lib/components/facility/shared/facilityCrudPageActions.svelte.js';
  import FacilityCrudListPage from '$lib/components/facility/shared/FacilityCrudListPage.svelte';
  import HistoryTimelineDialog from '$lib/components/history/HistoryTimelineDialog.svelte';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();
  const actions = createApparatActions();
  let historyItem = $state<Apparat | null>(null);
  let historyOpen = $state(false);
</script>

{#if historyItem}
  <HistoryTimelineDialog
    bind:open={historyOpen}
    title={`${$t('history.title')}: ${historyItem.short_name ?? historyItem.name}`}
    entityTable="apparats"
    entityId={historyItem.id}
    onRestored={() => apparatsStore.reload()}
  />
{/if}

<FacilityCrudListPage
  title={$t('facility.apparats_title')}
  description={$t('facility.apparats_desc')}
  createLabel={$t('facility.new_apparat')}
  permissionResource="apparat"
  store={apparatsStore}
  {actions}
  form={ApparatForm}
  columns={[
    { key: 'short_name', label: $t('facility.short_name') },
    { key: 'name', label: $t('common.name') },
    { key: 'description', label: $t('common.description') },
    { key: 'actions', label: '', width: 'w-[100px]' }
  ]}
  searchPlaceholder={$t('facility.search_apparats')}
  emptyMessage={$t('facility.no_apparats_found')}
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
          <DropdownMenu.Item
            onclick={() => {
              historyItem = item;
              historyOpen = true;
            }}
          >
            {$t('history.open')}
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
</FacilityCrudListPage>
