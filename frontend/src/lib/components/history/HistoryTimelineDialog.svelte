<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Dialog from '$lib/components/ui/dialog/index.js';
  import { addToast } from '$lib/components/toast.svelte';
  import { getErrorMessage } from '$lib/api/client.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { t as translate } from '$lib/i18n/index.js';
  import { historyRepository } from '$lib/infrastructure/api/historyRepository.js';
  import type { ChangeEvent, HistoryAction } from '$lib/domain/history.js';
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

  const dialogTitle = $derived(title ?? $t('history.title'));

  $effect(() => {
    if (!open) return;

    const controller = new AbortController();
    void loadTimeline(controller.signal);

    return () => controller.abort();
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

  function actionLabel(action: HistoryAction): string {
    return translate(`history.actions.${action}`);
  }

  function actionVariant(
    action: HistoryAction
  ): 'default' | 'secondary' | 'destructive' | 'outline' | 'success' | 'warning' {
    if (action === 'delete') return 'destructive';
    if (action === 'create') return 'success';
    if (action === 'restore') return 'warning';
    return 'secondary';
  }

  function tableLabel(table: string): string {
    const key = `history.tables.${table}`;
    const label = translate(key);
    return label === key ? table : label;
  }

  function formatDate(value: string): string {
    return new Intl.DateTimeFormat('de-CH', {
      dateStyle: 'medium',
      timeStyle: 'short'
    }).format(new Date(value));
  }

  function formatValue(value: unknown): string {
    if (value === null || value === undefined || value === '') return '∅';
    if (typeof value === 'object') return JSON.stringify(value);
    return String(value);
  }

  function visibleDiffEntries(
    event: ChangeEvent
  ): Array<[string, { before: unknown; after: unknown }]> {
    return Object.entries(event.diff_json ?? {})
      .filter(([field]) => field !== 'created_at' && field !== 'updated_at')
      .slice(0, 5);
  }

  function hiddenDiffCount(event: ChangeEvent): number {
    const count = Object.keys(event.diff_json ?? {}).filter(
      (field) => field !== 'created_at' && field !== 'updated_at'
    ).length;
    return Math.max(count - 5, 0);
  }
</script>

<Dialog.Root bind:open>
  <Dialog.Content class="max-h-[88vh] overflow-hidden sm:max-w-4xl">
    <Dialog.Header>
      <Dialog.Title>{dialogTitle}</Dialog.Title>
      <Dialog.Description>{$t('history.description')}</Dialog.Description>
    </Dialog.Header>

    <div class="flex max-h-[64vh] min-h-72 flex-col gap-3 overflow-y-auto pr-1">
      {#if loading && events.length === 0}
        <div class="space-y-3">
          {#each Array(4) as _}
            <div class="h-24 rounded-md border bg-muted/40"></div>
          {/each}
        </div>
      {:else if error}
        <div
          class="rounded-md border border-destructive/30 bg-destructive/10 px-4 py-3 text-sm text-destructive"
        >
          {error}
        </div>
      {:else if events.length === 0}
        <div
          class="rounded-md border bg-muted/40 px-4 py-8 text-center text-sm text-muted-foreground"
        >
          {$t('history.empty')}
        </div>
      {:else}
        {#each events as event (event.id)}
          {@const diffEntries = visibleDiffEntries(event)}
          <div class="rounded-md border bg-card p-4">
            <div class="flex flex-wrap items-start justify-between gap-3">
              <div class="min-w-0 space-y-1">
                <div class="flex flex-wrap items-center gap-2">
                  <Badge variant={actionVariant(event.action)}>{actionLabel(event.action)}</Badge>
                  <span class="font-medium">{tableLabel(event.entity_table)}</span>
                  <span class="text-xs text-muted-foreground">{formatDate(event.occurred_at)}</span>
                </div>
                <div class="text-xs break-all text-muted-foreground">
                  {event.actor_id
                    ? `${$t('history.actor')}: ${event.actor_id}`
                    : $t('history.system')}
                </div>
              </div>

              <div class="flex flex-wrap justify-end gap-2">
                {#if controlCabinetId}
                  <Button
                    variant="outline"
                    size="sm"
                    disabled={restoringEventId !== null}
                    onclick={() => void restoreHierarchy(event)}
                  >
                    <RotateCcwIcon class="size-4" />
                    {$t('history.restore_hierarchy')}
                  </Button>
                {/if}
                <Button
                  variant="outline"
                  size="sm"
                  disabled={restoringEventId !== null}
                  onclick={() => void restoreEvent(event)}
                >
                  <RotateCcwIcon class="size-4" />
                  {$t('history.restore')}
                </Button>
              </div>
            </div>

            {#if diffEntries.length > 0}
              <div class="mt-3 overflow-hidden rounded-md border">
                <div
                  class="grid grid-cols-[minmax(8rem,0.8fr)_minmax(0,1fr)_minmax(0,1fr)] bg-muted px-3 py-2 text-xs font-medium text-muted-foreground"
                >
                  <span>{$t('history.field')}</span>
                  <span>{$t('history.before')}</span>
                  <span>{$t('history.after')}</span>
                </div>
                {#each diffEntries as [field, diff]}
                  <div
                    class="grid grid-cols-[minmax(8rem,0.8fr)_minmax(0,1fr)_minmax(0,1fr)] border-t px-3 py-2 text-xs"
                  >
                    <span class="font-medium">{field}</span>
                    <span class="truncate text-muted-foreground" title={formatValue(diff.before)}>
                      {formatValue(diff.before)}
                    </span>
                    <span class="truncate" title={formatValue(diff.after)}>
                      {formatValue(diff.after)}
                    </span>
                  </div>
                {/each}
              </div>
              {#if hiddenDiffCount(event) > 0}
                <p class="mt-2 text-xs text-muted-foreground">
                  {$t('history.more_fields', { count: hiddenDiffCount(event) })}
                </p>
              {/if}
            {/if}
          </div>
        {/each}
      {/if}
    </div>

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
