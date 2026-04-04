<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import { X } from '@lucide/svelte';
  import BuildingSelect from '../selects/BuildingSelect.svelte';
  import ControlCabinetSelect from '../selects/ControlCabinetSelect.svelte';
  import SPSControllerSelect from '../selects/SPSControllerSelect.svelte';
  import SPSControllerSystemTypeSelect from '../selects/SPSControllerSystemTypeSelect.svelte';
  import ProjectSelect from '$lib/components/project/ProjectSelect.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import type { FieldDeviceFilters } from './state/types.js';
  import { useFieldDeviceState } from './state/context.svelte.js';

  interface Props {
    showProjectFilter?: boolean;
  }

  let { showProjectFilter = false }: Props = $props();

  const t = createTranslator();
  const fieldDeviceState = useFieldDeviceState();

  let buildingId = $state('');
  let controlCabinetId = $state('');
  let spsControllerId = $state('');
  let spsControllerSystemTypeId = $state('');
  let projectId = $state('');

  const hasActiveFilters = $derived(
    buildingId ||
      controlCabinetId ||
      spsControllerId ||
      spsControllerSystemTypeId ||
      (showProjectFilter && projectId)
  );

  const gridClass = $derived(
    showProjectFilter
      ? 'grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-5'
      : 'grid grid-cols-1 gap-4 md:grid-cols-2 lg:grid-cols-4'
  );

  function applyFilters() {
    const filters: FieldDeviceFilters = {
      buildingId: buildingId || undefined,
      controlCabinetId: controlCabinetId || undefined,
      spsControllerId: spsControllerId || undefined,
      spsControllerSystemTypeId: spsControllerSystemTypeId || undefined,
      projectId: showProjectFilter && projectId ? projectId : undefined
    };

    void fieldDeviceState.applyFilters(filters);
  }

  function clearFilters() {
    buildingId = '';
    controlCabinetId = '';
    spsControllerId = '';
    spsControllerSystemTypeId = '';
    projectId = '';
    void fieldDeviceState.clearFilters();
  }
</script>

<Card.Root>
  <Card.Content>
    <div class={gridClass}>
      <div class="flex flex-col gap-2">
        <label for="building-filter" class="text-sm font-medium">
          {$t('field_device.filters.building')}
        </label>
        <BuildingSelect bind:value={buildingId} width="w-full" />
      </div>
      <div class="flex flex-col gap-2">
        <label for="control-cabinet-filter" class="text-sm font-medium">
          {$t('field_device.filters.control_cabinet')}
        </label>
        <ControlCabinetSelect bind:value={controlCabinetId} width="w-full" />
      </div>
      <div class="flex flex-col gap-2">
        <label for="sps-controller-filter" class="text-sm font-medium">
          {$t('field_device.filters.sps_controller')}
        </label>
        <SPSControllerSelect bind:value={spsControllerId} width="w-full" />
      </div>
      <div class="flex flex-col gap-2">
        <label for="sps-controller-system-type-filter" class="text-sm font-medium">
          {$t('field_device.filters.sps_system_type')}
        </label>
        <SPSControllerSystemTypeSelect bind:value={spsControllerSystemTypeId} width="w-full" />
      </div>
      {#if showProjectFilter}
        <div class="flex flex-col gap-2">
          <label for="project-filter" class="text-sm font-medium">
            {$t('field_device.filters.project')}
          </label>
          <ProjectSelect bind:value={projectId} width="w-full" />
        </div>
      {/if}
    </div>
    <div class="mt-4 flex justify-end gap-2">
      <Button onclick={applyFilters}>{$t('field_device.filters.apply')}</Button>
      {#if hasActiveFilters}
        <Button variant="outline" onclick={clearFilters}>
          <X class="mr-2 size-4" />
          {$t('field_device.filters.clear')}
        </Button>
      {/if}
    </div>
  </Card.Content>
</Card.Root>
