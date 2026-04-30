<script lang="ts">
  import { goto } from '$app/navigation';
  import { getErrorMessage } from '$lib/api/client.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Popover from '$lib/components/ui/popover/index.js';
  import type { SystemNotification } from '$lib/domain/notification/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { systemNotificationRepository } from '$lib/infrastructure/api/systemNotificationRepository.js';
  import BellIcon from '@lucide/svelte/icons/bell';
  import CheckIcon from '@lucide/svelte/icons/check';
  import InboxIcon from '@lucide/svelte/icons/inbox';
  import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';
  import { onDestroy, onMount } from 'svelte';

  const t = createTranslator();

  let open = $state(false);
  let items = $state<SystemNotification[]>([]);
  let unreadCount = $state(0);
  let isLoading = $state(false);
  let error = $state<string | null>(null);
  let refreshTimer: ReturnType<typeof setInterval> | undefined;

  async function loadNotifications() {
    isLoading = true;
    error = null;
    try {
      const result = await systemNotificationRepository.list({ page: 1, limit: 5 });
      items = result.items;
      unreadCount = result.unread_count;
    } catch (err) {
      error = getErrorMessage(err);
    } finally {
      isLoading = false;
    }
  }

  async function markRead(notification: SystemNotification) {
    if (notification.read_at) return;
    try {
      await systemNotificationRepository.markRead(notification.id);
      await loadNotifications();
    } catch (err) {
      error = getErrorMessage(err);
    }
  }

  async function markAllRead() {
    try {
      await systemNotificationRepository.markAllRead();
      await loadNotifications();
    } catch (err) {
      error = getErrorMessage(err);
    }
  }

  function formatDateTime(value: string): string {
    return new Intl.DateTimeFormat('de-CH', {
      dateStyle: 'short',
      timeStyle: 'short'
    }).format(new Date(value));
  }

  onMount(() => {
    loadNotifications();
    refreshTimer = setInterval(loadNotifications, 60000);
  });

  onDestroy(() => {
    if (refreshTimer) clearInterval(refreshTimer);
  });
</script>

<Popover.Root bind:open>
  <Popover.Trigger>
    {#snippet child({ props })}
      <Button
        variant="ghost"
        size="icon"
        class="relative"
        aria-label={$t('notifications.inbox.open')}
        {...props}
      >
        <BellIcon class="size-5" />
        {#if unreadCount > 0}
          <span
            class="text-destructive-foreground absolute -top-0.5 -right-0.5 flex min-w-4 items-center justify-center rounded-full bg-destructive px-1 text-[10px] leading-4 font-semibold"
          >
            {unreadCount > 99 ? '99+' : unreadCount}
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
          {$t('notifications.inbox.unread_count', { count: unreadCount })}
        </p>
      </div>
      <div class="flex items-center gap-1">
        <Button variant="ghost" size="icon-sm" onclick={loadNotifications} disabled={isLoading}>
          <RefreshCwIcon class={`size-4${isLoading ? ' animate-spin' : ''}`} />
        </Button>
        <Button variant="ghost" size="icon-sm" onclick={markAllRead} disabled={unreadCount === 0}>
          <CheckIcon class="size-4" />
        </Button>
      </div>
    </div>

    {#if error}
      <div class="px-3 py-2 text-sm text-destructive">{error}</div>
    {:else if items.length === 0}
      <div
        class="flex flex-col items-center gap-2 px-4 py-8 text-center text-sm text-muted-foreground"
      >
        <InboxIcon class="size-8" />
        <span>{$t('notifications.inbox.empty')}</span>
      </div>
    {:else}
      <div class="max-h-96 overflow-y-auto">
        {#each items as notification (notification.id)}
          <button
            type="button"
            class="flex w-full gap-3 border-b px-3 py-3 text-left text-sm transition-colors hover:bg-muted/60"
            onclick={() => markRead(notification)}
          >
            <span
              class={`mt-1 size-2 shrink-0 rounded-full ${notification.read_at ? 'bg-muted' : 'bg-primary'}`}
            ></span>
            <span class="min-w-0 flex-1">
              <span class="block truncate font-medium">{notification.title}</span>
              {#if notification.body}
                <span class="mt-0.5 line-clamp-2 block text-xs text-muted-foreground">
                  {notification.body}
                </span>
              {/if}
              <span class="mt-1 block text-xs text-muted-foreground">
                {formatDateTime(notification.created_at)}
              </span>
            </span>
          </button>
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
