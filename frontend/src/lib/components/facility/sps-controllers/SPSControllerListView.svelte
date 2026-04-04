<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import SPSControllerList from './SPSControllerList.svelte';
  import { provideSPSControllerState } from './state/context.svelte.js';

  interface Props {
    projectId?: string;
    refreshKey?: string | number;
    controlCabinetRefreshKey?: string | number;
    onChanged?: () => void;
  }

  const { projectId, refreshKey, controlCabinetRefreshKey, onChanged }: Props = $props();

  const spsControllerState = provideSPSControllerState({
    projectId: () => projectId,
    controlCabinetRefreshKey: () => controlCabinetRefreshKey,
    onChanged: () => onChanged?.()
  });

  let initialized = $state(false);
  let lastRefreshKey: string | number | undefined = $state(undefined);

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
</script>

<SPSControllerList />
