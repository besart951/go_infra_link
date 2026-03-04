<script lang="ts">
  import * as Card from '$lib/components/ui/card/index.js';
  import { Badge } from '$lib/components/ui/badge/index.js';
  import { Separator } from '$lib/components/ui/separator/index.js';
  import UserAvatar from '$lib/components/user-avatar.svelte';
  import type { DashboardTeam, DashboardTeamSummary } from '$lib/domain/dashboard/index.js';
  import { createTranslator } from '$lib/i18n/translator.js';

  type Props = {
    primaryTeam?: DashboardTeam;
    teams: DashboardTeamSummary[];
  };

  let { primaryTeam, teams }: Props = $props();
  const t = createTranslator();

  const otherTeams = $derived(
    primaryTeam ? teams.filter((team) => team.id !== primaryTeam.id) : teams
  );
</script>

<Card.Root>
  <Card.Header>
    <Card.Title>{$t('dashboard.team_title')}</Card.Title>
    <Card.Description>{$t('dashboard.team_desc')}</Card.Description>
  </Card.Header>
  <Card.Content>
    {#if primaryTeam}
      <div class="space-y-4">
        <div class="flex items-center justify-between gap-3">
          <div>
            <p class="text-base font-semibold">{primaryTeam.name}</p>
            <p class="text-xs text-muted-foreground">{$t('dashboard.your_role')}: {primaryTeam.role}</p>
          </div>
          <Badge variant="outline">{$t('dashboard.members_count', { count: primaryTeam.members.length })}</Badge>
        </div>

        <div class="space-y-2">
          {#each primaryTeam.members as member (member.user_id)}
            <div class="flex items-center justify-between gap-2">
              <div class="flex min-w-0 items-center gap-2">
                <UserAvatar firstName={member.first_name} lastName={member.last_name} />
                <div class="min-w-0">
                  <p class="truncate text-sm">{member.first_name} {member.last_name}</p>
                  <p class="truncate text-xs text-muted-foreground">{member.role}</p>
                </div>
              </div>
              {#if member.is_online}
                <Badge variant="success">{$t('dashboard.online_badge')}</Badge>
              {/if}
            </div>
          {/each}
        </div>

        {#if otherTeams.length > 0}
          <Separator />
          <div>
            <p class="mb-2 text-xs font-medium uppercase tracking-wide text-muted-foreground">
              {$t('dashboard.other_memberships')}
            </p>
            <div class="flex flex-wrap gap-2">
              {#each otherTeams as team (team.id)}
                <Badge variant="secondary">{team.name} ({team.role})</Badge>
              {/each}
            </div>
          </div>
        {/if}
      </div>
    {:else}
      <p class="text-sm text-muted-foreground">{$t('dashboard.no_team')}</p>
    {/if}
  </Card.Content>
</Card.Root>
