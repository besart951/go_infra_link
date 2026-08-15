<script lang="ts">
  import type { Snippet } from 'svelte';
  import type { PermissionName, ProjectPermissionName } from '$lib/api/generated/permissions.js';
  import type { ProjectCapabilities } from '$lib/domain/project/capabilities.js';
  import { can, canProject } from '$lib/utils/permissions.js';

  interface Props {
    permission?: PermissionName;
    projectCapabilities?: ProjectCapabilities | null;
    projectPermission?: ProjectPermissionName;
    children: Snippet;
  }

  let { permission, projectCapabilities, projectPermission, children }: Props = $props();

  const allowed = $derived(
    permission
      ? can(permission)
      : projectPermission
        ? canProject(projectCapabilities, projectPermission)
        : false
  );
</script>

{#if allowed}
  {@render children()}
{/if}
