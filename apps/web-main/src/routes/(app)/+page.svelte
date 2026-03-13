<script lang="ts">
  import {
    DashboardLastProjectCard,
    DashboardOnlineUsersCard,
    DashboardTeamCard
  } from '$lib/components/dashboard/index.js';
  import type { DashboardSnapshot } from '$lib/domain/dashboard/index.js';
  import { createTranslator } from '@i18n/translator.js';

  let { data }: { data: { dashboard: DashboardSnapshot | null; loadError: string | null } } =
    $props();

  const t = createTranslator();
</script>

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
