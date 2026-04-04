<script lang="ts">
  import { onMount } from 'svelte';
  import { controlCabinetsStore } from '$lib/stores/list/entityStores.js';
  import type { Building, ControlCabinet } from '$lib/domain/facility/index.js';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import ControlCabinetList from '$lib/components/facility/control-cabinets/ControlCabinetList.svelte';
  import { addToast } from '$lib/components/toast.svelte';
  import { confirm } from '$lib/stores/confirm-dialog.js';
  import { ManageControlCabinetUseCase } from '$lib/application/useCases/facility/manageControlCabinetUseCase.js';
  import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
  import { buildingRepository } from '$lib/infrastructure/api/buildingRepository.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();
  const manageControlCabinet = new ManageControlCabinetUseCase(controlCabinetRepository);

  let showForm = $state(false);
  let editingItem: ControlCabinet | undefined = $state(undefined);
  let buildingMap = $state(new Map<string, string>());
  const buildingRequests = new Set<string>();

  function formatBuildingLabel(building: Building): string {
    return `${building.iws_code}-${building.building_group}`;
  }

  function getBuildingLabel(buildingId: string): string {
    return buildingMap.get(buildingId) ?? buildingId;
  }

  function updateBuildingMap(buildings: Building[]) {
    const next = new Map(buildingMap);
    for (const building of buildings) {
      next.set(building.id, formatBuildingLabel(building));
    }
    buildingMap = next;
  }

  async function ensureBuildingLabels(items: ControlCabinet[]) {
    const uniqueIds = new Set(
      items.map((item) => item.building_id).filter((id): id is string => Boolean(id))
    );
    const missingIds = Array.from(uniqueIds).filter(
      (id) => !buildingMap.has(id) && !buildingRequests.has(id)
    );

    if (missingIds.length === 0) return;

    missingIds.forEach((id) => buildingRequests.add(id));

    try {
      const buildings = await buildingRepository.getBulk(missingIds);
      updateBuildingMap(buildings);
    } catch (err) {
      console.error('Failed to load buildings:', err);
    } finally {
      missingIds.forEach((id) => buildingRequests.delete(id));
    }
  }

  function handleEdit(item: ControlCabinet) {
    editingItem = item;
    showForm = true;
  }

  function handleCreate() {
    editingItem = undefined;
    showForm = true;
  }

  function handleSuccess() {
    showForm = false;
    editingItem = undefined;
    controlCabinetsStore.reload();
  }

  function handleCancel() {
    showForm = false;
    editingItem = undefined;
  }

  async function handleDelete(item: ControlCabinet) {
    try {
      const impact = await manageControlCabinet.getDeleteImpact(item.id);

      if (impact.sps_controllers_count > 0) {
        const ok1 = await confirm({
          title: $t('facility.delete_control_cabinet_confirm'),
          message: $t('facility.delete_control_cabinet_message').replace(
            '{count}',
            impact.sps_controllers_count.toString()
          ),
          confirmText: $t('common.confirm'),
          cancelText: $t('common.cancel'),
          variant: 'destructive'
        });
        if (!ok1) return;

        const ok2 = await confirm({
          title: $t('facility.confirm_cascading_delete'),
          message: $t('facility.cascading_delete_message')
            .replace('{systemTypes}', impact.sps_controller_system_types_count.toString())
            .replace('{fieldDevices}', impact.field_devices_count.toString())
            .replace('{bacnetObjects}', impact.bacnet_objects_count.toString()),
          confirmText: $t('facility.delete_everything'),
          cancelText: $t('common.cancel'),
          variant: 'destructive'
        });
        if (!ok2) return;
      }

      await manageControlCabinet.delete(item.id);
      addToast($t('facility.control_cabinet_deleted'), 'success');
      controlCabinetsStore.reload();
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : $t('facility.delete_control_cabinet_failed'),
        'error'
      );
    }
  }

  async function handleDuplicate(item: ControlCabinet) {
    try {
      await manageControlCabinet.copy(item.id);
      addToast($t('facility.control_cabinet_copied'), 'success');
      controlCabinetsStore.reload();
    } catch (err) {
      addToast(err instanceof Error ? err.message : $t('facility.copy_failed'), 'error');
    }
  }

  async function handleCopy(value: string) {
    try {
      await navigator.clipboard.writeText(value);
    } catch (error) {
      console.error('Failed to copy to clipboard:', error);
    }
  }

  onMount(() => {
    controlCabinetsStore.load();
  });

  $effect(() => {
    const items = $controlCabinetsStore.items;
    if (items.length > 0) {
      void ensureBuildingLabels(items);
    }
  });
</script>

<svelte:head>
  <title>{$t('facility.control_cabinets_title')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
  <div>
    <h1 class="text-2xl font-semibold tracking-tight">{$t('facility.control_cabinets_title')}</h1>
    <p class="text-sm text-muted-foreground">{$t('facility.control_cabinets_desc')}</p>
  </div>

  <ControlCabinetList
    state={$controlCabinetsStore}
    {showForm}
    {editingItem}
    searchPlaceholder={$t('facility.search_control_cabinets')}
    emptyMessage={$t('facility.no_control_cabinets_found')}
    cabinetColumnLabel={$t('facility.forms.control_cabinet.number_label')}
    buildingColumnLabel={$t('facility.building')}
    newLabel={$t('facility.new_control_cabinet')}
    canCreate={canPerform('create', 'controlcabinet')}
    canDuplicate={canPerform('create', 'controlcabinet')}
    canUpdate={canPerform('update', 'controlcabinet')}
    canDelete={canPerform('delete', 'controlcabinet')}
    {getBuildingLabel}
    onCreate={handleCreate}
    onSearch={(text) => controlCabinetsStore.search(text)}
    onPageChange={(page) => controlCabinetsStore.goToPage(page)}
    onReload={() => controlCabinetsStore.reload()}
    onFormSuccess={handleSuccess}
    onFormCancel={handleCancel}
    onEdit={handleEdit}
    onDelete={handleDelete}
    onDuplicate={handleDuplicate}
    onCopy={handleCopy}
  />
</div>
