<script lang="ts">
  import { goto } from '$app/navigation';
  import NotificationPreviewItem from '$lib/components/notifications/NotificationPreviewItem.svelte';
  import { systemNotificationState } from '$lib/components/notifications/SystemNotificationState.svelte.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Popover from '$lib/components/ui/popover/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import BellIcon from '@lucide/svelte/icons/bell';
  import CheckIcon from '@lucide/svelte/icons/check';
  import InboxIcon from '@lucide/svelte/icons/inbox';
  import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
  import { onDestroy, onMount } from 'svelte';

  const t = createTranslator();
  const notifications = systemNotificationState;

  let open = $state(false);

  onMount(() => {
    notifications.connect();
    void notifications.loadPreview();
  });

  onDestroy(() => {
    notifications.disconnect();
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
        {#if notifications.unreadCount > 0}
          <span
            class="absolute -top-0.5 -right-0.5 flex min-w-4 items-center justify-center rounded-full bg-destructive px-1 text-[10px] leading-4 font-semibold text-destructive-foreground"
          >
            {notifications.unreadCount > 99 ? '99+' : notifications.unreadCount}
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
          {$t('notifications.inbox.unread_count', { count: notifications.unreadCount })}
        </p>
      </div>
      <div class="flex items-center gap-1">
        <Button
          variant="ghost"
          size="icon-sm"
          onclick={() => notifications.loadPreview()}
          disabled={notifications.isPreviewLoading}
        >
          <RefreshCwIcon class={`size-4${notifications.isPreviewLoading ? ' animate-spin' : ''}`} />
        </Button>
        <Button
          variant="ghost"
          size="icon-sm"
          onclick={() => notifications.markAllRead()}
          disabled={notifications.unreadCount === 0}
        >
          <CheckIcon class="size-4" />
        </Button>
      </div>
    </div>

    {#if notifications.previewError}
      <div class="px-3 py-2 text-sm text-destructive">{notifications.previewError}</div>
    {:else if notifications.previewItems.length === 0}
      <div
        class="flex flex-col items-center gap-2 px-4 py-8 text-center text-sm text-muted-foreground"
      >
        <InboxIcon class="size-8" />
        <span>{$t('notifications.inbox.empty')}</span>
      </div>
    {:else}
      <div class="max-h-96 overflow-y-auto">
        {#each notifications.previewItems as notification (notification.id)}
          <NotificationPreviewItem
            {notification}
            dateLabel={notifications.formatDateTime(notification.created_at)}
            onOpen={(item) => notifications.markRead(item)}
            onToggleRead={(item) => notifications.toggleRead(item)}
            onToggleImportant={(item) => notifications.toggleImportant(item)}
            onDelete={(item) => notifications.deleteNotification(item)}
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
