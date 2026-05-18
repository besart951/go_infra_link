type CanPerform = (action: string, resource: string) => boolean;

export interface RouteAccessRule {
  path: string;
  permission: string;
}

export interface FacilityNavAccess {
  titleKey: string;
  url: string;
  permission: string;
  dividerAfter?: boolean;
}

export const APP_ROUTE_ACCESS_RULES: RouteAccessRule[] = [
  { path: '/admin/notifications', permission: 'notification.smtp.manage' },
  { path: '/excel', permission: 'objectdata.create' },
  { path: '/timeline', permission: 'timeline.read' },

  { path: '/users/directory', permission: 'user.read' },
  { path: '/teams/new', permission: 'team.create' },
  { path: '/teams', permission: 'team.read' },

  { path: '/projects/new', permission: 'project.create' },
  { path: '/projects/phases/new', permission: 'phase.create' },
  { path: '/projects/phases', permission: 'phase.read' },
  { path: '/projects/*/settings', permission: 'project.update' },

  { path: '/facility/buildings', permission: 'building.read' },
  { path: '/facility/control-cabinets', permission: 'controlcabinet.read' },
  { path: '/facility/sps-controllers', permission: 'spscontroller.read' },
  { path: '/facility/sps-controller-system-type', permission: 'spscontrollersystemtype.read' },
  { path: '/facility/field-devices', permission: 'fielddevice.read' },
  { path: '/facility/system-types', permission: 'systemtype.read' },
  { path: '/facility/system-parts', permission: 'systempart.read' },
  { path: '/facility/apparats', permission: 'apparat.read' },
  { path: '/facility/object-data', permission: 'objectdata.read' },
  { path: '/facility/state-texts', permission: 'statetext.read' },
  { path: '/facility/alarm-definitions', permission: 'alarmdefinition.read' },
  { path: '/facility/alarm-catalog', permission: 'alarmtype.read' },
  { path: '/facility/notification-classes', permission: 'notificationclass.read' },
  { path: '/facility/specifications', permission: 'specification.read' }
];

export const FACILITY_NAV_ACCESS: FacilityNavAccess[] = [
  { titleKey: 'navigation.buildings', url: '/facility/buildings', permission: 'building.read' },
  {
    titleKey: 'navigation.control_cabinets',
    url: '/facility/control-cabinets',
    permission: 'controlcabinet.read'
  },
  {
    titleKey: 'navigation.sps_controllers',
    url: '/facility/sps-controllers',
    permission: 'spscontroller.read'
  },
  {
    titleKey: 'navigation.field_devices',
    url: '/facility/field-devices',
    permission: 'fielddevice.read',
    dividerAfter: true
  },
  {
    titleKey: 'navigation.system_types',
    url: '/facility/system-types',
    permission: 'systemtype.read'
  },
  {
    titleKey: 'navigation.system_parts',
    url: '/facility/system-parts',
    permission: 'systempart.read'
  },
  { titleKey: 'navigation.apparats', url: '/facility/apparats', permission: 'apparat.read' },
  {
    titleKey: 'navigation.object_data',
    url: '/facility/object-data',
    permission: 'objectdata.read'
  },
  {
    titleKey: 'navigation.state_texts',
    url: '/facility/state-texts',
    permission: 'statetext.read'
  },
  {
    titleKey: 'navigation.alarm_definitions',
    url: '/facility/alarm-definitions',
    permission: 'alarmdefinition.read'
  },
  {
    titleKey: 'navigation.alarm_catalog',
    url: '/facility/alarm-catalog',
    permission: 'alarmtype.read'
  },
  {
    titleKey: 'navigation.notification_classes',
    url: '/facility/notification-classes',
    permission: 'notificationclass.read'
  }
];

export const NAV_PERMISSION = {
  excelImporter: 'objectdata.create',
  timeline: 'timeline.read',
  phaseList: 'phase.read',
  notificationSMTP: 'notification.smtp.manage'
} as const;

export function canPerformPermission(canPerform: CanPerform, permission: string): boolean {
  const lastDot = permission.lastIndexOf('.');
  if (lastDot <= 0) return false;

  const resource = permission.slice(0, lastDot);
  const action = permission.slice(lastDot + 1);
  return canPerform(action, resource);
}
