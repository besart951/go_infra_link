<script lang="ts">
  import type { PageData } from './$types.js';
  import { goto, invalidateAll } from '$app/navigation';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import { confirm } from '$lib/stores/confirm-dialog.js';
  import { addToast } from '$lib/components/toast.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';

  import ControlCabinetForm from '$lib/components/facility/forms/ControlCabinetForm.svelte';
  import SPSControllerForm from '$lib/components/facility/forms/SPSControllerForm.svelte';
  import ControlCabinetDetailHeader from '$lib/components/facility/control-cabinet-detail/ControlCabinetDetailHeader.svelte';
  import ControlCabinetOverviewCard from '$lib/components/facility/control-cabinet-detail/ControlCabinetOverviewCard.svelte';
  import ControlCabinetSPSOverview from '$lib/components/facility/control-cabinet-detail/ControlCabinetSPSOverview.svelte';
  import { provideControlCabinetDetailState } from '$lib/components/facility/control-cabinet-detail/state/context.svelte.js';

  import type { ControlCabinetDetailData } from '$lib/components/facility/control-cabinet-detail/state/ControlCabinetDetailState.svelte.js';

  let { data }: { data: PageData } = $props();

  const t = createTranslator();
  const state = provideControlCabinetDetailState({
    data: function (): ControlCabinetDetailData {
      return data;
    },
    confirmAction: confirm,
    toastAction: addToast,
    gotoAction: goto,
    invalidateAllAction: invalidateAll
  });

  async function handleRefreshAfterChange(): Promise<void> {
    await state.refreshAfterChange();
  }

  function handleCabinetEditCancel(): void {
    state.cancelCabinetEdit();
  }

  function handleSpsCreateCancel(): void {
    state.cancelSpsCreate();
  }

  function handleSpsEditCancel(): void {
    state.cancelSpsEdit();
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

  {#if state.showCabinetEdit}
    <ControlCabinetForm
      initialData={data.cabinet}
      onSuccess={handleRefreshAfterChange}
      onCancel={handleCabinetEditCancel}
    />
  {/if}

  {#if state.showSpsCreate}
    <SPSControllerForm
      fixedControlCabinetId={data.cabinet.id}
      onSuccess={handleRefreshAfterChange}
      onCancel={handleSpsCreateCancel}
    />
  {/if}

  {#if state.editingSps}
    <SPSControllerForm
      initialData={state.editingSps}
      fixedControlCabinetId={data.cabinet.id}
      onSuccess={handleRefreshAfterChange}
      onCancel={handleSpsEditCancel}
    />
  {/if}

  <ControlCabinetOverviewCard cabinet={state.cabinet} building={state.building} />

  <ControlCabinetSPSOverview />
</div>
