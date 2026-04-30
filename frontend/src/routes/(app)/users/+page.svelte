<script lang="ts">
  import ModuleCardGrid, {
    type ModuleCardItem
  } from '$lib/components/navigation/ModuleCardGrid.svelte';
  import { Button } from '$lib/components/ui/button/index.js';
  import { auth } from '$lib/stores/auth.svelte.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import ArrowLeftIcon from '@lucide/svelte/icons/arrow-left';
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
  <header class="flex flex-col gap-4 border-b pb-5 sm:flex-row sm:items-end sm:justify-between">
    <div class="min-w-0 space-y-1">
      <h1 class="text-2xl font-semibold tracking-tight sm:text-3xl">
        {$t('hub.users.title')}
      </h1>
      <p class="max-w-3xl text-sm leading-6 text-muted-foreground">
        {$t('hub.users.description')}
      </p>
    </div>
    <Button variant="outline" href="/" class="w-full sm:w-auto">
      <ArrowLeftIcon class="size-4" />
      {$t('hub.back_to_dashboard')}
    </Button>
  </header>

  <ModuleCardGrid items={userCards} emptyMessage={$t('hub.no_access')} />
</div>
