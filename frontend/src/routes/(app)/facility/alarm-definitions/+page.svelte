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
  import { alarmDefinitionsStore } from '$lib/stores/list/entityStores.js';
  import type { AlarmDefinition } from '$lib/domain/facility/index.js';
  import AlarmDefinitionForm from '$lib/components/facility/forms/AlarmDefinitionForm.svelte';
  import { createAlarmDefinitionActions } from '$lib/components/facility/shared/facilityCrudPageActions.svelte.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();

  const actions = createAlarmDefinitionActions();

  onMount(() => {
    alarmDefinitionsStore.load();
  });
</script>

<svelte:head>
  <title>{$t('facility.alarm_definitions')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">
        {$t('facility.alarm_definitions_title')}
      </h1>
      <p class="text-sm text-muted-foreground">{$t('facility.alarm_definitions_desc')}</p>
    </div>
    {#if !actions.showForm && canPerform('create', 'alarmdefinition')}
      <Button onclick={() => actions.create()}>
        <Plus class="mr-2 size-4" />
        {$t('facility.new_alarm_definition')}
      </Button>
    {/if}
  </div>

  {#if actions.showForm}
    <AlarmDefinitionForm
      initialData={actions.editingItem}
      onSuccess={() => actions.success()}
      onCancel={() => actions.cancel()}
    />
  {/if}

  <PaginatedList
    state={$alarmDefinitionsStore}
    columns={[
      { key: 'name', label: $t('common.name') },
      { key: 'alarm_note', label: $t('facility.alarm_note') },
      { key: 'actions', label: '', width: 'w-[100px]' }
    ]}
    searchPlaceholder={$t('facility.search_alarm_definitions')}
    emptyMessage={$t('facility.no_alarm_definitions_found')}
    onSearch={(text) => alarmDefinitionsStore.search(text)}
    onPageChange={(page) => alarmDefinitionsStore.goToPage(page)}
    onReload={() => alarmDefinitionsStore.reload()}
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
  </PaginatedList>
</div>
