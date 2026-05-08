<script lang="ts">
  import type { Snippet } from 'svelte';
  import { cn } from '$lib/utils.js';

  let {
    class: className,
    children,
    resizable = true,
    minResizeWidth,
    ...restProps
  } = $props<{
    class?: string;
    children?: Snippet;
    resizable?: boolean;
    minResizeWidth?: number;
    [key: string]: any;
  }>();
</script>

<th
  data-table-resizable={resizable ? 'true' : 'false'}
  data-table-resize-min-width={minResizeWidth}
  class={cn(
    'h-10 px-4 py-2 text-left align-middle font-medium text-muted-foreground [&:has([role=checkbox])]:pr-0',
    className
  )}
  {...restProps}
>
  {@render children?.()}
</th>
