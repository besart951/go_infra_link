<script lang="ts">
  import {
    DashboardLastProjectCard,
    DashboardOnlineUsersCard,
    DashboardTeamCard
  } from '$lib/components/dashboard/index.js';
  import ModuleCardGrid, {
    type ModuleCardItem
  } from '$lib/components/navigation/ModuleCardGrid.svelte';
  import type { DashboardSnapshot } from '$lib/domain/dashboard/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';
  import { auth } from '$lib/stores/auth.svelte.js';
  import { canPerform } from '$lib/utils/permissions.js';
  import BellRingIcon from '@lucide/svelte/icons/bell-ring';
  import Building2Icon from '@lucide/svelte/icons/building-2';
  import FolderKanbanIcon from '@lucide/svelte/icons/folder-kanban';
  import UsersIcon from '@lucide/svelte/icons/users';

  let { data }: { data: { dashboard: DashboardSnapshot | null; loadError: string | null } } =
    $props();

  const t = createTranslator();

  const hasFacilityAccess = () =>
    [
      'building',
      'controlcabinet',
      'spscontroller',
      'fielddevice',
      'systemtype',
      'systempart',
      'apparat',
      'objectdata',
      'statetext',
      'alarmdefinition',
      'alarmtype',
      'notificationclass',
      'specification'
    ].some((resource) => canPerform('read', resource));

  const dashboardCards = $derived.by<ModuleCardItem[]>(() =>
    [
      {
        title: $t('hub.users.title'),
        description: $t('hub.users.dashboard_desc'),
        href: '/users',
        icon: UsersIcon,
        tone: 'user',
        hasAccess:
          auth.canAccessUserDirectory || canPerform('read', 'team') || canPerform('read', 'role')
      },
      {
        title: $t('navigation.facility'),
        description: $t('hub.facility.dashboard_desc'),
        href: '/facility',
        icon: Building2Icon,
        tone: 'facility',
        hasAccess: hasFacilityAccess()
      },
      {
        title: $t('hub.projects.title'),
        description: $t('hub.projects.dashboard_desc'),
        href: '/projects',
        icon: FolderKanbanIcon,
        tone: 'project',
        hasAccess: true
      },
      {
        title: $t('hub.notifications.title'),
        description: $t('hub.notifications.dashboard_desc'),
        href: '/notifications',
        icon: BellRingIcon,
        tone: 'notification',
        hasAccess: true
      }
    ].filter((item) => item.hasAccess)
  );
</script>

<div class="flex flex-col gap-6">
  <section class="space-y-3">
    <div class="space-y-1">
      <h1 class="text-2xl font-semibold tracking-tight">{$t('navigation.dashboard')}</h1>
      <p class="text-sm text-muted-foreground">{$t('hub.dashboard_description')}</p>
    </div>
    <ModuleCardGrid items={dashboardCards} />
  </section>

  {#if data.loadError}
    <div
      class="rounded-md border border-destructive/30 bg-destructive/10 p-4 text-sm text-destructive"
    >
      {data.loadError}
    </div>
  {:else if data.dashboard}
    <div class="grid gap-4 md:grid-cols-2 xl:grid-cols-3">
      <DashboardLastProjectCard project={data.dashboard.last_project} />
      <DashboardOnlineUsersCard users={data.dashboard.online_users} />
      <DashboardTeamCard primaryTeam={data.dashboard.primary_team} teams={data.dashboard.teams} />
    </div>
  {:else}
    <div class="rounded-md border bg-muted/40 p-4 text-sm text-muted-foreground">
      {$t('dashboard.unavailable')}
    </div>
  {/if}
</div>
