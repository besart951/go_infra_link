<script lang="ts">
  import * as Sidebar from '$lib/components/ui/sidebar/index.js';
  import { NavMain, NavProjects, NavUser, TeamSwitcher } from '$lib/components/sidebar/index.js';
  import { page } from '$app/stores';
  import { goto } from '$app/navigation';
  import { buildAppNavItems } from '$lib/navigation/appNavigation.js';
  import type { User } from '$lib/domain/user/index.js';
  import type { Team } from '$lib/domain/team/index.js';
  import type { Project } from '$lib/domain/project/index.js';
  import { createTranslator } from '$lib/i18n/translator';
  import { canPerform } from '$lib/utils/permissions.js';

  const t = createTranslator();

  // Props from layout
  let {
    user,
    teams = [],
    projects = []
  }: {
    user: User;
    teams?: Team[];
    projects?: Project[];
  } = $props();

  // Active team state - use $state with effect to update when teams change
  let activeTeam = $state<Team | undefined>(undefined);
  $effect(() => {
    if (teams && teams.length > 0 && !activeTeam) {
      activeTeam = teams[0];
    }
  });

  const navItems = $derived.by(() => {
    return buildAppNavItems({
      pathname: $page.url.pathname,
      user,
      translate: (key) => $t(key),
      canPerform
    });
  });

  // Transform projects for NavProjects component
  const projectItems = $derived(
    Array.isArray(projects)
      ? projects.map((p) => ({
          id: p.id,
          name: p.name,
          url: `/projects/${p.id}`,
          status: p.status
        }))
      : []
  );

  const handleViewProject = (id: string) => {
    goto(`/projects/${id}`);
  };

  const handleShareProject = (id: string) => {
    // TODO: Implement share modal
    console.log('Share project:', id);
  };

  const handleCreateProject = () => {
    goto('/projects/list');
  };

  const handleCreateTeam = () => {
    goto('/teams/new');
  };
</script>

<Sidebar.Root collapsible="icon">
  <Sidebar.Header>
    <TeamSwitcher {teams} bind:activeTeam onCreateTeam={handleCreateTeam} />
  </Sidebar.Header>

  <Sidebar.Content>
    <NavMain items={navItems} />
    <NavProjects
      projects={projectItems}
      onViewProject={handleViewProject}
      onShareProject={handleShareProject}
      onCreate={handleCreateProject}
    />
  </Sidebar.Content>

  <Sidebar.Footer>
    <NavUser {user} />
  </Sidebar.Footer>

  <Sidebar.Rail />
</Sidebar.Root>
