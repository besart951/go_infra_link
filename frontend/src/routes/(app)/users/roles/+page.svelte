<script lang="ts">
  import { onMount } from 'svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Skeleton } from '$lib/components/ui/skeleton/index.js';
  import * as Tabs from '$lib/components/ui/tabs/index.js';
  import * as Dialog from '$lib/components/ui/dialog/index.js';
  import * as Sheet from '$lib/components/ui/sheet/index.js';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import {
    RoleCard,
    PermissionForm,
    PermissionTable,
    PermissionMatrix,
    RolePermissionEditor,
    PhasePermissionRules
  } from '$lib/components/roles/index.js';
  import { RolesPageState } from '$lib/components/roles/state/RolesPageState.svelte.js';
  import {
    GitBranch,
    RefreshCw,
    Grid3X3,
    LayoutGrid,
    Shield,
    Users
  } from '@lucide/svelte';

  const t = createTranslator();
  const state = new RolesPageState();

  onMount(() => {
    state.loadData();
  });
</script>

<svelte:head>
  <title>{$t('roles.page.title')}</title>
</svelte:head>

<ConfirmDialog />

<div class="flex flex-col gap-6">
  <EntityListHeader
    title={$t('roles.page.title')}
    description={$t('roles.page.description')}
    infoLabel={$t('common.info')}
    backHref="/users"
    backLabel={$t('hub.back_to_overview')}
    createLabel={$t('roles.actions.create_permission')}
    canCreate={state.canManageRoles}
    createActive={state.createPermissionDialogOpen}
    onCreateClick={state.openCreatePermissionDialog}
  >
    <Button variant="outline" size="icon" onclick={state.loadData} disabled={state.isLoading} aria-label={$t('roles.actions.refresh')}>
      <RefreshCw class="h-4 w-4 {state.isLoading ? 'animate-spin' : ''}" />
    </Button>
  </EntityListHeader>

  <div class="grid gap-4 sm:grid-cols-3">
    <div class="rounded-lg border bg-card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-md bg-primary/10 p-2">
          <Users class="h-5 w-5 text-primary" />
        </div>
        <div>
          <p class="text-sm text-muted-foreground">{$t('roles.stats.total_roles')}</p>
          {#if state.isLoading}
            <Skeleton class="h-7 w-12" />
          {:else}
            <p class="text-2xl font-bold">{state.totalRoles}</p>
          {/if}
        </div>
      </div>
    </div>

    <div class="rounded-lg border bg-card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-md bg-green-500/10 p-2">
          <Shield class="h-5 w-5 text-green-600" />
        </div>
        <div>
          <p class="text-sm text-muted-foreground">{$t('roles.stats.total_permissions')}</p>
          {#if state.isLoading}
            <Skeleton class="h-7 w-12" />
          {:else}
            <p class="text-2xl font-bold">{state.totalPermissions}</p>
          {/if}
        </div>
      </div>
    </div>

    <div class="rounded-lg border bg-card p-4">
      <div class="flex items-center gap-3">
        <div class="rounded-md bg-blue-500/10 p-2">
          <Grid3X3 class="h-5 w-5 text-blue-600" />
        </div>
        <div>
          <p class="text-sm text-muted-foreground">{$t('roles.stats.resources')}</p>
          {#if state.isLoading}
            <Skeleton class="h-7 w-12" />
          {:else}
            <p class="text-2xl font-bold">{state.uniqueResources}</p>
          {/if}
        </div>
      </div>
    </div>
  </div>

  {#if state.error}
    <div class="rounded-md border border-destructive/50 bg-destructive/10 p-4 text-destructive">
      <p class="font-medium">{$t('roles.errors.load_title')}</p>
      <p class="text-sm">{state.error}</p>
    </div>
  {/if}

  <Tabs.Root bind:value={state.activeTab}>
    <Tabs.List>
      <Tabs.Trigger value="roles" class="gap-2">
        <LayoutGrid class="h-4 w-4" />
        {$t('roles.tabs.roles')}
      </Tabs.Trigger>
      <Tabs.Trigger value="permissions" class="gap-2">
        <Shield class="h-4 w-4" />
        {$t('roles.tabs.permissions')}
      </Tabs.Trigger>
      <Tabs.Trigger value="matrix" class="gap-2">
        <Grid3X3 class="h-4 w-4" />
        {$t('roles.tabs.matrix')}
      </Tabs.Trigger>
      {#if state.canManagePhaseRules}
        <Tabs.Trigger value="phase-rules" class="gap-2">
          <GitBranch class="h-4 w-4" />
          {$t('roles.tabs.phase_rules')}
        </Tabs.Trigger>
      {/if}
    </Tabs.List>

    <Tabs.Content value="roles" class="mt-6">
      {#if state.isLoading}
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {#each Array(7) as _}
            <div class="rounded-lg border p-4">
              <Skeleton class="mb-3 h-6 w-24" />
              <Skeleton class="mb-2 h-4 w-full" />
              <Skeleton class="h-4 w-3/4" />
              <div class="mt-4 space-y-2">
                <Skeleton class="h-4 w-16" />
                <div class="flex gap-1">
                  <Skeleton class="h-5 w-16" />
                  <Skeleton class="h-5 w-20" />
                </div>
              </div>
            </div>
          {/each}
        </div>
      {:else}
        <div class="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
          {#each state.roles as role (role.id)}
            <RoleCard
              {role}
              onEdit={state.canManageRoles ? state.openEditRole : undefined}
              onViewPermissions={state.viewRolePermissions}
              canEdit={state.canManageRoles}
            />
          {/each}
        </div>
      {/if}
    </Tabs.Content>

    <Tabs.Content value="permissions" class="mt-6">
      {#if state.isLoading}
        <div class="space-y-4">
          <div class="flex items-center justify-between">
            <Skeleton class="h-10 w-64" />
            <Skeleton class="h-10 w-40" />
          </div>
          <div class="rounded-lg border">
            {#each Array(5) as _}
              <div class="flex items-center gap-4 border-b p-4">
                <Skeleton class="h-4 w-32" />
                <Skeleton class="h-5 w-16" />
                <Skeleton class="h-5 w-16" />
                <Skeleton class="h-4 flex-1" />
              </div>
            {/each}
          </div>
        </div>
      {:else}
        <PermissionTable
          permissions={state.permissions}
          onEdit={state.canManageRoles ? state.openEditPermission : undefined}
          onDelete={state.canManageRoles ? state.deletePermission : undefined}
          onCreate={state.canManageRoles ? state.openCreatePermissionDialog : undefined}
          canManage={state.canManageRoles}
        />
      {/if}
    </Tabs.Content>

    <Tabs.Content value="matrix" class="mt-6">
      {#if state.isLoading}
        <div class="space-y-4">
          <Skeleton class="h-4 w-48" />
          <div class="rounded-lg border">
            <div class="overflow-x-auto">
              <div class="flex gap-4 p-4">
                {#each Array(8) as _}
                  <Skeleton class="h-75 w-30" />
                {/each}
              </div>
            </div>
          </div>
        </div>
      {:else}
        <PermissionMatrix roles={state.roles} permissions={state.permissions} />
      {/if}
    </Tabs.Content>

    <Tabs.Content value="phase-rules" class="mt-6">
      {#if state.isLoading}
        <div class="space-y-4">
          <Skeleton class="h-10 w-full" />
          <Skeleton class="h-40 w-full" />
          <Skeleton class="h-40 w-full" />
        </div>
      {:else if state.canManagePhaseRules}
        <PhasePermissionRules
          roles={state.roles}
          phases={state.phases}
          permissions={state.permissions}
          rules={state.phasePermissions}
          canManage={state.canManagePhaseRules}
          onRulesChange={state.reloadPhaseRules}
        />
      {/if}
    </Tabs.Content>
  </Tabs.Root>
</div>

<Dialog.Root bind:open={state.createPermissionDialogOpen}>
  <Dialog.Content class="sm:max-w-125">
    <PermissionForm
      onSubmit={state.createPermission}
      onCancel={state.closeCreatePermissionDialog}
      isSubmitting={state.isSubmittingPermission}
      error={state.permissionError}
    />
  </Dialog.Content>
</Dialog.Root>

<Dialog.Root bind:open={state.editPermissionDialogOpen}>
  <Dialog.Content class="sm:max-w-125">
    <PermissionForm
      permission={state.selectedPermission}
      onSubmit={state.updatePermission}
      onCancel={state.closeEditPermissionDialog}
      isSubmitting={state.isSubmittingPermission}
      error={state.permissionError}
    />
  </Dialog.Content>
</Dialog.Root>

<Sheet.Root bind:open={state.editRoleSheetOpen}>
  <Sheet.Content side="right" class="flex h-full w-full flex-col p-0 sm:max-w-2xl">
    {#if state.selectedRole}
      <RolePermissionEditor
        role={state.selectedRole}
        allPermissions={state.permissions}
        onSubmit={state.updateSelectedRolePermissions}
        onCancel={state.closeRoleSheet}
        isSubmitting={state.isSubmittingRole}
        error={state.roleError}
      />
    {/if}
  </Sheet.Content>
</Sheet.Root>
