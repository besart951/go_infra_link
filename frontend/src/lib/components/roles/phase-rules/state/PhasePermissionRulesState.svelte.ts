import { getErrorMessage } from '$lib/api/client.js';
import { addToast } from '$lib/components/toast.svelte';
import type { Phase, PhasePermission } from '$lib/domain/phase/index.js';
import type { Permission, Role } from '$lib/domain/role/index.js';
import type { UserRole } from '$lib/domain/user/index.js';
import { t as translate } from '$lib/i18n/index.js';
import {
  createPhasePermission,
  deletePhasePermission,
  updatePhasePermission
} from '$lib/infrastructure/api/phasePermission.adapter.js';

export type PhaseRulePreset = 'none' | 'read' | 'edit' | 'full';

type PermissionGroups = Record<string, Permission[]>;

interface PhasePermissionRulesStateOptions {
  roles: () => Role[];
  phases: () => Phase[];
  permissions: () => Permission[];
  rules: () => PhasePermission[];
  canManage: () => boolean;
  onRulesChange?: () => Promise<void> | void;
}

const RESOURCE_DISPLAY_NAMES: Record<string, string> = {
  controlcabinet: 'roles.resources.project.controlcabinet',
  spscontroller: 'roles.resources.project.spscontroller',
  'spscontroller.systemtype': 'roles.resources.project.spscontroller_systemtype',
  fielddevice: 'roles.resources.project.fielddevice',
  fielddevice_specification: 'roles.resources.project.fielddevice_specification',
  'fielddevice.bacnetobjects': 'roles.resources.project.fielddevice_bacnetobjects'
};

function createSavingSet(values?: Iterable<string>): Set<string> {
  return new Set(values);
}

function actionOrder(action: string): number {
  switch (action) {
    case 'read':
      return 1;
    case 'update':
      return 2;
    case 'create':
      return 3;
    case 'delete':
      return 4;
    default:
      return 10;
  }
}

function isPhaseRulePermission(permission: Permission): boolean {
  return (
    permission.name.startsWith('project.') &&
    permission.name !== 'project.create' &&
    permission.name !== 'project.listAll' &&
    permission.action !== 'edit'
  );
}

function ruleKey(phaseID: string, role: UserRole): string {
  return `${phaseID}:${role}`;
}

export class PhasePermissionRulesState {
  savingKeys = $state<Set<string>>(createSavingSet());

  private readonly resolveRoles: () => Role[];
  private readonly resolvePhases: () => Phase[];
  private readonly resolvePermissions: () => Permission[];
  private readonly resolveRules: () => PhasePermission[];
  private readonly resolveCanManage: () => boolean;
  private readonly onRulesChange?: () => Promise<void> | void;

  private sortedRolesValue = $derived.by(() => {
    return this.roles
      .filter((role) => role.name !== 'superadmin')
      .sort((a, b) => b.level - a.level);
  });

  private phaseRulePermissionsValue = $derived.by(() => {
    return this.permissions.filter(isPhaseRulePermission).sort((a, b) => {
      const resourceCompare = a.resource.localeCompare(b.resource);
      if (resourceCompare !== 0) return resourceCompare;
      return actionOrder(a.action) - actionOrder(b.action);
    });
  });

  private permissionsByResourceValue = $derived.by(() => {
    const grouped: PermissionGroups = {};
    for (const permission of this.phaseRulePermissions) {
      grouped[permission.resource] ??= [];
      grouped[permission.resource].push(permission);
    }
    return grouped;
  });

  private resourceNamesValue = $derived.by(() => {
    return Object.keys(this.permissionsByResource).sort();
  });

  constructor(options: PhasePermissionRulesStateOptions) {
    this.resolveRoles = options.roles;
    this.resolvePhases = options.phases;
    this.resolvePermissions = options.permissions;
    this.resolveRules = options.rules;
    this.resolveCanManage = options.canManage;
    this.onRulesChange = options.onRulesChange;
  }

  get roles(): Role[] {
    return this.resolveRoles();
  }

  get phases(): Phase[] {
    return this.resolvePhases();
  }

  get permissions(): Permission[] {
    return this.resolvePermissions();
  }

  get rules(): PhasePermission[] {
    return this.resolveRules();
  }

  get canManage(): boolean {
    return this.resolveCanManage();
  }

  get sortedRoles(): Role[] {
    return this.sortedRolesValue;
  }

  get phaseRulePermissions(): Permission[] {
    return this.phaseRulePermissionsValue;
  }

  get permissionsByResource(): PermissionGroups {
    return this.permissionsByResourceValue;
  }

  get resourceNames(): string[] {
    return this.resourceNamesValue;
  }

  get hasPhases(): boolean {
    return this.phases.length > 0;
  }

  get hasPhaseRulePermissions(): boolean {
    return this.phaseRulePermissions.length > 0;
  }

  get emptyMessage(): string {
    if (!this.hasPhases) {
      return translate('roles.phase_rules.empty_phases');
    }

    return translate('roles.phase_rules.empty_permissions');
  }

  getRule(phaseID: string, role: UserRole): PhasePermission | undefined {
    return this.rules.find((rule) => rule.phase_id === phaseID && rule.role === role);
  }

  getPermissionsForResource(resource: string): Permission[] {
    return this.permissionsByResource[resource] ?? [];
  }

  getRoleRuleCount(role: UserRole): number {
    return this.rules.filter((rule) => rule.role === role).length;
  }

  getRoleSummary(role: UserRole): string {
    return translate('roles.phase_rules.role_summary', {
      count: this.getRoleRuleCount(role),
      total: this.phases.length
    });
  }

  getRuleBadgeLabel(rule: PhasePermission | undefined): string {
    if (rule) {
      return translate('roles.phase_rules.custom');
    }

    return translate('roles.phase_rules.default');
  }

  isSaving(phaseID: string, role: UserRole): boolean {
    return this.savingKeys.has(ruleKey(phaseID, role));
  }

  isPermissionSelected(rule: PhasePermission, permissionName: string): boolean {
    return rule.permissions.includes(permissionName);
  }

  async applyPreset(phaseID: string, role: UserRole, preset: PhaseRulePreset): Promise<void> {
    await this.saveRulePermissions(phaseID, role, this.getPresetPermissions(preset));
  }

  async createEmptyRule(phaseID: string, role: UserRole): Promise<void> {
    await this.saveRulePermissions(phaseID, role, []);
  }

  async removeRule(rule: PhasePermission): Promise<void> {
    if (!this.canManage) return;

    this.setSaving(rule.phase_id, rule.role, true);
    try {
      await deletePhasePermission(rule.id);
      await this.refreshRules();
    } catch (err) {
      addToast(getErrorMessage(err), 'error');
    } finally {
      this.setSaving(rule.phase_id, rule.role, false);
    }
  }

  async togglePermission(phaseID: string, role: UserRole, permissionName: string): Promise<void> {
    const existing = this.getRule(phaseID, role);
    const current = new Set(existing?.permissions ?? []);

    if (current.has(permissionName)) {
      current.delete(permissionName);
    } else {
      current.add(permissionName);
    }

    await this.saveRulePermissions(phaseID, role, Array.from(current));
  }

  getActionLabel(action: string): string {
    if (action === 'read') return translate('common.view');
    if (action === 'update') return translate('common.edit');
    if (action === 'create') return translate('common.create');
    if (action === 'delete') return translate('common.delete');
    return action;
  }

  getResourceLabel(resource: string): string {
    const key = resource.replace('project.', '');
    if (resource === 'project') return translate('roles.resources.project_self');

    return RESOURCE_DISPLAY_NAMES[key] ? translate(RESOURCE_DISPLAY_NAMES[key]) : key;
  }

  private async saveRulePermissions(
    phaseID: string,
    role: UserRole,
    nextPermissions: string[]
  ): Promise<void> {
    if (!this.canManage) return;

    this.setSaving(phaseID, role, true);
    try {
      const permissions = this.normalizePermissions(nextPermissions);
      const existingRule = this.getRule(phaseID, role);
      if (existingRule) {
        await updatePhasePermission(existingRule.id, { permissions });
      } else {
        await createPhasePermission({ phase_id: phaseID, role, permissions });
      }
      await this.refreshRules();
    } catch (err) {
      addToast(getErrorMessage(err), 'error');
    } finally {
      this.setSaving(phaseID, role, false);
    }
  }

  private getPresetPermissions(preset: PhaseRulePreset): string[] {
    switch (preset) {
      case 'none':
        return [];
      case 'read':
        return this.phaseRulePermissions
          .filter((permission) => permission.action === 'read')
          .map((permission) => permission.name);
      case 'edit':
        return this.phaseRulePermissions
          .filter((permission) => permission.action === 'read' || permission.action === 'update')
          .map((permission) => permission.name);
      case 'full':
        return this.phaseRulePermissions.map((permission) => permission.name);
    }
  }

  private normalizePermissions(permissionNames: string[]): string[] {
    const allowed = new Set(this.phaseRulePermissions.map((permission) => permission.name));
    const unique = new Set(permissionNames.filter((permissionName) => allowed.has(permissionName)));

    return this.phaseRulePermissions
      .filter((permission) => unique.has(permission.name))
      .map((permission) => permission.name);
  }

  private setSaving(phaseID: string, role: UserRole, saving: boolean): void {
    const next = createSavingSet(this.savingKeys);
    const key = ruleKey(phaseID, role);

    if (saving) {
      next.add(key);
    } else {
      next.delete(key);
    }

    this.savingKeys = next;
  }

  private async refreshRules(): Promise<void> {
    await this.onRulesChange?.();
  }
}

export type { PhasePermissionRulesStateOptions };
