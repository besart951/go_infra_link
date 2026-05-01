<script lang="ts">
  import { buttonVariants } from '$lib/components/ui/button/index.js';
  import * as ButtonGroup from '$lib/components/ui/button-group/index.js';
  import * as Tooltip from '$lib/components/ui/tooltip/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import CheckIcon from '@lucide/svelte/icons/check';
  import MailIcon from '@lucide/svelte/icons/mail';
  import StarIcon from '@lucide/svelte/icons/star';
  import TrashIcon from '@lucide/svelte/icons/trash-2';

  interface Props {
    isRead: boolean;
    isImportant: boolean;
    onToggleRead: () => void | Promise<void>;
    onToggleImportant: () => void | Promise<void>;
    onDelete: () => void | Promise<void>;
    class?: string;
  }

  let {
    isRead,
    isImportant,
    onToggleRead,
    onToggleImportant,
    onDelete,
    class: className
  }: Props = $props();

  const t = createTranslator();
</script>

<Tooltip.Provider>
  <ButtonGroup.Root class={className}>
    <Tooltip.Root>
      <Tooltip.Trigger
        class={buttonVariants({ variant: isRead ? 'outline' : 'secondary', size: 'icon-sm' })}
        aria-label={isRead
          ? $t('notifications.inbox.mark_unread')
          : $t('notifications.inbox.mark_read')}
        onclick={onToggleRead}
      >
        {#if isRead}
          <MailIcon class="size-3.5" />
        {:else}
          <CheckIcon class="size-3.5" />
        {/if}
      </Tooltip.Trigger>
      <Tooltip.Content>
        {isRead ? $t('notifications.inbox.mark_unread') : $t('notifications.inbox.mark_read')}
      </Tooltip.Content>
    </Tooltip.Root>

    <Tooltip.Root>
      <Tooltip.Trigger
        class={buttonVariants({ variant: isImportant ? 'secondary' : 'outline', size: 'icon-sm' })}
        aria-label={isImportant
          ? $t('notifications.inbox.unmark_important')
          : $t('notifications.inbox.mark_important')}
        onclick={onToggleImportant}
      >
        <StarIcon class={`size-3.5${isImportant ? ' fill-current' : ''}`} />
      </Tooltip.Trigger>
      <Tooltip.Content>
        {isImportant
          ? $t('notifications.inbox.unmark_important')
          : $t('notifications.inbox.mark_important')}
      </Tooltip.Content>
    </Tooltip.Root>

    <Tooltip.Root>
      <Tooltip.Trigger
        class={buttonVariants({ variant: 'outline', size: 'icon-sm' })}
        aria-label={$t('notifications.inbox.delete')}
        onclick={onDelete}
      >
        <TrashIcon class="size-3.5" />
      </Tooltip.Trigger>
      <Tooltip.Content>{$t('notifications.inbox.delete')}</Tooltip.Content>
    </Tooltip.Root>
  </ButtonGroup.Root>
</Tooltip.Provider>
