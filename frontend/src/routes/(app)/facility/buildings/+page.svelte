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
  import { addToast } from '$lib/components/toast.svelte';
  import { confirm } from '$lib/stores/confirm-dialog.js';
  import { buildingsStore } from '$lib/stores/list/entityStores.js';
  import type { Building } from '$lib/domain/facility/index.js';
  import BuildingForm from '$lib/components/facility/forms/BuildingForm.svelte';
  import { ManageBuildingUseCase } from '$lib/application/useCases/facility/manageBuildingUseCase.js';
  import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
  import { CrudPageActions } from '$lib/components/facility/shared/crudPageActions.svelte.js';
  import { canPerform } from '$lib/utils/permissions.js';
  const manageBuilding = new ManageBuildingUseCase(buildingRepository);
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();

  const actions = new CrudPageActions<Building>({
    reload: () => buildingsStore.reload(),
    deleteItem: (building) => manageBuilding.delete(building.id),
    confirmDelete: confirm,
    addToast,
    getDeleteTitle: () => $t('common.delete'),
    getDeleteMessage: (building) =>
      $t('facility.delete_building_confirm').replace('{name}', building.iws_code),
    getDeleteConfirmText: () => $t('common.delete'),
    getDeleteCancelText: () => $t('common.cancel'),
    getDeleteSuccessMessage: () => $t('facility.building_deleted'),
    getDeleteFailureMessage: () => $t('facility.delete_building_failed')
  });

  onMount(() => {
    buildingsStore.load();
  });
</script>

<svelte:head>
  <title>{$t('facility.buildings_title')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">{$t('facility.buildings_title')}</h1>
      <p class="text-sm text-muted-foreground">{$t('facility.buildings_desc')}</p>
    </div>
    {#if !actions.showForm && canPerform('create', 'building')}
      <Button onclick={() => actions.create()}>
        <Plus class="mr-2 size-4" />
        {$t('facility.new_building')}
      </Button>
    {/if}
  </div>

  {#if actions.showForm}
    <BuildingForm
      initialData={actions.editingItem}
      onSuccess={() => actions.success()}
      onCancel={() => actions.cancel()}
    />
  {/if}

  <PaginatedList
    state={$buildingsStore}
    columns={[
      { key: 'iws_code', label: $t('facility.iws_code') },
      { key: 'building_group', label: $t('facility.building_group') },
      { key: 'actions', label: '', width: 'w-[100px]' }
    ]}
    searchPlaceholder={$t('facility.search_buildings')}
    emptyMessage={$t('facility.no_buildings_found')}
    onSearch={(text) => buildingsStore.search(text)}
    onPageChange={(page) => buildingsStore.goToPage(page)}
    onReload={() => buildingsStore.reload()}
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
  </PaginatedList>
</div>
