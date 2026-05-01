<script lang="ts">
  import { Button, buttonVariants } from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as ButtonGroup from '$lib/components/ui/button-group/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
  import BellIcon from '@lucide/svelte/icons/bell';
  import CheckIcon from '@lucide/svelte/icons/check';
  import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';

  interface Props {
    title: string;
    description: string;
    backLabel: string;
    unreadCountLabel: string;
    unreadOnlyLabel: string;
    markAllReadLabel: string;
    refreshLabel: string;
    unreadCount: number;
    unreadOnly: boolean;
    isLoading: boolean;
    onToggleUnreadOnly: () => void;
    onMarkAllRead: () => void;
    onRefresh: () => void;
  }

  let {
    title,
    description,
    backLabel,
    unreadCountLabel,
    unreadOnlyLabel,
    markAllReadLabel,
    refreshLabel,
    unreadCount,
    unreadOnly,
    isLoading,
    onToggleUnreadOnly,
    onMarkAllRead,
    onRefresh
  }: Props = $props();
</script>

<Tooltip.Provider>
  <header class="flex flex-col gap-4 border-b pb-5 sm:flex-row sm:items-center sm:justify-between">
    <div class="flex min-w-0 items-center gap-3">
      <Button variant="ghost" size="icon" href="/notifications" aria-label={backLabel}>
        <ArrowLeftIcon class="size-4" />
      </Button>
      <div class="flex min-w-0 items-start gap-2">
        <BellIcon class="size-5 shrink-0 text-muted-foreground" />
        <div class="min-w-0">
          <h1 class="truncate text-2xl font-semibold tracking-tight">
            {title}
          </h1>
          <p class="text-sm text-muted-foreground">
            {description}
          </p>
        </div>
      </div>
    </div>

    <div class="flex shrink-0 flex-wrap items-center gap-2 sm:justify-end">
      <Badge variant={unreadCount > 0 ? 'default' : 'secondary'} class="h-9 px-3">
        {unreadCountLabel}
      </Badge>

      <ButtonGroup.Root>
        <Tooltip.Root>
          <Tooltip.Trigger
            class={buttonVariants({ variant: unreadOnly ? 'default' : 'outline', size: 'icon' })}
            onclick={onToggleUnreadOnly}
            aria-label={unreadOnlyLabel}
          >
            <BellIcon class="size-4" />
          </Tooltip.Trigger>
          <Tooltip.Content>{unreadOnlyLabel}</Tooltip.Content>
        </Tooltip.Root>

        <Tooltip.Root>
          <Tooltip.Trigger
            class={buttonVariants({ variant: 'outline', size: 'icon' })}
            onclick={onMarkAllRead}
            disabled={unreadCount === 0}
            aria-label={markAllReadLabel}
          >
            <CheckIcon class="size-4" />
          </Tooltip.Trigger>
          <Tooltip.Content>{markAllReadLabel}</Tooltip.Content>
        </Tooltip.Root>

        <Tooltip.Root>
          <Tooltip.Trigger
            class={buttonVariants({ variant: 'outline', size: 'icon' })}
            onclick={onRefresh}
            disabled={isLoading}
            aria-label={refreshLabel}
          >
            <RefreshCwIcon class={`size-4${isLoading ? ' animate-spin' : ''}`} />
          </Tooltip.Trigger>
          <Tooltip.Content>{refreshLabel}</Tooltip.Content>
        </Tooltip.Root>
      </ButtonGroup.Root>
    </div>
  </header>
</Tooltip.Provider>
