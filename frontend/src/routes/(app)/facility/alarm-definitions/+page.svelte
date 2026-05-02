<script lang="ts">
  import { goto } from '$app/navigation';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
  import { alarmDefinitionsStore } from '$lib/stores/list/entityStores.js';
  import type { AlarmDefinition } from '$lib/domain/facility/index.js';
  import AlarmDefinitionForm from '$lib/components/facility/forms/AlarmDefinitionForm.svelte';
  import { createAlarmDefinitionActions } from '$lib/components/facility/shared/facilityCrudPageActions.svelte.js';
  import FacilityCrudListPage from '$lib/components/facility/shared/FacilityCrudListPage.svelte';
  import HistoryTimelineDialog from '$lib/components/history/HistoryTimelineDialog.svelte';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();
  const actions = createAlarmDefinitionActions();
  let historyItem = $state<AlarmDefinition | null>(null);
  let historyOpen = $state(false);
</script>

{#if historyItem}
  <HistoryTimelineDialog
    bind:open={historyOpen}
    title={`${$t('history.title')}: ${historyItem.name}`}
    entityTable="alarm_definitions"
    entityId={historyItem.id}
    onRestored={() => alarmDefinitionsStore.reload()}
  />
{/if}

<FacilityCrudListPage
  title={$t('facility.alarm_definitions_title')}
  description={$t('facility.alarm_definitions_desc')}
  createLabel={$t('facility.new_alarm_definition')}
  permissionResource="alarmdefinition"
  store={alarmDefinitionsStore}
  {actions}
  form={AlarmDefinitionForm}
  columns={[
    { key: 'name', label: $t('common.name') },
    { key: 'alarm_note', label: $t('facility.alarm_note') },
    { key: 'actions', label: '', width: 'w-[100px]' }
  ]}
  searchPlaceholder={$t('facility.search_alarm_definitions')}
  emptyMessage={$t('facility.no_alarm_definitions_found')}
  documentTitle={$t('facility.alarm_definitions')}
>
  {#snippet rowSnippet(item: AlarmDefinition)}
    <Table.Cell class="font-medium">{item.name}</Table.Cell>
    <Table.Cell>{item.alarm_note ?? $t('common.not_available')}</Table.Cell>
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
          <DropdownMenu.Item onclick={() => goto(`/facility/alarm-definitions/${item.id}`)}>
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
          {#if canPerform('update', 'alarmdefinition')}
            <DropdownMenu.Item onclick={() => actions.edit(item)}
              >{$t('common.edit')}</DropdownMenu.Item
            >
          {/if}
          {#if canPerform('delete', 'alarmdefinition')}
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
