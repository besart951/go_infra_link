<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import { ActivityTimelineController } from '$lib/activity/activityTimelineController.svelte.js';
  import { globalActivityDataSource } from '$lib/activity/historyActivityDataSource.js';
  import ActivityFeed from '$lib/components/activity/ActivityFeed.svelte';
  import { addToast } from '$lib/components/toast.svelte';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import StaticCombobox from '$lib/components/ui/combobox/StaticCombobox.svelte';
  import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
  import HistoryFieldMultiSelect from '$lib/components/history/HistoryFieldMultiSelect.svelte';
  import TimelineDateTimePicker from '$lib/components/history/TimelineDateTimePicker.svelte';
  import { historyActionLabel, historyFieldLabel } from '$lib/components/history/historyLabels.js';
  import {
    historyEntityFilterOptions,
    historyFieldFilterOptions
  } from '$lib/components/history/historyTimelineFilters.js';
  import {
    fetchTimelineUser,
    fetchTimelineUsers,
    restoreTimelineEvent,
    timelineErrorMessage
  } from '$lib/components/history/historyTimelinePageData.js';
  import { buildTimelineDateTimeISO } from '$lib/components/history/historyTimelineDateTime.js';
  import { can } from '$lib/utils/permissions.js';
  import { t as translate } from '$lib/i18n/index.js';
  import type { ChangeEvent, HistoryAction, HistoryTimelineParams } from '$lib/domain/history.js';
  import type { User } from '$lib/domain/user/index.js';
  import FilterXIcon from '@lucide/svelte/icons/filter-x';
  import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
  import SearchIcon from '@lucide/svelte/icons/search';

  const limit = 50;
  const entityOptions = historyEntityFilterOptions();
  const actionOptions: Array<{ id: HistoryAction; label: string }> = (
    ['create', 'update', 'delete', 'restore'] as const
  ).map((action) => ({ id: action, label: historyActionLabel(action) }));
  const timeline = new ActivityTimelineController(globalActivityDataSource);

  let actorId = $state('');
  let entityTable = $state('');
  let selectedFields = $state<string[]>([]);
  let selectedActions = $state<HistoryAction[]>([]);
  let occurredFromDate = $state('');
  let occurredFromTime = $state('');
  let occurredToDate = $state('');
  let occurredToTime = $state('');
  let restoringEventId = $state<string | null>(null);

  const fieldOptions = $derived(historyFieldFilterOptions(entityTable || undefined));
  const occurredFrom = $derived(
    buildTimelineDateTimeISO(occurredFromDate, occurredFromTime, 'start')
  );
  const occurredTo = $derived(buildTimelineDateTimeISO(occurredToDate, occurredToTime, 'end'));
  const activeFilterCount = $derived(
    [actorId, entityTable, occurredFrom, occurredTo].filter(Boolean).length +
      selectedFields.length +
      selectedActions.length
  );
  const canRestoreTimeline = $derived(can('timeline.restore'));
  const filters = $derived<HistoryTimelineParams>({
    limit,
    actorId: actorId || undefined,
    entityTable: entityTable || undefined,
    occurredFrom: occurredFrom || undefined,
    occurredTo: occurredTo || undefined,
    actions: selectedActions,
    fields: selectedFields
  });

  $effect(() => {
    if (!entityTable && selectedFields.length > 0) selectedFields = [];
  });

  onMount(() => {
    void loadTimeline();
  });

  onDestroy(() => timeline.dispose());

  async function loadTimeline(options: { append?: boolean; force?: boolean } = {}): Promise<void> {
    await timeline.load(filters, options);
  }

  function applyFilters(): void {
    void loadTimeline();
  }

  function resetFilters(): void {
    actorId = '';
    entityTable = '';
    selectedFields = [];
    selectedActions = [];
    occurredFromDate = '';
    occurredFromTime = '';
    occurredToDate = '';
    occurredToTime = '';
    void loadTimeline();
  }

  function setEntityFilter(value: string): void {
    if (value !== entityTable) selectedFields = [];
    entityTable = value;
  }

  async function undoEvent(event: ChangeEvent): Promise<void> {
    if (!canRestoreTimeline) return;
    restoringEventId = event.id;
    try {
      const result = await restoreTimelineEvent(event.id);
      const changedCount = result.restored_count + result.deleted_count;
      addToast(translateUndoSuccess(changedCount), 'success');
      await loadTimeline({ force: true });
    } catch (error) {
      addToast(timelineErrorMessage(error), 'error');
    } finally {
      restoringEventId = null;
    }
  }

  async function fetchUsers(search: string): Promise<User[]> {
    return fetchTimelineUsers(search);
  }

  async function fetchUser(id: string): Promise<User | null> {
    return fetchTimelineUser(id);
  }

  function userLabel(user: User): string {
    const name = [user.first_name, user.last_name].filter(Boolean).join(' ').trim();
    return name || user.email;
  }

  function translateUndoSuccess(count: number): string {
    return translate('history.timeline.undo_success', { count });
  }
</script>

<svelte:head>
  <title>{translate('history.timeline.page_title')} | Infra Link</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-6xl flex-col gap-5">
  <header class="flex flex-wrap items-start justify-between gap-3">
    <div>
      <h1 class="text-2xl font-semibold tracking-tight">
        {translate('history.timeline.page_title')}
      </h1>
      <p class="mt-1 text-sm text-muted-foreground">
        {translate('history.activity.global_description')}
      </p>
    </div>
    <div class="flex items-center gap-2">
      <Badge variant="secondary">
        {translate('history.activity.cursor_loaded_count', { count: timeline.events.length })}
      </Badge>
      {#if activeFilterCount > 0}
        <Badge variant="outline">
          {translate('history.timeline.active_filters', { count: activeFilterCount })}
        </Badge>
      {/if}
    </div>
  </header>

  <section class="rounded-lg border bg-card p-4">
    <div class="grid gap-3 md:grid-cols-2 xl:grid-cols-3">
      <div class="space-y-1.5">
        <Label for="timeline_user">{translate('history.timeline.filters.user')}</Label>
        <AsyncCombobox
          id="timeline_user"
          bind:value={actorId}
          fetcher={fetchUsers}
          fetchById={fetchUser}
          labelKey="email"
          labelFormatter={userLabel}
          clearable
          placeholder={translate('history.timeline.filters.all_users')}
          searchPlaceholder={translate('history.timeline.filters.search_user')}
          emptyText={translate('history.timeline.filters.no_users')}
          width="w-full"
          popupWidth="w-72"
        />
      </div>

      <div class="space-y-1.5">
        <Label for="timeline_entity">{translate('history.timeline.filters.entity')}</Label>
        <StaticCombobox
          id="timeline_entity"
          items={entityOptions}
          bind:value={entityTable}
          labelKey="label"
          clearable
          clearText={translate('history.activity.all_entity_types')}
          placeholder={translate('history.activity.all_entity_types')}
          searchPlaceholder={translate('history.activity.search_entity_type')}
          emptyText={translate('history.activity.no_entity_types')}
          width="w-full"
          onValueChange={setEntityFilter}
        />
      </div>

      <div class="space-y-1.5">
        <Label>{translate('common.actions')}</Label>
        <HistoryFieldMultiSelect
          bind:value={selectedActions}
          items={actionOptions}
          placeholder={translate('history.activity.all_actions')}
          searchPlaceholder={translate('history.activity.search_actions')}
          emptyText={translate('history.activity.no_actions')}
          selectedText={translate('history.activity.actions_selected', {
            count: selectedActions.length
          })}
          showSelectedBadges={false}
        />
      </div>

      {#if entityTable}
        <div class="space-y-1.5">
          <Label>{translate('history.timeline.filters.fields')}</Label>
          <HistoryFieldMultiSelect
            bind:value={selectedFields}
            items={fieldOptions}
            placeholder={translate('history.timeline.filters.select_fields')}
            searchPlaceholder={translate('history.timeline.filters.search_field')}
            emptyText={translate('history.timeline.filters.no_fields')}
            selectedText={translate('history.timeline.filters.fields_selected', {
              count: selectedFields.length
            })}
            showSelectedBadges={false}
          />
        </div>
      {/if}

      <TimelineDateTimePicker
        id="timeline_from"
        label={translate('history.timeline.filters.from')}
        bind:date={occurredFromDate}
        bind:time={occurredFromTime}
        timeLabel={translate('history.when')}
        defaultTime="00:00"
      />
      <TimelineDateTimePicker
        id="timeline_to"
        label={translate('history.timeline.filters.to')}
        bind:date={occurredToDate}
        bind:time={occurredToTime}
        timeLabel={translate('history.when')}
        defaultTime="23:59"
      />
    </div>

    <div class="mt-4 flex flex-wrap gap-2 border-t pt-4">
      <Button onclick={applyFilters} disabled={timeline.loading || timeline.loadingMore}>
        <SearchIcon class="size-4" />
        {translate('history.activity.apply_filters')}
      </Button>
      <Button
        variant="outline"
        onclick={resetFilters}
        disabled={timeline.loading || timeline.loadingMore || activeFilterCount === 0}
      >
        <FilterXIcon class="size-4" />
        {translate('common.reset')}
      </Button>
      <Button
        variant="outline"
        size="icon"
        onclick={() => void loadTimeline({ force: true })}
        disabled={timeline.loading || timeline.loadingMore}
        aria-label={translate('history.activity.refresh_aria')}
      >
        <RefreshCwIcon class={['size-4', timeline.loading ? 'animate-spin' : ''].join(' ')} />
      </Button>
    </div>
  </section>

  <ActivityFeed
    events={timeline.events}
    loading={timeline.loading}
    error={timeline.error}
    canRestore={canRestoreTimeline}
    {restoringEventId}
    onRestore={undoEvent}
  />

  <div class="flex flex-wrap items-center justify-between gap-3">
    <p class="text-sm text-muted-foreground">
      {translate('history.activity.cursor_loaded_count', { count: timeline.events.length })}
    </p>
    {#if timeline.hasMore}
      <Button
        variant="outline"
        disabled={timeline.loading || timeline.loadingMore}
        onclick={() => void loadTimeline({ append: true })}
      >
        {#if timeline.loadingMore}
          <RefreshCwIcon class="size-4 animate-spin" />
          {translate('history.activity.loading')}
        {:else}
          {translate('history.activity.load_more')}
        {/if}
      </Button>
    {/if}
  </div>
</div>
