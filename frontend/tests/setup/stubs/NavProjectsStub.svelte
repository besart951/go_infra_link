<script lang="ts">
  type ProjectItem = {
    id: string;
    name: string;
    url: string;
  };

  let { projects = [] } = $props<{ projects?: ProjectItem[] }>();
</script>

<div data-testid="nav-projects">
  {#each projects as project}
    <a href={project.url} data-testid={`project-link:${project.id}`}>{project.name}</a>
  {/each}
</div>
