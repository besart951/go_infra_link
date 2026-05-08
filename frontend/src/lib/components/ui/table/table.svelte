<script lang="ts">
  import type { Snippet } from 'svelte';
  import { resizableTableColumns } from '$lib/actions/resizableTableColumns.js';
  import { cn } from '$lib/utils.js';

  let {
    class: className,
    children,
    ...restProps
  } = $props<{
    class?: string;
    children?: Snippet;
    [key: string]: any;
  }>();
</script>

<div class="relative w-full max-w-full min-w-0 overflow-x-auto overscroll-auto">
  <table
    use:resizableTableColumns
    class={cn('w-full caption-bottom text-sm', className)}
    {...restProps}
  >
    {@render children?.()}
  </table>
</div>
