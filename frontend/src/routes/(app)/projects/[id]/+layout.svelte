<script lang="ts">
  import type { Snippet } from 'svelte';
  import type { LayoutData } from './$types.js';
  import ProjectCollaborationProvider from '$lib/components/project/ProjectCollaborationProvider.svelte';

  let { children, data }: { children: Snippet; data: LayoutData } = $props();
</script>

<ProjectCollaborationProvider projectId={data.project.id}>
  {@render children()}
</ProjectCollaborationProvider>
