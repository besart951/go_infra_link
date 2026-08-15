<script lang="ts">
  import ConfirmDialog from '$lib/components/confirm-dialog.svelte';
  import { confirm } from '$lib/stores/confirm-dialog.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Collapsible from '$lib/components/ui/collapsible/index.js';
  import * as Sheet from '$lib/components/ui/sheet/index.js';
  import {
    formatHistoryDate,
    formatHistoryValue,
    historyActionLabel,
    historyActionVariant,
    historyFieldLabel,
    historyTableLabel
  } from '$lib/components/history/historyLabels.js';
  import type { ActivityItem } from '$lib/activity/contract.js';
  import { t as translate } from '$lib/i18n/index.js';
  import RotateCcwIcon from '@lucide/svelte/icons/rotate-ccw';
  import ChevronDownIcon from '@lucide/svelte/icons/chevron-down';

  interface Props {
    open?: boolean;
    item?: ActivityItem | null;
    canRestore?: boolean;
    restoring?: boolean;
    onRestore?: (item: ActivityItem) => void | Promise<void>;
  }

  let {
    open = $bindable(false),
    item = null,
    canRestore = false,
    restoring = false,
    onRestore
  }: Props = $props();

  const actionLabel = $derived(
    item?.action === 'relation_changed'
      ? translate('history.activity.relation_changed')
      : item
        ? historyActionLabel(item.event.action)
        : ''
  );
  const actionVariant = $derived(
    item?.action === 'relation_changed'
      ? 'outline'
      : item
        ? historyActionVariant(item.event.action)
        : 'secondary'
  );
  const restoreAvailable = $derived(Boolean(item && canRestore && item.event.action !== 'restore'));
  let technicalOpen = $state(false);

  async function restore(): Promise<void> {
    if (!item || !onRestore || !restoreAvailable || restoring) return;
    const approved = await confirm({
      title: translate('history.activity.restore_title'),
      message: translate('history.activity.restore_message'),
      confirmText: translate('history.restore'),
      variant: 'destructive'
    });
    if (!approved) return;
    await onRestore(item);
  }
</script>

<ConfirmDialog />
<Sheet.Root bind:open>
  <Sheet.Content class="w-full gap-0 overflow-y-auto p-0 sm:max-w-xl">
    {#if item}
      <Sheet.Header class="border-b px-6 py-5 pr-12">
        <div class="flex flex-wrap items-center gap-2">
          <Badge variant={actionVariant}>{actionLabel}</Badge>
          <span class="text-sm font-medium">{historyTableLabel(item.entity.table)}</span>
        </div>
        <Sheet.Title class="mt-2">{translate('history.activity.detail_title')}</Sheet.Title>
        <Sheet.Description>
          {item.actorName ?? translate('history.system')} · {formatHistoryDate(item.occurredAt)}
        </Sheet.Description>
      </Sheet.Header>

      <div class="space-y-6 px-6 py-5">
        {#if item.scopes.length > 0}
          <section>
            <h3 class="text-sm font-medium">{translate('history.activity.context')}</h3>
            <p class="mt-1 text-sm text-muted-foreground">
              {item.scopes.map((scope) => scope.label ?? scope.id.slice(0, 8)).join(' / ')}
            </p>
          </section>
        {/if}

        {#if item.summary}
          <section>
            <h3 class="text-sm font-medium">{translate('history.activity.summary')}</h3>
            <p class="mt-1 text-sm text-muted-foreground">{item.summary}</p>
          </section>
        {/if}

        <section>
          <h3 class="text-sm font-medium">{translate('history.activity.changed_fields')}</h3>
          {#if item.changes.length > 0}
            <div class="mt-3 divide-y overflow-hidden rounded-md border">
              {#each item.changes as change (change.field)}
                <div class="space-y-2 px-3 py-3">
                  <div class="text-sm font-medium">{historyFieldLabel(change.field)}</div>
                  <div
                    class="grid grid-cols-[minmax(0,1fr)_1rem_minmax(0,1fr)] items-start gap-2 text-sm"
                  >
                    <span class="min-w-0 break-words text-muted-foreground">
                      {formatHistoryValue(change.before)}
                    </span>
                    <span class="text-center text-muted-foreground">→</span>
                    <span class="min-w-0 font-medium break-words">
                      {formatHistoryValue(change.after)}
                    </span>
                  </div>
                </div>
              {/each}
            </div>
          {:else}
            <p class="mt-2 rounded-md border bg-muted/40 px-3 py-2 text-sm text-muted-foreground">
              {translate('history.activity.no_field_values')}
            </p>
          {/if}
        </section>

        <Collapsible.Root bind:open={technicalOpen}>
          <Collapsible.Trigger
            class="flex w-full items-center justify-between rounded-md border bg-muted/30 px-3 py-2 text-left text-xs font-medium text-muted-foreground hover:bg-muted/60"
          >
            {translate('history.activity.technical_details')}
            <ChevronDownIcon
              class={['size-4 transition-transform', technicalOpen ? 'rotate-180' : ''].join(' ')}
            />
          </Collapsible.Trigger>
          <Collapsible.Content
            class="rounded-b-md border-x border-b bg-muted/20 px-3 py-2 text-xs text-muted-foreground"
          >
            <div>{translate('history.activity.entry')}: {item.id}</div>
            <div>{translate('history.object')}: {item.entity.id}</div>
            {#if item.batchId}<div>Batch: {item.batchId}</div>{/if}
          </Collapsible.Content>
        </Collapsible.Root>
      </div>

      {#if restoreAvailable}
        <Sheet.Footer class="border-t px-6 py-4">
          <Button variant="outline" disabled={restoring} onclick={() => void restore()}>
            <RotateCcwIcon class="size-4" />
            {translate('history.restore')}
          </Button>
        </Sheet.Footer>
      {/if}
    {/if}
  </Sheet.Content>
</Sheet.Root>
