<script lang="ts">
	import * as Sidebar from '$lib/components/ui/sidebar/index.js';
	import { NavMain, NavProjects, NavUser, TeamSwitcher } from '$lib/components/sidebar/index.js';
	import { page } from '$app/stores';
	import { goto } from '$app/navigation';
	import type { User } from '$lib/domain/user/index.js';
	import type { Team } from '$lib/domain/team/index.js';
	import type { Project } from '$lib/domain/project/index.js';
	import { createTranslator } from '$lib/i18n/translator';

	// Icons
	import {UsersIcon, Building2Icon, FolderKanbanIcon, SheetIcon} from '@lucide/svelte';

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

	import { canPerform } from '$lib/utils/permissions.js';

	// Navigation items with collapsible sub-menus
	const navItems = $derived.by(() => {
		const items = [
			{
				title: $t('navigation.users'),
				url: '/users',
				icon: UsersIcon,
				isActive:
					$page.url.pathname.startsWith('/users') ||
					$page.url.pathname.startsWith('/auth') ||
					$page.url.pathname.startsWith('/teams'),
				items: [
					{ title: $t('navigation.all_users'), url: '/users', hasAccess: canPerform('read', 'user') },
					{ title: $t('navigation.teams'), url: '/teams', hasAccess: canPerform('read', 'team') },
					{ title: $t('navigation.roles_permissions'), url: '/users/roles', hasAccess: canPerform('read', 'role') }
				].filter(item => item.hasAccess)
			},
			{
				title: $t('navigation.facility'),
				url: '/facility',
				icon: Building2Icon,
				isActive: $page.url.pathname.startsWith('/facility'),
				items: [
					{ title: $t('navigation.buildings'), url: '/facility/buildings', hasAccess: canPerform('read', 'building') },
					{ title: $t('navigation.control_cabinets'), url: '/facility/control-cabinets', hasAccess: canPerform('read', 'controlcabinet') },
					{ title: $t('navigation.sps_controllers'), url: '/facility/sps-controllers', hasAccess: canPerform('read', 'spscontroller') },
					{ title: $t('navigation.field_devices'), url: '/facility/field-devices', hasAccess: canPerform('read', 'fielddevice') },
					{ title: $t('navigation.system_types'), url: '/facility/system-types', hasAccess: canPerform('read', 'systemtype') },
					{ title: $t('navigation.system_parts'), url: '/facility/system-parts', hasAccess: canPerform('read', 'systempart') },
					{ title: $t('navigation.apparats'), url: '/facility/apparats', hasAccess: canPerform('read', 'apparat') },
					{ title: $t('navigation.object_data'), url: '/facility/object-data', hasAccess: canPerform('read', 'objectdata') },
					{ title: $t('navigation.state_texts'), url: '/facility/state-texts', hasAccess: canPerform('read', 'statetext') },
					{ title: $t('navigation.alarm_definitions'), url: '/facility/alarm-definitions', hasAccess: canPerform('read', 'alarmdefinition') },
					{ title: $t('navigation.alarm_catalog'), url: '/facility/alarm-catalog', hasAccess: canPerform('read', 'alarmtype') },
					{ title: $t('navigation.notification_classes'), url: '/facility/notification-classes', hasAccess: canPerform('read', 'notificationclass') }
				].filter(item => item.hasAccess)
			},
			{
				title: $t('navigation.projects'),
				url: '/projects',
				icon: FolderKanbanIcon,
				isActive: $page.url.pathname.startsWith('/projects'),
				items: [
					{ title: $t('navigation.projects'), url: '/projects', hasAccess: canPerform('read', 'project') },
					{ title: $t('phase.phases'), url: '/projects/phases', hasAccess: canPerform('read', 'phase') }
				].filter(item => item.hasAccess)
			},
			{
				title: $t('navigation.excel_importer'),
				url: '/excel',
				icon: SheetIcon,
				isActive: $page.url.pathname.startsWith('/excel'),
				hasAccess: canPerform('read', 'objectdata')
			}
		];

		return items.filter(group => {
			if (group.items !== undefined) {
				return group.items.length > 0;
			}
			return group.hasAccess !== false;
		});
	});

	// Transform projects for NavProjects component
	const projectItems = $derived(
		Array.isArray(projects) && canPerform('read', 'project')
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
		goto('/projects');
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
