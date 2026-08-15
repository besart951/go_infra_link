<script lang="ts">
  import { historyFieldLabel, formatHistoryValue } from '$lib/components/history/historyLabels.js';
  import { t as translate } from '$lib/i18n/index.js';
  import type { ActivityChange } from '$lib/activity/contract.js';

  interface Props {
    changes: ActivityChange[];
    limit?: number;
  }

  let { changes, limit = 2 }: Props = $props();
  const visibleChanges = $derived(changes.slice(0, limit));
  const moreCount = $derived(Math.max(changes.length - visibleChanges.length, 0));
</script>

{#if visibleChanges.length > 0}
  <div class="space-y-1.5">
    {#each visibleChanges as change (change.field)}
      <div
        class="grid grid-cols-[minmax(0,1fr)_1rem_minmax(0,1fr)] items-center gap-2 rounded-md border bg-background px-2.5 py-2 text-xs"
      >
        <span class="col-span-3 font-medium text-foreground">{historyFieldLabel(change.field)}</span
        >
        <span
          class="min-w-0 truncate text-muted-foreground"
          title={formatHistoryValue(change.before)}
        >
          {formatHistoryValue(change.before)}
        </span>
        <span class="text-center text-muted-foreground">→</span>
        <span class="min-w-0 truncate font-medium" title={formatHistoryValue(change.after)}>
          {formatHistoryValue(change.after)}
        </span>
      </div>
    {/each}
    {#if moreCount > 0}
      <p class="text-xs text-muted-foreground">
        {translate('history.more_fields', { count: moreCount })}
      </p>
    {/if}
  </div>
{/if}
