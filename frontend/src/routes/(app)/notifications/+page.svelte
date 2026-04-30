<script lang="ts">
  import { onMount } from 'svelte';
  import { getErrorMessage } from '$lib/api/client.js';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import type { SystemNotification } from '$lib/domain/notification/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { systemNotificationRepository } from '$lib/infrastructure/api/systemNotificationRepository.js';
  import BellIcon from '@lucide/svelte/icons/bell';
  import CheckIcon from '@lucide/svelte/icons/check';
  import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';

  const t = createTranslator();

  let notifications = $state<SystemNotification[]>([]);
  let unreadCount = $state(0);
  let page = $state(1);
  let totalPages = $state(1);
  let unreadOnly = $state(false);
  let isLoading = $state(true);
  let error = $state<string | null>(null);

  async function loadNotifications(nextPage = page) {
    isLoading = true;
    error = null;
    try {
      const result = await systemNotificationRepository.list({
        page: nextPage,
        limit: 20,
        unread_only: unreadOnly
      });
      notifications = result.items;
      unreadCount = result.unread_count;
      page = result.page;
      totalPages = result.total_pages || 1;
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
      await loadNotifications(1);
    } catch (err) {
      error = getErrorMessage(err);
    }
  }

  function formatDateTime(value: string): string {
    return new Intl.DateTimeFormat('de-CH', {
      dateStyle: 'medium',
      timeStyle: 'short'
    }).format(new Date(value));
  }

  onMount(() => {
    loadNotifications(1);
  });
</script>

<svelte:head>
  <title>{$t('notifications.inbox.page_title')} | {$t('app.brand')}</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-5xl flex-col gap-5">
  <header class="flex flex-col gap-3 border-b pb-5 sm:flex-row sm:items-end sm:justify-between">
    <div class="space-y-1">
      <div class="flex items-center gap-2">
        <BellIcon class="size-5 text-muted-foreground" />
        <h1 class="text-2xl font-semibold tracking-tight">
          {$t('notifications.inbox.page_title')}
        </h1>
      </div>
      <p class="text-sm text-muted-foreground">{$t('notifications.inbox.page_description')}</p>
    </div>
    <div class="flex flex-wrap items-center gap-2">
      <Badge variant={unreadCount > 0 ? 'default' : 'secondary'}>
        {$t('notifications.inbox.unread_count', { count: unreadCount })}
      </Badge>
      <Button
        variant={unreadOnly ? 'default' : 'outline'}
        onclick={() => {
          unreadOnly = !unreadOnly;
          loadNotifications(1);
        }}
      >
        {$t('notifications.inbox.unread_only')}
      </Button>
      <Button variant="outline" onclick={markAllRead} disabled={unreadCount === 0}>
        <CheckIcon class="size-4" />
        {$t('notifications.inbox.mark_all_read')}
      </Button>
      <Button variant="outline" onclick={() => loadNotifications()} disabled={isLoading}>
        <RefreshCwIcon class={`size-4${isLoading ? ' animate-spin' : ''}`} />
        {$t('common.refresh')}
      </Button>
    </div>
  </header>

  {#if error}
    <Card.Root>
      <Card.Content class="py-4 text-sm text-destructive">{error}</Card.Content>
    </Card.Root>
  {/if}

  <div class="flex flex-col gap-3">
    {#if isLoading && notifications.length === 0}
      <Card.Root>
        <Card.Content class="py-8 text-center text-sm text-muted-foreground">
          {$t('common.loading')}
        </Card.Content>
      </Card.Root>
    {:else if notifications.length === 0}
      <Card.Root>
        <Card.Content class="py-10 text-center text-sm text-muted-foreground">
          {$t('notifications.inbox.empty')}
        </Card.Content>
      </Card.Root>
    {:else}
      {#each notifications as notification (notification.id)}
        <Card.Root class={notification.read_at ? '' : 'border-primary/40'}>
          <Card.Header class="gap-2">
            <div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
              <div class="min-w-0">
                <Card.Title class="text-base leading-6">{notification.title}</Card.Title>
                <Card.Description>
                  {formatDateTime(notification.created_at)} · {notification.event_key}
                </Card.Description>
              </div>
              {#if !notification.read_at}
                <Button variant="outline" size="sm" onclick={() => markRead(notification)}>
                  <CheckIcon class="size-4" />
                  {$t('notifications.inbox.mark_read')}
                </Button>
              {/if}
            </div>
          </Card.Header>
          {#if notification.body}
            <Card.Content class="text-sm leading-6 text-muted-foreground">
              {notification.body}
            </Card.Content>
          {/if}
        </Card.Root>
      {/each}
    {/if}
  </div>

  <footer class="flex items-center justify-between">
    <Button
      variant="outline"
      disabled={page <= 1 || isLoading}
      onclick={() => loadNotifications(page - 1)}
    >
      {$t('common.previous')}
    </Button>
    <span class="text-sm text-muted-foreground">
      {$t('messages.page_of', { page, total: totalPages })}
    </span>
    <Button
      variant="outline"
      disabled={page >= totalPages || isLoading}
      onclick={() => loadNotifications(page + 1)}
    >
      {$t('common.next')}
    </Button>
  </footer>
</div>
