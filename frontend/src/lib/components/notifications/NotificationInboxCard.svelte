<script lang="ts">
  import NotificationActions from '$lib/components/notifications/NotificationActions.svelte';
  import * as Card from '$lib/components/ui/card/index.js';
  import type { SystemNotification } from '$lib/domain/notification/index.js';
  import { cn } from '$lib/utils.js';
  import StarIcon from '@lucide/svelte/icons/star';

  interface Props {
    notification: SystemNotification;
    dateLabel: string;
    onToggleRead: (notification: SystemNotification) => void | Promise<void>;
    onToggleImportant: (notification: SystemNotification) => void | Promise<void>;
    onDelete: (notification: SystemNotification) => void | Promise<void>;
  }

  let { notification, dateLabel, onToggleRead, onToggleImportant, onDelete }: Props = $props();
</script>

<Card.Root
  class={cn(
    'group/notification relative',
    !notification.read_at && 'border-primary/40',
    notification.is_important && 'border-warning/50'
  )}
>
  <Card.Header class="gap-2">
    <div class="flex flex-col gap-2 pr-20 sm:flex-row sm:items-start sm:justify-between">
      <div class="min-w-0">
        <Card.Title class="flex min-w-0 items-center gap-2 text-base leading-6">
          {#if notification.is_important}
            <StarIcon class="size-4 shrink-0 fill-current text-warning" />
          {/if}
          <span class="truncate">{notification.title}</span>
        </Card.Title>
        <Card.Description>
          {dateLabel} · {notification.event_key}
        </Card.Description>
      </div>
    </div>
  </Card.Header>

  <NotificationActions
    isRead={Boolean(notification.read_at)}
    isImportant={notification.is_important}
    onToggleRead={() => onToggleRead(notification)}
    onToggleImportant={() => onToggleImportant(notification)}
    onDelete={() => onDelete(notification)}
    class="absolute top-4 right-4 opacity-0 transition-opacity group-hover/notification:opacity-100 focus-within:opacity-100"
  />

  {#if notification.body}
    <Card.Content class="text-sm leading-6 text-muted-foreground">
      {notification.body}
    </Card.Content>
  {/if}
</Card.Root>
