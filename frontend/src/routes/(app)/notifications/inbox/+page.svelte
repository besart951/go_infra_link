<script lang="ts">
  import { onMount } from 'svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import NotificationInboxHeader from '$lib/components/notifications/NotificationInboxHeader.svelte';
  import { NotificationInboxPageState } from '$lib/components/notifications/NotificationInboxPageState.svelte.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import CheckIcon from '@lucide/svelte/icons/check';

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
  <NotificationInboxHeader
    title={$t('notifications.inbox.page_title')}
    description={$t('notifications.inbox.page_description')}
    backLabel={$t('hub.back_to_overview')}
    unreadCountLabel={$t('notifications.inbox.unread_count', { count: state.unreadCount })}
    unreadOnlyLabel={$t('notifications.inbox.unread_only')}
    markAllReadLabel={$t('notifications.inbox.mark_all_read')}
    refreshLabel={$t('common.refresh')}
    infoLabel={$t('common.info')}
    unreadCount={state.unreadCount}
    unreadOnly={state.unreadOnly}
    isLoading={state.isLoading}
    onToggleUnreadOnly={() => state.toggleUnreadOnly()}
    onMarkAllRead={() => state.markAllRead()}
    onRefresh={() => state.loadNotifications()}
  />

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
