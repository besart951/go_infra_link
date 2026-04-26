<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import SPSControllerList from './SPSControllerList.svelte';
  import { provideSPSControllerState } from './state/context.svelte.js';

  interface Props {
    projectId?: string;
    refreshKey?: string | number;
    refreshRequest?: import('../shared/entityRefresh.js').EntityRefreshRequest;
    deltaRequest?: import('../shared/entityRefresh.js').EntityDeltaRequest<import('$lib/domain/facility/index.js').SPSController>;
    controlCabinetLabelRefreshRequest?: import('../shared/entityRefresh.js').EntityRefreshRequest;
    controlCabinetLabelDeltaRequest?: import('../shared/entityRefresh.js').EntityDeltaRequest<import('$lib/domain/facility/index.js').ControlCabinet>;
    controlCabinetRefreshKey?: string | number;
    onChanged?: (
      event?: import('../shared/entityRefresh.js').EntityChangeEvent<
        import('$lib/domain/facility/index.js').SPSController
      >
    ) => void;
  }

  const {
    projectId,
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
    controlCabinetRefreshKey: () => controlCabinetRefreshKey,
    onChanged: (event) => onChanged?.(event)
  });

  let initialized = $state(false);
  let lastRefreshKey: string | number | undefined = $state(undefined);
  let lastRefreshRequestKey: string | number | undefined = $state(undefined);
  let lastDeltaRequestKey: string | number | undefined = $state(undefined);
  let lastControlCabinetLabelRefreshKey: string | number | undefined = $state(undefined);
  let lastControlCabinetLabelDeltaKey: string | number | undefined = $state(undefined);

  onMount(() => {
    initialized = true;
    lastRefreshKey = refreshKey;
    void spsControllerState.initialize();
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

<SPSControllerList />
