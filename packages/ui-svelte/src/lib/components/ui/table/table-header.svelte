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

<thead class={cn('[&_tr]:border-b', className)} {...restProps}>
  {@render children?.()}
</thead>
