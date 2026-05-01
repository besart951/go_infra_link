import type { Component } from 'svelte';
import {
  BellRingIcon,
  Building2Icon,
  FolderKanbanIcon,
  SheetIcon,
  UsersIcon
} from '@lucide/svelte';
import type { User } from '$lib/domain/user/index.js';

type Translate = (key: string) => string;
type CanPerform = (action: string, resource: string) => boolean;

export interface AppNavSubItem {
  title: string;
  url: string;
  dividerAfter?: boolean;
  isActive?: boolean;
}

export interface AppNavItem {
  title: string;
  url: string;
  icon?: Component;
  isActive?: boolean;
  hasAccess?: boolean;
  items?: AppNavSubItem[];
}

export interface BreadcrumbItem {
  title: string;
  href: string;
}

export interface AppBreadcrumb {
  current: string;
  parent?: BreadcrumbItem;
}

interface NavContext {
  pathname: string;
  user: User;
  translate: Translate;
  canPerform: CanPerform;
}

interface RouteEntry {
  path: string;
  titleKey: string;
  parent?: {
    titleKey: string;
    href: string;
  };
}

interface NavDefinition {
  titleKey: string;
  url: string;
  icon?: Component;
  activePaths?: string[];
  children?: NavChildDefinition[];
  hideWhenOnlyOverview?: boolean;
  hasAccess?: (context: NavContext) => boolean;
}

interface NavChildDefinition {
  titleKey: string;
  url: string;
  activePaths?: string[];
  dividerAfter?: boolean;
  hasAccess?: (context: NavContext) => boolean;
}

const facilityParent = { titleKey: 'navigation.facility', href: '/facility' };
const projectParent = { titleKey: 'navigation.projects', href: '/projects' };
const userParent = { titleKey: 'navigation.users', href: '/users' };
const notificationParent = { titleKey: 'navigation.notifications', href: '/notifications' };

const breadcrumbRoutes: RouteEntry[] = [
  { path: '/users/directory', titleKey: 'navigation.all_users', parent: userParent },
  { path: '/users/roles', titleKey: 'navigation.roles_permissions', parent: userParent },
  { path: '/teams', titleKey: 'navigation.teams', parent: userParent },
  { path: '/users', titleKey: 'navigation.users' },
  { path: '/projects/list', titleKey: 'hub.projects.list_title', parent: projectParent },
  { path: '/projects/phases', titleKey: 'phase.phases', parent: projectParent },
  { path: '/projects', titleKey: 'navigation.projects' },
  {
    path: '/notifications/inbox',
    titleKey: 'notifications.inbox.page_title',
    parent: notificationParent
  },
  {
    path: '/admin/notifications',
    titleKey: 'notifications.page.title',
    parent: notificationParent
  },
  { path: '/notifications', titleKey: 'navigation.notifications' },
  { path: '/account', titleKey: 'navigation.account' },
  { path: '/errors', titleKey: 'pages.http_error.title' },
  { path: '/facility/buildings', titleKey: 'navigation.buildings', parent: facilityParent },
  {
    path: '/facility/control-cabinets',
    titleKey: 'navigation.control_cabinets',
    parent: facilityParent
  },
  {
    path: '/facility/sps-controllers',
    titleKey: 'navigation.sps_controllers',
    parent: facilityParent
  },
  {
    path: '/facility/field-devices',
    titleKey: 'navigation.field_devices',
    parent: facilityParent
  },
  { path: '/facility/system-types', titleKey: 'navigation.system_types', parent: facilityParent },
  { path: '/facility/system-parts', titleKey: 'navigation.system_parts', parent: facilityParent },
  { path: '/facility/apparats', titleKey: 'navigation.apparats', parent: facilityParent },
  { path: '/facility/object-data', titleKey: 'navigation.object_data', parent: facilityParent },
  { path: '/facility/state-texts', titleKey: 'navigation.state_texts', parent: facilityParent },
  {
    path: '/facility/alarm-definitions',
    titleKey: 'navigation.alarm_definitions',
    parent: facilityParent
  },
  { path: '/facility/alarm-catalog', titleKey: 'navigation.alarm_catalog', parent: facilityParent },
  {
    path: '/facility/notification-classes',
    titleKey: 'navigation.notification_classes',
    parent: facilityParent
  },
  { path: '/facility/specifications', titleKey: 'facility.specifications', parent: facilityParent },
  { path: '/facility', titleKey: 'navigation.facility' }
];

function canReadFacility(context: NavContext, resource: string): boolean {
  return context.canPerform('read', resource);
}

function isPathActive(pathname: string, url: string): boolean {
  return pathname === url || pathname.startsWith(`${url}/`);
}

function isProjectsIndexActive(pathname: string): boolean {
  return (
    isPathActive(pathname, '/projects/list') ||
    (pathname !== '/projects' &&
      isPathActive(pathname, '/projects') &&
      !isPathActive(pathname, '/projects/phases'))
  );
}

const navDefinitions: NavDefinition[] = [
  {
    titleKey: 'navigation.users',
    url: '/users',
    icon: UsersIcon,
    activePaths: ['/users', '/auth', '/teams'],
    hideWhenOnlyOverview: true,
    children: [
      { titleKey: 'hub.overview', url: '/users' },
      {
        titleKey: 'navigation.all_users',
        url: '/users/directory',
        hasAccess: ({ user }) => Boolean(user.can_access_user_directory)
      },
      {
        titleKey: 'navigation.teams',
        url: '/teams',
        hasAccess: ({ canPerform }) => canPerform('read', 'team')
      },
      {
        titleKey: 'navigation.roles_permissions',
        url: '/users/roles',
        hasAccess: ({ canPerform }) => canPerform('read', 'role')
      }
    ]
  },
  {
    titleKey: 'navigation.facility',
    url: '/facility',
    icon: Building2Icon,
    hideWhenOnlyOverview: true,
    children: [
      { titleKey: 'hub.overview', url: '/facility' },
      {
        titleKey: 'navigation.buildings',
        url: '/facility/buildings',
        hasAccess: (context) => canReadFacility(context, 'building')
      },
      {
        titleKey: 'navigation.control_cabinets',
        url: '/facility/control-cabinets',
        hasAccess: (context) => canReadFacility(context, 'controlcabinet')
      },
      {
        titleKey: 'navigation.sps_controllers',
        url: '/facility/sps-controllers',
        hasAccess: (context) => canReadFacility(context, 'spscontroller')
      },
      {
        titleKey: 'navigation.field_devices',
        url: '/facility/field-devices',
        dividerAfter: true,
        hasAccess: (context) => canReadFacility(context, 'fielddevice')
      },
      {
        titleKey: 'navigation.system_types',
        url: '/facility/system-types',
        hasAccess: (context) => canReadFacility(context, 'systemtype')
      },
      {
        titleKey: 'navigation.system_parts',
        url: '/facility/system-parts',
        hasAccess: (context) => canReadFacility(context, 'systempart')
      },
      {
        titleKey: 'navigation.apparats',
        url: '/facility/apparats',
        hasAccess: (context) => canReadFacility(context, 'apparat')
      },
      {
        titleKey: 'navigation.object_data',
        url: '/facility/object-data',
        hasAccess: (context) => canReadFacility(context, 'objectdata')
      },
      {
        titleKey: 'navigation.state_texts',
        url: '/facility/state-texts',
        hasAccess: (context) => canReadFacility(context, 'statetext')
      },
      {
        titleKey: 'navigation.alarm_definitions',
        url: '/facility/alarm-definitions',
        hasAccess: (context) => canReadFacility(context, 'alarmdefinition')
      },
      {
        titleKey: 'navigation.alarm_catalog',
        url: '/facility/alarm-catalog',
        hasAccess: (context) => canReadFacility(context, 'alarmtype')
      },
      {
        titleKey: 'navigation.notification_classes',
        url: '/facility/notification-classes',
        hasAccess: (context) => canReadFacility(context, 'notificationclass')
      }
    ]
  },
  {
    titleKey: 'navigation.projects',
    url: '/projects',
    icon: FolderKanbanIcon,
    children: [
      { titleKey: 'hub.overview', url: '/projects' },
      {
        titleKey: 'navigation.projects',
        url: '/projects/list',
        activePaths: ['/projects/list'],
        hasAccess: () => true
      },
      {
        titleKey: 'phase.phases',
        url: '/projects/phases',
        hasAccess: ({ canPerform }) => canPerform('read', 'phase')
      }
    ]
  },
  {
    titleKey: 'navigation.excel_importer',
    url: '/excel',
    icon: SheetIcon,
    hasAccess: ({ canPerform }) => canPerform('read', 'objectdata')
  },
  {
    titleKey: 'navigation.notifications',
    url: '/notifications',
    icon: BellRingIcon,
    activePaths: ['/notifications', '/admin/notifications'],
    children: [
      { titleKey: 'hub.overview', url: '/notifications' },
      { titleKey: 'notifications.inbox.page_title', url: '/notifications/inbox' },
      {
        titleKey: 'notifications.page.title',
        url: '/admin/notifications/smtp',
        activePaths: ['/admin/notifications'],
        hasAccess: ({ canPerform }) => canPerform('manage', 'notification.smtp')
      }
    ]
  }
];

function routeMatches(pathname: string, routePath: string): boolean {
  return pathname === routePath || pathname.startsWith(`${routePath}/`);
}

function childIsActive(pathname: string, child: NavChildDefinition): boolean {
  if (child.url === '/projects/list') {
    return isProjectsIndexActive(pathname);
  }

  return (child.activePaths ?? [child.url]).some((path) => isPathActive(pathname, path));
}

function itemIsActive(pathname: string, item: NavDefinition): boolean {
  return (item.activePaths ?? [item.url]).some((path) => isPathActive(pathname, path));
}

export function getBreadcrumbForPath(pathname: string, translate: Translate): AppBreadcrumb {
  if (pathname === '/') {
    return { current: translate('navigation.dashboard') };
  }

  const match = breadcrumbRoutes.find((route) => routeMatches(pathname, route.path));
  if (!match) {
    return { current: translate('navigation.app') };
  }

  return {
    current: translate(match.titleKey),
    parent: match.parent
      ? {
          title: translate(match.parent.titleKey),
          href: match.parent.href
        }
      : undefined
  };
}

export function buildAppNavItems(context: NavContext): AppNavItem[] {
  return navDefinitions
    .map((definition) => {
      const children = definition.children
        ?.filter((child) => child.hasAccess?.(context) ?? true)
        .map((child) => ({
          title: context.translate(child.titleKey),
          url: child.url,
          dividerAfter: child.dividerAfter,
          isActive: childIsActive(context.pathname, child)
        }));

      const visibleChildren =
        definition.hideWhenOnlyOverview &&
        children?.length === 1 &&
        children[0]?.url === definition.url
          ? []
          : children;

      return {
        title: context.translate(definition.titleKey),
        url: definition.url,
        icon: definition.icon,
        isActive: itemIsActive(context.pathname, definition),
        hasAccess: definition.hasAccess?.(context) ?? true,
        items: visibleChildren
      };
    })
    .filter((item) => {
      if (item.items !== undefined) {
        return item.items.length > 0;
      }
      return item.hasAccess;
    });
}
