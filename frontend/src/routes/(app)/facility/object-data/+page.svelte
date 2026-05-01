<script lang="ts">
  import { goto } from '$app/navigation';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
  import { objectDataStore } from '$lib/stores/list/entityStores.js';
  import type { ObjectData } from '$lib/domain/facility/index.js';
  import { createObjectDataActions } from '$lib/components/facility/shared/facilityCrudPageActions.svelte.js';
  import FacilityCrudListPage from '$lib/components/facility/shared/FacilityCrudListPage.svelte';
  import { canPerform } from '$lib/utils/permissions.js';
  import ObjectDataForm from '$lib/components/facility/forms/ObjectDataForm.svelte';
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();
  const actions = createObjectDataActions();
</script>

<FacilityCrudListPage
  title={$t('facility.object_data_title')}
  description={$t('facility.object_data_desc')}
  createLabel={$t('facility.new_object_data')}
  permissionResource="objectdata"
  store={objectDataStore}
  {actions}
  form={ObjectDataForm}
  columns={[
    { key: 'description', label: $t('common.description') },
    { key: 'version', label: $t('facility.version') },
    { key: 'is_active', label: $t('common.status') },
    { key: 'actions', label: '', width: 'w-[100px]' }
  ]}
  searchPlaceholder={$t('facility.search_object_data')}
  emptyMessage={$t('facility.no_object_data_found')}
  documentTitle={$t('facility.object_data')}
>
  {#snippet rowSnippet(item: ObjectData)}
    <Table.Cell class="font-medium">{item.description}</Table.Cell>
    <Table.Cell>
      <code class="rounded-md bg-muted px-1.5 py-0.5 text-sm">{item.version}</code>
    </Table.Cell>
    <Table.Cell>
      <span
        class="inline-flex items-center rounded-full px-2 py-1 text-xs font-medium {item.is_active
          ? 'bg-success-muted text-success-muted-foreground'
          : 'bg-muted/50 text-muted-foreground'}"
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
            <DropdownMenu.Item onclick={() => actions.editFresh(item)}
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
</FacilityCrudListPage>
