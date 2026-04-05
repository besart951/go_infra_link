<script lang="ts">
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { ChevronDown, ChevronRight } from '@lucide/svelte';
  import PermissionItem from './PermissionItem.svelte';
  import { useRolePermissionEditorState } from './state/context.svelte.js';

  import type { CategoryId } from './state/RolePermissionEditorState.svelte.js';

  interface Props {
    categoryId: CategoryId;
    resource: string;
  }

  let { categoryId, resource }: Props = $props();

  const state = useRolePermissionEditorState();

  const displayName = $derived(state.getResourceDisplayName(resource));
  const permissions = $derived(state.getPermissionsForResource(categoryId, resource));
  const counts = $derived(state.getResourceSelectionCounts(categoryId, resource));
  const selectedCount = $derived(counts.selected);
  const totalCount = $derived(counts.total);
  const isFullySelected = $derived(totalCount > 0 && selectedCount === totalCount);
  const isPartiallySelected = $derived(selectedCount > 0 && selectedCount < totalCount);

  function handleToggleExpand(): void {
    state.toggleResource(resource);
  }

  function handleToggleAll(): void {
    state.toggleAllInResource(categoryId, resource);
  }
</script>

<div class="rounded-md border bg-card">
  <!-- Resource Header -->
  <div class="flex items-center gap-2 px-3 py-2">
    <Checkbox
      checked={isFullySelected}
      indeterminate={isPartiallySelected}
      onCheckedChange={handleToggleAll}
      aria-label={`Select all ${resource} permissions`}
    />

    <button
      type="button"
      class="flex flex-1 items-center gap-2 rounded px-1 py-0.5 text-left hover:bg-accent/50"
      onclick={handleToggleExpand}
    >
      {#if state.isResourceExpanded(resource)}
        <ChevronDown class="h-3.5 w-3.5 text-muted-foreground" />
      {:else}
        <ChevronRight class="h-3.5 w-3.5 text-muted-foreground" />
      {/if}
      <span class="text-sm font-medium capitalize">{displayName}</span>
      <Badge variant="secondary" class="ml-auto text-xs">
        {selectedCount}/{totalCount}
      </Badge>
    </button>
  </div>

  <!-- Permission List -->
  {#if state.isResourceExpanded(resource)}
    <div class="border-t px-3 pb-2">
      <div class="mt-1 space-y-0.5">
        {#each permissions as perm (perm.id)}
          <PermissionItem permission={perm} />
        {/each}
      </div>
    </div>
  {/if}
</div>
