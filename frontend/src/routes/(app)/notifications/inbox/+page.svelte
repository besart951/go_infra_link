<script lang="ts">
  import { onMount } from 'svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import { NotificationInboxPageState } from '$lib/components/notifications/NotificationInboxPageState.svelte.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
  import BellIcon from '@lucide/svelte/icons/bell';
  import CheckIcon from '@lucide/svelte/icons/check';
  import RefreshCwIcon from '@lucide/svelte/icons/refresh-cw';

  const t = createTranslator();
  const state = new NotificationInboxPageState();

  onMount(() => {
    void state.loadNotifications(1);
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
      <Button variant="outline" href="/notifications">
        <ArrowLeftIcon class="size-4" />
        {$t('hub.back_to_overview')}
      </Button>
      <Badge variant={state.unreadCount > 0 ? 'default' : 'secondary'}>
        {$t('notifications.inbox.unread_count', { count: state.unreadCount })}
      </Badge>
      <Button
        variant={state.unreadOnly ? 'default' : 'outline'}
        onclick={() => state.toggleUnreadOnly()}
      >
        {$t('notifications.inbox.unread_only')}
      </Button>
      <Button
        variant="outline"
        onclick={() => state.markAllRead()}
        disabled={state.unreadCount === 0}
      >
        <CheckIcon class="size-4" />
        {$t('notifications.inbox.mark_all_read')}
      </Button>
      <Button
        variant="outline"
        onclick={() => state.loadNotifications()}
        disabled={state.isLoading}
      >
        <RefreshCwIcon class={`size-4${state.isLoading ? ' animate-spin' : ''}`} />
        {$t('common.refresh')}
      </Button>
    </div>
  </header>

  {#if state.error}
    <Card.Root>
      <Card.Content class="py-4 text-sm text-destructive">{state.error}</Card.Content>
    </Card.Root>
  {/if}

  <div class="flex flex-col gap-3">
    {#if state.isLoading && state.notifications.length === 0}
      <Card.Root>
        <Card.Content class="py-8 text-center text-sm text-muted-foreground">
          {$t('common.loading')}
        </Card.Content>
      </Card.Root>
    {:else if state.notifications.length === 0}
      <Card.Root>
        <Card.Content class="py-10 text-center text-sm text-muted-foreground">
          {$t('notifications.inbox.empty')}
        </Card.Content>
      </Card.Root>
    {:else}
      {#each state.notifications as notification (notification.id)}
        <Card.Root class={notification.read_at ? '' : 'border-primary/40'}>
          <Card.Header class="gap-2">
            <div class="flex flex-col gap-2 sm:flex-row sm:items-start sm:justify-between">
              <div class="min-w-0">
                <Card.Title class="text-base leading-6">{notification.title}</Card.Title>
                <Card.Description>
                  {state.formatDateTime(notification.created_at)} · {notification.event_key}
                </Card.Description>
              </div>
              {#if !notification.read_at}
                <Button variant="outline" size="sm" onclick={() => state.markRead(notification)}>
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
      disabled={state.page <= 1 || state.isLoading}
      onclick={() => state.loadNotifications(state.page - 1)}
    >
      {$t('common.previous')}
    </Button>
    <span class="text-sm text-muted-foreground">
      {$t('messages.page_of', { page: state.page, total: state.totalPages })}
    </span>
    <Button
      variant="outline"
      disabled={state.page >= state.totalPages || state.isLoading}
      onclick={() => state.loadNotifications(state.page + 1)}
    >
      {$t('common.next')}
    </Button>
  </footer>
</div>
