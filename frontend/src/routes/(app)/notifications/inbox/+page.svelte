<script lang="ts">
  import { onDestroy, onMount } from 'svelte';
  import NotificationActions from '$lib/components/notifications/NotificationActions.svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import * as Card from '$lib/components/ui/card/index.js';
  import { Input } from '$lib/components/ui/input/index.js';
  import NotificationInboxHeader from '$lib/components/notifications/NotificationInboxHeader.svelte';
  import { systemNotificationState } from '$lib/components/notifications/SystemNotificationState.svelte.js';
  import type { SystemNotification } from '$lib/domain/notification/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { cn } from '$lib/utils.js';
  import InboxIcon from '@lucide/svelte/icons/inbox';
  import SearchIcon from '@lucide/svelte/icons/search';
  import StarIcon from '@lucide/svelte/icons/star';
  import XIcon from '@lucide/svelte/icons/x';

  const t = createTranslator();
  const notifications = systemNotificationState;
  let selectedNotificationId = $state<string | null>(null);
  let searchQuery = $state('');

  const filteredInboxItems = $derived.by(() =>
    notifications.inboxItems.filter((notification) => notificationMatchesSearch(notification))
  );

  const selectedNotification = $derived.by(
    () =>
      filteredInboxItems.find((notification) => notification.id === selectedNotificationId) ??
      filteredInboxItems[0] ??
      null
  );

  function notificationMatchesSearch(notification: SystemNotification): boolean {
    const query = searchQuery.trim().toLocaleLowerCase('de-CH');
    if (!query) return true;

    const metadataEntries = Object.entries(notification.metadata ?? {}).flat();
    return [
      notification.title,
      notification.body,
      notification.event_key,
      notification.resource_type,
      notification.resource_id,
      ...metadataEntries
    ]
      .filter(Boolean)
      .join('\n')
      .toLocaleLowerCase('de-CH')
      .includes(query);
  }

  function selectNotification(notification: SystemNotification): void {
    selectedNotificationId = notification.id;
    void notifications.markRead(notification);
  }

  onMount(() => {
    void notifications.loadInbox(1);
  });

  onDestroy(() => {
    notifications.deactivateInbox();
  });
</script>

<svelte:head>
  <title>{$t('notifications.inbox.page_title')} | {$t('app.brand')}</title>
</svelte:head>

<div class="mx-auto flex w-full max-w-7xl flex-col gap-5">
  <NotificationInboxHeader
    title={$t('notifications.inbox.page_title')}
    description={$t('notifications.inbox.page_description')}
    backLabel={$t('common.back')}
    unreadCountLabel={$t('notifications.inbox.unread_count', { count: notifications.unreadCount })}
    unreadOnlyLabel={$t('notifications.inbox.unread_only')}
    markAllReadLabel={$t('notifications.inbox.mark_all_read')}
    refreshLabel={$t('common.refresh')}
    unreadCount={notifications.unreadCount}
    unreadOnly={notifications.inboxUnreadOnly}
    isLoading={notifications.isInboxLoading}
    onToggleUnreadOnly={() => notifications.toggleInboxUnreadOnly()}
    onMarkAllRead={() => notifications.markAllRead()}
    onRefresh={() => notifications.loadInbox()}
  />

  {#if notifications.inboxError}
    <Card.Root>
      <Card.Content class="py-4 text-sm text-destructive">{notifications.inboxError}</Card.Content>
    </Card.Root>
  {/if}

  <Card.Root
    class="grid min-h-[calc(100vh-13rem)] overflow-hidden rounded-lg border bg-background p-0 shadow-sm lg:grid-cols-[22rem_minmax(0,1fr)]"
  >
    <section class="flex min-h-96 flex-col border-b lg:min-h-0 lg:border-r lg:border-b-0">
      <div class="flex items-center justify-between gap-3 border-b bg-card px-4 py-3">
        <div class="min-w-0">
          <h2 class="truncate text-sm font-semibold">{$t('notifications.inbox.messages')}</h2>
          <p class="text-xs text-muted-foreground">
            {$t('notifications.inbox.unread_count', { count: notifications.unreadCount })}
          </p>
        </div>
        <Badge variant={notifications.inboxUnreadOnly ? 'default' : 'secondary'} class="shrink-0">
          {notifications.inboxUnreadOnly
            ? $t('notifications.inbox.unread_only')
            : $t('notifications.inbox.all_messages')}
        </Badge>
      </div>

      <div class="border-b bg-card px-4 py-3">
        <div class="relative">
          <SearchIcon
            class="pointer-events-none absolute top-1/2 left-2.5 size-4 -translate-y-1/2 text-muted-foreground"
          />
          <Input
            bind:value={searchQuery}
            type="text"
            role="searchbox"
            class="h-8 pr-9 pl-8"
            placeholder={$t('notifications.inbox.search_placeholder')}
            aria-label={$t('notifications.inbox.search_placeholder')}
          />
          {#if searchQuery}
            <Button
              variant="ghost"
              size="icon-sm"
              class="absolute top-1/2 right-0.5 size-7 -translate-y-1/2"
              aria-label={$t('notifications.inbox.clear_search')}
              onclick={() => (searchQuery = '')}
            >
              <XIcon class="size-3.5" />
            </Button>
          {/if}
        </div>
      </div>

      <div class="min-h-0 flex-1 overflow-y-auto">
        {#if notifications.isInboxLoading && notifications.inboxItems.length === 0}
          <div class="px-4 py-10 text-center text-sm text-muted-foreground">
            {$t('common.loading')}
          </div>
        {:else if notifications.inboxItems.length === 0}
          <div
            class="flex min-h-72 flex-col items-center justify-center gap-3 px-5 text-center text-sm text-muted-foreground"
          >
            <InboxIcon class="size-9" />
            <span>{$t('notifications.inbox.empty')}</span>
          </div>
        {:else if filteredInboxItems.length === 0}
          <div
            class="flex min-h-72 flex-col items-center justify-center gap-3 px-5 text-center text-sm text-muted-foreground"
          >
            <SearchIcon class="size-9" />
            <span>{$t('notifications.inbox.no_search_results')}</span>
          </div>
        {:else}
          {#each filteredInboxItems as notification (notification.id)}
            {@const isSelected = selectedNotification?.id === notification.id}
            <div class="group/notification relative border-b">
              <Button
                type="button"
                variant="ghost"
                class={cn(
                  'h-auto w-full items-start justify-start gap-3 rounded-none px-4 py-3 text-left font-normal whitespace-normal transition-colors hover:bg-muted/70 focus-visible:bg-muted/80 focus-visible:ring-2 focus-visible:ring-ring/50 focus-visible:outline-none',
                  !notification.read_at && 'bg-primary/5',
                  isSelected && 'bg-muted'
                )}
                aria-pressed={isSelected}
                onclick={() => selectNotification(notification)}
              >
                <span
                  class={cn(
                    'mt-2 size-2 shrink-0 rounded-full',
                    notification.read_at ? 'bg-muted-foreground/30' : 'bg-primary'
                  )}
                ></span>
                <span class="min-w-0 flex-1 space-y-1">
                  <span class="flex min-w-0 items-start justify-between gap-2">
                    <span class="flex min-w-0 items-center gap-1.5">
                      {#if notification.is_important}
                        <StarIcon class="size-3.5 shrink-0 fill-current text-warning" />
                      {/if}
                      <span class="truncate text-sm leading-5 font-medium">
                        {notification.title}
                      </span>
                    </span>
                    <span class="shrink-0 text-[11px] leading-5 text-muted-foreground">
                      {notifications.formatInboxDateTime(notification.created_at)}
                    </span>
                  </span>
                  {#if notification.body}
                    <span class="line-clamp-2 block pr-24 text-xs leading-5 text-muted-foreground">
                      {notification.body}
                    </span>
                  {:else}
                    <span class="line-clamp-2 block pr-24 text-xs leading-5 text-muted-foreground">
                      {notification.event_key}
                    </span>
                  {/if}
                </span>
              </Button>

              <NotificationActions
                isRead={Boolean(notification.read_at)}
                isImportant={notification.is_important}
                onToggleRead={() => notifications.toggleRead(notification)}
                onToggleImportant={() => notifications.toggleImportant(notification)}
                onDelete={() => notifications.deleteNotification(notification)}
                class="pointer-events-none absolute right-3 bottom-3 z-10 opacity-0 transition-opacity group-hover/notification:pointer-events-auto group-hover/notification:opacity-100 focus-within:pointer-events-auto focus-within:opacity-100"
              />
            </div>
          {/each}
        {/if}
      </div>

      <footer class="flex items-center justify-between gap-3 border-t bg-card p-3">
        <Button
          variant="outline"
          size="sm"
          disabled={notifications.inboxPage <= 1 || notifications.isInboxLoading}
          onclick={() => notifications.loadInbox(notifications.inboxPage - 1)}
        >
          {$t('common.previous')}
        </Button>
        <span class="min-w-0 text-center text-xs text-muted-foreground">
          {$t('messages.page_of', {
            page: notifications.inboxPage,
            total: notifications.inboxTotalPages
          })}
        </span>
        <Button
          variant="outline"
          size="sm"
          disabled={notifications.inboxPage >= notifications.inboxTotalPages ||
            notifications.isInboxLoading}
          onclick={() => notifications.loadInbox(notifications.inboxPage + 1)}
        >
          {$t('common.next')}
        </Button>
      </footer>
    </section>

    <section class="flex min-h-[32rem] flex-col">
      {#if selectedNotification}
        {@const metadataEntries = Object.entries(selectedNotification.metadata ?? {})}
        <div
          class="flex flex-col gap-4 border-b bg-card p-5 md:flex-row md:items-start md:justify-between"
        >
          <div class="min-w-0 space-y-3">
            <div class="flex flex-wrap items-center gap-2">
              <Badge variant={selectedNotification.read_at ? 'secondary' : 'default'}>
                {selectedNotification.read_at
                  ? $t('notifications.inbox.read')
                  : $t('notifications.inbox.unread')}
              </Badge>
              {#if selectedNotification.is_important}
                <Badge variant="warning">{$t('notifications.inbox.important')}</Badge>
              {/if}
            </div>
            <div class="min-w-0">
              <h2 class="text-xl font-semibold tracking-tight break-words">
                {selectedNotification.title}
              </h2>
              <p class="mt-1 text-sm text-muted-foreground">
                {notifications.formatInboxDateTime(selectedNotification.created_at)}
              </p>
            </div>
          </div>
        </div>

        <div class="min-h-0 flex-1 overflow-y-auto p-5">
          {#if selectedNotification.body}
            <p class="max-w-3xl text-sm leading-7 whitespace-pre-wrap">
              {selectedNotification.body}
            </p>
          {:else}
            <p class="text-sm text-muted-foreground">{$t('notifications.inbox.no_body')}</p>
          {/if}
        </div>

        <div class="grid gap-4 border-t bg-card p-5 text-sm sm:grid-cols-2">
          <div class="min-w-0">
            <span class="text-xs font-medium text-muted-foreground">
              {$t('notifications.inbox.event')}
            </span>
            <p class="mt-1 break-words">{selectedNotification.event_key}</p>
          </div>
          <div class="min-w-0">
            <span class="text-xs font-medium text-muted-foreground">
              {$t('notifications.inbox.resource')}
            </span>
            <p class="mt-1 break-words">
              {#if selectedNotification.resource_type || selectedNotification.resource_id}
                {selectedNotification.resource_type || $t('common.not_available')}
                {#if selectedNotification.resource_id}
                  · {selectedNotification.resource_id}
                {/if}
              {:else}
                {$t('common.not_available')}
              {/if}
            </p>
          </div>

          {#if metadataEntries.length > 0}
            <div class="min-w-0 sm:col-span-2">
              <span class="text-xs font-medium text-muted-foreground">
                {$t('notifications.inbox.metadata')}
              </span>
              <dl class="mt-2 grid gap-2 sm:grid-cols-2">
                {#each metadataEntries as [key, value]}
                  <div class="min-w-0 rounded-md border bg-background px-3 py-2">
                    <dt class="truncate text-xs text-muted-foreground">{key}</dt>
                    <dd class="mt-1 text-sm break-words">{value}</dd>
                  </div>
                {/each}
              </dl>
            </div>
          {/if}
        </div>
      {:else}
        <div
          class="flex flex-1 flex-col items-center justify-center gap-3 px-6 text-center text-sm text-muted-foreground"
        >
          <InboxIcon class="size-10" />
          <p>{$t('notifications.inbox.no_selection')}</p>
        </div>
      {/if}
    </section>
  </Card.Root>
</div>
