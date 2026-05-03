<script lang="ts">
  import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Dialog from '$lib/components/ui/dialog/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import { addToast } from '$lib/components/toast.svelte';
  import { getErrorMessage } from '$lib/api/client.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { t as translate } from '$lib/i18n/index.js';
  import { historyRepository } from '$lib/infrastructure/api/historyRepository.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import type { ChangeEvent } from '$lib/domain/history.js';
  import {
    formatHistoryDate,
    formatHistoryValue,
    historyActionLabel,
    historyActionVariant,
    historyActorLabel,
    historyFieldLabel,
    historyTableLabel
  } from './historyLabels.js';
  import {
    buildHistoryTimelineView,
    type HistoryTimelineGroup,
    type HistoryTimelineRow
  } from './historyTimelineView.js';
  import ChevronsDownUpIcon from '@lucide/svelte/icons/chevrons-down-up';
  import ChevronsUpDownIcon from '@lucide/svelte/icons/chevrons-up-down';
  import ChevronLeftIcon from '@lucide/svelte/icons/chevron-left';
  import ChevronRightIcon from '@lucide/svelte/icons/chevron-right';
  import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
  import RotateCcwIcon from '@lucide/svelte/icons/rotate-ccw';

  interface Props {
    open?: boolean;
    title?: string;
    scopeType?: string;
    scopeId?: string;
    entityTable?: string;
    entityId?: string;
    projectId?: string;
    controlCabinetId?: string;
    onRestored?: () => void | Promise<void>;
  }

  let {
    open = $bindable(false),
    title,
    scopeType,
    scopeId,
    entityTable,
    entityId,
    projectId,
    controlCabinetId,
    onRestored
  }: Props = $props();

  const t = createTranslator();
  const limit = 25;

  let events = $state<ChangeEvent[]>([]);
  let loading = $state(false);
  let error = $state<string | null>(null);
  let page = $state(1);
  let totalPages = $state(1);
  let restoringEventId = $state<string | null>(null);
  let directSectionOpen = $state(true);
  let childSectionOpen = $state(true);
  let expandedGroupKeys = $state<Set<string>>(new Set());
  let initializedExpansionKey = $state('');

  const timelineView = $derived.by(() =>
    buildHistoryTimelineView(events, { scopeType, scopeId, controlCabinetId })
  );
  const canRestoreTimeline = $derived(canPerform('restore', 'timeline'));
  const visibleRowsEmpty = $derived.by(() =>
    timelineView.isHierarchicalView
      ? timelineView.directRows.length === 0 && timelineView.childGroups.length === 0
      : timelineView.flatRows.length === 0
  );
  const expansionKey = $derived.by(() => {
    const eventKey = events.map((event) => event.id).join('|');
    return `${scopeType ?? ''}:${scopeId ?? ''}:${page}:${eventKey}`;
  });

  $effect(() => {
    if (!open) return;

    const controller = new AbortController();
    void loadTimeline(controller.signal);

    return () => controller.abort();
  });

  $effect(() => {
    if (!open || expansionKey === initializedExpansionKey) return;

    directSectionOpen = true;
    childSectionOpen = true;
    expandedGroupKeys = new Set(collectGroupKeys(timelineView.childGroups));
    initializedExpansionKey = expansionKey;
  });

  async function loadTimeline(signal?: AbortSignal): Promise<void> {
    loading = true;
    error = null;

    try {
      const params = {
        scopeType,
        scopeId,
        entityTable,
        entityId,
        page,
        limit
      };
      const response = projectId
        ? await historyRepository.listProjectTimeline(projectId, params, signal)
        : await historyRepository.listTimeline(params, signal);

      events = response.items;
      totalPages = Math.max(response.total_pages || 1, 1);
      page = response.page || page;
    } catch (loadError) {
      if (loadError instanceof DOMException && loadError.name === 'AbortError') return;
      error = getErrorMessage(loadError);
    } finally {
      loading = false;
    }
  }

  async function restoreEvent(event: ChangeEvent): Promise<void> {
    restoringEventId = event.id;
    try {
      const mode = event.action === 'delete' ? 'before' : 'after';
      const result = await historyRepository.restoreEvent(event.id, mode);
      addToast(translate('history.restored', { count: result.restored_count }), 'success');
      await onRestored?.();
      await loadTimeline();
    } catch (restoreError) {
      addToast(getErrorMessage(restoreError), 'error');
    } finally {
      restoringEventId = null;
    }
  }

  async function restoreHierarchy(event: ChangeEvent): Promise<void> {
    if (!controlCabinetId) return;

    restoringEventId = event.id;
    try {
      const result = projectId
        ? await historyRepository.restoreProjectControlCabinet(
            projectId,
            controlCabinetId,
            event.id
          )
        : await historyRepository.restoreControlCabinet(controlCabinetId, event.id);

      addToast(
        translate('history.hierarchy_restored', { count: result.restored_count }),
        'success'
      );
      await onRestored?.();
      await loadTimeline();
    } catch (restoreError) {
      addToast(getErrorMessage(restoreError), 'error');
    } finally {
      restoringEventId = null;
    }
  }

  function nextPage(): void {
    if (page >= totalPages) return;
    page += 1;
  }

  function previousPage(): void {
    if (page <= 1) return;
    page -= 1;
  }

  function rowBefore(row: HistoryTimelineRow): string {
    if (!row.summaryOnly) return formatHistoryValue(row.before);
    return row.event.action === 'create' ? '∅' : historyActionLabel(row.event.action);
  }

  function rowAfter(row: HistoryTimelineRow): string {
    if (!row.summaryOnly) return formatHistoryValue(row.after);
    return row.event.action === 'delete' ? '∅' : historyActionLabel(row.event.action);
  }

  function groupLabel(group: HistoryTimelineGroup): string {
    const name = group.label ?? translate('history.unknown_entity');
    return translate(`history.groups.${group.kind}`, { name });
  }

  function directTableTitle(): string {
    const key = `history.direct_tables.${scopeType ?? 'default'}`;
    const label = translate(key);
    return label === key ? translate('history.direct_changes') : label;
  }

  function groupTableTitle(groups: HistoryTimelineGroup[]): string {
    const kind = commonGroupKind(groups);
    const key = kind ? `history.group_tables.${kind}` : 'history.child_changes';
    const label = translate(key);
    return label === key ? translate('history.child_changes') : label;
  }

  function commonGroupKind(groups: HistoryTimelineGroup[]): HistoryTimelineGroup['kind'] | null {
    const [first] = groups;
    if (!first) return null;
    return groups.every((group) => group.kind === first.kind) ? first.kind : null;
  }

  function collectGroupKeys(groups: HistoryTimelineGroup[]): string[] {
    return groups.flatMap((group) => [group.key, ...collectGroupKeys(group.children)]);
  }

  function expandWholeHierarchy(): void {
    directSectionOpen = true;
    childSectionOpen = true;
    expandedGroupKeys = new Set(collectGroupKeys(timelineView.childGroups));
  }

  function collapseWholeHierarchy(): void {
    directSectionOpen = false;
    childSectionOpen = false;
    expandedGroupKeys = new Set();
  }

  function isGroupOpen(group: HistoryTimelineGroup): boolean {
    return expandedGroupKeys.has(group.key);
  }

  function setGroupOpen(group: HistoryTimelineGroup, isOpen: boolean): void {
    const next = new Set(expandedGroupKeys);
    if (isOpen) {
      next.add(group.key);
    } else {
      next.delete(group.key);
    }
    expandedGroupKeys = next;
  }

  function groupChangeCount(group: HistoryTimelineGroup): number {
    return (
      group.rows.length +
      group.children.reduce((total, child) => total + groupChangeCount(child), 0)
    );
  }

  function groupLatestTimestamp(group: HistoryTimelineGroup): number {
    const ownLatest = group.rows.reduce(
      (latest, row) => Math.max(latest, new Date(row.event.occurred_at).getTime()),
      0
    );
    return group.children.reduce(
      (latest, child) => Math.max(latest, groupLatestTimestamp(child)),
      ownLatest
    );
  }

  function groupLatestDate(group: HistoryTimelineGroup): string {
    const timestamp = groupLatestTimestamp(group);
    return timestamp > 0 ? formatHistoryDate(new Date(timestamp).toISOString()) : '∅';
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="max-h-[88vh] overflow-hidden sm:max-w-4xl">
    <Tooltip.Provider>
      <div class="flex max-h-[64vh] min-h-72 flex-col gap-3 overflow-y-auto pr-1">
        {#if loading && events.length === 0}
          <div class="space-y-2">
            {#each Array(5) as _}
              <div class="h-12 rounded-md border bg-muted/40"></div>
            {/each}
          </div>
        {:else if error}
          <div
            class="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive"
          >
            {error}
          </div>
        {:else if events.length === 0 || visibleRowsEmpty}
          <div
            class="rounded-md border bg-muted/40 px-4 py-8 text-center text-sm text-muted-foreground"
          >
            {$t('history.empty')}
          </div>
        {:else if timelineView.isHierarchicalView}
          <div class="flex items-center gap-1">
            <Tooltip.Root>
              <Tooltip.Trigger
                class={buttonVariants({ variant: 'outline', size: 'icon-sm' })}
                aria-label={$t('history.expand_all')}
                onclick={expandWholeHierarchy}
              >
                <ChevronsUpDownIcon class="size-4" />
              </Tooltip.Trigger>
              <Tooltip.Content>{$t('history.expand_all')}</Tooltip.Content>
            </Tooltip.Root>
            <Tooltip.Root>
              <Tooltip.Trigger
                class={buttonVariants({ variant: 'outline', size: 'icon-sm' })}
                aria-label={$t('history.collapse_all')}
                onclick={collapseWholeHierarchy}
              >
                <ChevronsDownUpIcon class="size-4" />
              </Tooltip.Trigger>
              <Tooltip.Content>{$t('history.collapse_all')}</Tooltip.Content>
            </Tooltip.Root>
          </div>

          {#if timelineView.directRows.length > 0}
            <details
              class="overflow-hidden rounded-md border bg-card"
              open={directSectionOpen}
              ontoggle={(event) =>
                (directSectionOpen = (event.currentTarget as HTMLDetailsElement).open)}
            >
              <summary
                class="flex cursor-pointer items-center justify-between gap-3 bg-muted/40 px-3 py-2 text-sm font-medium select-none"
              >
                <span>{directTableTitle()}</span>
                <Badge variant="secondary">
                  {$t('history.change_count', { count: timelineView.directRows.length })}
                </Badge>
              </summary>
              <div class="border-t">
                {@render rowStack(timelineView.directRows, false)}
              </div>
            </details>
          {/if}

          {#if timelineView.childGroups.length > 0}
            <details
              class="overflow-hidden rounded-md border bg-card"
              open={childSectionOpen}
              ontoggle={(event) =>
                (childSectionOpen = (event.currentTarget as HTMLDetailsElement).open)}
            >
              <summary
                class="flex cursor-pointer items-center justify-between gap-3 bg-muted/40 px-3 py-2 text-sm font-medium select-none"
              >
                <span>{groupTableTitle(timelineView.childGroups)}</span>
                <Badge variant="secondary">
                  {$t('history.change_count', {
                    count: timelineView.childGroups.reduce(
                      (total, group) => total + groupChangeCount(group),
                      0
                    )
                  })}
                </Badge>
              </summary>
              <div class="border-t p-2">
                {@render groupTable(timelineView.childGroups, false)}
              </div>
            </details>
          {/if}
        {:else}
          {@render rowStack(timelineView.flatRows)}
        {/if}
      </div>
    </Tooltip.Provider>

    <Dialog.Footer class="items-center justify-between gap-3 sm:justify-between">
      <div class="text-sm text-muted-foreground">
        {$t('history.page', { page, total: totalPages })}
      </div>
      <div class="flex items-center gap-2">
        <Button
          variant="outline"
          size="icon-sm"
          disabled={loading || page <= 1}
          onclick={previousPage}
        >
          <ChevronLeftIcon class="size-4" />
        </Button>
        <Button
          variant="outline"
          size="icon-sm"
          disabled={loading || page >= totalPages}
          onclick={nextPage}
        >
          <ChevronRightIcon class="size-4" />
        </Button>
        <Button
          variant="outline"
          size="icon-sm"
          disabled={loading}
          onclick={() => void loadTimeline()}
        >
          <RefreshCwIcon class="size-4" />
        </Button>
      </div>
    </Dialog.Footer>
  </Dialog.Content>
</Dialog.Root>

{#snippet rowStack(rows: HistoryTimelineRow[], framed = true)}
  <div class:rounded-md={framed} class:border={framed} class="overflow-x-auto bg-card">
    <div class="min-w-[68rem] divide-y">
      <div
        class="grid grid-cols-[minmax(10rem,0.85fr)_7rem_minmax(11rem,1fr)_9rem_minmax(11rem,1fr)_16rem] items-center gap-3 bg-muted/60 px-3 py-2 text-xs font-medium text-muted-foreground"
      >
        <span>{$t('history.field')}</span>
        <span>{$t('history.action')}</span>
        <span>{$t('history.before')}</span>
        <span>{$t('history.when')}</span>
        <span>{$t('history.after')}</span>
        <span class="text-right">{$t('common.actions')}</span>
      </div>
      {#each rows as row (row.event.id)}
        <div
          class="grid grid-cols-[minmax(10rem,0.85fr)_7rem_minmax(11rem,1fr)_9rem_minmax(11rem,1fr)_16rem] items-center gap-3 px-3 py-2 text-sm"
        >
          <div class="min-w-0">
            <div class="truncate font-medium" title={historyFieldLabel(row.field)}>
              {historyFieldLabel(row.field)}
            </div>
            <div
              class="truncate text-xs text-muted-foreground"
              title={historyTableLabel(row.event.entity_table)}
            >
              {historyTableLabel(row.event.entity_table)}
              {#if row.moreFields > 0}
                · {$t('history.more_fields', { count: row.moreFields })}
              {/if}
            </div>
          </div>
          <div>
            <Badge variant={historyActionVariant(row.event.action)}>
              {historyActionLabel(row.event.action)}
            </Badge>
          </div>
          <span class="truncate text-muted-foreground" title={rowBefore(row)}>{rowBefore(row)}</span
          >
          <div class="min-w-0 text-xs text-muted-foreground">
            <div class="truncate" title={formatHistoryDate(row.event.occurred_at)}>
              {formatHistoryDate(row.event.occurred_at)}
            </div>
            <div class="truncate" title={row.event.actor_id}>{historyActorLabel(row.event)}</div>
          </div>
          <span class="truncate font-medium" title={rowAfter(row)}>{rowAfter(row)}</span>
          <div class="flex justify-end gap-1.5">
            {#if canRestoreTimeline && controlCabinetId}
              <Tooltip.Root>
                <Tooltip.Trigger
                  class={buttonVariants({ variant: 'outline', size: 'sm' })}
                  disabled={restoringEventId !== null}
                  aria-label={$t('history.restore_hierarchy_tooltip')}
                  onclick={() => void restoreHierarchy(row.event)}
                >
                  <RotateCcwIcon class="size-3.5" />
                  {$t('history.restore_hierarchy')}
                </Tooltip.Trigger>
                <Tooltip.Content class="max-w-xs">
                  {$t('history.restore_hierarchy_tooltip')}
                </Tooltip.Content>
              </Tooltip.Root>
            {/if}
            {#if canRestoreTimeline}
              <Tooltip.Root>
                <Tooltip.Trigger
                  class={buttonVariants({ variant: 'outline', size: 'sm' })}
                  disabled={restoringEventId !== null}
                  aria-label={$t('history.restore_tooltip')}
                  onclick={() => void restoreEvent(row.event)}
                >
                  <RotateCcwIcon class="size-3.5" />
                  {$t('history.restore')}
                </Tooltip.Trigger>
                <Tooltip.Content class="max-w-xs">
                  {$t('history.restore_tooltip')}
                </Tooltip.Content>
              </Tooltip.Root>
            {/if}
          </div>
        </div>
      {/each}
    </div>
  </div>
{/snippet}

{#snippet groupTable(groups: HistoryTimelineGroup[], showTitle = true)}
  <div class="space-y-2">
    {#if showTitle}
      <div class="text-xs font-medium text-muted-foreground">
        {groupTableTitle(groups)}
      </div>
    {/if}
    <div class="overflow-x-auto rounded-md border bg-card">
      <div class="min-w-[48rem] divide-y">
        <div
          class="grid grid-cols-[minmax(16rem,1fr)_10rem_9rem] items-center gap-3 bg-muted/60 px-3 py-2 text-xs font-medium text-muted-foreground"
        >
          <span>{$t('history.object')}</span>
          <span>{$t('history.latest_change')}</span>
          <span class="text-right">{$t('history.change_count_header')}</span>
        </div>
        {#each groups as group (group.key)}
          <details
            class="group"
            open={isGroupOpen(group)}
            ontoggle={(event) =>
              setGroupOpen(group, (event.currentTarget as HTMLDetailsElement).open)}
          >
            <summary
              class="grid cursor-pointer grid-cols-[minmax(16rem,1fr)_10rem_9rem] items-center gap-3 px-3 py-2 text-sm select-none hover:bg-muted/30"
            >
              <span class="min-w-0 truncate font-medium">{groupLabel(group)}</span>
              <span class="truncate text-xs text-muted-foreground">{groupLatestDate(group)}</span>
              <span class="text-right">
                <Badge variant="secondary">
                  {$t('history.change_count', { count: groupChangeCount(group) })}
                </Badge>
              </span>
            </summary>
            <div class="space-y-3 border-t bg-muted/10 p-2 pl-4">
              {#if group.rows.length > 0}
                {@render rowStack(group.rows, false)}
              {/if}
              {#if group.children.length > 0}
                {@render groupTable(group.children)}
              {/if}
            </div>
          </details>
        {/each}
      </div>
    </div>
  </div>
{/snippet}
