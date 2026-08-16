<script lang="ts">
  import type { PageData } from './$types.js';
  import { goto, invalidateAll } from '$app/navigation';
  import { onMount } from 'svelte';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import { confirm } from '$lib/stores/confirm-dialog.js';
  import { addToast } from '$lib/components/toast.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { Button } from '$lib/components/ui/button/index.js';

  import ControlCabinetForm from '$lib/components/facility/forms/ControlCabinetForm.svelte';
  import SPSControllerForm from '$lib/components/facility/forms/SPSControllerForm.svelte';
  import ControlCabinetDetailHeader from '$lib/components/facility/control-cabinet-detail/ControlCabinetDetailHeader.svelte';
  import ControlCabinetOverviewCard from '$lib/components/facility/control-cabinet-detail/ControlCabinetOverviewCard.svelte';
  import ControlCabinetSPSOverview from '$lib/components/facility/control-cabinet-detail/ControlCabinetSPSOverview.svelte';
  import { provideControlCabinetDetailState } from '$lib/components/facility/control-cabinet-detail/state/context.svelte.js';

  import type { ControlCabinetDetailData } from '$lib/components/facility/control-cabinet-detail/state/ControlCabinetDetailState.svelte.js';
  import { facilityReferenceDataCache } from '$lib/services/facilityReferenceDataCache.js';

  let { data }: { data: PageData } = $props();

  const t = createTranslator();
  const detailState = provideControlCabinetDetailState({
    data: function (): ControlCabinetDetailData {
      return data;
    },
    confirmAction: confirm,
    toastAction: addToast,
    gotoAction: goto,
    invalidateAllAction: invalidateAll
  });
  let remoteChangePending = $state(false);

  onMount(() =>
    facilityReferenceDataCache.subscribeFacilityChanges((event) => {
      if (
        event.resource !== 'control_cabinets' &&
        event.resource !== 'sps_controllers' &&
        event.resource !== 'sps_controller_system_types' &&
        event.resource !== 'all'
      )
        return;
      if (detailState.showCabinetEdit || detailState.showSpsCreate || detailState.editingSps) {
        if (!facilityReferenceDataCache.isChangeFromCurrentUser(event)) {
          remoteChangePending = true;
        }
        return;
      }
      void invalidateAll();
    })
  );

  async function handleRefreshAfterChange(): Promise<void> {
    await detailState.refreshAfterChange();
  }

  function handleCabinetEditCancel(): void {
    detailState.cancelCabinetEdit();
  }

  function handleSpsCreateCancel(): void {
    detailState.cancelSpsCreate();
  }

  function handleSpsEditCancel(): void {
    detailState.cancelSpsEdit();
  }

  async function reloadRemoteChange(): Promise<void> {
    await invalidateAll();
    remoteChangePending = false;
  }
</script>

<svelte:head>
  <title
    >#{data.cabinet.control_cabinet_nr} | {$t('facility.control_cabinets_title')} | Infra Link</title
  >
</svelte:head>

<ConfirmDialog />

<div class="space-y-6">
  <ControlCabinetDetailHeader />

  {#if remoteChangePending}
    <div
      class="flex items-center justify-between gap-3 rounded-md border border-warning/40 bg-warning/10 p-4 text-sm"
    >
      <span>{$t('facility.realtime_change_pending')}</span>
      <Button variant="outline" size="sm" onclick={reloadRemoteChange}
        >{$t('common.refresh')}</Button
      >
    </div>
  {/if}

  {#if detailState.showCabinetEdit}
    <ControlCabinetForm
      initialData={data.cabinet}
      onSuccess={handleRefreshAfterChange}
      onCancel={handleCabinetEditCancel}
    />
  {/if}

  {#if detailState.showSpsCreate}
    <SPSControllerForm
      fixedControlCabinetId={data.cabinet.id}
      onSuccess={handleRefreshAfterChange}
      onCancel={handleSpsCreateCancel}
    />
  {/if}

  {#if detailState.editingSps}
    <SPSControllerForm
      initialData={detailState.editingSps}
      fixedControlCabinetId={data.cabinet.id}
      onSuccess={handleRefreshAfterChange}
      onCancel={handleSpsEditCancel}
    />
  {/if}

  <ControlCabinetOverviewCard cabinet={detailState.cabinet} building={detailState.building} />

  <ControlCabinetSPSOverview />
</div>
