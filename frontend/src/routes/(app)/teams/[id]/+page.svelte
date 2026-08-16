<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import * as Dialog from '$lib/components/ui/dialog/index.js';
  import StaticCombobox from '$lib/components/ui/combobox/StaticCombobox.svelte';
  import * as Table from '$lib/components/ui/table/index.js';
  import * as Popover from '$lib/components/ui/popover/index.js';
  import * as Command from '$lib/components/ui/command/index.js';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import UserAvatar from '$lib/components/user-avatar.svelte';
  import { Pencil, UserMinus, UserPlus } from '@lucide/svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { TeamDetailPageState } from '$lib/components/teams/TeamDetailPageState.svelte.js';

  const teamId = $derived($page.params.id ?? '');
  const state = new TeamDetailPageState(() => teamId);

  const t = createTranslator();

  type TeamRole = 'member' | 'manager' | 'owner';

  const roleOptions = $derived([
    { id: 'member' satisfies TeamRole, label: $t('teams.roles.member') },
    { id: 'manager' satisfies TeamRole, label: $t('teams.roles.manager') },
    { id: 'owner' satisfies TeamRole, label: $t('teams.roles.owner') }
  ]);

  $effect(() => {
    state.scheduleUserSearch();
  });

  $effect(() => {
    if (state.addMemberOpen) {
      void state.searchUsers('');
    }
  });

  onMount(() => {
    void state.load();
  });
</script>

<ConfirmDialog />

<div class="flex flex-col gap-6">
  <EntityListHeader
    title={state.team?.name ?? $t('team.team')}
    description={$t('teams.detail.description')}
    backHref="/teams"
    backLabel={$t('common.back')}
  >
    {#if state.canUpdateTeam()}
      <Button
        variant="outline"
        size="icon"
        aria-label={$t('common.edit')}
        onclick={() => state.openEdit()}
      >
        <Pencil class="h-4 w-4" />
      </Button>
    {/if}

    {#if state.canUpdateTeam()}
      <Popover.Root bind:open={state.addMemberOpen}>
        <Popover.Trigger>
          {#snippet child({ props })}
            <Button
              {...props}
              size="icon"
              class="bg-primary text-primary-foreground shadow-xs hover:bg-primary/90"
              aria-label={$t('teams.detail.add_member')}
            >
              <UserPlus class="h-4 w-4" />
            </Button>
          {/snippet}
        </Popover.Trigger>
        <Popover.Content class="w-72 p-0" align="end">
          <Command.Root shouldFilter={false}>
            <Command.Input
              placeholder={$t('teams.detail.search_users_placeholder')}
              bind:value={state.addMemberSearch}
            />
            <Command.List>
              <Command.Empty>
                {state.addMemberLoading
                  ? $t('teams.detail.searching')
                  : $t('teams.detail.no_users_found')}
              </Command.Empty>
              <Command.Group>
                {#each state.addMemberResults as user (user.id)}
                  <Command.Item value={user.id} onSelect={() => state.handleAddMember(user.id)}>
                    <div class="flex items-center gap-2">
                      <UserAvatar
                        firstName={user.first_name}
                        lastName={user.last_name}
                        class="h-6 w-6"
                      />
                      <div class="flex flex-col">
                        <span class="text-sm">{user.first_name} {user.last_name}</span>
                        <span class="text-xs text-muted-foreground">{user.email}</span>
                      </div>
                    </div>
                  </Command.Item>
                {/each}
              </Command.Group>
            </Command.List>
          </Command.Root>
        </Popover.Content>
      </Popover.Root>
    {/if}
  </EntityListHeader>

  {#if state.team?.description}
    <div class="text-sm text-muted-foreground">{state.team.description}</div>
  {/if}

  {#if state.error}
    <div class="rounded-md border bg-muted px-4 py-3 text-muted-foreground">
      <p class="font-medium">{$t('teams.errors.load_title')}</p>
      <p class="text-sm">{state.error}</p>
    </div>
  {/if}

  <div class="rounded-lg border bg-background">
    <Table.Root>
      <Table.Header>
        <Table.Row>
          <Table.Head>{$t('common.user')}</Table.Head>
          <Table.Head>{$t('common.role')}</Table.Head>
          <Table.Head class="w-30"></Table.Head>
        </Table.Row>
      </Table.Header>
      <Table.Body>
        {#if state.loading && state.members.length === 0}
          <Table.LoadingRows loading rowCount={6}>
            {#snippet children(_rowIndex)}
              <Table.Cell><div class="h-4 w-70 rounded-md bg-muted/40"></div></Table.Cell>
              <Table.Cell><div class="h-4 w-30 rounded-md bg-muted/40"></div></Table.Cell>
              <Table.Cell><div class="h-8 w-24 rounded-md bg-muted/40"></div></Table.Cell>
            {/snippet}
          </Table.LoadingRows>
        {:else if state.members.length === 0}
          <Table.Row>
            <Table.Cell colspan={3}>
              <div class="flex flex-col items-center justify-center gap-2 py-10 text-center">
                <div class="text-sm font-medium">{$t('teams.detail.empty_title')}</div>
                <p class="text-sm text-muted-foreground">
                  {$t('teams.detail.empty_description')}
                </p>
              </div>
            </Table.Cell>
          </Table.Row>
        {:else}
          {#each state.members as m (m.user_id)}
            <Table.Row>
              <Table.Cell>
                {#if state.userById(m.user_id)}
                  {@const u = state.userById(m.user_id)!}
                  <div class="flex items-center gap-3">
                    <UserAvatar firstName={u.first_name} lastName={u.last_name} />
                    <div class="flex flex-col">
                      <div class="font-medium">
                        {u.first_name}
                        {u.last_name}
                      </div>
                      <div class="text-sm text-muted-foreground">{u.email}</div>
                    </div>
                  </div>
                {:else}
                  <div class="font-medium">{m.user_id}</div>
                {/if}
              </Table.Cell>
              <Table.Cell>
                {#if state.canUpdateTeam()}
                  <StaticCombobox
                    items={roleOptions}
                    value={m.role}
                    labelKey="label"
                    width="h-8 w-32 px-2"
                    popupWidth="w-40"
                    searchPlaceholder={$t('teams.detail.search_roles_placeholder')}
                    emptyText={$t('teams.detail.no_roles_found')}
                    disabled={state.busy}
                    onValueChange={(role) => {
                      if (role === m.role) return;
                      void state.changeRole(m.user_id, role as TeamRole);
                    }}
                  />
                {:else}
                  <span class="text-sm text-muted-foreground">{m.role}</span>
                {/if}
              </Table.Cell>
              <Table.Cell class="text-right">
                {#if state.canUpdateTeam()}
                  <Button
                    variant="outline"
                    onclick={() => state.remove(m.user_id)}
                    disabled={state.busy}
                  >
                    <UserMinus class="mr-2 h-4 w-4" />
                    {$t('teams.detail.remove_member')}
                  </Button>
                {/if}
              </Table.Cell>
            </Table.Row>
          {/each}
        {/if}
      </Table.Body>
    </Table.Root>
  </div>
</div>

<Dialog.Root bind:open={state.editOpen}>
  <Dialog.Content class="sm:max-w-xl">
    <Dialog.Header>
      <Dialog.Title>{$t('common.edit')} {$t('team.team')}</Dialog.Title>
      <Dialog.Description>{$t('teams.detail.description')}</Dialog.Description>
    </Dialog.Header>

    <form
      class="grid gap-4"
      onsubmit={(event) => {
        event.preventDefault();
        void state.submitEdit();
      }}
    >
      <div class="grid gap-2">
        <label class="text-sm font-medium" for="team_edit_name">{$t('common.name')}</label>
        <Input id="team_edit_name" bind:value={state.editName} disabled={state.editBusy} />
      </div>
      <div class="grid gap-2">
        <label class="text-sm font-medium" for="team_edit_description">
          {$t('common.description')}
        </label>
        <Input
          id="team_edit_description"
          bind:value={state.editDescription}
          disabled={state.editBusy}
        />
      </div>
      <Dialog.Footer>
        <Button
          type="button"
          variant="outline"
          onclick={() => (state.editOpen = false)}
          disabled={state.editBusy}
        >
          {$t('common.cancel')}
        </Button>
        <Button type="submit" disabled={!state.canSubmitEdit()}>
          {state.editBusy ? $t('common.saving') : $t('common.save_changes')}
        </Button>
      </Dialog.Footer>
    </form>
  </Dialog.Content>
</Dialog.Root>
