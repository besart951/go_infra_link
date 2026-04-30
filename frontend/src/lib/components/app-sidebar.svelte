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
  import {
    UsersIcon,
    Building2Icon,
    FolderKanbanIcon,
    SheetIcon,
    BellRingIcon
  } from '@lucide/svelte';

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

  const canReadFacility = (resource: string) => {
    return canPerform('read', resource);
  };

  const isPathActive = (pathname: string, url: string) => {
    return pathname === url || pathname.startsWith(`${url}/`);
  };

  const isProjectsIndexActive = (pathname: string) => {
    return (
      isPathActive(pathname, '/projects/list') ||
      (pathname !== '/projects' &&
        isPathActive(pathname, '/projects') &&
        !isPathActive(pathname, '/projects/phases'))
    );
  };

  // Navigation items with collapsible sub-menus
  const navItems = $derived.by(() => {
    const pathname = $page.url.pathname;
    const userSubItems = [
      {
        title: $t('navigation.all_users'),
        url: '/users/directory',
        isActive: isPathActive(pathname, '/users/directory'),
        hasAccess: Boolean(user.can_access_user_directory)
      },
      {
        title: $t('navigation.teams'),
        url: '/teams',
        isActive: isPathActive(pathname, '/teams'),
        hasAccess: canPerform('read', 'team')
      },
      {
        title: $t('navigation.roles_permissions'),
        url: '/users/roles',
        isActive: isPathActive(pathname, '/users/roles'),
        hasAccess: canPerform('read', 'role')
      }
    ].filter((item) => item.hasAccess);
    const facilitySubItems = [
      {
        title: $t('navigation.buildings'),
        url: '/facility/buildings',
        isActive: isPathActive(pathname, '/facility/buildings'),
        hasAccess: canReadFacility('building')
      },
      {
        title: $t('navigation.control_cabinets'),
        url: '/facility/control-cabinets',
        isActive: isPathActive(pathname, '/facility/control-cabinets'),
        hasAccess: canReadFacility('controlcabinet')
      },
      {
        title: $t('navigation.sps_controllers'),
        url: '/facility/sps-controllers',
        isActive: isPathActive(pathname, '/facility/sps-controllers'),
        hasAccess: canReadFacility('spscontroller')
      },
      {
        title: $t('navigation.field_devices'),
        url: '/facility/field-devices',
        isActive: isPathActive(pathname, '/facility/field-devices'),
        hasAccess: canReadFacility('fielddevice'),
        dividerAfter: true
      },
      {
        title: $t('navigation.system_types'),
        url: '/facility/system-types',
        isActive: isPathActive(pathname, '/facility/system-types'),
        hasAccess: canReadFacility('systemtype')
      },
      {
        title: $t('navigation.system_parts'),
        url: '/facility/system-parts',
        isActive: isPathActive(pathname, '/facility/system-parts'),
        hasAccess: canReadFacility('systempart')
      },
      {
        title: $t('navigation.apparats'),
        url: '/facility/apparats',
        isActive: isPathActive(pathname, '/facility/apparats'),
        hasAccess: canReadFacility('apparat')
      },
      {
        title: $t('navigation.object_data'),
        url: '/facility/object-data',
        isActive: isPathActive(pathname, '/facility/object-data'),
        hasAccess: canReadFacility('objectdata')
      },
      {
        title: $t('navigation.state_texts'),
        url: '/facility/state-texts',
        isActive: isPathActive(pathname, '/facility/state-texts'),
        hasAccess: canReadFacility('statetext')
      },
      {
        title: $t('navigation.alarm_definitions'),
        url: '/facility/alarm-definitions',
        isActive: isPathActive(pathname, '/facility/alarm-definitions'),
        hasAccess: canReadFacility('alarmdefinition')
      },
      {
        title: $t('navigation.alarm_catalog'),
        url: '/facility/alarm-catalog',
        isActive: isPathActive(pathname, '/facility/alarm-catalog'),
        hasAccess: canReadFacility('alarmtype')
      },
      {
        title: $t('navigation.notification_classes'),
        url: '/facility/notification-classes',
        isActive: isPathActive(pathname, '/facility/notification-classes'),
        hasAccess: canReadFacility('notificationclass')
      }
    ].filter((item) => item.hasAccess);
    const projectSubItems = [
      {
        title: $t('navigation.projects'),
        url: '/projects/list',
        isActive: isProjectsIndexActive(pathname),
        hasAccess: true
      },
      {
        title: $t('phase.phases'),
        url: '/projects/phases',
        isActive: isPathActive(pathname, '/projects/phases'),
        hasAccess: canPerform('read', 'phase')
      }
    ].filter((item) => item.hasAccess);
    const notificationSubItems = [
      {
        title: $t('notifications.inbox.page_title'),
        url: '/notifications/inbox',
        isActive: isPathActive(pathname, '/notifications/inbox'),
        hasAccess: true
      },
      {
        title: $t('notifications.page.title'),
        url: '/admin/notifications/smtp',
        isActive: isPathActive(pathname, '/admin/notifications'),
        hasAccess: canPerform('manage', 'notification.smtp')
      }
    ].filter((item) => item.hasAccess);
    const items = [
      {
        title: $t('navigation.users'),
        url: '/users',
        icon: UsersIcon,
        isActive:
          isPathActive(pathname, '/users') ||
          isPathActive(pathname, '/auth') ||
          isPathActive(pathname, '/teams'),
        items:
          userSubItems.length > 0
            ? [
                {
                  title: $t('hub.overview'),
                  url: '/users',
                  isActive: pathname === '/users'
                },
                ...userSubItems
              ]
            : []
      },
      {
        title: $t('navigation.facility'),
        url: '/facility',
        icon: Building2Icon,
        isActive: isPathActive(pathname, '/facility'),
        items:
          facilitySubItems.length > 0
            ? [
                {
                  title: $t('hub.overview'),
                  url: '/facility',
                  isActive: pathname === '/facility'
                },
                ...facilitySubItems
              ]
            : []
      },
      {
        title: $t('navigation.projects'),
        url: '/projects',
        icon: FolderKanbanIcon,
        isActive: isPathActive(pathname, '/projects'),
        items: [
          {
            title: $t('hub.overview'),
            url: '/projects',
            isActive: pathname === '/projects'
          },
          ...projectSubItems
        ]
      },
      {
        title: $t('navigation.excel_importer'),
        url: '/excel',
        icon: SheetIcon,
        isActive: isPathActive(pathname, '/excel'),
        hasAccess: canPerform('read', 'objectdata')
      },
      {
        title: $t('navigation.notifications'),
        url: '/notifications',
        icon: BellRingIcon,
        isActive:
          isPathActive(pathname, '/notifications') ||
          isPathActive(pathname, '/admin/notifications'),
        items: [
          {
            title: $t('hub.overview'),
            url: '/notifications',
            isActive: pathname === '/notifications'
          },
          ...notificationSubItems
        ]
      }
    ];

    return items.filter((group) => {
      if (group.items !== undefined) {
        return group.items.length > 0;
      }
      return group.hasAccess !== false;
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
