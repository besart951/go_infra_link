<script lang="ts">
  import type { Component } from 'svelte';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import { ChevronDown, ChevronRight } from '@lucide/svelte';
  import ResourceSection from './ResourceSection.svelte';
  import { useRolePermissionEditorState } from './state/context.svelte.js';

  import type { CategoryId } from './state/RolePermissionEditorState.svelte.js';

  interface Props {
    categoryId: CategoryId;
    label: string;
    icon: Component<{ class?: string }>;
  }

  let { categoryId, label, icon: Icon }: Props = $props();

  const state = useRolePermissionEditorState();

  const resources = $derived(state.getResourcesForCategory(categoryId));
  const counts = $derived(state.getCategorySelectionCounts(categoryId));
  const selectedCount = $derived(counts.selected);
  const totalCount = $derived(counts.total);
  const isFullySelected = $derived(totalCount > 0 && selectedCount === totalCount);
  const isPartiallySelected = $derived(selectedCount > 0 && selectedCount < totalCount);

  function handleToggleExpand(): void {
    state.toggleCategory(categoryId);
  }

  function handleToggleAll(): void {
    state.toggleAllInCategory(categoryId);
  }
</script>

{#if resources.length > 0}
  <div class="rounded-lg border bg-background">
    <!-- Category Header -->
    <div class="flex items-center gap-2 border-b bg-muted/50 px-4 py-3">
      <Checkbox
        checked={isFullySelected}
        indeterminate={isPartiallySelected}
        onCheckedChange={handleToggleAll}
        aria-label={`Select all ${label} permissions`}
      />

      <Button
        type="button"
        variant="ghost"
        class="flex flex-1 items-center gap-2 text-left"
        onclick={handleToggleExpand}
      >
        {#if state.isCategoryExpanded(categoryId)}
          <ChevronDown class="h-4 w-4 text-muted-foreground" />
        {:else}
          <ChevronRight class="h-4 w-4 text-muted-foreground" />
        {/if}
        <Icon class="h-5 w-5 text-primary" />
        <span class="font-semibold">{label}</span>
        <Badge variant="outline" class="ml-auto">
          {selectedCount}/{totalCount}
        </Badge>
      </Button>
    </div>

    <!-- Resources within Category -->
    {#if state.isCategoryExpanded(categoryId)}
      <div class="space-y-1 p-2">
        {#each resources as resource (resource)}
          <ResourceSection {categoryId} {resource} />
        {/each}
      </div>
    {/if}
  </div>
{/if}
