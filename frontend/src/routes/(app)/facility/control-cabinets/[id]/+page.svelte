<script lang="ts">
  import type { PageData } from './$types.js';
  import { goto, invalidateAll } from '$app/navigation';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import { confirm } from '$lib/stores/confirm-dialog.js';
  import { addToast } from '$lib/components/toast.svelte';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator.js';

  import ControlCabinetForm from '$lib/components/facility/forms/ControlCabinetForm.svelte';
  import SPSControllerForm from '$lib/components/facility/forms/SPSControllerForm.svelte';
  import ControlCabinetDetailHeader from '$lib/components/facility/control-cabinet-detail/ControlCabinetDetailHeader.svelte';
  import ControlCabinetOverviewCard from '$lib/components/facility/control-cabinet-detail/ControlCabinetOverviewCard.svelte';
  import ControlCabinetSPSOverview from '$lib/components/facility/control-cabinet-detail/ControlCabinetSPSOverview.svelte';

  import { ManageControlCabinetUseCase } from '$lib/application/useCases/facility/manageControlCabinetUseCase.js';
  import { controlCabinetRepository } from '$lib/infrastructure/api/controlCabinetRepository.js';
  import { ManageSPSControllerUseCase } from '$lib/application/useCases/facility/manageSPSControllerUseCase.js';
  import { spsControllerRepository } from '$lib/infrastructure/api/spsControllerRepository.js';
  import type { SPSController, SPSControllerSystemType } from '$lib/domain/facility/index.js';

  let { data }: { data: PageData } = $props();

  const t = createTranslator();
  const manageControlCabinet = new ManageControlCabinetUseCase(controlCabinetRepository);
  const manageSpsController = new ManageSPSControllerUseCase(spsControllerRepository);

  let showCabinetEdit = $state(false);
  let showSpsCreate = $state(false);
  let editingSps = $state<SPSController | undefined>(undefined);

  const controllers = $derived(data.spsControllers ?? []);
  const systemTypesByController = $derived(data.systemTypesByController ?? {});

  const canReadSps = canPerform('read', 'spscontroller');
  const canCreateSps = canPerform('create', 'spscontroller');
  const canUpdateSps = canPerform('update', 'spscontroller');
  const canDeleteSps = canPerform('delete', 'spscontroller');
  const canUpdateCabinet = canPerform('update', 'controlcabinet');
  const canDeleteCabinet = canPerform('delete', 'controlcabinet');

  function resetForms() {
    showCabinetEdit = false;
    showSpsCreate = false;
    editingSps = undefined;
  }

  async function refreshAfterChange() {
    await invalidateAll();
    resetForms();
  }

  async function handleDeleteCabinet() {
    try {
      const impact = await manageControlCabinet.getDeleteImpact(data.cabinet.id);

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

      await manageControlCabinet.delete(data.cabinet.id);
      addToast($t('facility.control_cabinet_deleted'), 'success');
      await goto('/facility/control-cabinets');
    } catch (error) {
      addToast(
        error instanceof Error ? error.message : $t('facility.delete_control_cabinet_failed'),
        'error'
      );
    }
  }

  async function handleDeleteSps(controller: SPSController) {
    const ok = await confirm({
      title: $t('common.delete'),
      message: $t('facility.delete_sps_controller_confirm').replace(
        '{name}',
        controller.device_name
      ),
      confirmText: $t('common.delete'),
      cancelText: $t('common.cancel'),
      variant: 'destructive'
    });

    if (!ok) return;

    try {
      await manageSpsController.delete(controller.id);
      addToast($t('facility.sps_controller_deleted'), 'success');
      await refreshAfterChange();
    } catch (error) {
      addToast(
        error instanceof Error ? error.message : $t('facility.delete_sps_controller_failed'),
        'error'
      );
    }
  }

  async function handleCopySps(controller: SPSController) {
    try {
      await manageSpsController.copy(controller.id);
      addToast($t('facility.sps_controller_copied'), 'success');
      await refreshAfterChange();
    } catch (error) {
      addToast(error instanceof Error ? error.message : $t('facility.copy_failed'), 'error');
    }
  }
</script>

<svelte:head>
  <title
    >#{data.cabinet.control_cabinet_nr} | {$t('facility.control_cabinets_title')} | Infra Link</title
  >
</svelte:head>

<ConfirmDialog />

<div class="space-y-6">
  <ControlCabinetDetailHeader
    cabinet={data.cabinet}
    canUpdate={canUpdateCabinet}
    canDelete={canDeleteCabinet}
    onEdit={() => {
      showCabinetEdit = true;
      showSpsCreate = false;
      editingSps = undefined;
    }}
    onDelete={handleDeleteCabinet}
  />

  {#if showCabinetEdit}
    <ControlCabinetForm
      initialData={data.cabinet}
      onSuccess={refreshAfterChange}
      onCancel={() => (showCabinetEdit = false)}
    />
  {/if}

  {#if showSpsCreate}
    <SPSControllerForm
      fixedControlCabinetId={data.cabinet.id}
      onSuccess={refreshAfterChange}
      onCancel={() => (showSpsCreate = false)}
    />
  {/if}

  {#if editingSps}
    <SPSControllerForm
      initialData={editingSps}
      fixedControlCabinetId={data.cabinet.id}
      onSuccess={refreshAfterChange}
      onCancel={() => (editingSps = undefined)}
    />
  {/if}

  <ControlCabinetOverviewCard cabinet={data.cabinet} building={data.building ?? null} />

  <ControlCabinetSPSOverview
    {controllers}
    {systemTypesByController}
    canRead={canReadSps}
    canCreate={canCreateSps}
    canUpdate={canUpdateSps}
    canDelete={canDeleteSps}
    onCreate={() => {
      showSpsCreate = true;
      showCabinetEdit = false;
      editingSps = undefined;
    }}
    onEdit={(controller) => {
      editingSps = controller;
      showCabinetEdit = false;
      showSpsCreate = false;
    }}
    onCopy={handleCopySps}
    onDelete={handleDeleteSps}
  />
</div>
