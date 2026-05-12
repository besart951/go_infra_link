<script lang="ts">
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import RoleBadge from '$lib/components/role-badge.svelte';
  import UserAvatar from '$lib/components/user-avatar.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import RegistrationProcessStepper from './RegistrationProcessStepper.svelte';
  import type { UserDirectoryPageState } from './UserDirectoryPageState.svelte.js';
  import UserDirectoryRowActions from './UserDirectoryRowActions.svelte';
  import UserStatusBadge from './UserStatusBadge.svelte';

  type Props = {
    state: UserDirectoryPageState;
  };

  let { state }: Props = $props();
  const t = createTranslator();
</script>

{#if state.query.error}
  <div
    role="alert"
    class="rounded-xl border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive"
  >
    <p>{state.query.error}</p>
  </div>
{/if}

<div class="overflow-hidden rounded-xl border bg-card shadow-sm">
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
      {#if state.query.loading && state.users.length === 0}
        <Table.LoadingRows loading rowCount={5}>
          {#snippet children(_rowIndex)}
            {#each Array(7) as _, colIndex (colIndex)}
              <Table.Cell><div class="h-8 w-full rounded-md bg-muted/40"></div></Table.Cell>
            {/each}
          {/snippet}
        </Table.LoadingRows>
      {:else if state.users.length === 0}
        <Table.Row>
          <Table.Cell colspan={7} class="h-32 text-center">
            <div class="flex flex-col items-center justify-center gap-2 text-muted-foreground">
              <p class="font-medium">{$t('messages.no_users_found')}</p>
              {#if state.query.searchText}
                <p class="text-sm">{$t('messages.try_adjusting_search')}</p>
              {/if}
            </div>
          </Table.Cell>
        </Table.Row>
      {:else}
        {#each state.users as user (user.id)}
          <Table.Row class={state.query.loading ? 'opacity-60' : undefined}>
            <Table.Cell>
              <div class="flex items-center gap-3">
                <UserAvatar firstName={user.first_name} lastName={user.last_name} />
                <div class="flex min-w-0 flex-col">
                  <div class="font-medium">
                    {user.first_name}
                    {user.last_name}
                  </div>
                  <div class="truncate text-sm text-muted-foreground">
                    {#if user.is_deleted && !state.pageCapabilities.can_read_deleted}
                      -
                    {:else if user.email}
                      {user.email}
                    {:else}
                      -
                    {/if}
                  </div>
                </div>
              </div>
            </Table.Cell>
            <Table.Cell>
              {@const teamNames = user.teams.map((team) => team.name)}
              {#if teamNames.length === 0}
                <span class="text-sm text-muted-foreground">-</span>
              {:else}
                <div class="flex items-center gap-2">
                  <span class="text-sm font-medium">{teamNames[0]}</span>
                  {#if teamNames.length > 1}
                    <Tooltip.Root>
                      <Tooltip.Trigger class="inline-flex">
                        <Badge variant="outline">+{teamNames.length - 1}</Badge>
                      </Tooltip.Trigger>
                      <Tooltip.Content class="max-w-xs">
                        <div class="text-sm">{teamNames.join(', ')}</div>
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
              <UserStatusBadge {user} />
            </Table.Cell>
            <Table.Cell>
              <span class="text-sm">{state.formatDate(user.last_login_at)}</span>
            </Table.Cell>
            <Table.Cell class="text-right">
              <UserDirectoryRowActions {state} {user} />
            </Table.Cell>
          </Table.Row>
        {/each}
      {/if}
    </Table.Body>
  </Table.Root>
</div>
