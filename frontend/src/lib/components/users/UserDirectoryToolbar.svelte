<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import { Checkbox } from '$lib/components/ui/checkbox/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import type { UserDirectoryPageState } from './UserDirectoryPageState.svelte.js';

  type Props = {
    state: UserDirectoryPageState;
  };

  let { state }: Props = $props();
  const t = createTranslator();

  function handleSearchKeydown(event: KeyboardEvent): void {
    if (event.key === 'Enter') {
      void state.refreshDirectory();
    }
  }

  function handleDeletedUsersChange(checked: boolean | 'indeterminate'): void {
    void state.setShowDeletedUsers(checked === true);
  }

  function handleTeamChange(event: Event): void {
    const target = event.currentTarget;
    if (!(target instanceof HTMLSelectElement)) return;

    void state.setTeamFilter(target.value);
  }

  function handleRoleChange(event: Event): void {
    const target = event.currentTarget;
    if (!(target instanceof HTMLSelectElement)) return;

    void state.setRoleFilter(
      target.value as Parameters<UserDirectoryPageState['setRoleFilter']>[0]
    );
  }
</script>

<section class="rounded-xl border bg-card p-4 shadow-sm">
  <div class="flex flex-col gap-4 xl:flex-row xl:items-center xl:justify-between">
    <div class="flex min-w-0 flex-1 flex-col gap-3 sm:flex-row sm:items-center">
      <Input
        class="min-w-0 flex-1"
        bind:value={state.query.searchText}
        placeholder={$t('messages.search_users')}
        aria-label={$t('messages.search_users')}
        onkeydown={handleSearchKeydown}
      />
      <Button
        class="w-full sm:w-auto"
        variant="outline"
        onclick={() => void state.refreshDirectory()}
        disabled={state.query.loading}
      >
        {$t('messages.refresh')}
      </Button>
    </div>

    <div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between xl:justify-end">
      <div class="text-sm text-muted-foreground">
        {#if state.selectedTeamId === 'all'}
          {state.query.total}
          {state.query.total === 1 ? $t('common.user') : $t('common.users')}
          {$t('common.total')}
        {:else}
          {state.users.length} {$t('common.shown')} • {state.query.total} {$t('common.total')}
        {/if}
      </div>

      {#if state.pageCapabilities.can_read_deleted}
        <Label class="inline-flex items-center gap-2 text-muted-foreground">
          <Checkbox
            checked={state.showDeletedUsers}
            disabled={state.query.loading}
            onCheckedChange={handleDeletedUsersChange}
          />
          {$t('common.show_deleted_users')}
        </Label>
      {/if}

      <div class="flex flex-col gap-2 sm:flex-row sm:items-center">
        <Label for="user_team_filter" class="text-muted-foreground">{$t('common.team')}</Label>
        <select
          id="user_team_filter"
          class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm font-medium shadow-xs transition-[color,box-shadow] outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50 sm:w-55"
          value={state.selectedTeamId}
          disabled={state.query.loading}
          onchange={handleTeamChange}
        >
          <option value="all">{$t('common.all_teams')}</option>
          {#each state.teamFilters as teamFilter (teamFilter.id)}
            <option value={teamFilter.id}>{teamFilter.name}</option>
          {/each}
        </select>
      </div>

      <div class="flex flex-col gap-2 sm:flex-row sm:items-center">
        <Label for="user_role_filter" class="text-muted-foreground">{$t('common.role')}</Label>
        <select
          id="user_role_filter"
          class="h-9 w-full rounded-md border border-input bg-background px-3 text-sm font-medium shadow-xs transition-[color,box-shadow] outline-none focus-visible:border-ring focus-visible:ring-[3px] focus-visible:ring-ring/50 disabled:cursor-not-allowed disabled:opacity-50 sm:w-55"
          value={state.selectedRole}
          disabled={state.query.loading}
          onchange={handleRoleChange}
        >
          <option value="all">{$t('common.all_roles')}</option>
          {#each state.roleFilters as roleFilter (roleFilter.role)}
            <option value={roleFilter.role}>{roleFilter.display_name}</option>
          {/each}
        </select>
      </div>
    </div>
  </div>
</section>
