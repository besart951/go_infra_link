<script lang="ts">
  import type { PageData } from './$types.js';
  import { goto, invalidateAll } from '$app/navigation';
  import { onMount } from 'svelte';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import { confirm } from '$lib/stores/confirm-dialog.js';
  import { addToast } from '$lib/components/toast.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { Button } from '$lib/components/ui/button/index.js';

  import SPSControllerForm from '$lib/components/facility/forms/SPSControllerForm.svelte';
  import SPSControllerDetailHeader from '$lib/components/facility/sps-controller-detail/SPSControllerDetailHeader.svelte';
  import SPSControllerOverviewCard from '$lib/components/facility/sps-controller-detail/SPSControllerOverviewCard.svelte';
  import SPSControllerSystemTypesOverview from '$lib/components/facility/sps-controller-detail/SPSControllerSystemTypesOverview.svelte';
  import { provideSPSControllerDetailState } from '$lib/components/facility/sps-controller-detail/state/context.svelte.js';

  import type { SPSControllerDetailData } from '$lib/components/facility/sps-controller-detail/state/SPSControllerDetailState.svelte.js';
  import { facilityReferenceDataCache } from '$lib/services/facilityReferenceDataCache.js';

  let { data }: { data: PageData } = $props();

  const t = createTranslator();
  const detailState = provideSPSControllerDetailState({
    data: function (): SPSControllerDetailData {
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
        event.resource !== 'sps_controllers' &&
        event.resource !== 'sps_controller_system_types' &&
        event.resource !== 'all'
      )
        return;
      if (detailState.showEdit) {
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

  function handleEditCancel(): void {
    detailState.cancelEdit();
  }

  async function reloadRemoteChange(): Promise<void> {
    await invalidateAll();
    remoteChangePending = false;
  }
</script>

<svelte:head>
  <title>{data.controller.device_name} | {$t('facility.sps_controllers_title')} | Infra Link</title>
</svelte:head>

<ConfirmDialog />

<div class="space-y-6">
  <SPSControllerDetailHeader />

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

  {#if detailState.showEdit}
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
