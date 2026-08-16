<script lang="ts">
  import type { HTMLAttributes } from 'svelte/elements';
  import { cn, type WithElementRef } from '$lib/utils.js';

  let {
    ref = $bindable(null),
    class: className,
    children,
    ...restProps
  }: WithElementRef<HTMLAttributes<HTMLElement>> = $props();
</script>

<div
  bind:this={ref}
  data-slot="sidebar-content"
  data-sidebar="content"
  class={cn(
    'flex min-h-0 flex-1 flex-col gap-2 overflow-auto group-data-[collapsible=icon]:overflow-visible',
    className
  )}
  {...restProps}
>
  {@render children?.()}
</div>

<style>
  [data-sidebar='content'] {
    scrollbar-color: color-mix(in oklab, var(--sidebar-foreground) 38%, transparent) transparent;
    scrollbar-width: thin;
  }

  [data-sidebar='content']::-webkit-scrollbar {
    width: 8px;
  }

  [data-sidebar='content']::-webkit-scrollbar-track {
    background: transparent;
  }

  [data-sidebar='content']::-webkit-scrollbar-thumb {
    background: color-mix(in oklab, var(--sidebar-foreground) 38%, transparent);
    background-clip: padding-box;
    border: 2px solid transparent;
    border-radius: 9999px;
  }

  [data-sidebar='content']::-webkit-scrollbar-thumb:hover {
    background-color: color-mix(in oklab, var(--sidebar-foreground) 55%, transparent);
  }

  [data-sidebar='content']::-webkit-scrollbar-button {
    display: none;
    height: 0;
    width: 0;
  }
</style>
