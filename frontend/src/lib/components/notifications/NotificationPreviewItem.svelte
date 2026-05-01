<script lang="ts">
  import NotificationActions from '$lib/components/notifications/NotificationActions.svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import type { SystemNotification } from '$lib/domain/notification/index.js';
  import StarIcon from '@lucide/svelte/icons/star';

  interface Props {
    notification: SystemNotification;
    dateLabel: string;
    onOpen: (notification: SystemNotification) => void | Promise<void>;
    onToggleRead: (notification: SystemNotification) => void | Promise<void>;
    onToggleImportant: (notification: SystemNotification) => void | Promise<void>;
    onDelete: (notification: SystemNotification) => void | Promise<void>;
  }

  let { notification, dateLabel, onOpen, onToggleRead, onToggleImportant, onDelete }: Props =
    $props();
</script>

<div class="group/notification relative border-b">
  <Button
    type="button"
    variant="ghost"
    class="h-auto w-full items-start justify-start gap-3 rounded-none px-3 py-3 pr-20 text-left text-sm whitespace-normal transition-colors hover:bg-muted/60"
    onclick={() => onOpen(notification)}
  >
    <span
      class={`mt-1 size-2 shrink-0 rounded-full ${notification.read_at ? 'bg-muted' : 'bg-primary'}`}
    ></span>
    <span class="min-w-0 flex-1 space-y-1">
      <span class="flex min-w-0 items-center gap-1.5">
        {#if notification.is_important}
          <StarIcon class="size-3.5 shrink-0 fill-current text-warning" />
        {/if}
        <span class="block truncate leading-5 font-medium">{notification.title}</span>
      </span>
      {#if notification.body}
        <span class="line-clamp-2 block leading-4 text-muted-foreground">
          {notification.body}
        </span>
      {/if}
      <span class="block leading-4 text-muted-foreground">
        {dateLabel}
      </span>
    </span>
  </Button>

  <NotificationActions
    isRead={Boolean(notification.read_at)}
    isImportant={notification.is_important}
    onToggleRead={() => onToggleRead(notification)}
    onToggleImportant={() => onToggleImportant(notification)}
    onDelete={() => onDelete(notification)}
    class="absolute top-2 right-2 opacity-0 transition-opacity group-hover/notification:opacity-100 focus-within:opacity-100"
  />
</div>
