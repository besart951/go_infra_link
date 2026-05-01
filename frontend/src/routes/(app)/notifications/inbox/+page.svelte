<script lang="ts">
  import { onMount } from 'svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import NotificationInboxCard from '$lib/components/notifications/NotificationInboxCard.svelte';
  import NotificationInboxHeader from '$lib/components/notifications/NotificationInboxHeader.svelte';
  import { NotificationInboxPageState } from '$lib/components/notifications/NotificationInboxPageState.svelte.js';
  import { createTranslator } from '$lib/i18n/translator.js';

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
        <NotificationInboxCard
          {notification}
          dateLabel={state.formatDateTime(notification.created_at)}
          onToggleRead={(item) => state.toggleRead(item)}
          onToggleImportant={(item) => state.toggleImportant(item)}
          onDelete={(item) => state.deleteNotification(item)}
        />
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
