<script lang="ts">
  import type { PageData } from './$types.js';
  import { goto, invalidateAll } from '$app/navigation';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import { confirm } from '$lib/stores/confirm-dialog.js';
  import { addToast } from '$lib/components/toast.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';

  import SPSControllerForm from '$lib/components/facility/forms/SPSControllerForm.svelte';
  import SPSControllerDetailHeader from '$lib/components/facility/sps-controller-detail/SPSControllerDetailHeader.svelte';
  import SPSControllerOverviewCard from '$lib/components/facility/sps-controller-detail/SPSControllerOverviewCard.svelte';
  import SPSControllerSystemTypesOverview from '$lib/components/facility/sps-controller-detail/SPSControllerSystemTypesOverview.svelte';
  import { provideSPSControllerDetailState } from '$lib/components/facility/sps-controller-detail/state/context.svelte.js';

  import type { SPSControllerDetailData } from '$lib/components/facility/sps-controller-detail/state/SPSControllerDetailState.svelte.js';

  let { data }: { data: PageData } = $props();

  const t = createTranslator();
  const state = provideSPSControllerDetailState({
    data: function (): SPSControllerDetailData {
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

  function handleEditCancel(): void {
    state.cancelEdit();
  }
</script>

<svelte:head>
  <title>{data.controller.device_name} | {$t('facility.sps_controllers_title')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="space-y-6">
  <SPSControllerDetailHeader />

  {#if state.showEdit}
    <SPSControllerForm
      initialData={data.controller}
      onSuccess={handleRefreshAfterChange}
      onCancel={handleEditCancel}
    />
  {/if}

  <div class="grid gap-6 xl:grid-cols-[minmax(0,380px)_minmax(0,1fr)]">
    <SPSControllerOverviewCard />
    <SPSControllerSystemTypesOverview />
  </div>
</div>
