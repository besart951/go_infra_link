import type { User, UserRole } from '$lib/domain/user/index.js';

type TestUser = User & {
  permissions?: string[];
  role_display_name?: string;
};

const DEFAULT_TIMESTAMP = '2026-01-01T00:00:00.000Z';

const DEFAULT_USER: TestUser = {
  id: 'user-1',
  first_name: 'Test',
  last_name: 'User',
  email: 'test@example.com',
  is_active: true,
  role: 'planer',
  created_at: DEFAULT_TIMESTAMP,
  updated_at: DEFAULT_TIMESTAMP,
  failed_login_attempts: 0,
  last_login_at: undefined,
  disabled_at: undefined,
  locked_until: undefined
};

const ADMIN_RESOURCES = [
  'user',
  'team',
  'role',
  'permission',
  'project',
  'phase',
  'building',
  'controlcabinet',
  'spscontroller',
  'fielddevice',
  'systemtype',
  'systempart',
  'apparat',
  'objectdata',
  'statetext',
  'alarmdefinition',
  'alarmtype',
  'notificationclass'
] as const;

const ADMIN_ACTIONS = ['read', 'create', 'update', 'delete'] as const;

export function permission(resource: string, action = 'read'): string {
  return `${resource}.${action}`;
}

export function buildUser(overrides: Partial<TestUser> = {}): TestUser {
  return {
    ...DEFAULT_USER,
    ...overrides,
    permissions: [...(overrides.permissions ?? DEFAULT_USER.permissions ?? [])]
  };
}

export function buildPermissionUser(
  resource: string,
  actions: string[] = ['read'],
  overrides: Partial<TestUser> = {}
): TestUser {
  return buildUser({
    ...overrides,
    permissions: actions.map((action) => permission(resource, action))
  });
}

export function buildAdminUser(overrides: Partial<TestUser> = {}): TestUser {
  const permissions =
    overrides.permissions && overrides.permissions.length > 0
      ? overrides.permissions
      : ADMIN_RESOURCES.flatMap((resource) =>
          ADMIN_ACTIONS.map((action) => permission(resource, action))
        );

  return buildUser({
    role: 'superadmin' satisfies UserRole,
    role_display_name: 'Superadmin',
    permissions,
    ...overrides
  });
}
