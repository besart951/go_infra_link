<script lang="ts">
  import * as Sheet from "@ui-svelte/components/ui/sheet/index.js";
  import { Button } from "@ui-svelte/components/ui/button/index.js";
  import { cn, type WithElementRef } from "@ui-svelte/utils.js";
  import XIcon from "@lucide/svelte/icons/x";
  import type { HTMLAttributes } from "svelte/elements";
  import { SIDEBAR_WIDTH_MOBILE } from "./constants.js";
  import { useSidebar } from "./context.svelte.js";

  let {
    ref = $bindable(null),
    side = "left",
    variant = "sidebar",
    collapsible = "offcanvas",
    style,
    class: className,
    children,
    ...restProps
  }: WithElementRef<HTMLAttributes<HTMLDivElement>> & {
    side?: "left" | "right";
    variant?: "sidebar" | "floating" | "inset";
    collapsible?: "offcanvas" | "icon" | "none";
  } = $props();

  const sidebar = useSidebar();
  const collapsedDesktopWidth = $derived.by(() =>
    variant === "floating" || variant === "inset"
      ? "calc(var(--sidebar-width-icon) + 1rem + 2px)"
      : "var(--sidebar-width-icon)",
  );
  const desktopGapWidth = $derived.by(() => {
    if (sidebar.state !== "collapsed") {
      return "var(--sidebar-width)";
    }

    if (collapsible === "offcanvas") {
      return "0px";
    }

    if (collapsible === "icon") {
      return collapsedDesktopWidth;
    }

    return "var(--sidebar-width)";
  });
  const desktopContainerWidth = $derived.by(() =>
    sidebar.state === "collapsed" && collapsible === "icon"
      ? collapsedDesktopWidth
      : "var(--sidebar-width)",
  );
  const rootWidthStyle = $derived.by(() =>
    `width: var(--sidebar-width); ${style ?? ""}`.trim(),
  );
  const mobileWidthStyle = `--sidebar-width: ${SIDEBAR_WIDTH_MOBILE}; width: var(--sidebar-width);`;
</script>

{#if collapsible === "none"}
  <div
    class={cn(
      "bg-sidebar text-sidebar-foreground flex h-full flex-col",
      className,
    )}
    style={rootWidthStyle}
    bind:this={ref}
    {...restProps}
  >
    {@render children?.()}
  </div>
{:else if sidebar.isMobile}
  <Sheet.Root
    bind:open={() => sidebar.openMobile, (v) => sidebar.setOpenMobile(v)}
    {...restProps}
  >
    <Sheet.Content
      onInteractOutside={() => sidebar.setOpenMobile(false)}
      data-sidebar="sidebar"
      data-slot="sidebar"
      data-mobile="true"
      class="bg-sidebar text-sidebar-foreground p-0 [&>button]:hidden"
      style={mobileWidthStyle}
      {side}
    >
      <Sheet.Header class="sr-only">
        <Sheet.Title>Sidebar</Sheet.Title>
        <Sheet.Description>Displays the mobile sidebar.</Sheet.Description>
      </Sheet.Header>
      <div class="flex h-full w-full flex-col">
        <div class="flex items-center justify-end p-2">
          <Button
            variant="ghost"
            size="icon"
            type="button"
            onclick={() => sidebar.setOpenMobile(false)}
          >
            <XIcon class="size-4" />
            <span class="sr-only">Close sidebar</span>
          </Button>
        </div>
        {@render children?.()}
      </div>
    </Sheet.Content>
  </Sheet.Root>
{:else}
  <div
    bind:this={ref}
    class="group peer text-sidebar-foreground hidden md:block"
    data-state={sidebar.state}
    data-collapsible={sidebar.state === "collapsed" ? collapsible : ""}
    data-variant={variant}
    data-side={side}
    data-slot="sidebar"
  >
    <!-- This is what handles the sidebar gap on desktop -->
    <div
      data-slot="sidebar-gap"
      class={cn(
        "relative bg-transparent transition-[width] duration-200 ease-linear",
        "group-data-[side=right]:rotate-180",
      )}
      style={`width: ${desktopGapWidth};`}
    ></div>
    <div
      data-slot="sidebar-container"
      class={cn(
        "fixed inset-y-0 z-10 hidden h-svh transition-[left,right,width] duration-200 ease-linear md:flex",
        side === "left"
          ? "start-0 group-data-[collapsible=offcanvas]:start-[calc(var(--sidebar-width)*-1)]"
          : "end-0 group-data-[collapsible=offcanvas]:end-[calc(var(--sidebar-width)*-1)]",
        // Adjust the padding for floating and inset variants.
        variant === "floating" || variant === "inset"
          ? "p-2"
          : "group-data-[side=left]:border-e group-data-[side=right]:border-s",
        className,
      )}
      style={`width: ${desktopContainerWidth}; ${style ?? ""}`.trim()}
      {...restProps}
    >
      <div
        data-sidebar="sidebar"
        data-slot="sidebar-inner"
        class="bg-sidebar group-data-[variant=floating]:border-sidebar-border flex h-full w-full flex-col group-data-[variant=floating]:rounded-lg group-data-[variant=floating]:border group-data-[variant=floating]:shadow-sm"
      >
        {@render children?.()}
      </div>
    </div>
  </div>
{/if}
