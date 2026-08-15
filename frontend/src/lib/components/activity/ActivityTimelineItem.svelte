<script lang="ts">
  import * as Avatar from '$lib/components/ui/avatar/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import {
    historyActionLabel,
    historyActionVariant,
    historyTableLabel
  } from '$lib/components/history/historyLabels.js';
  import type { ActivityItem } from '$lib/activity/contract.js';
  import { t as translate } from '$lib/i18n/index.js';
  import ActivityDiffPreview from './ActivityDiffPreview.svelte';
  import ActivityEventIcon from './ActivityEventIcon.svelte';

  interface Props {
    item: ActivityItem;
    isLast?: boolean;
    onclick?: (item: ActivityItem) => void;
  }

  let { item, isLast = false, onclick }: Props = $props();
  const actorInitials = $derived(initials(item.actorName));
  const contextLabel = $derived(
    item.scopes
      .map((scope) => scope.label)
      .filter((label): label is string => Boolean(label))
      .join(' / ')
  );
  const actionLabel = $derived(
    item.action === 'relation_changed'
      ? translate('history.activity.relation_changed')
      : historyActionLabel(item.event.action)
  );
  const actionVariant = $derived(
    item.action === 'relation_changed' ? 'outline' : historyActionVariant(item.event.action)
  );

  function initials(name?: string): string {
    if (!name) return 'S';
    return (
      name
        .split(/\s+/u)
        .filter(Boolean)
        .slice(0, 2)
        .map((part) => part[0]?.toLocaleUpperCase('de-CH'))
        .join('') || 'S'
    );
  }

  function formatActivityTime(value: string): string {
    return new Intl.DateTimeFormat('de-CH', { hour: '2-digit', minute: '2-digit' }).format(
      new Date(value)
    );
  }
</script>

<article class="group relative grid grid-cols-[2rem_minmax(0,1fr)] gap-3 py-3 first:pt-0 last:pb-0">
  <div class="relative flex justify-center">
    {#if !isLast}
      <div class="absolute top-8 bottom-[-0.75rem] w-px bg-border"></div>
    {/if}
    <ActivityEventIcon action={item.action} />
  </div>

  <div class="min-w-0 rounded-md px-2 py-1 transition-colors hover:bg-muted/60">
    <div class="flex min-w-0 flex-wrap items-start justify-between gap-x-3 gap-y-1">
      <div class="min-w-0 flex-1">
        <div class="flex min-w-0 flex-wrap items-center gap-2">
          <Badge variant={actionVariant}>{actionLabel}</Badge>
          {#if onclick}
            <button
              type="button"
              class="min-w-0 truncate text-left text-sm font-medium outline-none hover:underline focus-visible:rounded-sm focus-visible:ring-2 focus-visible:ring-ring"
              onclick={() => onclick?.(item)}
            >
              {historyTableLabel(item.entity.table)}
            </button>
          {:else}
            <span class="min-w-0 truncate text-sm font-medium"
              >{historyTableLabel(item.entity.table)}</span
            >
          {/if}
        </div>
        <div class="mt-1 flex min-w-0 items-center gap-2 text-xs text-muted-foreground">
          <Avatar.Root class="size-5 border border-border">
            <Avatar.Fallback class="bg-muted text-[9px] text-muted-foreground">
              {actorInitials}
            </Avatar.Fallback>
          </Avatar.Root>
          <span class="truncate">{item.actorName ?? translate('history.system')}</span>
          {#if contextLabel}
            <span aria-hidden="true">·</span>
            <span class="min-w-0 truncate">{contextLabel}</span>
          {/if}
        </div>
      </div>
      <time class="shrink-0 text-xs text-muted-foreground" datetime={item.occurredAt}>
        {formatActivityTime(item.occurredAt)}
      </time>
    </div>

    {#if item.summary}
      <p class="mt-2 text-sm text-muted-foreground">{item.summary}</p>
    {/if}
    <div class="mt-2">
      <ActivityDiffPreview changes={item.changes} />
    </div>
  </div>
</article>
