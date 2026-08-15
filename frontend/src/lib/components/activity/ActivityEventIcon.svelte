<script lang="ts">
  import CheckCircle2Icon from '@lucide/svelte/icons/check-circle-2';
  import CirclePlusIcon from '@lucide/svelte/icons/circle-plus';
  import Link2Icon from '@lucide/svelte/icons/link-2';
  import PencilIcon from '@lucide/svelte/icons/pencil';
  import RotateCcwIcon from '@lucide/svelte/icons/rotate-ccw';
  import Trash2Icon from '@lucide/svelte/icons/trash-2';
  import type { ActivityAction } from '$lib/activity/contract.js';

  interface Props {
    action: ActivityAction;
  }

  let { action }: Props = $props();

  const colorClass = $derived(
    action === 'delete'
      ? 'border-destructive/30 bg-destructive/10 text-destructive'
      : action === 'create'
        ? 'border-success-border bg-success-muted text-success-muted-foreground'
        : action === 'restore'
          ? 'border-warning-border bg-warning-muted text-warning-muted-foreground'
          : action === 'relation_changed'
            ? 'border-info-border bg-info-muted text-info-muted-foreground'
            : 'border-border bg-muted text-muted-foreground'
  );
</script>

<span
  class={['flex size-8 shrink-0 items-center justify-center rounded-full border', colorClass].join(
    ' '
  )}
>
  {#if action === 'create'}
    <CirclePlusIcon class="size-4" />
  {:else if action === 'delete'}
    <Trash2Icon class="size-4" />
  {:else if action === 'restore'}
    <RotateCcwIcon class="size-4" />
  {:else if action === 'relation_changed'}
    <Link2Icon class="size-4" />
  {:else}
    <PencilIcon class="size-4" />
  {/if}
</span>
