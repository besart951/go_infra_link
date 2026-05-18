<script lang="ts">
  import { onMount } from 'svelte';
  import { addToast } from '$lib/components/toast.svelte';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { t as translate } from '$lib/i18n/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import { Label } from '$lib/components/ui/label/index.js';
  import StaticCombobox from '$lib/components/ui/combobox/StaticCombobox.svelte';
  import AsyncCombobox from '$lib/components/ui/combobox/AsyncCombobox.svelte';
  import HistoryFieldMultiSelect from '$lib/components/history/HistoryFieldMultiSelect.svelte';
  import TimelineDateTimePicker from '$lib/components/history/TimelineDateTimePicker.svelte';
  import {
    formatHistoryDate,
    formatHistoryValue,
    historyActionLabel,
    historyActionVariant,
    historyActorLabel,
    historyFieldLabel,
    historyPrimaryField,
    historyTableLabel,
    historyVisibleDiffEntries
  } from '$lib/components/history/historyLabels.js';
  import {
    historyEntityFilterOptions,
    historyFieldFilterOptions
  } from '$lib/components/history/historyTimelineFilters.js';
  import {
    fetchTimelineUser,
    fetchTimelineUsers,
    loadHistoryTimeline,
    restoreTimelineEvent,
    timelineErrorMessage
  } from '$lib/components/history/historyTimelinePageData.js';
  import { buildTimelineDateTimeISO } from '$lib/components/history/historyTimelineDateTime.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import type { ChangeEvent, HistoryTimelineParams } from '$lib/domain/history.js';
  import type { User } from '$lib/domain/user/index.js';
  import FilterXIcon from '@lucide/svelte/icons/filter-x';
  import InfoIcon from '@lucide/svelte/icons/info';
  import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
  import RotateCcwIcon from '@lucide/svelte/icons/rotate-ccw';
  import SearchIcon from '@lucide/svelte/icons/search';
  import XIcon from '@lucide/svelte/icons/x';

  const limit = 50;
  const entityOptions = historyEntityFilterOptions();
  const t = createTranslator();

  let actorId = $state('');
  let entityTable = $state('');
  let selectedFields = $state<string[]>([]);
  let occurredFromDate = $state('');
  let occurredFromTime = $state('');
  let occurredToDate = $state('');
  let occurredToTime = $state('');
  let events = $state<ChangeEvent[]>([]);
  let loading = $state(false);
  let loadingMore = $state(false);
  let error = $state<string | null>(null);
  let page = $state(0);
  let total = $state(0);
  let totalPages = $state(1);
  let restoringEventId = $state<string | null>(null);
  let requestId = 0;

  const fieldOptions = $derived(historyFieldFilterOptions(entityTable || undefined));
  const occurredFrom = $derived(
    buildTimelineDateTimeISO(occurredFromDate, occurredFromTime, 'start')
  );
  const occurredTo = $derived(buildTimelineDateTimeISO(occurredToDate, occurredToTime, 'end'));
  const activeFilterCount = $derived(
    [actorId, entityTable, occurredFrom, occurredTo].filter(Boolean).length + selectedFields.length
  );
  const selectedEntityLabel = $derived(
    entityOptions.find((option) => option.id === entityTable)?.label
  );
  const hasMore = $derived(page < totalPages);
  const loadedCount = $derived(events.length);
  const canRestoreTimeline = $derived(canPerform('restore', 'timeline'));

  $effect(() => {
    if (!entityTable && selectedFields.length > 0) {
      selectedFields = [];
    }
  });

  onMount(() => {
    void loadTimeline('reset');
  });

  async function loadTimeline(mode: 'reset' | 'append' = 'reset'): Promise<void> {
    if (mode === 'append' && (loading || loadingMore || !hasMore)) return;

    const currentRequest = ++requestId;
    const targetPage = mode === 'append' ? page + 1 : 1;
    if (mode === 'append') {
      loadingMore = true;
    } else {
      loading = true;
      loadingMore = false;
      page = 0;
      events = [];
    }
    error = null;

    try {
      const params: HistoryTimelineParams = {
        page: targetPage,
        limit,
        actorId: actorId || undefined,
        entityTable: entityTable || undefined,
        occurredFrom: occurredFrom || undefined,
        occurredTo: occurredTo || undefined,
        fields: selectedFields
      };
      const response = await loadHistoryTimeline(params);
      if (currentRequest !== requestId) return;
      events = mode === 'append' ? [...events, ...response.items] : response.items;
      total = response.total;
      page = response.page || targetPage;
      totalPages = Math.max(response.total_pages || 1, 1);
    } catch (loadError) {
      if (currentRequest !== requestId) return;
      error = timelineErrorMessage(loadError);
    } finally {
      if (currentRequest === requestId) {
        loading = false;
        loadingMore = false;
      }
    }
  }

  function applyFilters(): void {
    void loadTimeline('reset');
  }

  function resetFilters(): void {
    actorId = '';
    entityTable = '';
    selectedFields = [];
    occurredFromDate = '';
    occurredFromTime = '';
    occurredToDate = '';
    occurredToTime = '';
    void loadTimeline('reset');
  }

  function setEntityFilter(value: string): void {
    if (value !== entityTable) {
      selectedFields = [];
    }
    entityTable = value;
  }

  function clearEntityFilter(): void {
    entityTable = '';
    selectedFields = [];
  }

  function removeFieldFilter(field: string): void {
    selectedFields = selectedFields.filter((selectedField) => selectedField !== field);
  }

  function loadNextPage(force = false): void {
    if (!force && error && events.length > 0) return;
    void loadTimeline('append');
  }

  async function undoEvent(event: ChangeEvent): Promise<void> {
    restoringEventId = event.id;
    try {
      const result = await restoreTimelineEvent(event.id);
      const changedCount = result.restored_count + result.deleted_count;
      addToast(translate('history.timeline.undo_success', { count: changedCount }), 'success');
      await loadTimeline('reset');
    } catch (restoreError) {
      addToast(timelineErrorMessage(restoreError), 'error');
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

  function eventTime(value: string): string {
    return new Intl.DateTimeFormat('de-CH', { timeStyle: 'short' }).format(new Date(value));
  }

  function eventDay(value: string): string {
    return new Intl.DateTimeFormat('de-CH', { dateStyle: 'medium' }).format(new Date(value));
  }

  function eventTitle(event: ChangeEvent): string {
    return `${historyTableLabel(event.entity_table)} · ${historyFieldLabel(historyPrimaryField(event))}`;
  }

  function scopeLabel(event: ChangeEvent): string {
    return (
      event.scopes?.map((scope) => scope.label || scope.scope_id.slice(0, 8)).join(' / ') ?? ''
    );
  }

  function eventRestoreDisabled(event: ChangeEvent): boolean {
    return !canRestoreTimeline || restoringEventId !== null || event.action === 'restore';
  }

  function eventTooltipDetails(event: ChangeEvent): Array<{ label: string; value: string }> {
    const details = [
      {
        label: historyTableLabel(event.entity_table),
        value: formatHistoryDate(event.occurred_at)
      },
      {
        label: translate('history.actor'),
        value: historyActorLabel(event).replace(`${translate('history.actor')}: `, '')
      },
      {
        label: 'ID',
        value: event.entity_id
      }
    ];
    if (scopeLabel(event)) {
      details.push({ label: translate('history.object'), value: scopeLabel(event) });
    }
    return details;
  }

  function loadMoreAction(node: HTMLElement): { destroy: () => void } {
    const observer = new IntersectionObserver(
      (entries) => {
        if (entries.some((entry) => entry.isIntersecting)) {
          loadNextPage();
        }
      },
      { rootMargin: '360px 0px' }
    );
    observer.observe(node);
    return {
      destroy() {
        observer.disconnect();
      }
    };
  }
</script>

<svelte:head>
  <title>{$t('history.timeline.page_title')} | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-5">
  <div class="flex flex-wrap items-start justify-between gap-3">
    <div class="min-w-0">
      <h1 class="text-2xl font-semibold tracking-tight">{$t('history.timeline.page_title')}</h1>
      <p class="text-sm text-muted-foreground">
        {$t('history.timeline.page_description')}
      </p>
    </div>
    <div class="flex items-center gap-2">
      <Badge variant="secondary">
        {$t('history.timeline.result_count', { count: total })}
      </Badge>
      {#if activeFilterCount > 0}
        <Badge variant="outline">
          {$t('history.timeline.active_filters', { count: activeFilterCount })}
        </Badge>
      {/if}
    </div>
  </div>

  <section class="rounded-md border bg-card p-3">
    <div class="timeline-filter-grid">
      <div class="min-w-0 space-y-1.5">
        <Label for="timeline_user">{$t('history.timeline.filters.user')}</Label>
        <AsyncCombobox
          id="timeline_user"
          bind:value={actorId}
          fetcher={fetchUsers}
          fetchById={fetchUser}
          labelKey="email"
          labelFormatter={userLabel}
          clearable
          placeholder={$t('history.timeline.filters.all_users')}
          searchPlaceholder={$t('history.timeline.filters.search_user')}
          emptyText={$t('history.timeline.filters.no_users')}
          width="w-full"
          popupWidth="w-72"
        />
      </div>

      <div class="min-w-0 space-y-1.5">
        <Label for="timeline_entity">{$t('history.timeline.filters.entity')}</Label>
        <StaticCombobox
          id="timeline_entity"
          items={entityOptions}
          bind:value={entityTable}
          labelKey="label"
          clearable
          clearText={$t('history.timeline.filters.all_entities')}
          placeholder={$t('history.timeline.filters.all_entities')}
          searchPlaceholder={$t('history.timeline.filters.search_entity')}
          emptyText={$t('history.timeline.filters.no_entities')}
          width="w-full"
          onValueChange={setEntityFilter}
        />
      </div>

      {#if entityTable}
        <div class="min-w-0 space-y-1.5">
          <Label>{$t('history.timeline.filters.fields')}</Label>
          <HistoryFieldMultiSelect
            bind:value={selectedFields}
            items={fieldOptions}
            placeholder={$t('history.timeline.filters.select_fields')}
            searchPlaceholder={$t('history.timeline.filters.search_field')}
            emptyText={$t('history.timeline.filters.no_fields')}
            selectedText={$t('history.timeline.filters.fields_selected', {
              count: selectedFields.length
            })}
            showSelectedBadges={false}
          />
        </div>
      {/if}

      <TimelineDateTimePicker
        id="timeline_from"
        label={$t('history.timeline.filters.from')}
        bind:date={occurredFromDate}
        bind:time={occurredFromTime}
        timeLabel={$t('history.timeline.filters.from_time')}
        defaultTime="00:00"
      />

      <TimelineDateTimePicker
        id="timeline_to"
        label={$t('history.timeline.filters.to')}
        bind:date={occurredToDate}
        bind:time={occurredToTime}
        timeLabel={$t('history.timeline.filters.to_time')}
        defaultTime="23:59"
      />

      <div class="flex min-w-0 flex-wrap items-end gap-2">
        <Button
          class="min-w-0 flex-1 whitespace-normal xl:flex-none xl:whitespace-nowrap"
          onclick={applyFilters}
          disabled={loading || loadingMore}
        >
          <SearchIcon class="size-4" />
          {$t('history.timeline.filters.apply')}
        </Button>
        <Button
          variant="outline"
          size="icon"
          onclick={resetFilters}
          disabled={loading || loadingMore || activeFilterCount === 0}
          aria-label={$t('history.timeline.filters.reset')}
        >
          <FilterXIcon class="size-4" />
        </Button>
      </div>
    </div>

    {#if selectedEntityLabel || selectedFields.length > 0}
      <div class="mt-3 flex flex-wrap items-center gap-1.5 border-t pt-3">
        {#if selectedEntityLabel}
          <Badge variant="outline" class="group max-w-full gap-1.5 pr-1">
            <span class="min-w-0 break-words">{selectedEntityLabel}</span>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              class="size-4 rounded-full p-0 opacity-0 transition-opacity group-hover:opacity-100 hover:bg-muted focus-visible:opacity-100 focus-visible:ring-2 focus-visible:ring-ring"
              onclick={clearEntityFilter}
              aria-label={$t('history.timeline.filters.remove_entity', {
                name: selectedEntityLabel
              })}
            >
              <XIcon class="size-3" />
            </Button>
          </Badge>
        {/if}
        {#each selectedFields as field (field)}
          {@const fieldLabel = historyFieldLabel(field)}
          <Badge variant="secondary" class="group max-w-full gap-1.5 pr-1">
            <span class="min-w-0 break-words">{fieldLabel}</span>
            <Button
              type="button"
              variant="ghost"
              size="icon-sm"
              class="size-4 rounded-full p-0 opacity-0 transition-opacity group-hover:opacity-100 hover:bg-secondary-foreground/20 focus-visible:opacity-100 focus-visible:ring-2 focus-visible:ring-ring"
              onclick={() => removeFieldFilter(field)}
              aria-label={$t('history.timeline.filters.remove_field', { name: fieldLabel })}
            >
              <XIcon class="size-3" />
            </Button>
          </Badge>
        {/each}
      </div>
    {/if}
  </section>

  <Tooltip.Provider>
    <section class="overflow-hidden rounded-md border bg-card">
      {#if loading && events.length === 0}
        <div class="space-y-0 divide-y">
          {#each Array(8) as _}
            <div class="grid grid-cols-[7rem_1rem_minmax(0,1fr)] gap-3 px-4 py-3">
              <div class="h-8 animate-pulse rounded bg-muted"></div>
              <div class="mx-auto h-8 w-px bg-muted"></div>
              <div class="space-y-2">
                <div class="h-4 w-1/3 animate-pulse rounded bg-muted"></div>
                <div class="h-4 w-3/4 animate-pulse rounded bg-muted"></div>
              </div>
            </div>
          {/each}
        </div>
      {:else if error && events.length === 0}
        <div class="border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive">
          {error}
        </div>
      {:else if events.length === 0}
        <div class="px-4 py-10 text-center text-sm text-muted-foreground">
          {$t('history.timeline.empty')}
        </div>
      {:else}
        <div class="divide-y">
          {#each events as event, index (event.id)}
            {@const entries = historyVisibleDiffEntries(event)}
            <article
              class="grid grid-cols-[6.5rem_1.25rem_minmax(0,1fr)] gap-3 px-3 py-3 transition-colors hover:bg-muted/30 md:grid-cols-[8rem_1.5rem_minmax(0,1fr)]"
            >
              <div class="pt-1 text-right">
                <div class="text-xs font-medium">{eventDay(event.occurred_at)}</div>
                <div class="text-xs text-muted-foreground">{eventTime(event.occurred_at)}</div>
              </div>

              <div class="relative flex justify-center">
                {#if index > 0}
                  <div class="absolute top-0 h-3 w-px bg-border"></div>
                {/if}
                {#if index < events.length - 1}
                  <div class="absolute top-6 bottom-[-0.75rem] w-px bg-border"></div>
                {/if}
                <div
                  class="z-10 mt-1.5 size-3 rounded-full border-2 border-background bg-primary"
                ></div>
              </div>

              <div class="min-w-0 space-y-2">
                <div class="flex flex-wrap items-start justify-between gap-2">
                  <div class="min-w-0 flex-1 basis-64 space-y-1">
                    <div class="flex min-w-0 flex-wrap items-center gap-2">
                      <Badge variant={historyActionVariant(event.action)}>
                        {historyActionLabel(event.action)}
                      </Badge>
                      <Tooltip.Root>
                        <Tooltip.Trigger
                          class="max-w-full min-w-0 text-left text-sm font-medium break-words outline-none focus-visible:ring-2 focus-visible:ring-ring"
                        >
                          {eventTitle(event)}
                        </Tooltip.Trigger>
                        <Tooltip.Content class="max-w-sm space-y-1.5" side="top">
                          <div class="font-medium">{eventTitle(event)}</div>
                          {#each eventTooltipDetails(event) as detail}
                            <div
                              class="grid grid-cols-[5.5rem_minmax(0,1fr)] gap-2 text-muted-foreground"
                            >
                              <span>{detail.label}</span>
                              <span class="break-all">{detail.value}</span>
                            </div>
                          {/each}
                        </Tooltip.Content>
                      </Tooltip.Root>
                    </div>
                    <div
                      class="flex min-w-0 flex-wrap gap-x-3 gap-y-1 text-xs text-muted-foreground"
                    >
                      <span class="min-w-0 break-words">{historyActorLabel(event)}</span>
                      {#if scopeLabel(event)}
                        <Tooltip.Root>
                          <Tooltip.Trigger
                            class="max-w-full min-w-0 text-left break-words outline-none focus-visible:ring-2 focus-visible:ring-ring"
                          >
                            {scopeLabel(event)}
                          </Tooltip.Trigger>
                          <Tooltip.Content class="max-w-sm break-all" side="top">
                            {scopeLabel(event)}
                          </Tooltip.Content>
                        </Tooltip.Root>
                      {/if}
                    </div>
                  </div>

                  <div class="flex shrink-0 items-center gap-1.5">
                    <Tooltip.Root>
                      <Tooltip.Trigger
                        class={buttonVariants({ variant: 'ghost', size: 'icon-sm' })}
                        aria-label={$t('history.timeline.details')}
                      >
                        <InfoIcon class="size-4" />
                      </Tooltip.Trigger>
                      <Tooltip.Content class="max-w-sm space-y-2" side="left">
                        <div class="font-medium">{historyTableLabel(event.entity_table)}</div>
                        <div class="space-y-1 text-muted-foreground">
                          <div>{formatHistoryDate(event.occurred_at)}</div>
                          <div>{historyActorLabel(event)}</div>
                          <div class="break-all">ID: {event.entity_id}</div>
                          {#if event.summary}
                            <div>{event.summary}</div>
                          {/if}
                        </div>
                        {#if entries.length > 0}
                          <div class="space-y-1 border-t pt-2">
                            {#each entries.slice(0, 5) as [field, diff]}
                              <div>
                                <span class="font-medium">{historyFieldLabel(field)}:</span>
                                {formatHistoryValue(diff.before)} -> {formatHistoryValue(
                                  diff.after
                                )}
                              </div>
                            {/each}
                          </div>
                        {/if}
                      </Tooltip.Content>
                    </Tooltip.Root>

                    {#if canRestoreTimeline}
                      <Tooltip.Root>
                        <Tooltip.Trigger
                          class={buttonVariants({ variant: 'outline', size: 'sm' })}
                          disabled={eventRestoreDisabled(event)}
                          aria-label={$t('history.timeline.undo_tooltip')}
                          onclick={() => void undoEvent(event)}
                        >
                          <RotateCcwIcon class="size-3.5" />
                          {$t('history.timeline.undo')}
                        </Tooltip.Trigger>
                        <Tooltip.Content class="max-w-xs">
                          {$t(
                            event.action === 'restore'
                              ? 'history.timeline.undo_restore_disabled'
                              : 'history.timeline.undo_tooltip'
                          )}
                        </Tooltip.Content>
                      </Tooltip.Root>
                    {/if}
                  </div>
                </div>

                {#if entries.length > 0}
                  <div class="grid gap-1.5 xl:grid-cols-2">
                    {#each entries.slice(0, 4) as [field, diff]}
                      <div class="rounded-md border bg-background px-2 py-1.5 text-xs">
                        <div class="mb-1 font-medium">{historyFieldLabel(field)}</div>
                        <div class="grid grid-cols-[minmax(0,1fr)_1rem_minmax(0,1fr)] gap-2">
                          <Tooltip.Root>
                            <Tooltip.Trigger
                              class="min-w-0 text-left break-words text-muted-foreground outline-none focus-visible:ring-2 focus-visible:ring-ring"
                            >
                              {formatHistoryValue(diff.before)}
                            </Tooltip.Trigger>
                            <Tooltip.Content class="max-w-sm break-all" side="top">
                              <span class="font-medium">{historyFieldLabel(field)} vorher:</span>
                              {formatHistoryValue(diff.before)}
                            </Tooltip.Content>
                          </Tooltip.Root>
                          <span class="text-center text-muted-foreground">-&gt;</span>
                          <Tooltip.Root>
                            <Tooltip.Trigger
                              class="min-w-0 text-left font-medium break-words outline-none focus-visible:ring-2 focus-visible:ring-ring"
                            >
                              {formatHistoryValue(diff.after)}
                            </Tooltip.Trigger>
                            <Tooltip.Content class="max-w-sm break-all" side="top">
                              <span class="font-medium">{historyFieldLabel(field)} nachher:</span>
                              {formatHistoryValue(diff.after)}
                            </Tooltip.Content>
                          </Tooltip.Root>
                        </div>
                      </div>
                    {/each}
                  </div>
                  {#if entries.length > 4}
                    <div class="text-xs text-muted-foreground">
                      {$t('history.more_fields', { count: entries.length - 4 })}
                    </div>
                  {/if}
                {:else}
                  <div
                    class="rounded-md border bg-background px-2 py-1.5 text-xs break-words text-muted-foreground"
                  >
                    {event.summary || $t('history.record')}
                  </div>
                {/if}
              </div>
            </article>
          {/each}
        </div>
      {/if}
    </section>
  </Tooltip.Provider>

  <div class="flex flex-wrap items-center justify-between gap-3">
    <div class="text-sm text-muted-foreground">
      {$t('history.timeline.loaded_count', { loaded: loadedCount, total })}
    </div>
    <div class="flex items-center gap-2">
      <Button
        variant="outline"
        size="icon-sm"
        disabled={loading || loadingMore}
        onclick={() => void loadTimeline('reset')}
        aria-label={$t('common.refresh')}
      >
        <RefreshCwIcon class="size-4" />
      </Button>
    </div>
  </div>

  {#if events.length > 0}
    <div
      use:loadMoreAction
      class="flex min-h-12 items-center justify-center rounded-md border border-dashed bg-muted/20 px-4 py-3 text-sm text-muted-foreground"
    >
      {#if loadingMore}
        <span class="inline-flex items-center gap-2">
          <span
            class="size-4 animate-spin rounded-full border-2 border-current border-r-transparent"
          ></span>
          {$t('history.timeline.loading_more')}
        </span>
      {:else if error}
        <div class="flex flex-wrap items-center justify-center gap-2 text-destructive">
          <span>{error}</span>
          <Button variant="outline" size="sm" onclick={() => loadNextPage(true)}>
            {$t('history.timeline.retry')}
          </Button>
        </div>
      {:else if hasMore}
        {$t('history.timeline.scroll_hint')}
      {:else}
        {$t('history.timeline.end_reached')}
      {/if}
    </div>
  {/if}
</div>

<style>
  .timeline-filter-grid {
    display: grid;
    grid-template-columns: repeat(auto-fit, minmax(min(100%, 18rem), 1fr));
    gap: 0.75rem;
    align-items: end;
  }
</style>
