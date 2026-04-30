<script lang="ts">
  import { onMount } from 'svelte';
  import { goto } from '$app/navigation';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import * as Table from '$lib/components/ui/table/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import { ArrowLeft, Plus, Trash2 } from '@lucide/svelte';
  import PaginatedList from '$lib/components/list/PaginatedList.svelte';
  import { teamsStore } from '$lib/stores/list/entityStores.js';
  import type { Team } from '$lib/domain/entities/team.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator';
  import { TeamListPageState } from '$lib/components/teams/TeamListPageState.svelte.js';

  const t = createTranslator();
  const state = new TeamListPageState();

  $effect(() => {
    const items = $teamsStore.items;
    if (items.length > 0) {
      void state.loadMemberCounts(items);
    }
  });

  onMount(() => {
    state.initialize();
  });
</script>

<svelte:head>
  <title>{$t('navigation.teams')} | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
  <div class="flex items-start justify-between gap-4">
    <div>
      <h1 class="text-3xl font-bold tracking-tight">{$t('navigation.teams')}</h1>
      <p class="mt-1 text-muted-foreground">{$t('pages.teams_desc')}</p>
    </div>
    <div class="flex flex-col gap-2 sm:flex-row">
      <Button variant="outline" href="/users">
        <ArrowLeft class="size-4" />
        {$t('hub.back_to_overview')}
      </Button>
      {#if canPerform('create', 'team')}
        <Button variant="outline" onclick={() => (state.createOpen = !state.createOpen)}>
          <Plus class="mr-2 h-4 w-4" />
          {$t('pages.create_team')}
        </Button>
      {/if}
    </div>
  </div>

  {#if state.createOpen}
    <div class="rounded-lg border bg-background p-4">
      <div class="grid gap-3 md:grid-cols-3">
        <div class="md:col-span-1">
          <label class="text-sm font-medium" for="team_name">{$t('common.name')}</label>
          <Input
            id="team_name"
            placeholder={$t('messages.team_name_placeholder')}
            bind:value={state.form.name}
            disabled={state.createBusy}
          />
        </div>
        <div class="md:col-span-2">
          <label class="text-sm font-medium" for="team_desc"
            >{$t('messages.team_description')}</label
          >
          <Input
            id="team_desc"
            placeholder={$t('pages.optional')}
            bind:value={state.form.description}
            disabled={state.createBusy}
          />
        </div>
      </div>
      <div class="mt-4 flex items-center justify-end gap-2">
        <Button
          variant="outline"
          onclick={() => (state.createOpen = false)}
          disabled={state.createBusy}>{$t('common.cancel')}</Button
        >
        <Button onclick={() => state.submitCreate()} disabled={!state.canSubmitCreate()}
          >{$t('common.create')}</Button
        >
      </div>
    </div>
  {/if}

  <PaginatedList
    state={$teamsStore}
    columns={[
      { key: 'name', label: $t('common.name') },
      { key: 'description', label: $t('common.description') },
      { key: 'members', label: $t('common.members'), width: 'w-24' },
      { key: 'actions', label: '', width: 'w-40' }
    ]}
    searchPlaceholder={$t('messages.search_teams')}
    emptyMessage={$t('messages.no_teams_found')}
    onSearch={(text) => teamsStore.search(text)}
    onPageChange={(page) => teamsStore.goToPage(page)}
    onReload={() => teamsStore.reload()}
  >
    {#snippet rowSnippet(team: Team)}
      <Table.Cell class="font-medium">{team.name}</Table.Cell>
      <Table.Cell class="text-muted-foreground">{team.description ?? ''}</Table.Cell>
      <Table.Cell>
        {@const count = state.memberCounts.get(team.id)}
        {#if count !== undefined}
          <Badge variant="secondary">{count}</Badge>
        {:else}
          <span class="text-sm text-muted-foreground">&mdash;</span>
        {/if}
      </Table.Cell>
      <Table.Cell class="text-right">
        <div class="flex items-center justify-end gap-2">
          <Button variant="outline" onclick={() => goto(`/teams/${team.id}`)}
            >{$t('common.manage')}</Button
          >
          {#if canPerform('delete', 'team')}
            <Button variant="outline" size="icon" onclick={() => state.handleDeleteTeam(team)}>
              <Trash2 class="h-4 w-4 text-destructive" />
            </Button>
          {/if}
        </div>
      </Table.Cell>
    {/snippet}
  </PaginatedList>
</div>

<ConfirmDialog />
