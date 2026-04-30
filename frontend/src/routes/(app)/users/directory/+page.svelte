<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import * as Dialog from '$lib/components/ui/dialog/index.js';
  import { addToast } from '$lib/components/toast.svelte';
  import { confirm } from '$lib/stores/confirm-dialog.js';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import RoleBadge from '$lib/components/role-badge.svelte';
  import UserAvatar from '$lib/components/user-avatar.svelte';
  import UserManagementForm from '$lib/components/user-management-form.svelte';
  import {
    listUserDirectory,
    setUserRole,
    disableUser,
    enableUser,
    deleteUser,
    type UserRole,
    type UserDirectoryUser,
    type UserDirectoryTeamFilter,
    type UserDirectoryPageCapabilities
  } from '$lib/api/users.js';
  import { getErrorMessage } from '$lib/api/client.js';
  import { getAllowedRolesForCreation, auth } from '$lib/stores/auth.svelte.js';
  import {
    MoreVertical,
    UserMinus,
    UserCheck,
    Trash2,
    BadgeCheck,
    BadgeX,
    KeyRound,
    UserPlus,
    ArrowLeft
  } from '@lucide/svelte';
  import { createTranslator } from '$lib/i18n/translator';

  const t = createTranslator();

  let users = $state<UserDirectoryUser[]>([]);
  let total = $state(0);
  let page = $state(1);
  let totalPages = $state(1);
  let searchText = $state('');
  let selectedTeamId = $state<string>('all');
  let teamFilters = $state<UserDirectoryTeamFilter[]>([]);
  let pageCapabilities = $state<UserDirectoryPageCapabilities>({ can_create_user: false });
  let isLoading = $state(true);
  let error = $state<string | null>(null);
  let createDialogOpen = $state(false);

  async function loadDirectory(
    nextPage = page,
    nextSearch = searchText,
    nextTeamId = selectedTeamId
  ) {
    isLoading = true;
    error = null;
    try {
      const result = await listUserDirectory({
        page: nextPage,
        limit: 10,
        search: nextSearch || undefined,
        team_id: nextTeamId === 'all' ? undefined : nextTeamId
      });
      users = result.items;
      total = result.total;
      page = result.page;
      totalPages = result.total_pages;
      teamFilters = result.teams;
      pageCapabilities = result.capabilities;
    } catch (err) {
      error = getErrorMessage(err);
    } finally {
      isLoading = false;
    }
  }

  async function handleRoleChange(userId: string, newRole: UserRole) {
    try {
      await setUserRole(userId, newRole);
      await loadDirectory();
      addToast($t('messages.role_updated_success'), 'success');
    } catch (err) {
      addToast(err instanceof Error ? err.message : $t('errors.change_role_failed'), 'error');
    }
  }

  async function handleToggleActive(userId: string, isActive: boolean) {
    try {
      if (isActive) {
        await disableUser(userId);
        addToast($t('messages.user_disabled_success'), 'success');
      } else {
        await enableUser(userId);
        addToast($t('messages.user_enabled_success'), 'success');
      }
      await loadDirectory();
    } catch (err) {
      addToast(
        err instanceof Error ? err.message : $t('errors.toggle_user_status_failed'),
        'error'
      );
    }
  }

  async function handleDeleteUser(userId: string, userName: string) {
    const confirmed = await confirm({
      title: $t('common.delete_user'),
      message: $t('messages.delete_user_confirm', { name: userName }),
      confirmText: $t('common.delete'),
      cancelText: $t('common.cancel'),
      variant: 'destructive'
    });

    if (!confirmed) return;

    try {
      await deleteUser(userId);
      await loadDirectory();
      addToast($t('messages.user_deleted_success'), 'success');
    } catch (err) {
      addToast(err instanceof Error ? err.message : $t('errors.delete_user_failed'), 'error');
    }
  }

  function formatDate(dateString: string | null | undefined): string {
    if (!dateString) return $t('messages.never');
    const date = new Date(dateString);
    const now = new Date();
    const diffInMs = now.getTime() - date.getTime();
    const diffInDays = Math.floor(diffInMs / (1000 * 60 * 60 * 24));

    if (diffInDays === 0) return $t('messages.today');
    if (diffInDays === 1) return $t('messages.yesterday');
    if (diffInDays < 7) return $t('messages.days_ago').replace('{count}', String(diffInDays));
    if (diffInDays < 30)
      return $t('messages.weeks_ago').replace('{count}', String(Math.floor(diffInDays / 7)));
    if (diffInDays < 365)
      return $t('messages.months_ago').replace('{count}', String(Math.floor(diffInDays / 30)));
    return $t('messages.years_ago').replace('{count}', String(Math.floor(diffInDays / 365)));
  }

  function authVerified(user: UserDirectoryUser): boolean {
    return Boolean(user.is_active && !user.disabled_at);
  }

  function twoFactorEnabled(_user: UserDirectoryUser): boolean {
    return false;
  }

  function roleOptionsFor(user: UserDirectoryUser) {
    return getAllowedRolesForCreation().filter((roleObj) => roleObj.role !== user.role);
  }

  onMount(() => {
    if (!auth.canAccessUserDirectory) {
      void goto('/');
      return;
    }
    void loadDirectory();
  });
</script>

<svelte:head>
  <title>{$t('navigation.users')} | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
  <!-- Header -->
  <div class="flex items-center justify-between">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">{$t('pages.user_management')}</h1>
      <p class="mt-1 text-muted-foreground">{$t('pages.user_management_desc')}</p>
    </div>
    <div class="flex flex-col gap-2 sm:flex-row">
      <Button variant="outline" href="/users">
        <ArrowLeft class="size-4" />
        {$t('hub.back_to_overview')}
      </Button>
      {#if pageCapabilities.can_create_user}
        <Button onclick={() => (createDialogOpen = true)}>
          <UserPlus class="mr-2 h-4 w-4" />
          {$t('common.create_user')}
        </Button>
      {/if}
    </div>
  </div>

  <div class="flex flex-wrap items-center justify-between gap-3">
    <div class="flex flex-1 items-center gap-3">
      <input
        class="h-9 min-w-55 flex-1 rounded-md border bg-background px-3 text-sm"
        bind:value={searchText}
        placeholder={$t('messages.search_users')}
        onkeydown={(event) => {
          if (event.key === 'Enter') {
            void loadDirectory(1, searchText, selectedTeamId);
          }
        }}
      />
      <Button
        variant="outline"
        onclick={() => void loadDirectory(1, searchText, selectedTeamId)}
        disabled={isLoading}
      >
        {$t('messages.refresh')}
      </Button>
    </div>
    <div class="text-sm text-muted-foreground">
      {#if selectedTeamId === 'all'}
        {total}
        {total === 1 ? $t('common.user') : $t('common.users')}
        {$t('common.total')}
      {:else}
        {users.length} {$t('common.shown')} • {total} {$t('common.total')}
      {/if}
    </div>
    <div class="flex items-center gap-2">
      <span class="text-sm text-muted-foreground">{$t('common.team')}</span>
      <select
        class="h-9 rounded-md border bg-background px-3 text-sm"
        bind:value={selectedTeamId}
        disabled={isLoading || teamFilters.length === 0}
        onchange={() => void loadDirectory(1, searchText, selectedTeamId)}
      >
        <option value="all">{$t('common.all_teams')}</option>
        {#each teamFilters as t (t.id)}
          <option value={t.id}>{t.name}</option>
        {/each}
      </select>
    </div>
  </div>

  {#if error}
    <div class="rounded-md border bg-muted px-4 py-3 text-muted-foreground">
      <p class="text-sm">{error}</p>
    </div>
  {/if}

  <div class="overflow-hidden rounded-lg border bg-background">
    <Table.Root>
      <Table.Header>
        <Table.Row>
          <Table.Head>{$t('common.name_email')}</Table.Head>
          <Table.Head>{$t('common.team')}</Table.Head>
          <Table.Head>{$t('common.role')}</Table.Head>
          <Table.Head>{$t('common.auth')}</Table.Head>
          <Table.Head>{$t('common.status')}</Table.Head>
          <Table.Head>{$t('common.last_active')}</Table.Head>
          <Table.Head class="text-right">{$t('common.actions')}</Table.Head>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {#if isLoading && users.length === 0}
          {#each Array(5) as _, rowIndex (rowIndex)}
            <Table.Row>
              {#each Array(7) as _, colIndex (colIndex)}
                <Table.Cell><div class="h-8 w-full rounded bg-muted/40"></div></Table.Cell>
              {/each}
            </Table.Row>
          {/each}
        {:else if users.length === 0}
          <Table.Row>
            <Table.Cell colspan={7} class="h-24 text-center text-muted-foreground">
              {$t('messages.no_users_found')}
            </Table.Cell>
          </Table.Row>
        {:else}
          {#each users as user (user.id)}
            <Table.Row>
              <Table.Cell>
                <div class="flex items-center gap-3">
                  <UserAvatar firstName={user.first_name} lastName={user.last_name} />
                  <div class="flex flex-col">
                    <div class="font-medium">
                      {user.first_name}
                      {user.last_name}
                    </div>
                    <div class="text-sm text-muted-foreground">{user.email}</div>
                  </div>
                </div>
              </Table.Cell>
              <Table.Cell>
                {@const tnames = user.teams.map((team) => team.name)}
                {#if tnames.length === 0}
                  <span class="text-sm text-muted-foreground">&mdash;</span>
                {:else}
                  <div class="flex items-center gap-2">
                    <span class="text-sm font-medium">{tnames[0]}</span>
                    {#if tnames.length > 1}
                      <Tooltip.Root>
                        <Tooltip.Trigger class="inline-flex">
                          <Badge variant="outline">+{tnames.length - 1}</Badge>
                        </Tooltip.Trigger>
                        <Tooltip.Content class="max-w-xs">
                          <div class="text-sm">{tnames.join(', ')}</div>
                        </Tooltip.Content>
                      </Tooltip.Root>
                    {/if}
                  </div>
                {/if}
              </Table.Cell>
              <Table.Cell>
                <RoleBadge role={user.role} label={user.role_display_name ?? ''} />
              </Table.Cell>
              <Table.Cell>
                <div class="flex items-center gap-2">
                  <Tooltip.Root>
                    <Tooltip.Trigger class="inline-flex">
                      {#if authVerified(user)}
                        <Badge variant="success">
                          <BadgeCheck class="mr-1 h-3 w-3" />
                          {$t('common.verified')}
                        </Badge>
                      {:else}
                        <Badge variant="outline">
                          <BadgeX class="mr-1 h-3 w-3" />
                          {$t('common.unverified')}
                        </Badge>
                      {/if}
                    </Tooltip.Trigger>
                    <Tooltip.Content>
                      <div class="text-sm">{$t('messages.email_verification_info')}</div>
                    </Tooltip.Content>
                  </Tooltip.Root>

                  <Tooltip.Root>
                    <Tooltip.Trigger class="inline-flex">
                      {#if twoFactorEnabled(user)}
                        <Badge variant="secondary">
                          <KeyRound class="mr-1 h-3 w-3" />
                          {$t('common.2fa')}
                        </Badge>
                      {:else}
                        <Badge variant="outline">
                          <KeyRound class="mr-1 h-3 w-3" />
                          {$t('common.2fa_off')}
                        </Badge>
                      {/if}
                    </Tooltip.Trigger>
                    <Tooltip.Content>
                      <div class="text-sm">{$t('messages.2fa_not_implemented')}</div>
                    </Tooltip.Content>
                  </Tooltip.Root>
                </div>
              </Table.Cell>
              <Table.Cell>
                {#if user.disabled_at}
                  <Badge variant="destructive">{$t('common.disabled')}</Badge>
                {:else if user.locked_until}
                  <Badge variant="warning">{$t('common.locked')}</Badge>
                {:else if user.is_active}
                  <Badge variant="success">{$t('common.active')}</Badge>
                {:else}
                  <Badge variant="outline">{$t('common.inactive')}</Badge>
                {/if}
              </Table.Cell>
              <Table.Cell>
                <span class="text-sm">{formatDate(user.last_login_at)}</span>
              </Table.Cell>
              <Table.Cell class="text-right">
                {#if user.capabilities.can_change_role || user.capabilities.can_disable || user.capabilities.can_enable || user.capabilities.can_delete}
                  <DropdownMenu.Root>
                    <DropdownMenu.Trigger>
                      {#snippet child({ props })}
                        <Button variant="ghost" size="sm" {...props}>
                          <MoreVertical class="h-4 w-4" />
                        </Button>
                      {/snippet}
                    </DropdownMenu.Trigger>
                    <DropdownMenu.Content align="end" class="w-56">
                      {#if user.capabilities.can_change_role}
                        <DropdownMenu.Label>{$t('common.change_role')}</DropdownMenu.Label>
                        <DropdownMenu.Separator />
                        {#each roleOptionsFor(user) as roleObj (roleObj.role)}
                          <DropdownMenu.Item
                            onclick={() => handleRoleChange(user.id, roleObj.role)}
                          >
                            {roleObj.display_name}
                          </DropdownMenu.Item>
                        {/each}
                      {/if}

                      {#if user.capabilities.can_disable || user.capabilities.can_enable}
                        <DropdownMenu.Separator />
                        <DropdownMenu.Item
                          onclick={() => handleToggleActive(user.id, user.is_active)}
                        >
                          {#if user.is_active}
                            <UserMinus class="mr-2 h-4 w-4" />
                            {$t('actions.disable_user')}
                          {:else}
                            <UserCheck class="mr-2 h-4 w-4" />
                            {$t('actions.enable_user')}
                          {/if}
                        </DropdownMenu.Item>
                      {/if}

                      {#if user.capabilities.can_delete}
                        <DropdownMenu.Separator />
                        <DropdownMenu.Item
                          class="text-destructive"
                          onclick={() =>
                            handleDeleteUser(user.id, `${user.first_name} ${user.last_name}`)}
                        >
                          <Trash2 class="mr-2 h-4 w-4" />
                          {$t('actions.delete_user')}
                        </DropdownMenu.Item>
                      {/if}
                    </DropdownMenu.Content>
                  </DropdownMenu.Root>
                {/if}
              </Table.Cell>
            </Table.Row>
          {/each}
        {/if}
      </Table.Body>
    </Table.Root>
  </div>

  {#if totalPages > 1}
    <div class="flex items-center justify-between">
      <div class="text-sm text-muted-foreground">
        {$t('messages.page_of')
          .replace('{page}', String(page))
          .replace('{total}', String(totalPages))}
        • {$t('messages.total_items').replace('{count}', String(total))}
      </div>
      <div class="flex items-center gap-2">
        <Button
          variant="outline"
          size="sm"
          disabled={page <= 1 || isLoading}
          onclick={() => void loadDirectory(page - 1, searchText, selectedTeamId)}
        >
          {$t('messages.previous')}
        </Button>
        <Button
          variant="outline"
          size="sm"
          disabled={page >= totalPages || isLoading}
          onclick={() => void loadDirectory(page + 1, searchText, selectedTeamId)}
        >
          {$t('messages.next')}
        </Button>
      </div>
    </div>
  {/if}
</div>

<Dialog.Root bind:open={createDialogOpen}>
  <Dialog.Content class="sm:max-w-lg">
    <Dialog.Header>
      <Dialog.Title>{$t('common.create_user')}</Dialog.Title>
      <Dialog.Description>{$t('messages.add_new_user')}</Dialog.Description>
    </Dialog.Header>
    <UserManagementForm
      onSuccess={() => {
        createDialogOpen = false;
        void loadDirectory();
        addToast($t('messages.user_created_success'), 'success');
      }}
      onCancel={() => (createDialogOpen = false)}
    />
  </Dialog.Content>
</Dialog.Root>

<ConfirmDialog />
