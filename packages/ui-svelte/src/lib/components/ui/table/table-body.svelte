<script lang="ts">
  import type { Snippet } from 'svelte';
  import { cn } from '@ui-svelte/utils.js';

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

<tbody class={cn('[&_tr:last-child]:border-0', className)} {...restProps}>
  {@render children?.()}
</tbody>
