<script lang="ts">
  type ProjectItem = {
    id: string;
    name: string;
    url: string;
  };

  let { projects = [], onCreate } = $props<{
    projects?: ProjectItem[];
    onCreate?: () => void;
  }>();
</script>

<div data-testid="nav-projects">
  {#each projects as project}
    <a href={project.url} data-testid={`project-link:${project.id}`}>{project.name}</a>
  {/each}
  {#if onCreate}
    <button type="button" data-testid="project-create" onclick={onCreate}>
      navigation.new_project
    </button>
  {/if}
</div>
