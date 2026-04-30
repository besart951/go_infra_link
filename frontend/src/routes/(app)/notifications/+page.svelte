<script lang="ts">
  import ModuleCardGrid, {
    type ModuleCardItem
  } from '$lib/components/navigation/ModuleCardGrid.svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
  import InboxIcon from '@lucide/svelte/icons/inbox';
  import ServerCogIcon from '@lucide/svelte/icons/server-cog';

  const t = createTranslator();

  const notificationCards = $derived.by<ModuleCardItem[]>(() =>
    [
      {
        title: $t('notifications.inbox.page_title'),
        description: $t('notifications.inbox.page_description'),
        href: '/notifications/inbox',
        icon: InboxIcon,
        tone: 'notification',
        hasAccess: true
      },
      {
        title: $t('notifications.page.title'),
        description: $t('hub.notifications.smtp_desc'),
        href: '/admin/notifications/smtp',
        icon: ServerCogIcon,
        tone: 'notification',
        hasAccess: canPerform('manage', 'notification.smtp')
      }
    ].filter((item) => item.hasAccess)
  );
</script>

<svelte:head>
  <title>{$t('navigation.notifications')} | {$t('app.brand')}</title>
</svelte:head>

<div class="flex flex-col gap-6">
  <header class="flex flex-col gap-4 border-b pb-5 sm:flex-row sm:items-end sm:justify-between">
    <div class="min-w-0 space-y-1">
      <h1 class="text-2xl font-semibold tracking-tight sm:text-3xl">
        {$t('hub.notifications.title')}
      </h1>
      <p class="max-w-3xl text-sm leading-6 text-muted-foreground">
        {$t('hub.notifications.description')}
      </p>
    </div>
    <Button variant="outline" href="/" class="w-full sm:w-auto">
      <ArrowLeftIcon class="size-4" />
      {$t('hub.back_to_dashboard')}
    </Button>
  </header>

  <ModuleCardGrid items={notificationCards} emptyMessage={$t('hub.no_access')} />
</div>
