import { t as translate } from '$lib/i18n/index.js';

import type { Role, Permission } from '$lib/domain/role/index.js';

export type CategoryId = 'users' | 'facility' | 'projects';

type PermissionGroups = Record<CategoryId, Record<string, Permission[]>>;

const CATEGORY_IDS: CategoryId[] = ['users', 'facility', 'projects'];

const CATEGORY_RESOURCE_MAP: Record<CategoryId, string[]> = {
  users: ['user', 'team', 'role', 'permission'],
  facility: [
    'building',
    'controlcabinet',
    'spscontroller',
    'spscontrollersystemtype',
    'fielddevice',
    'bacnetobject',
    'systemtype',
    'systempart',
    'apparat',
    'objectdata',
    'specification',
    'statetext',
    'alarmdefinition',
    'notificationclass'
  ],
  projects: []
};

const RESOURCE_DISPLAY_NAMES: Record<string, string> = {
  user: 'roles.resources.user',
  team: 'roles.resources.team',
  role: 'roles.resources.role',
  permission: 'roles.resources.permission',
  project: 'roles.resources.project_self',
  building: 'roles.resources.building',
  controlcabinet: 'roles.resources.controlcabinet',
  spscontroller: 'roles.resources.spscontroller',
  spscontrollersystemtype: 'roles.resources.spscontrollersystemtype',
  fielddevice: 'roles.resources.fielddevice',
  bacnetobject: 'roles.resources.bacnetobject',
  systemtype: 'roles.resources.systemtype',
  systempart: 'roles.resources.systempart',
  apparat: 'roles.resources.apparat',
  objectdata: 'roles.resources.objectdata',
  specification: 'roles.resources.specification',
  statetext: 'roles.resources.statetext',
  alarmdefinition: 'roles.resources.alarmdefinition',
  notificationclass: 'roles.resources.notificationclass',
  'project.controlcabinet': 'roles.resources.project.controlcabinet',
  'project.spscontroller': 'roles.resources.project.spscontroller',
  'project.spscontroller.systemtype': 'roles.resources.project.spscontroller_systemtype',
  'project.fielddevice': 'roles.resources.project.fielddevice',
  'project.fielddevice_specification': 'roles.resources.project.fielddevice_specification',
  'project.fielddevice.bacnetobjects': 'roles.resources.project.fielddevice_bacnetobjects'
};

export interface RolePermissionEditorStateOptions {
  role: () => Role;
  allPermissions: () => Permission[];
}

export class RolePermissionEditorState {
  searchQuery = $state('');
  selectedPermissions = $state<Set<string>>(new Set());
  expandedCategories = $state<Set<CategoryId>>(new Set(CATEGORY_IDS));
  expandedResources = $state<Set<string>>(new Set());

  private readonly resolveRole: () => Role;
  private readonly resolveAllPermissions: () => Permission[];
  private roleSignature = $state('');

  private permissionsByCategoryValue = $derived.by(() => {
    return this.buildPermissionsByCategory();
  });

  private filteredPermissionsByCategoryValue = $derived.by(() => {
    return this.filterPermissionsByCategory(this.permissionsByCategoryValue);
  });

  constructor(options: RolePermissionEditorStateOptions) {
    this.resolveRole = options.role;
    this.resolveAllPermissions = options.allPermissions;

    this.setupEffects();
  }

  get role(): Role {
    return this.resolveRole();
  }

  get allPermissions(): Permission[] {
    return this.resolveAllPermissions();
  }

  get selectedCount(): number {
    return this.selectedPermissions.size;
  }

  get totalCount(): number {
    return this.allPermissions.length;
  }

  get hasAnyPermissions(): boolean {
    for (const categoryId of CATEGORY_IDS) {
      const resources = this.filteredPermissionsByCategoryValue[categoryId];
      if (Object.keys(resources).length > 0) {
        return true;
      }
    }

    return false;
  }

  getResourcesForCategory(categoryId: CategoryId): string[] {
    const byResource = this.filteredPermissionsByCategoryValue[categoryId];
    return Object.keys(byResource).sort();
  }

  getPermissionsForResource(categoryId: CategoryId, resource: string): Permission[] {
    return this.filteredPermissionsByCategoryValue[categoryId][resource] ?? [];
  }

  isCategoryExpanded(categoryId: CategoryId): boolean {
    return this.expandedCategories.has(categoryId);
  }

  isResourceExpanded(resource: string): boolean {
    return this.expandedResources.has(resource);
  }

  isPermissionSelected(permissionName: string): boolean {
    return this.selectedPermissions.has(permissionName);
  }

  getCategorySelectionCounts(categoryId: CategoryId): { selected: number; total: number } {
    const byResource = this.filteredPermissionsByCategoryValue[categoryId];
    const permissions = this.flattenPermissions(byResource);

    let selected = 0;
    for (const permission of permissions) {
      if (this.selectedPermissions.has(permission.name)) {
        selected += 1;
      }
    }

    return { selected, total: permissions.length };
  }

  getResourceSelectionCounts(
    categoryId: CategoryId,
    resource: string
  ): {
    selected: number;
    total: number;
  } {
    const permissions = this.getPermissionsForResource(categoryId, resource);
    let selected = 0;

    for (const permission of permissions) {
      if (this.selectedPermissions.has(permission.name)) {
        selected += 1;
      }
    }

    return { selected, total: permissions.length };
  }

  getResourceDisplayName(resource: string): string {
    if (RESOURCE_DISPLAY_NAMES[resource]) {
      return translate(RESOURCE_DISPLAY_NAMES[resource]);
    }

    if (resource.startsWith('project.')) {
      const subResource = resource.replace('project.', '');
      if (RESOURCE_DISPLAY_NAMES[subResource]) {
        return translate(RESOURCE_DISPLAY_NAMES[subResource]);
      }

      return subResource;
    }

    return resource;
  }

  setSearchQuery(value: string): void {
    this.searchQuery = value;
  }

  togglePermission(permissionName: string): void {
    const next = new Set(this.selectedPermissions);
    if (next.has(permissionName)) {
      next.delete(permissionName);
    } else {
      next.add(permissionName);
    }

    this.selectedPermissions = next;
  }

  toggleCategory(categoryId: CategoryId): void {
    const next = new Set(this.expandedCategories);
    if (next.has(categoryId)) {
      next.delete(categoryId);
    } else {
      next.add(categoryId);
    }

    this.expandedCategories = next;
  }

  toggleResource(resource: string): void {
    const next = new Set(this.expandedResources);
    if (next.has(resource)) {
      next.delete(resource);
    } else {
      next.add(resource);
    }

    this.expandedResources = next;
  }

  toggleAllInResource(categoryId: CategoryId, resource: string): void {
    const permissions = this.permissionsByCategoryValue[categoryId][resource] ?? [];
    const allSelected = this.areAllSelected(permissions);

    const next = new Set(this.selectedPermissions);
    for (const permission of permissions) {
      if (allSelected) {
        next.delete(permission.name);
      } else {
        next.add(permission.name);
      }
    }

    this.selectedPermissions = next;
  }

  toggleAllInCategory(categoryId: CategoryId): void {
    const byResource = this.permissionsByCategoryValue[categoryId];
    const permissions = this.flattenPermissions(byResource);
    const allSelected = this.areAllSelected(permissions);

    const next = new Set(this.selectedPermissions);
    for (const permission of permissions) {
      if (allSelected) {
        next.delete(permission.name);
      } else {
        next.add(permission.name);
      }
    }

    this.selectedPermissions = next;
  }

  selectAll(): void {
    const next = new Set<string>();
    for (const permission of this.allPermissions) {
      next.add(permission.name);
    }

    this.selectedPermissions = next;
  }

  deselectAll(): void {
    this.selectedPermissions = new Set();
  }

  buildSubmitPayload(): { permissions: string[] } {
    return {
      permissions: Array.from(this.selectedPermissions)
    };
  }

  private setupEffects(): void {
    $effect(() => {
      const nextSignature = this.buildRoleSignature();
      if (nextSignature === this.roleSignature) {
        return;
      }

      this.roleSignature = nextSignature;
      this.selectedPermissions = new Set(this.role.permissions);
    });

    $effect(() => {
      if (!this.searchQuery.trim()) {
        return;
      }

      const resources = new Set<string>();
      for (const categoryId of CATEGORY_IDS) {
        for (const resource of this.getResourcesForCategory(categoryId)) {
          resources.add(resource);
        }
      }

      this.expandedResources = resources;
      this.expandedCategories = new Set(CATEGORY_IDS);
    });
  }

  private buildRoleSignature(): string {
    const sortedPermissions = [...this.role.permissions].sort();
    return `${this.role.id}:${sortedPermissions.join('|')}`;
  }

  private categorizePermission(permission: Permission): CategoryId {
    if (permission.name.startsWith('project.') || permission.resource.startsWith('project.')) {
      return 'projects';
    }

    for (const categoryId of CATEGORY_IDS) {
      if (CATEGORY_RESOURCE_MAP[categoryId].includes(permission.resource)) {
        return categoryId;
      }
    }

    return 'facility';
  }

  private buildPermissionsByCategory(): PermissionGroups {
    const result: PermissionGroups = {
      users: {},
      facility: {},
      projects: {}
    };

    for (const permission of this.allPermissions) {
      const categoryId = this.categorizePermission(permission);
      const resource = permission.resource;

      if (!result[categoryId][resource]) {
        result[categoryId][resource] = [];
      }

      result[categoryId][resource].push(permission);
    }

    for (const categoryId of CATEGORY_IDS) {
      for (const resource of Object.keys(result[categoryId])) {
        result[categoryId][resource].sort(this.sortPermissionsByAction);
      }
    }

    return result;
  }

  private filterPermissionsByCategory(source: PermissionGroups): PermissionGroups {
    if (!this.searchQuery.trim()) {
      return source;
    }

    const query = this.searchQuery.toLowerCase();
    const filtered: PermissionGroups = {
      users: {},
      facility: {},
      projects: {}
    };

    for (const categoryId of CATEGORY_IDS) {
      const entries = Object.entries(source[categoryId]);
      for (const [resource, permissions] of entries) {
        const matches: Permission[] = [];

        for (const permission of permissions) {
          if (this.matchesSearch(permission, query)) {
            matches.push(permission);
          }
        }

        if (matches.length > 0) {
          filtered[categoryId][resource] = matches;
        }
      }
    }

    return filtered;
  }

  private matchesSearch(permission: Permission, query: string): boolean {
    return (
      permission.name.toLowerCase().includes(query) ||
      permission.description.toLowerCase().includes(query) ||
      permission.resource.toLowerCase().includes(query) ||
      permission.action.toLowerCase().includes(query)
    );
  }

  private flattenPermissions(byResource: Record<string, Permission[]>): Permission[] {
    const flat: Permission[] = [];
    for (const permissions of Object.values(byResource)) {
      for (const permission of permissions) {
        flat.push(permission);
      }
    }

    return flat;
  }

  private areAllSelected(permissions: Permission[]): boolean {
    if (permissions.length === 0) {
      return false;
    }

    for (const permission of permissions) {
      if (!this.selectedPermissions.has(permission.name)) {
        return false;
      }
    }

    return true;
  }

  private sortPermissionsByAction(a: Permission, b: Permission): number {
    return a.action.localeCompare(b.action);
  }
}
