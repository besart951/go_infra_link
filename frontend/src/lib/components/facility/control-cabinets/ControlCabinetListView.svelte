<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import ControlCabinetList from './ControlCabinetList.svelte';
  import { provideControlCabinetState } from './state/context.svelte.js';

  interface Props {
    projectId?: string;
    refreshKey?: string | number;
    onChanged?: () => void;
  }

  const { projectId, refreshKey, onChanged }: Props = $props();

  const controlCabinetState = provideControlCabinetState({
    projectId: () => projectId,
    onChanged: () => onChanged?.()
  });

  let initialized = $state(false);
  let lastRefreshKey: string | number | undefined = $state(undefined);

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
</script>

<ControlCabinetList />
