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

  let buildingId = $state(fieldDeviceState.filters.buildingId ?? '');
  let controlCabinetId = $state(fieldDeviceState.filters.controlCabinetId ?? '');
  let spsControllerId = $state(fieldDeviceState.filters.spsControllerId ?? '');
  let spsControllerSystemTypeId = $state(fieldDeviceState.filters.spsControllerSystemTypeId ?? '');
  let projectId = $state(fieldDeviceState.filters.projectId ?? '');

  const showBuildingFilter = $derived(!fieldDeviceState.isFilterFixed('buildingId'));
  const showControlCabinetFilter = $derived(!fieldDeviceState.isFilterFixed('controlCabinetId'));
  const showSpsControllerFilter = $derived(!fieldDeviceState.isFilterFixed('spsControllerId'));
  const showSpsSystemTypeFilter = $derived(
    !fieldDeviceState.isFilterFixed('spsControllerSystemTypeId')
  );
  const showProjectSelect = $derived(
    showProjectFilter && !fieldDeviceState.isFilterFixed('projectId')
  );

  const effectiveProjectId = $derived(fieldDeviceState.fixedFilterValue('projectId') ?? projectId);
  const effectiveBuildingId = $derived(
    fieldDeviceState.fixedFilterValue('buildingId') ?? buildingId
  );
  const effectiveControlCabinetId = $derived(
    fieldDeviceState.fixedFilterValue('controlCabinetId') ?? controlCabinetId
  );
  const effectiveSpsControllerId = $derived(
    fieldDeviceState.fixedFilterValue('spsControllerId') ?? spsControllerId
  );
  const scopedProjectId = $derived(effectiveProjectId || undefined);
  const scopedBuildingId = $derived(effectiveBuildingId || undefined);
  const scopedControlCabinetId = $derived(effectiveControlCabinetId || undefined);
  const scopedSpsControllerId = $derived(effectiveSpsControllerId || undefined);

  const spsControllerDisabled = $derived(
    Boolean(effectiveBuildingId && !effectiveControlCabinetId)
  );
  const spsSystemTypeDisabled = $derived(
    Boolean((effectiveBuildingId || effectiveControlCabinetId) && !effectiveSpsControllerId)
  );

  const hasActiveFilters = $derived(
    (showBuildingFilter && buildingId) ||
      (showControlCabinetFilter && controlCabinetId) ||
      (showSpsControllerFilter && spsControllerId) ||
      (showSpsSystemTypeFilter && spsControllerSystemTypeId) ||
      (showProjectSelect && projectId)
  );

  function handleProjectChange(value: string) {
    projectId = value;
    buildingId = '';
    controlCabinetId = '';
    spsControllerId = '';
    spsControllerSystemTypeId = '';
  }

  function handleBuildingChange(value: string) {
    buildingId = value;
    controlCabinetId = '';
    spsControllerId = '';
    spsControllerSystemTypeId = '';
  }

  function handleControlCabinetChange(value: string) {
    controlCabinetId = value;
    spsControllerId = '';
    spsControllerSystemTypeId = '';
  }

  function handleSpsControllerChange(value: string) {
    spsControllerId = value;
    spsControllerSystemTypeId = '';
  }

  function applyFilters() {
    const filters: FieldDeviceFilters = {
      buildingId: showBuildingFilter && buildingId ? buildingId : undefined,
      controlCabinetId: showControlCabinetFilter && controlCabinetId ? controlCabinetId : undefined,
      spsControllerId: showSpsControllerFilter && spsControllerId ? spsControllerId : undefined,
      spsControllerSystemTypeId:
        showSpsSystemTypeFilter && spsControllerSystemTypeId
          ? spsControllerSystemTypeId
          : undefined,
      projectId: showProjectSelect && projectId ? projectId : undefined
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
    <div class="grid grid-cols-1 gap-4 md:grid-cols-2 xl:grid-cols-4">
      {#if showProjectSelect}
        <div class="flex flex-col gap-2">
          <label for="project-filter" class="text-sm font-medium">
            {$t('field_device.filters.project')}
          </label>
          <ProjectSelect
            bind:value={projectId}
            width="w-full"
            onValueChange={handleProjectChange}
          />
        </div>
      {/if}
      {#if showBuildingFilter}
        <div class="flex flex-col gap-2">
          <label for="building-filter" class="text-sm font-medium">
            {$t('field_device.filters.building')}
          </label>
          <BuildingSelect
            bind:value={buildingId}
            width="w-full"
            projectId={scopedProjectId}
            onValueChange={handleBuildingChange}
          />
        </div>
      {/if}
      {#if showControlCabinetFilter}
        <div class="flex flex-col gap-2">
          <label for="control-cabinet-filter" class="text-sm font-medium">
            {$t('field_device.filters.control_cabinet')}
          </label>
          <ControlCabinetSelect
            bind:value={controlCabinetId}
            width="w-full"
            projectId={scopedProjectId}
            buildingId={scopedBuildingId}
            onValueChange={handleControlCabinetChange}
          />
        </div>
      {/if}
      {#if showSpsControllerFilter}
        <div class="flex flex-col gap-2">
          <label for="sps-controller-filter" class="text-sm font-medium">
            {$t('field_device.filters.sps_controller')}
          </label>
          <SPSControllerSelect
            bind:value={spsControllerId}
            width="w-full"
            projectId={scopedProjectId}
            controlCabinetId={scopedControlCabinetId}
            disabled={spsControllerDisabled}
            onValueChange={handleSpsControllerChange}
          />
        </div>
      {/if}
      {#if showSpsSystemTypeFilter}
        <div class="flex flex-col gap-2">
          <label for="sps-controller-system-type-filter" class="text-sm font-medium">
            {$t('field_device.filters.sps_system_type')}
          </label>
          <SPSControllerSystemTypeSelect
            bind:value={spsControllerSystemTypeId}
            width="w-full"
            projectId={scopedProjectId}
            spsControllerId={scopedSpsControllerId}
            disabled={spsSystemTypeDisabled}
          />
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
