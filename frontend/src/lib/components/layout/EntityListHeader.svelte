<script lang="ts">
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import type { Snippet } from 'svelte';
  import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
  import InfoIcon from '@lucide/svelte/icons/info';
  import PlusIcon from '@lucide/svelte/icons/plus';

  interface Props {
    title: string;
    description?: string;
    infoLabel: string;
    backHref: string;
    backLabel: string;
    createLabel?: string;
    canCreate?: boolean;
    createActive?: boolean;
    onCreateClick?: () => void;
    children?: Snippet;
  }

  let {
    title,
    description,
    infoLabel,
    backHref,
    backLabel,
    createLabel,
    canCreate = false,
    createActive = false,
    onCreateClick,
    children
  }: Props = $props();
</script>

<Tooltip.Provider>
  <header class="flex flex-col gap-4 border-b pb-5 sm:flex-row sm:items-center sm:justify-between">
    <div class="flex min-w-0 items-center gap-3">
      <Tooltip.Root>
        <Tooltip.Trigger>
          <Button variant="ghost" size="icon" href={backHref} aria-label={backLabel}>
            <ArrowLeftIcon class="size-4" />
          </Button>
        </Tooltip.Trigger>
        <Tooltip.Content>{backLabel}</Tooltip.Content>
      </Tooltip.Root>

      <div class="flex min-w-0 items-center gap-2">
        <h1 class="truncate text-2xl font-semibold tracking-tight">{title}</h1>
        {#if description}
          <Tooltip.Root>
            <Tooltip.Trigger class="inline-flex shrink-0 text-muted-foreground hover:text-foreground">
              <InfoIcon class="size-4" />
              <span class="sr-only">{infoLabel}</span>
            </Tooltip.Trigger>
            <Tooltip.Content class="max-w-xs">{description}</Tooltip.Content>
          </Tooltip.Root>
        {/if}
      </div>
    </div>

    <div class="flex shrink-0 items-center justify-end gap-2">
    {@render children?.()}

    {#if canCreate && createLabel && onCreateClick}
      <Tooltip.Root>
        <Tooltip.Trigger>
          <Button
            variant="default"
            size="icon"
            class={`bg-blue-600 text-white shadow-xs hover:bg-blue-700 ${createActive ? 'ring-2 ring-blue-500/30' : ''}`}
            onclick={onCreateClick}
            aria-label={createLabel}
          >
            <PlusIcon class="size-4" />
          </Button>
        </Tooltip.Trigger>
        <Tooltip.Content>{createLabel}</Tooltip.Content>
      </Tooltip.Root>
    {/if}
    </div>
  </header>
</Tooltip.Provider>
