<script lang="ts">
  let { teams = [], onCreateTeam } = $props<{
    teams?: Array<{ name: string }>;
    onCreateTeam?: () => void;
  }>();
</script>

<div data-testid="team-switcher">
  {#each teams as team}
    <span>{team.name}</span>
  {/each}
  {#if onCreateTeam}
    <button type="button" data-testid="team-switcher-create" onclick={onCreateTeam}>
      sidebar.team_switcher.add_team
    </button>
  {/if}
</div>
