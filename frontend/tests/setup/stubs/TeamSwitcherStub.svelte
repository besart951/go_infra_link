<script lang="ts">
  let { teams = [] } = $props<{ teams?: Array<{ name: string }> }>();
</script>

<div data-testid="team-switcher">
  {#each teams as team}
    <span>{team.name}</span>
  {/each}
</div>
