<script lang="ts">
  import { toActivityItems } from '$lib/activity/activityMapper.js';
  import type { ChangeEvent } from '$lib/domain/history.js';
  import type { ActivityItem } from '$lib/activity/contract.js';
  import ActivityDetailSheet from './ActivityDetailSheet.svelte';
  import ActivityTimeline from './ActivityTimeline.svelte';

  interface Props {
    events: ChangeEvent[];
    loading?: boolean;
    error?: string | null;
    emptyText?: string;
    canRestore?: boolean;
    restoringEventId?: string | null;
    onRestore?: (event: ChangeEvent) => void | Promise<void>;
  }

  let {
    events,
    loading = false,
    error = null,
    emptyText,
    canRestore = false,
    restoringEventId = null,
    onRestore
  }: Props = $props();

  let selectedItem = $state<ActivityItem | null>(null);
  let detailOpen = $state(false);
  const items = $derived(toActivityItems(events));

  function select(item: ActivityItem): void {
    selectedItem = item;
    detailOpen = true;
  }

  async function restore(item: ActivityItem): Promise<void> {
    await onRestore?.(item.event);
  }
</script>

<ActivityTimeline {items} {loading} {error} {emptyText} onSelect={select} />
<ActivityDetailSheet
  bind:open={detailOpen}
  item={selectedItem}
  {canRestore}
  restoring={restoringEventId === selectedItem?.id}
  onRestore={restore}
/>
