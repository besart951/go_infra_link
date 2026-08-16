<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import SPSControllerList from './SPSControllerList.svelte';
  import { provideSPSControllerState } from './state/context.svelte.js';
  import type { ProjectCapabilities } from '$lib/domain/project/capabilities.js';
  import { facilityReferenceDataCache } from '$lib/services/facilityReferenceDataCache.js';
  import { createTranslator } from '$lib/i18n/translator.js';

  interface Props {
    projectId?: string;
    projectCapabilities?: ProjectCapabilities;
    refreshKey?: string | number;
    refreshRequest?: import('../shared/entityRefresh.js').EntityRefreshRequest;
    deltaRequest?: import('../shared/entityRefresh.js').EntityDeltaRequest<
      import('$lib/domain/facility/index.js').SPSController
    >;
    controlCabinetLabelRefreshRequest?: import('../shared/entityRefresh.js').EntityRefreshRequest;
    controlCabinetLabelDeltaRequest?: import('../shared/entityRefresh.js').EntityDeltaRequest<
      import('$lib/domain/facility/index.js').ControlCabinet
    >;
    controlCabinetRefreshKey?: string | number;
    onChanged?: (
      event?: import('../shared/entityRefresh.js').EntityChangeEvent<
        import('$lib/domain/facility/index.js').SPSController
      >
    ) => void;
  }

  const {
    projectId,
    projectCapabilities,
    refreshKey,
    refreshRequest,
    deltaRequest,
    controlCabinetLabelRefreshRequest,
    controlCabinetLabelDeltaRequest,
    controlCabinetRefreshKey,
    onChanged
  }: Props = $props();

  const spsControllerState = provideSPSControllerState({
    projectId: () => projectId,
    projectCapabilities: () => projectCapabilities,
    controlCabinetRefreshKey: () => controlCabinetRefreshKey,
    onChanged: (event) => onChanged?.(event)
  });
  const t = createTranslator();

  let initialized = $state(false);
  let lastRefreshKey: string | number | undefined = $state(undefined);
  let lastRefreshRequestKey: string | number | undefined = $state(undefined);
  let lastDeltaRequestKey: string | number | undefined = $state(undefined);
  let lastControlCabinetLabelRefreshKey: string | number | undefined = $state(undefined);
  let lastControlCabinetLabelDeltaKey: string | number | undefined = $state(undefined);
  let remoteChangePending = $state(false);

  onMount(() => {
    initialized = true;
    lastRefreshKey = refreshKey;
    void spsControllerState.initialize();
    return facilityReferenceDataCache.subscribeFacilityChanges((event) => {
      if (projectId || (event.resource !== 'sps_controllers' && event.resource !== 'all')) return;
      if (spsControllerState.showForm) {
        if (!facilityReferenceDataCache.isChangeFromCurrentUser(event)) {
          remoteChangePending = true;
        }
        return;
      }
      void spsControllerState.reload();
    });
  });

  onDestroy(() => {
    spsControllerState.dispose();
  });

  $effect(() => {
    const nextRefreshKey = refreshKey;

    if (!initialized) return;
    if (nextRefreshKey === undefined || nextRefreshKey === lastRefreshKey) {
      lastRefreshKey = nextRefreshKey;
      return;
    }

    lastRefreshKey = nextRefreshKey;
    void spsControllerState.reload();
  });

  $effect(() => {
    const nextRefreshRequest = refreshRequest;

    if (!initialized) return;
    if (!nextRefreshRequest || nextRefreshRequest.key === lastRefreshRequestKey) {
      lastRefreshRequestKey = nextRefreshRequest?.key;
      return;
    }

    lastRefreshRequestKey = nextRefreshRequest.key;
    void spsControllerState.refreshControllers(nextRefreshRequest.entityIds ?? []);
  });

  $effect(() => {
    const nextDeltaRequest = deltaRequest;

    if (!initialized) return;
    if (!nextDeltaRequest || nextDeltaRequest.key === lastDeltaRequestKey) {
      lastDeltaRequestKey = nextDeltaRequest?.key;
      return;
    }

    lastDeltaRequestKey = nextDeltaRequest.key;
    void spsControllerState.applyControllerDelta(nextDeltaRequest.items);
  });

  $effect(() => {
    const nextRefreshRequest = controlCabinetLabelRefreshRequest;

    if (!initialized) return;
    if (!nextRefreshRequest || nextRefreshRequest.key === lastControlCabinetLabelRefreshKey) {
      lastControlCabinetLabelRefreshKey = nextRefreshRequest?.key;
      return;
    }

    lastControlCabinetLabelRefreshKey = nextRefreshRequest.key;
    void spsControllerState.refreshCabinetLabels(nextRefreshRequest.entityIds ?? []);
  });

  $effect(() => {
    const nextDeltaRequest = controlCabinetLabelDeltaRequest;

    if (!initialized) return;
    if (!nextDeltaRequest || nextDeltaRequest.key === lastControlCabinetLabelDeltaKey) {
      lastControlCabinetLabelDeltaKey = nextDeltaRequest?.key;
      return;
    }

    lastControlCabinetLabelDeltaKey = nextDeltaRequest.key;
    spsControllerState.applyCabinetLabelDelta(nextDeltaRequest.items);
  });
</script>

{#if remoteChangePending}
  <div
    class="mb-3 flex items-center justify-between gap-3 rounded-md border border-warning-border bg-warning-muted px-3 py-2 text-sm text-warning-muted-foreground"
    role="status"
  >
    <span>{$t('facility.realtime_change_pending')}</span>
    <button
      class="font-medium underline underline-offset-2"
      type="button"
      onclick={() => {
        remoteChangePending = false;
        void spsControllerState.reload();
      }}
    >
      {$t('common.refresh')}
    </button>
  </div>
{/if}

<SPSControllerList />
