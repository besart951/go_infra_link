<script lang="ts">
  import { onMount } from 'svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import * as DropdownMenu from '$lib/components/ui/dropdown-menu/index.js';
  import * as Dialog from '$lib/components/ui/dialog/index.js';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import RoleBadge from '$lib/components/role-badge.svelte';
  import UserAvatar from '$lib/components/user-avatar.svelte';
  import UserManagementForm from '$lib/components/user-management-form.svelte';
  import RegistrationProcessStepper from '$lib/components/users/RegistrationProcessStepper.svelte';
  import { UserDirectoryPageState } from '$lib/components/users/UserDirectoryPageState.svelte.js';
  import { MoreVertical, UserMinus, UserCheck, Trash2, Send, RotateCcw } from '@lucide/svelte';
  import { createTranslator } from '$lib/i18n/translator';

  interface Props {
    data?: {
      user?: {
        can_access_user_directory?: boolean | null;
      } | null;
    };
  }

  let { data }: Props = $props();

  const t = createTranslator();
  const state = new UserDirectoryPageState();

  onMount(() => {
    void state.initialize(data?.user ? Boolean(data.user.can_access_user_directory) : undefined);
    return state.startResendClock();
  });
</script>

<svelte:head>
  <title>{$t('navigation.users')} | Infra Link</title>
</svelte:head>

<Tooltip.Provider>
  <div class="flex flex-col gap-6">
    <EntityListHeader
      title={$t('pages.user_management')}
      description={$t('pages.user_management_desc')}
      backHref="/users"
      backLabel={$t('common.back')}
      createLabel={$t('common.create_user')}
      canCreate={state.pageCapabilities.can_create_user}
      createActive={state.createDialogOpen}
      onCreateClick={() => (state.createDialogOpen = true)}
    />

    <div class="flex flex-wrap items-center justify-between gap-3">
      <div class="flex flex-1 items-center gap-3">
        <Input
          class="min-w-55 flex-1"
          bind:value={state.searchText}
          placeholder={$t('messages.search_users')}
          onkeydown={(event) => {
            if (event.key === 'Enter') {
              void state.loadDirectory(1, state.searchText, state.selectedTeamId);
            }
          }}
        />
        <Button
          variant="outline"
          onclick={() => void state.loadDirectory(1, state.searchText, state.selectedTeamId)}
          disabled={state.isLoading}
        >
          {$t('messages.refresh')}
        </Button>
      </div>
      <div class="text-sm text-muted-foreground">
        {#if state.selectedTeamId === 'all'}
          {state.total}
          {state.total === 1 ? $t('common.user') : $t('common.users')}
          {$t('common.total')}
        {:else}
          {state.users.length} {$t('common.shown')} • {state.total} {$t('common.total')}
        {/if}
      </div>
      {#if state.pageCapabilities.can_read_deleted}
        <div class="flex items-center gap-2">
          <label class="inline-flex items-center gap-2 text-sm text-muted-foreground">
            <input
              type="checkbox"
              bind:checked={state.showDeletedUsers}
              disabled={state.isLoading}
              onchange={() => void state.loadDirectory(1, state.searchText, state.selectedTeamId)}
            />
            {$t('common.show_deleted_users')}
          </label>
        </div>
      {/if}
      <div class="flex items-center gap-2">
        <span class="text-sm text-muted-foreground">{$t('common.team')}</span>
        <select
          class="h-9 rounded-md border bg-background px-3 text-sm"
          bind:value={state.selectedTeamId}
          disabled={state.isLoading || state.teamFilters.length === 0}
          onchange={() => void state.loadDirectory(1, state.searchText, state.selectedTeamId)}
        >
          <option value="all">{$t('common.all_teams')}</option>
          {#each state.teamFilters as teamFilter (teamFilter.id)}
            <option value={teamFilter.id}>{teamFilter.name}</option>
          {/each}
        </select>
      </div>
    </div>

    {#if state.error}
      <div class="rounded-md border bg-muted px-4 py-3 text-muted-foreground">
        <p class="text-sm">{state.error}</p>
      </div>
    {/if}

    <div class="overflow-hidden rounded-lg border bg-background">
      <Table.Root>
        <Table.Header>
          <Table.Row>
            <Table.Head>{$t('common.name_email')}</Table.Head>
            <Table.Head>{$t('common.team')}</Table.Head>
            <Table.Head>{$t('common.role')}</Table.Head>
            <Table.Head class="min-w-56 sm:min-w-76">
              {$t('user.registration_column')}
            </Table.Head>
            <Table.Head>{$t('common.status')}</Table.Head>
            <Table.Head>{$t('common.last_active')}</Table.Head>
            <Table.Head class="text-right">{$t('common.actions')}</Table.Head>
          </Table.Row>
        </Table.Header>
        <Table.Body>
          {#if state.isLoading && state.users.length === 0}
            <Table.LoadingRows loading rowCount={5}>
              {#snippet children(_rowIndex)}
                {#each Array(7) as _, colIndex (colIndex)}
                  <Table.Cell><div class="h-8 w-full rounded-md bg-muted/40"></div></Table.Cell>
                {/each}
              {/snippet}
            </Table.LoadingRows>
          {:else if state.users.length === 0}
            <Table.Row>
              <Table.Cell colspan={7} class="h-24 text-center text-muted-foreground">
                {$t('messages.no_users_found')}
              </Table.Cell>
            </Table.Row>
          {:else}
            {#each state.users as user (user.id)}
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
                <Table.Cell class="min-w-56 sm:min-w-76">
                  <RegistrationProcessStepper process={user.registration_process} />
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
                  <span class="text-sm">{state.formatDate(user.last_login_at)}</span>
                </Table.Cell>
                <Table.Cell class="text-right">
                  {@const hasRoleAction = user.capabilities.can_change_role}
                  {@const hasToggleAction =
                    user.capabilities.can_disable || user.capabilities.can_enable}
                  {@const hasInvitationAction = state.hasInvitationResendAction(user)}
                  {@const hasDeleteAction = user.capabilities.can_delete}
                  {@const hasRestoreAction = user.capabilities.can_restore}
                  {#if hasRoleAction || hasToggleAction || hasInvitationAction || hasDeleteAction || hasRestoreAction}
                    <DropdownMenu.Root>
                      <DropdownMenu.Trigger>
                        {#snippet child({ props })}
                          <Button
                            variant="ghost"
                            size="sm"
                            aria-label={$t('common.actions')}
                            {...props}
                          >
                            <MoreVertical class="h-4 w-4" />
                          </Button>
                        {/snippet}
                      </DropdownMenu.Trigger>
                      <DropdownMenu.Content align="end" class="w-56">
                        {#if hasRoleAction}
                          <DropdownMenu.Label>{$t('common.change_role')}</DropdownMenu.Label>
                          <DropdownMenu.Separator />
                          {#each state.roleOptionsFor(user) as roleObj (roleObj.role)}
                            <DropdownMenu.Item
                              onclick={() => state.handleRoleChange(user.id, roleObj.role)}
                            >
                              {roleObj.display_name}
                            </DropdownMenu.Item>
                          {/each}
                        {/if}

                        {#if hasToggleAction}
                          {#if hasRoleAction}
                            <DropdownMenu.Separator />
                          {/if}
                          <DropdownMenu.Item
                            onclick={() => state.handleToggleActive(user.id, user.is_active)}
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

                        {#if hasInvitationAction}
                          {#if hasRoleAction || hasToggleAction}
                            <DropdownMenu.Separator />
                          {/if}
                          {@const canResendInvitation = state.canResendInvitation(
                            user,
                            state.resendClock
                          )}
                          {@const resendDisabledReason = state.invitationResendDisabledReason(
                            user,
                            state.resendClock
                          )}
                          {#if canResendInvitation}
                            <DropdownMenu.Item
                              onclick={() => state.handleResendInvitation(user.id)}
                            >
                              <Send class="mr-2 h-4 w-4" />
                              {$t('user.resend_invitation')}
                            </DropdownMenu.Item>
                          {:else}
                            <Tooltip.Root>
                              <Tooltip.Trigger class="block w-full">
                                <DropdownMenu.Item
                                  disabled
                                  class="w-full cursor-not-allowed"
                                  title={resendDisabledReason ?? undefined}
                                  aria-label={`${$t('user.resend_invitation')}: ${resendDisabledReason ?? ''}`}
                                >
                                  <Send class="mr-2 h-4 w-4" />
                                  {$t('user.resend_invitation')}
                                </DropdownMenu.Item>
                              </Tooltip.Trigger>
                              {#if resendDisabledReason}
                                <Tooltip.Content side="left" class="max-w-xs text-sm">
                                  {resendDisabledReason}
                                </Tooltip.Content>
                              {/if}
                            </Tooltip.Root>
                          {/if}
                        {/if}

                        {#if hasRestoreAction}
                          {#if hasRoleAction || hasToggleAction || hasInvitationAction}
                            <DropdownMenu.Separator />
                          {/if}
                          <DropdownMenu.Item
                            onclick={() =>
                              state.handleRestoreUser(
                                user.id,
                                `${user.first_name} ${user.last_name}`
                              )}
                          >
                            <RotateCcw class="mr-2 h-4 w-4" />
                            {$t('actions.restore_user')}
                          </DropdownMenu.Item>
                        {/if}

                        {#if hasDeleteAction && !user.is_deleted}
                          {#if hasRoleAction || hasToggleAction || hasInvitationAction}
                            <DropdownMenu.Separator />
                          {/if}
                          <DropdownMenu.Item
                            class="text-destructive"
                            onclick={() =>
                              state.handleDeleteUser(
                                user.id,
                                `${user.first_name} ${user.last_name}`
                              )}
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

    {#if state.totalPages > 1}
      <div class="flex items-center justify-between">
        <div class="text-sm text-muted-foreground">
          {$t('messages.page_of')
            .replace('{page}', String(state.page))
            .replace('{total}', String(state.totalPages))}
          • {$t('messages.total_items').replace('{count}', String(state.total))}
        </div>
        <div class="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            disabled={state.page <= 1 || state.isLoading}
            onclick={() =>
              void state.loadDirectory(state.page - 1, state.searchText, state.selectedTeamId)}
          >
            {$t('messages.previous')}
          </Button>
          <Button
            variant="outline"
            size="sm"
            disabled={state.page >= state.totalPages || state.isLoading}
            onclick={() =>
              void state.loadDirectory(state.page + 1, state.searchText, state.selectedTeamId)}
          >
            {$t('messages.next')}
          </Button>
        </div>
      </div>
    {/if}
  </div>

  <Dialog.Root bind:open={state.createDialogOpen}>
    <Dialog.Content class="sm:max-w-lg">
      <Dialog.Header>
        <Dialog.Title>{$t('common.create_user')}</Dialog.Title>
        <Dialog.Description>{$t('user.invitation_create_description')}</Dialog.Description>
      </Dialog.Header>
      <UserManagementForm
        onSuccess={() => {
          void state.handleUserCreated();
        }}
        onCancel={() => (state.createDialogOpen = false)}
      />
    </Dialog.Content>
  </Dialog.Root>

  <ConfirmDialog />
</Tooltip.Provider>
