<script lang="ts">
  import { goto } from '$app/navigation';
  import { NotificationBellState } from '$lib/components/notifications/NotificationBellState.svelte.js';
  import NotificationPreviewItem from '$lib/components/notifications/NotificationPreviewItem.svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Popover from '$lib/components/ui/popover/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import BellIcon from '@lucide/svelte/icons/bell';
  import CheckIcon from '@lucide/svelte/icons/check';
  import InboxIcon from '@lucide/svelte/icons/inbox';
  import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
  import { onDestroy, onMount } from 'svelte';

  const t = createTranslator();
  const bellState = new NotificationBellState();

  let open = $state(false);

  onMount(() => {
    bellState.startPolling();
  });

  onDestroy(() => {
    bellState.stopPolling();
  });
</script>

<Popover.Root bind:open>
  <Popover.Trigger>
    {#snippet child({ props })}
      <Button
        {...props}
        variant="ghost"
        size="icon"
        class="relative"
        aria-label={$t('notifications.inbox.open')}
      >
        <BellIcon class="size-5" />
        {#if bellState.unreadCount > 0}
          <span
            class="absolute -top-0.5 -right-0.5 flex min-w-4 items-center justify-center rounded-full bg-destructive px-1 text-[10px] leading-4 font-semibold text-destructive-foreground"
          >
            {bellState.unreadCount > 99 ? '99+' : bellState.unreadCount}
          </span>
        {/if}
      </Button>
    {/snippet}
  </Popover.Trigger>

  <Popover.Content class="w-[min(24rem,calc(100vw-2rem))] p-0" align="end">
    <div class="flex items-center justify-between border-b px-3 py-2">
      <div class="min-w-0">
        <h2 class="text-sm font-semibold">{$t('notifications.inbox.title')}</h2>
        <p class="text-xs text-muted-foreground">
          {$t('notifications.inbox.unread_count', { count: bellState.unreadCount })}
        </p>
      </div>
      <div class="flex items-center gap-1">
        <Button
          variant="ghost"
          size="icon-sm"
          onclick={() => bellState.loadNotifications()}
          disabled={bellState.isLoading}
        >
          <RefreshCwIcon class={`size-4${bellState.isLoading ? ' animate-spin' : ''}`} />
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          onclick={() => bellState.markAllRead()}
          disabled={bellState.unreadCount === 0}
        >
          <CheckIcon class="size-4" />
        </Button>
      </div>
    </div>

    {#if bellState.error}
      <div class="px-3 py-2 text-sm text-destructive">{bellState.error}</div>
    {:else if bellState.items.length === 0}
      <div
        class="flex flex-col items-center gap-2 px-4 py-8 text-center text-sm text-muted-foreground"
      >
        <InboxIcon class="size-8" />
        <span>{$t('notifications.inbox.empty')}</span>
      </div>
    {:else}
      <div class="max-h-96 overflow-y-auto">
        {#each bellState.items as notification (notification.id)}
          <NotificationPreviewItem
            {notification}
            dateLabel={bellState.formatDateTime(notification.created_at)}
            onOpen={(item) => bellState.markRead(item)}
            onToggleRead={(item) => bellState.toggleRead(item)}
            onToggleImportant={(item) => bellState.toggleImportant(item)}
            onDelete={(item) => bellState.deleteNotification(item)}
          />
        {/each}
      </div>
    {/if}

    <div class="border-t p-2">
      <Button
        variant="ghost"
        class="w-full justify-center"
        onclick={() => {
          open = false;
          goto('/notifications/inbox');
        }}
      >
        {$t('notifications.inbox.view_all')}
      </Button>
    </div>
  </Popover.Content>
</Popover.Root>
