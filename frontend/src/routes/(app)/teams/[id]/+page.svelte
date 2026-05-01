<script lang="ts">
  import { onMount } from 'svelte';
  import { page } from '$app/stores';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import * as Popover from '$lib/components/ui/popover/index.js';
  import * as Command from '$lib/components/ui/command/index.js';
  import { Skeleton } from '$lib/components/ui/skeleton/index.js';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import UserAvatar from '$lib/components/user-avatar.svelte';
  import { UserMinus, UserPlus } from '@lucide/svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { TeamDetailPageState } from '$lib/components/teams/TeamDetailPageState.svelte.js';

  const teamId = $derived($page.params.id ?? '');
  const state = new TeamDetailPageState(() => teamId);

  const t = createTranslator();

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
    <Popover.Root bind:open={state.addMemberOpen}>
      <Popover.Trigger>
        {#snippet child({ props })}
          <Button
            {...props}
            size="icon"
            class="bg-blue-600 text-white shadow-xs hover:bg-blue-700"
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
        {#if state.loading}
          {#each Array(6) as _}
            <Table.Row>
              <Table.Cell><Skeleton class="h-4 w-70" /></Table.Cell>
              <Table.Cell><Skeleton class="h-4 w-30" /></Table.Cell>
              <Table.Cell><Skeleton class="h-8 w-24" /></Table.Cell>
            </Table.Row>
          {/each}
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
                <select
                  class="flex h-8 rounded-md border border-input bg-transparent px-2 text-sm shadow-sm"
                  onchange={(e) =>
                    state.changeRole(m.user_id, (e.target as HTMLSelectElement).value as any)}
                  disabled={state.busy}
                >
                  <option value="member" selected={m.role === 'member'}
                    >{$t('teams.roles.member')}</option
                  >
                  <option value="manager" selected={m.role === 'manager'}
                    >{$t('teams.roles.manager')}</option
                  >
                  <option value="owner" selected={m.role === 'owner'}
                    >{$t('teams.roles.owner')}</option
                  >
                </select>
              </Table.Cell>
              <Table.Cell class="text-right">
                <Button
                  variant="outline"
                  onclick={() => state.remove(m.user_id)}
                  disabled={state.busy}
                >
                  <UserMinus class="mr-2 h-4 w-4" />
                  {$t('teams.detail.remove_member')}
                </Button>
              </Table.Cell>
            </Table.Row>
          {/each}
        {/if}
      </Table.Body>
    </Table.Root>
  </div>
</div>
