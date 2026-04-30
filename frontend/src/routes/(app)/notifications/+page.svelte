<script lang="ts">
  import ModuleCardGrid, {
    type ModuleCardItem
  } from '$lib/components/navigation/ModuleCardGrid.svelte';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator.js';
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
  <EntityListHeader
    title={$t('hub.notifications.title')}
    description={$t('hub.notifications.description')}
    infoLabel={$t('common.info')}
    backHref="/"
    backLabel={$t('hub.back_to_dashboard')}
  />

  <ModuleCardGrid items={notificationCards} emptyMessage={$t('hub.no_access')} />
</div>
