<script lang="ts">
  import { cn, type WithElementRef } from "@ui-svelte/utils.js";
  import type { HTMLAttributes } from "svelte/elements";

  let {
    ref = $bindable(null),
    class: className,
    children,
    ...restProps
  }: WithElementRef<HTMLAttributes<HTMLElement>> = $props();
</script>

<main
  bind:this={ref}
  data-slot="sidebar-inset"
  class={cn(
    "bg-background relative flex w-full flex-1 flex-col",
    "md:peer-data-[variant=inset]:m-2 md:peer-data-[variant=inset]:ms-0 md:peer-data-[variant=inset]:rounded-xl md:peer-data-[variant=inset]:shadow-sm",
    className,
  )}
  {...restProps}
>
  {@render children?.()}
</main>

<style>
  @media (min-width: 768px) {
    :global([data-slot="sidebar"][data-variant="inset"][data-state="collapsed"])
      ~ [data-slot="sidebar-inset"] {
      margin-inline-start: 0.5rem;
    }
  }
</style>
