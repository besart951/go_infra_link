<script lang="ts">
  import { onDestroy } from 'svelte';
  import {
    provideProjectSyncCoordinator,
    type ProjectChange
  } from '$lib/services/projectCollaboration.svelte.js';
  import { facilityDetailCache } from '$lib/services/facilityDetailCache.js';

  let {
    projectId,
    children
  }: {
    projectId: string;
    children: import('svelte').Snippet;
  } = $props();

  const collaboration = provideProjectSyncCoordinator({
    onProjectChange: (_change: ProjectChange) => {
      facilityDetailCache.invalidateProject(projectId);
    },
    onResetRequired: () => facilityDetailCache.invalidateProject(projectId),
    onReconnect: () => facilityDetailCache.invalidateProject(projectId)
  });

  $effect(() => {
    if (!projectId) return;
    collaboration.connect(projectId);
  });

  onDestroy(() => collaboration.disconnect());
</script>

{@render children()}
