<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { NavMain, NavProjects, NavUser, TeamSwitcher } from '$lib/components/sidebar/index.js';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import type { User } from '$lib/domain/user/index.js';
	import type { Team } from '$lib/domain/team/index.js';
	import type { Project } from '$lib/domain/project/index.js';

	// Icons
	import UsersIcon from '@lucide/svelte/icons/users';
	import Building2Icon from '@lucide/svelte/icons/building-2';
	import FolderKanbanIcon from '@lucide/svelte/icons/folder-kanban';

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

	// Navigation items with collapsible sub-menus
	const navItems = $derived([
		{
			title: 'Users',
			url: '/users',
			icon: UsersIcon,
			isActive:
				$page.url.pathname.startsWith('/users') ||
				$page.url.pathname.startsWith('/auth') ||
				$page.url.pathname.startsWith('/teams'),
			items: [
				{ title: 'All Users', url: '/users' },
				{ title: 'Teams', url: '/teams' },
				{ title: 'Roles & Permissions', url: '/users/roles' }
			]
		},
		{
			title: 'Facility',
			url: '/facility',
			icon: Building2Icon,
			isActive: $page.url.pathname.startsWith('/facility'),
			items: [
				{ title: 'Buildings', url: '/facility/buildings' },
				{ title: 'Control Cabinets', url: '/facility/control-cabinets' },
				{ title: 'SPS Controllers', url: '/facility/sps-controllers' },
				{ title: 'Field Devices', url: '/facility/field-devices' }
			]
		},
		{
			title: 'Projects',
			url: '/projects',
			icon: FolderKanbanIcon,
			isActive: $page.url.pathname.startsWith('/projects')
		}
	]);

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
		goto('/projects/new');
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
