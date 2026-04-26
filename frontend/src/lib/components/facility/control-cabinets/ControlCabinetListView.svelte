<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import ControlCabinetList from './ControlCabinetList.svelte';
  import { provideControlCabinetState } from './state/context.svelte.js';

  interface Props {
    projectId?: string;
    refreshKey?: string | number;
    refreshRequest?: import('../shared/entityRefresh.js').EntityRefreshRequest;
    deltaRequest?: import('../shared/entityRefresh.js').EntityDeltaRequest<import('$lib/domain/facility/index.js').ControlCabinet>;
    onChanged?: (
      event?: import('../shared/entityRefresh.js').EntityChangeEvent<
        import('$lib/domain/facility/index.js').ControlCabinet
      >
    ) => void;
  }

  const { projectId, refreshKey, refreshRequest, deltaRequest, onChanged }: Props = $props();

  const controlCabinetState = provideControlCabinetState({
    projectId: () => projectId,
    onChanged: (event) => onChanged?.(event)
  });

  let initialized = $state(false);
  let lastRefreshKey: string | number | undefined = $state(undefined);
  let lastRefreshRequestKey: string | number | undefined = $state(undefined);
  let lastDeltaRequestKey: string | number | undefined = $state(undefined);

  onMount(() => {
    initialized = true;
    lastRefreshKey = refreshKey;
    void controlCabinetState.initialize();
  });

  onDestroy(() => {
    controlCabinetState.dispose();
  });

  $effect(() => {
    const nextRefreshKey = refreshKey;

    if (!initialized) return;
    if (nextRefreshKey === undefined || nextRefreshKey === lastRefreshKey) {
      lastRefreshKey = nextRefreshKey;
      return;
    }

    lastRefreshKey = nextRefreshKey;
    void controlCabinetState.reload();
  });

  $effect(() => {
    const nextRefreshRequest = refreshRequest;

    if (!initialized) return;
    if (!nextRefreshRequest || nextRefreshRequest.key === lastRefreshRequestKey) {
      lastRefreshRequestKey = nextRefreshRequest?.key;
      return;
    }

    lastRefreshRequestKey = nextRefreshRequest.key;
    void controlCabinetState.refreshCabinets(nextRefreshRequest.entityIds ?? []);
  });

  $effect(() => {
    const nextDeltaRequest = deltaRequest;

    if (!initialized) return;
    if (!nextDeltaRequest || nextDeltaRequest.key === lastDeltaRequestKey) {
      lastDeltaRequestKey = nextDeltaRequest?.key;
      return;
    }

    lastDeltaRequestKey = nextDeltaRequest.key;
    void controlCabinetState.applyCabinetDelta(nextDeltaRequest.items);
  });
</script>

<ControlCabinetList />
