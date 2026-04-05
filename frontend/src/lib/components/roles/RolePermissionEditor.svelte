<script lang="ts">
  import type { Role, Permission } from '$lib/domain/role/index.js';
  import type { Component } from 'svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Search } from '@lucide/svelte';
  import Users from '@lucide/svelte/icons/users';
  import Building2 from '@lucide/svelte/icons/building-2';
  import FolderKanban from '@lucide/svelte/icons/folder-kanban';
  import CategorySection from './CategorySection.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { provideRolePermissionEditorState } from './state/context.svelte.js';

  import type { CategoryId } from './state/RolePermissionEditorState.svelte.js';

  // ============================================================================
  // Props
  // ============================================================================

  interface Props {
    role: Role;
    allPermissions: Permission[];
    onSubmit: (data: { permissions: string[] }) => void;
    onCancel: () => void;
    isSubmitting?: boolean;
    error?: string | null;
  }

  let {
    role,
    allPermissions,
    onSubmit,
    onCancel,
    isSubmitting = false,
    error = null
  }: Props = $props();

  const t = createTranslator();

  interface CategoryView {
    id: CategoryId;
    label: string;
    icon: Component<{ class?: string }>;
  }

  const categories: CategoryView[] = [
    {
      id: 'users',
      label: 'roles.categories.users_access',
      icon: Users
    },
    {
      id: 'facility',
      label: 'roles.categories.facility',
      icon: Building2
    },
    {
      id: 'projects',
      label: 'roles.categories.projects',
      icon: FolderKanban
    }
  ];
  const state = provideRolePermissionEditorState({
    role: function (): Role {
      return role;
    },
    allPermissions: function (): Permission[] {
      return allPermissions;
    }
  });

  function handleSubmit(e: Event) {
    e.preventDefault();
    onSubmit(state.buildSubmitPayload());
  }

  function handleSelectAll(): void {
    state.selectAll();
  }

  function handleDeselectAll(): void {
    state.deselectAll();
  }
</script>

<form onsubmit={handleSubmit} class="flex h-full flex-col gap-4 p-6">
  <!-- Header -->
  <div class="shrink-0">
    <h3 class="text-lg font-semibold">{$t('roles.editor.title')}</h3>
    <p class="text-sm text-muted-foreground">
      {$t('roles.editor.description', { role: role.display_name })}
    </p>
  </div>

  <!-- Error Message -->
  {#if error}
    <div
      class="shrink-0 rounded-md border border-destructive/50 bg-destructive/10 p-3 text-sm text-destructive"
    >
      {error}
    </div>
  {/if}

  <!-- Full Access Warning -->
  <!-- Search and Bulk Actions -->
  <div class="flex shrink-0 items-center gap-4">
    <div class="relative flex-1">
      <Search class="absolute top-1/2 left-3 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
      <Input
        type="search"
        placeholder={$t('roles.permissions.search_placeholder')}
        class="pl-9"
        bind:value={state.searchQuery}
      />
    </div>
    <div class="flex gap-2">
      <Button type="button" variant="outline" size="sm" onclick={handleSelectAll}>
        {$t('roles.actions.select_all')}
      </Button>
      <Button type="button" variant="outline" size="sm" onclick={handleDeselectAll}>
        {$t('roles.actions.deselect_all')}
      </Button>
    </div>
  </div>

  <!-- Selected Count -->
  <div class="shrink-0 text-sm text-muted-foreground">
    {$t('roles.editor.selected_count', {
      selected: state.selectedCount,
      total: state.totalCount
    })}
  </div>

  <!-- Permission Categories -->
  <div class="min-h-0 flex-1 space-y-3 overflow-y-auto rounded-lg border bg-muted/30 p-3">
    {#each categories as category (category.id)}
      <CategorySection categoryId={category.id} label={$t(category.label)} icon={category.icon} />
    {/each}

    {#if !state.hasAnyPermissions}
      <div class="py-8 text-center text-muted-foreground">
        {#if state.searchQuery}
          {$t('roles.permissions.empty_match', { query: state.searchQuery })}
        {:else}
          {$t('roles.permissions.empty')}
        {/if}
      </div>
    {/if}
  </div>

  <!-- Actions -->
  <div class="flex shrink-0 justify-end gap-3 border-t pt-4">
    <Button type="button" variant="outline" onclick={onCancel} disabled={isSubmitting}>
      {$t('common.cancel')}
    </Button>
    <Button type="submit" disabled={isSubmitting}>
      {#if isSubmitting}
        <span class="mr-2 h-4 w-4 animate-spin">⟳</span>
      {/if}
      {$t('roles.actions.save_changes')}
    </Button>
  </div>
</form>
