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
  import { notificationClassesStore } from '$lib/stores/list/entityStores.js';
  import type { NotificationClass } from '$lib/domain/facility/index.js';
  import NotificationClassForm from '$lib/components/facility/forms/NotificationClassForm.svelte';
  import { createNotificationClassActions } from '$lib/components/facility/shared/facilityCrudPageActions.svelte.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();

  const actions = createNotificationClassActions();

  onMount(() => {
    notificationClassesStore.load();
  });
</script>

<svelte:head>
  <title>{$t('facility.notification_classes')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">
        {$t('facility.notification_classes_title')}
      </h1>
      <p class="text-sm text-muted-foreground">{$t('facility.notification_classes_desc')}</p>
    </div>
    {#if !actions.showForm && canPerform('create', 'notificationclass')}
      <Button onclick={() => actions.create()}>
        <Plus class="mr-2 size-4" />
        {$t('facility.new_notification_class')}
      </Button>
    {/if}
  </div>

  {#if actions.showForm}
    <NotificationClassForm
      initialData={actions.editingItem}
      onSuccess={() => actions.success()}
      onCancel={() => actions.cancel()}
    />
  {/if}

  <PaginatedList
    state={$notificationClassesStore}
    columns={[
      { key: 'event_category', label: $t('facility.event_category') },
      { key: 'nc', label: $t('facility.nc') },
      { key: 'object_description', label: $t('facility.object_description') },
      { key: 'meaning', label: $t('facility.meaning') },
      { key: 'actions', label: '', width: 'w-[100px]' }
    ]}
    searchPlaceholder={$t('facility.search_notification_classes')}
    emptyMessage={$t('facility.no_notification_classes_found')}
    onSearch={(text) => notificationClassesStore.search(text)}
    onPageChange={(page) => notificationClassesStore.goToPage(page)}
    onReload={() => notificationClassesStore.reload()}
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
  </PaginatedList>
</div>
