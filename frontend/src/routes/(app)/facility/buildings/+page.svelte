<script lang="ts">
  import { goto } from '$app/navigation';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import EllipsisIcon from '@lucide/svelte/icons/ellipsis';
  import { buildingsStore } from '$lib/stores/list/entityStores.js';
  import type { Building } from '$lib/domain/facility/index.js';
  import BuildingForm from '$lib/components/facility/forms/BuildingForm.svelte';
  import { createBuildingActions } from '$lib/components/facility/shared/facilityCrudPageActions.svelte.js';
  import FacilityCrudListPage from '$lib/components/facility/shared/FacilityCrudListPage.svelte';
  import HistoryTimelineDialog from '$lib/components/history/HistoryTimelineDialog.svelte';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();

  const actions = createBuildingActions();
  let historyItem = $state<Building | null>(null);
  let historyOpen = $state(false);
</script>

{#if historyItem}
  <HistoryTimelineDialog
    bind:open={historyOpen}
    title={`${$t('history.title')}: ${historyItem.iws_code}`}
    entityTable="buildings"
    entityId={historyItem.id}
    onRestored={() => buildingsStore.reload()}
  />
{/if}

<FacilityCrudListPage
  title={$t('facility.buildings_title')}
  description={$t('facility.buildings_desc')}
  createLabel={$t('facility.new_building')}
  permissionResource="building"
  store={buildingsStore}
  {actions}
  form={BuildingForm}
  columns={[
    { key: 'iws_code', label: $t('facility.iws_code') },
    { key: 'building_group', label: $t('facility.building_group') },
    { key: 'actions', label: '', width: 'w-[100px]' }
  ]}
  searchPlaceholder={$t('facility.search_buildings')}
  emptyMessage={$t('facility.no_buildings_found')}
>
  {#snippet rowSnippet(building: Building)}
    <Table.Cell class="font-medium">
      <a href="/facility/buildings/{building.id}" class="hover:underline">
        {building.iws_code}
      </a>
    </Table.Cell>
    <Table.Cell>{building.building_group}</Table.Cell>
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
          <DropdownMenu.Item onclick={() => actions.copy(building.iws_code)}>
            {$t('facility.copy')}
          </DropdownMenu.Item>
          <DropdownMenu.Item onclick={() => goto(`/facility/buildings/${building.id}`)}>
            {$t('facility.view')}
          </DropdownMenu.Item>
          <DropdownMenu.Item
            onclick={() => {
              historyItem = building;
              historyOpen = true;
            }}
          >
            {$t('history.open')}
          </DropdownMenu.Item>
          {#if canPerform('update', 'building')}
            <DropdownMenu.Item onclick={() => actions.edit(building)}
              >{$t('common.edit')}</DropdownMenu.Item
            >
          {/if}
          {#if canPerform('delete', 'building')}
            <DropdownMenu.Separator />
            <DropdownMenu.Item variant="destructive" onclick={() => actions.delete(building)}>
              {$t('common.delete')}
            </DropdownMenu.Item>
          {/if}
        </DropdownMenu.Content>
      </DropdownMenu.Root>
    </Table.Cell>
  {/snippet}
</FacilityCrudListPage>
