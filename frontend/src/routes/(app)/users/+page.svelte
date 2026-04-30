<script lang="ts">
  import ModuleCardGrid, {
    type ModuleCardItem
  } from '$lib/components/navigation/ModuleCardGrid.svelte';
  import EntityListHeader from '$lib/components/layout/EntityListHeader.svelte';
  import { auth } from '$lib/stores/auth.svelte.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import ShieldCheckIcon from '@lucide/svelte/icons/shield-check';
  import UserRoundCogIcon from '@lucide/svelte/icons/user-round-cog';
  import UsersIcon from '@lucide/svelte/icons/users';

  const t = createTranslator();

  const userCards = $derived.by<ModuleCardItem[]>(() =>
    [
      {
        title: $t('hub.users.directory_title'),
        description: $t('hub.users.directory_desc'),
        href: '/users/directory',
        icon: UsersIcon,
        tone: 'user',
        hasAccess: auth.canAccessUserDirectory
      },
      {
        title: $t('navigation.teams'),
        description: $t('hub.users.teams_desc'),
        href: '/teams',
        icon: UserRoundCogIcon,
        tone: 'user',
        hasAccess: canPerform('read', 'team')
      },
      {
        title: $t('navigation.roles_permissions'),
        description: $t('hub.users.roles_desc'),
        href: '/users/roles',
        icon: ShieldCheckIcon,
        tone: 'user',
        hasAccess: canPerform('read', 'role')
      }
    ].filter((item) => item.hasAccess)
  );
</script>

<svelte:head>
  <title>{$t('navigation.users')} | Infra Link</title>
</svelte:head>

<div class="flex flex-col gap-6">
  <EntityListHeader
    title={$t('hub.users.title')}
    description={$t('hub.users.description')}
    infoLabel={$t('common.info')}
    backHref="/"
    backLabel={$t('hub.back_to_dashboard')}
  />

  <ModuleCardGrid items={userCards} emptyMessage={$t('hub.no_access')} />
</div>
