<script lang="ts">
  import { goto } from '$app/navigation';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
  import { notificationClassesStore } from '$lib/stores/list/entityStores.js';
  import type { NotificationClass } from '$lib/domain/facility/index.js';
  import NotificationClassForm from '$lib/components/facility/forms/NotificationClassForm.svelte';
  import { createNotificationClassActions } from '$lib/components/facility/shared/facilityCrudPageActions.svelte.js';
  import FacilityCrudListPage from '$lib/components/facility/shared/FacilityCrudListPage.svelte';
  import HistoryTimelineDialog from '$lib/components/history/HistoryTimelineDialog.svelte';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();
  const actions = createNotificationClassActions();
  let historyItem = $state<NotificationClass | null>(null);
  let historyOpen = $state(false);
</script>

{#if historyItem}
  <HistoryTimelineDialog
    bind:open={historyOpen}
    title={`${$t('history.title')}: ${historyItem.event_category}`}
    entityTable="notification_classes"
    entityId={historyItem.id}
    onRestored={() => notificationClassesStore.reload()}
  />
{/if}

<FacilityCrudListPage
  title={$t('facility.notification_classes_title')}
  description={$t('facility.notification_classes_desc')}
  createLabel={$t('facility.new_notification_class')}
  permissionResource="notificationclass"
  store={notificationClassesStore}
  {actions}
  form={NotificationClassForm}
  columns={[
    { key: 'event_category', label: $t('facility.event_category') },
    { key: 'nc', label: $t('facility.nc') },
    { key: 'object_description', label: $t('facility.object_description') },
    { key: 'meaning', label: $t('facility.meaning') },
    { key: 'actions', label: '', width: 'w-[100px]' }
  ]}
  searchPlaceholder={$t('facility.search_notification_classes')}
  emptyMessage={$t('facility.no_notification_classes_found')}
  documentTitle={$t('facility.notification_classes')}
>
  {#snippet rowSnippet(item: NotificationClass)}
    <Table.Cell class="font-medium">{item.event_category}</Table.Cell>
    <Table.Cell>{item.nc}</Table.Cell>
    <Table.Cell>{item.object_description}</Table.Cell>
    <Table.Cell>{item.meaning}</Table.Cell>
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
          <DropdownMenu.Item onclick={() => actions.copy(item.event_category ?? item.id)}>
            {$t('facility.copy')}
          </DropdownMenu.Item>
          <DropdownMenu.Item onclick={() => goto(`/facility/notification-classes/${item.id}`)}>
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
          {#if canPerform('update', 'notificationclass')}
            <DropdownMenu.Item onclick={() => actions.edit(item)}
              >{$t('common.edit')}</DropdownMenu.Item
            >
          {/if}
          {#if canPerform('delete', 'notificationclass')}
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
