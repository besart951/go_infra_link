<script lang="ts">
  import type { Snippet } from 'svelte';
  import { Skeleton } from '$lib/components/ui/skeleton/index.js';
  import Cell from './table-cell.svelte';
  import Row from './table-row.svelte';

  interface Props {
    loading: boolean;
    rowCount?: number;
    columnCount?: number;
    delayMs?: number;
    rowClass?: string;
    cellClass?: string;
    skeletonClass?: string;
    children?: Snippet<[number]>;
  }

  let {
    loading,
    rowCount = 5,
    columnCount = 1,
    delayMs = 150,
    rowClass,
    cellClass,
    skeletonClass = 'h-8 w-full',
    children
  }: Props = $props();

  let visible = $state(false);

  const rows = $derived.by(() =>
    Array.from({ length: Math.max(0, rowCount) }, (_, index) => index)
  );
  const columns = $derived.by(() =>
    Array.from({ length: Math.max(0, columnCount) }, (_, index) => index)
  );

  $effect(() => {
    if (!loading) {
      visible = false;
      return;
    }

    if (delayMs <= 0) {
      visible = true;
      return;
    }

    const timeout = window.setTimeout(() => {
      visible = true;
    }, delayMs);

    return () => window.clearTimeout(timeout);
  });
</script>

{#if visible}
  {#each rows as rowIndex (rowIndex)}
    <Row class={rowClass} aria-hidden="true" data-table-loading-row="true">
      {#if children}
        {@render children(rowIndex)}
      {:else}
        {#each columns as columnIndex (columnIndex)}
          <Cell class={cellClass} data-table-loading-cell="true">
            <Skeleton class={skeletonClass} />
          </Cell>
        {/each}
      {/if}
    </Row>
  {/each}
{/if}
