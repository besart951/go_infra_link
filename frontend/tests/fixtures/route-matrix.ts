export type RouteDomain =
  | 'auth'
  | 'app'
  | 'admin'
  | 'user'
  | 'team'
  | 'project'
  | 'facility'
  | 'excel';

export type RouteAuthorizationMode = 'none' | 'route' | 'ui-only';
export type RouteAuditStatus = 'configured' | 'misconfigured';

export interface RouteAudit {
  path: string;
  domain: RouteDomain;
  files: string[];
  auth: 'public' | 'authenticated';
  authorization: RouteAuthorizationMode;
  status: RouteAuditStatus;
  expectedAccess: string;
  protectedUi: string[];
  notes: string;
}

export const routeAudits = [
  {
    path: '/login',
    domain: 'auth',
    files: ['src/routes/(auth)/login/+page.svelte'],
    auth: 'public',
    authorization: 'none',
    status: 'configured',
    expectedAccess: 'Public route for unauthenticated users.',
    protectedUi: [],
    notes: 'Authentication happens through the login API, not through route guards.'
  },
  {
    path: '/',
    domain: 'app',
    files: ['src/routes/(app)/+page.svelte', 'src/routes/(app)/+page.ts'],
    auth: 'authenticated',
    authorization: 'none',
    status: 'configured',
    expectedAccess: 'Any authenticated user.',
    protectedUi: ['dashboard cards'],
    notes: 'Guest protection comes from the shared app layout.'
  },
  {
    path: '/account',
    domain: 'app',
    files: ['src/routes/(app)/account/+page.svelte'],
    auth: 'authenticated',
    authorization: 'none',
    status: 'configured',
    expectedAccess: 'Any authenticated user viewing their own account.',
    protectedUi: ['profile form', 'password form', 'theme preferences'],
    notes: 'This is intentionally auth-only rather than permission-gated.'
  },
  {
    path: '/logout',
    domain: 'app',
    files: ['src/routes/(app)/logout/+page.svelte'],
    auth: 'authenticated',
    authorization: 'none',
    status: 'configured',
    expectedAccess: 'Any authenticated user ending their current session.',
    protectedUi: ['logout in progress indicator'],
    notes: 'Guest protection comes from the shared app layout.'
  },
  {
    path: '/projects/new',
    domain: 'project',
    files: ['src/routes/(app)/projects/new/+page.svelte'],
    auth: 'authenticated',
    authorization: 'none',
    status: 'configured',
    expectedAccess: 'Redirect helper into /projects.',
    protectedUi: [],
    notes: 'This route does not expose a separate form; it only forwards to the list page.'
  },
  {
    path: '/admin/notifications/smtp',
    domain: 'admin',
    files: [
      'src/routes/(app)/admin/notifications/smtp/+page.svelte',
      'src/routes/(app)/admin/notifications/smtp/+page.ts'
    ],
    auth: 'authenticated',
    authorization: 'route',
    status: 'configured',
    expectedAccess: 'Superadmin only.',
    protectedUi: ['notifications.page.title', 'SMTP settings form'],
    notes: 'This is the current reference implementation for load-level authorization.'
  },
  {
    path: '/users',
    domain: 'user',
    files: ['src/routes/(app)/users/+page.svelte'],
    auth: 'authenticated',
    authorization: 'ui-only',
    status: 'misconfigured',
    expectedAccess: 'Requires user.read for the route and user.create/update/delete for actions.',
    protectedUi: ['common.create_user', 'common.change_role', 'common.delete_user'],
    notes: 'The page renders for any authenticated user and only hides mutation controls.'
  },
  {
    path: '/users/roles',
    domain: 'user',
    files: ['src/routes/(app)/users/roles/+page.svelte'],
    auth: 'authenticated',
    authorization: 'ui-only',
    status: 'misconfigured',
    expectedAccess:
      'Requires role.read plus role.update or permission.update for management actions.',
    protectedUi: ['role and permission editor tabs', 'create permission dialog'],
    notes: 'The route is reachable without role.read and still reveals the role matrix.'
  },
  {
    path: '/teams',
    domain: 'team',
    files: ['src/routes/(app)/teams/+page.svelte'],
    auth: 'authenticated',
    authorization: 'ui-only',
    status: 'misconfigured',
    expectedAccess: 'Requires team.read for the route and team.create/delete for actions.',
    protectedUi: ['pages.create_team', 'common.delete_team'],
    notes: 'Any authenticated user can load the team listing today.'
  },
  {
    path: '/teams/:id',
    domain: 'team',
    files: ['src/routes/(app)/teams/[id]/+page.svelte'],
    auth: 'authenticated',
    authorization: 'none',
    status: 'misconfigured',
    expectedAccess: 'Requires team membership or an explicit team management permission.',
    protectedUi: ['teams.detail.add_member', 'member removal controls'],
    notes: 'The detail screen does not consult permissions before loading or mutating membership.'
  },
  {
    path: '/projects',
    domain: 'project',
    files: ['src/routes/(app)/projects/+page.svelte'],
    auth: 'authenticated',
    authorization: 'ui-only',
    status: 'misconfigured',
    expectedAccess: 'Requires project.read for the route and project.create for creation.',
    protectedUi: ['common.create', 'project creation form'],
    notes: 'The page still renders without project.read and only hides the create workflow.'
  },
  {
    path: '/projects/:id',
    domain: 'project',
    files: ['src/routes/(app)/projects/[id]/+page.svelte'],
    auth: 'authenticated',
    authorization: 'none',
    status: 'misconfigured',
    expectedAccess: 'Requires project membership or a project.read-style grant.',
    protectedUi: ['project collaboration surface', 'project details'],
    notes: 'The frontend does not verify project membership before rendering the page.'
  },
  {
    path: '/projects/:id/settings',
    domain: 'project',
    files: ['src/routes/(app)/projects/[id]/settings/+page.svelte'],
    auth: 'authenticated',
    authorization: 'none',
    status: 'misconfigured',
    expectedAccess: 'Requires project administration or ownership.',
    protectedUi: [
      'project settings form',
      'project user management',
      'project object-data assignment'
    ],
    notes: 'The settings route exposes project administration workflows without a route guard.'
  },
  {
    path: '/projects/phases',
    domain: 'project',
    files: ['src/routes/(app)/projects/phases/+page.svelte'],
    auth: 'authenticated',
    authorization: 'ui-only',
    status: 'misconfigured',
    expectedAccess:
      'Requires phase.read for the route and phase.create/update/delete for mutations.',
    protectedUi: ['phases.new.title', 'phase action buttons'],
    notes: 'The list page is visible without phase.read.'
  },
  {
    path: '/projects/phases/new',
    domain: 'project',
    files: ['src/routes/(app)/projects/phases/new/+page.svelte'],
    auth: 'authenticated',
    authorization: 'none',
    status: 'misconfigured',
    expectedAccess: 'Requires phase.create.',
    protectedUi: ['PhaseForm'],
    notes: 'The dedicated creation page exposes the form without checking phase.create.'
  },
  {
    path: '/projects/phases/:id',
    domain: 'project',
    files: ['src/routes/(app)/projects/phases/[id]/+page.svelte'],
    auth: 'authenticated',
    authorization: 'none',
    status: 'misconfigured',
    expectedAccess: 'Requires phase.read and phase.update/delete for mutations.',
    protectedUi: ['phase detail', 'phase delete action'],
    notes: 'The phase detail page does not enforce permission checks at load time.'
  },
  {
    path: '/facility',
    domain: 'facility',
    files: ['src/routes/(app)/facility/+page.svelte'],
    auth: 'authenticated',
    authorization: 'none',
    status: 'misconfigured',
    expectedAccess:
      'The overview should only surface links that match the user’s facility permissions.',
    protectedUi: [
      'facility.buildings',
      'facility.control_cabinets',
      'facility.sps_controllers',
      'facility.field_devices'
    ],
    notes: 'The overview links are static and currently ignore permission state completely.'
  },
  {
    path: '/facility/buildings',
    domain: 'facility',
    files: ['src/routes/(app)/facility/buildings/+page.svelte'],
    auth: 'authenticated',
    authorization: 'ui-only',
    status: 'misconfigured',
    expectedAccess:
      'Requires building.read for the route and building.create/update/delete for mutations.',
    protectedUi: ['building create button', 'building row actions'],
    notes: 'The route itself is still reachable without building.read.'
  },
  {
    path: '/facility/buildings/:id',
    domain: 'facility',
    files: [
      'src/routes/(app)/facility/buildings/[id]/+page.svelte',
      'src/routes/(app)/facility/buildings/[id]/+page.ts'
    ],
    auth: 'authenticated',
    authorization: 'none',
    status: 'misconfigured',
    expectedAccess: 'Requires building.read plus building.update for edits.',
    protectedUi: ['building detail form'],
    notes: 'Neither the page load nor the page component checks access.'
  },
  {
    path: '/facility/control-cabinets',
    domain: 'facility',
    files: ['src/routes/(app)/facility/control-cabinets/+page.svelte'],
    auth: 'authenticated',
    authorization: 'ui-only',
    status: 'misconfigured',
    expectedAccess: 'Requires controlcabinet.read for the route.',
    protectedUi: ['control cabinet listing'],
    notes: 'Sidebar visibility is permission-aware but the route is not.'
  },
  {
    path: '/facility/control-cabinets/:id',
    domain: 'facility',
    files: [
      'src/routes/(app)/facility/control-cabinets/[id]/+page.svelte',
      'src/routes/(app)/facility/control-cabinets/[id]/+page.ts'
    ],
    auth: 'authenticated',
    authorization: 'none',
    status: 'misconfigured',
    expectedAccess: 'Requires controlcabinet.read and the relevant edit permission for mutations.',
    protectedUi: ['control cabinet detail'],
    notes: 'The detail route lacks load-level authorization.'
  },
  {
    path: '/facility/sps-controllers',
    domain: 'facility',
    files: ['src/routes/(app)/facility/sps-controllers/+page.svelte'],
    auth: 'authenticated',
    authorization: 'ui-only',
    status: 'misconfigured',
    expectedAccess: 'Requires spscontroller.read for the route.',
    protectedUi: ['SPS controller list'],
    notes: 'Only sidebar visibility reflects the permission model today.'
  },
  {
    path: '/facility/sps-controllers/:id',
    domain: 'facility',
    files: [
      'src/routes/(app)/facility/sps-controllers/[id]/+page.svelte',
      'src/routes/(app)/facility/sps-controllers/[id]/+page.ts'
    ],
    auth: 'authenticated',
    authorization: 'none',
    status: 'misconfigured',
    expectedAccess: 'Requires spscontroller.read plus spscontroller.update for edits.',
    protectedUi: [
      'SPSControllerDetailHeader',
      'SPSControllerForm',
      'SPSControllerOverviewCard',
      'SPSControllerSystemTypesOverview'
    ],
    notes: 'The controller detail screen loads and wires edit actions without a route-level check.'
  },
  {
    path: '/facility/sps-controller-system-type/:id',
    domain: 'facility',
    files: [
      'src/routes/(app)/facility/sps-controller-system-type/[id]/+page.svelte',
      'src/routes/(app)/facility/sps-controller-system-type/[id]/+page.ts'
    ],
    auth: 'authenticated',
    authorization: 'none',
    status: 'misconfigured',
    expectedAccess:
      'Requires systemtype.read and the relevant update permission before exposing detail workflows.',
    protectedUi: ['facility.sps_controller_system_type_detail.overview_title'],
    notes: 'The system-type detail route resolves without a route-level permission check.'
  },
  {
    path: '/facility/field-devices',
    domain: 'facility',
    files: ['src/routes/(app)/facility/field-devices/+page.svelte'],
    auth: 'authenticated',
    authorization: 'ui-only',
    status: 'misconfigured',
    expectedAccess: 'Requires fielddevice.read for the route.',
    protectedUi: ['field device list view'],
    notes: 'The route loads regardless of fielddevice.read.'
  },
  {
    path: '/facility/system-types',
    domain: 'facility',
    files: ['src/routes/(app)/facility/system-types/+page.svelte'],
    auth: 'authenticated',
    authorization: 'ui-only',
    status: 'misconfigured',
    expectedAccess:
      'Requires systemtype.read for the route and systemtype.create/update/delete for actions.',
    protectedUi: ['system type create button', 'system type row actions'],
    notes: 'This follows the same UI-only pattern as the other CRUD list pages.'
  },
  {
    path: '/facility/system-parts',
    domain: 'facility',
    files: ['src/routes/(app)/facility/system-parts/+page.svelte'],
    auth: 'authenticated',
    authorization: 'ui-only',
    status: 'misconfigured',
    expectedAccess:
      'Requires systempart.read for the route and systempart.create/update/delete for actions.',
    protectedUi: ['system part create button', 'system part row actions'],
    notes: 'The page is still reachable without systempart.read.'
  },
  {
    path: '/facility/apparats',
    domain: 'facility',
    files: ['src/routes/(app)/facility/apparats/+page.svelte'],
    auth: 'authenticated',
    authorization: 'ui-only',
    status: 'misconfigured',
    expectedAccess:
      'Requires apparat.read for the route and apparat.create/update/delete for actions.',
    protectedUi: ['apparat create button', 'apparat row actions'],
    notes: 'List visibility and mutation visibility are decoupled today.'
  },
  {
    path: '/facility/object-data',
    domain: 'facility',
    files: ['src/routes/(app)/facility/object-data/+page.svelte'],
    auth: 'authenticated',
    authorization: 'ui-only',
    status: 'misconfigured',
    expectedAccess:
      'Requires objectdata.read for the route and objectdata.create/update/delete for actions.',
    protectedUi: ['object-data create button', 'object-data row actions'],
    notes: 'The route is rendered before any permission check happens.'
  },
  {
    path: '/facility/state-texts',
    domain: 'facility',
    files: ['src/routes/(app)/facility/state-texts/+page.svelte'],
    auth: 'authenticated',
    authorization: 'ui-only',
    status: 'misconfigured',
    expectedAccess:
      'Requires statetext.read for the route and statetext.create/update/delete for actions.',
    protectedUi: ['state text create button', 'state text row actions'],
    notes: 'The page uses UI gates but not a route guard.'
  },
  {
    path: '/facility/alarm-definitions',
    domain: 'facility',
    files: ['src/routes/(app)/facility/alarm-definitions/+page.svelte'],
    auth: 'authenticated',
    authorization: 'ui-only',
    status: 'misconfigured',
    expectedAccess:
      'Requires alarmdefinition.read for the route and alarmdefinition.create/update/delete for actions.',
    protectedUi: ['alarm definition create button', 'alarm definition row actions'],
    notes: 'The route can still be opened by users without alarmdefinition.read.'
  },
  {
    path: '/facility/alarm-catalog',
    domain: 'facility',
    files: ['src/routes/(app)/facility/alarm-catalog/+page.svelte'],
    auth: 'authenticated',
    authorization: 'ui-only',
    status: 'misconfigured',
    expectedAccess:
      'Requires alarmtype.read for the route and alarmtype.create/update/delete for actions.',
    protectedUi: ['alarm catalog create buttons', 'alarm catalog mutation controls'],
    notes: 'Mutation controls are permission-aware, but route access is not.'
  },
  {
    path: '/facility/notification-classes',
    domain: 'facility',
    files: ['src/routes/(app)/facility/notification-classes/+page.svelte'],
    auth: 'authenticated',
    authorization: 'ui-only',
    status: 'misconfigured',
    expectedAccess:
      'Requires notificationclass.read for the route and notificationclass.create/update/delete for actions.',
    protectedUi: ['notification class create button', 'notification class row actions'],
    notes: 'The route itself is not protected.'
  },
  {
    path: '/facility/specifications',
    domain: 'facility',
    files: ['src/routes/(app)/facility/specifications/+page.svelte'],
    auth: 'authenticated',
    authorization: 'none',
    status: 'misconfigured',
    expectedAccess:
      'Should be tied to a dedicated facility specification permission before exposure.',
    protectedUi: ['facility.specifications'],
    notes: 'The page has no permission checks at all today.'
  },
  {
    path: '/excel',
    domain: 'excel',
    files: ['src/routes/(app)/excel/+page.svelte'],
    auth: 'authenticated',
    authorization: 'ui-only',
    status: 'misconfigured',
    expectedAccess:
      'The route should require objectdata.create, and the sidebar should use the same permission.',
    protectedUi: ['ExcelUploadDropzone', 'ExcelReadProgressCard', 'ExcelSessionSummary'],
    notes:
      'The page hides the uploader without objectdata.create, but the route still resolves and the sidebar uses objectdata.read instead.'
  }
] satisfies RouteAudit[];

routeAudits.sort((left, right) => left.path.localeCompare(right.path));

export const configuredRoutePaths = routeAudits
  .filter((route) => route.status === 'configured')
  .map((route) => route.path);

export const misconfiguredRoutePaths = routeAudits
  .filter((route) => route.status === 'misconfigured')
  .map((route) => route.path);

export function getRouteAuditSummary() {
  return {
    totalRoutes: routeAudits.length,
    configured: configuredRoutePaths,
    misconfigured: misconfiguredRoutePaths
  };
}
