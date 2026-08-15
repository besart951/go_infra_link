<script lang="ts">
  import type { ActivityItem } from '$lib/activity/contract.js';
  import { t as translate } from '$lib/i18n/index.js';
  import ActivityTimelineItem from './ActivityTimelineItem.svelte';

  interface ActivityDayGroup {
    key: string;
    label: string;
    items: ActivityItem[];
  }

  interface Props {
    items: ActivityItem[];
    loading?: boolean;
    error?: string | null;
    emptyText?: string;
    onSelect?: (item: ActivityItem) => void;
  }

  let {
    items,
    loading = false,
    error = null,
    emptyText = translate('history.activity.empty'),
    onSelect
  }: Props = $props();

  const dayGroups = $derived.by(() => groupByDay(items));

  function groupByDay(items: ActivityItem[]): ActivityDayGroup[] {
    const groups = new Map<string, ActivityDayGroup>();
    for (const item of items) {
      const date = new Date(item.occurredAt);
      const key = new Intl.DateTimeFormat('en-CA').format(date);
      const existing = groups.get(key);
      if (existing) {
        existing.items.push(item);
      } else {
        groups.set(key, { key, label: activityDayLabel(date), items: [item] });
      }
    }
    return [...groups.values()];
  }

  function activityDayLabel(date: Date): string {
    const today = new Date();
    const yesterday = new Date(today);
    yesterday.setDate(today.getDate() - 1);
    const formatter = new Intl.DateTimeFormat('en-CA');
    const key = formatter.format(date);
    if (key === formatter.format(today)) return translate('history.activity.today');
    if (key === formatter.format(yesterday)) return translate('history.activity.yesterday');
    return new Intl.DateTimeFormat('de-CH', { dateStyle: 'long' }).format(date);
  }
</script>

<section class="rounded-lg border bg-card">
  {#if loading && items.length === 0}
    <div class="space-y-6 p-5">
      {#each Array(5) as _}
        <div class="grid grid-cols-[2rem_minmax(0,1fr)] gap-3">
          <div class="size-8 animate-pulse rounded-full bg-muted"></div>
          <div class="space-y-2 pt-1">
            <div class="h-4 w-1/3 animate-pulse rounded bg-muted"></div>
            <div class="h-3 w-2/3 animate-pulse rounded bg-muted"></div>
            <div class="h-12 w-full animate-pulse rounded bg-muted"></div>
          </div>
        </div>
      {/each}
    </div>
  {:else if error && items.length === 0}
    <div class="border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive">{error}</div>
  {:else if items.length === 0}
    <div class="p-10 text-center text-sm text-muted-foreground">{emptyText}</div>
  {:else}
    <div class="divide-y divide-border/70 px-5">
      {#each dayGroups as group (group.key)}
        <section class="py-5 first:pt-5">
          <h2 class="mb-4 text-xs font-semibold tracking-wide text-muted-foreground uppercase">
            {group.label}
          </h2>
          <div>
            {#each group.items as item, index (item.id)}
              <ActivityTimelineItem
                {item}
                isLast={index === group.items.length - 1}
                onclick={onSelect}
              />
            {/each}
          </div>
        </section>
      {/each}
    </div>
  {/if}
</section>
