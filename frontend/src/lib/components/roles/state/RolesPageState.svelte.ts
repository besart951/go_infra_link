import { getErrorMessage } from '$lib/api/client.js';
import { addToast } from '$lib/components/toast.svelte';
import type { Phase, PhaseListResponse, PhasePermission } from '$lib/domain/phase/index.js';
import type { Permission, Role } from '$lib/domain/role/index.js';
import { t as translate } from '$lib/i18n/index.js';
import { listPhases } from '$lib/infrastructure/api/phase.adapter.js';
import { listPhasePermissions } from '$lib/infrastructure/api/phasePermission.adapter.js';
import {
  createPermission as apiCreatePermission,
  deletePermission as apiDeletePermission,
  listPermissions,
  listRoles,
  updatePermission as apiUpdatePermission,
  updateRolePermissions as apiUpdateRolePermissions
} from '$lib/infrastructure/api/role.adapter.js';
import { confirm } from '$lib/stores/confirm-dialog.js';
import { canPerform } from '$lib/utils/permissions.js';

export interface PermissionFormData {
  name: string;
  description: string;
  resource: string;
  action: string;
}

function emptyPhaseResponse(): PhaseListResponse {
  return {
    items: [],
    total: 0,
    page: 1,
    limit: 100,
    total_pages: 1
  };
}

export class RolesPageState {
  roles = $state<Role[]>([]);
  permissions = $state<Permission[]>([]);
  phases = $state<Phase[]>([]);
  phasePermissions = $state<PhasePermission[]>([]);
  isLoading = $state(true);
  error = $state<string | null>(null);
  activeTab = $state('roles');

  createPermissionDialogOpen = $state(false);
  editPermissionDialogOpen = $state(false);
  editRoleSheetOpen = $state(false);
  selectedPermission = $state<Permission | null>(null);
  selectedRole = $state<Role | null>(null);

  isSubmittingPermission = $state(false);
  isSubmittingRole = $state(false);
  permissionError = $state<string | null>(null);
  roleError = $state<string | null>(null);

  canManageRoles = $derived(canPerform('update', 'role') || canPerform('update', 'permission'));
  canManagePhaseRules = $derived(canPerform('manage', 'phase_permission'));
  totalPermissions = $derived(this.permissions.length);
  totalRoles = $derived(this.roles.length);
  uniqueResources = $derived(
    new Set(this.permissions.map((permission) => permission.resource)).size
  );

  loadData = async (): Promise<void> => {
    this.isLoading = true;
    this.error = null;

    try {
      const phaseDataRequest = this.canManagePhaseRules
        ? listPhases({ page: 1, limit: 100 })
        : Promise.resolve(emptyPhaseResponse());
      const phasePermissionDataRequest = this.canManagePhaseRules
        ? listPhasePermissions()
        : Promise.resolve({ items: [] });

      const [rolesData, permissionsData, phasesData, phasePermissionsData] = await Promise.all([
        listRoles(),
        listPermissions(),
        phaseDataRequest,
        phasePermissionDataRequest
      ]);

      this.roles = rolesData;
      this.permissions = permissionsData;
      this.phases = phasesData.items;
      this.phasePermissions = phasePermissionsData.items;
    } catch (err) {
      this.error = err instanceof Error ? err.message : translate('roles.errors.load_failed');
    } finally {
      this.isLoading = false;
    }
  };

  reloadPhaseRules = async (): Promise<void> => {
    const phasePermissionsData = await listPhasePermissions();
    this.phasePermissions = phasePermissionsData.items;
  };

  openCreatePermissionDialog = (): void => {
    this.permissionError = null;
    this.createPermissionDialogOpen = true;
  };

  closeCreatePermissionDialog = (): void => {
    this.createPermissionDialogOpen = false;
  };

  openEditPermission = (permission: Permission): void => {
    this.selectedPermission = permission;
    this.permissionError = null;
    this.editPermissionDialogOpen = true;
  };

  closeEditPermissionDialog = (): void => {
    this.editPermissionDialogOpen = false;
    this.selectedPermission = null;
  };

  openEditRole = (role: Role): void => {
    this.selectedRole = role;
    this.roleError = null;
    this.editRoleSheetOpen = true;
  };

  viewRolePermissions = (role: Role): void => {
    this.selectedRole = role;
    this.editRoleSheetOpen = true;
  };

  closeRoleSheet = (): void => {
    this.editRoleSheetOpen = false;
    this.selectedRole = null;
  };

  createPermission = async (data: PermissionFormData): Promise<void> => {
    this.isSubmittingPermission = true;
    this.permissionError = null;

    try {
      await apiCreatePermission(data);
      addToast(translate('roles.toasts.permission_created'), 'success');
      this.createPermissionDialogOpen = false;
      await this.loadData();
    } catch (err) {
      this.permissionError = getErrorMessage(err);
    } finally {
      this.isSubmittingPermission = false;
    }
  };

  updatePermission = async (data: PermissionFormData): Promise<void> => {
    if (!this.selectedPermission) return;

    this.isSubmittingPermission = true;
    this.permissionError = null;

    try {
      await apiUpdatePermission(this.selectedPermission.id, { description: data.description });
      addToast(translate('roles.toasts.permission_updated'), 'success');
      this.closeEditPermissionDialog();
      await this.loadData();
    } catch (err) {
      this.permissionError = getErrorMessage(err);
    } finally {
      this.isSubmittingPermission = false;
    }
  };

  deletePermission = async (permission: Permission): Promise<void> => {
    const confirmed = await confirm({
      title: translate('roles.permissions.delete_confirm_title'),
      message: translate('roles.permissions.delete_confirm_message', { name: permission.name }),
      confirmText: translate('roles.permissions.delete_confirm_confirm'),
      cancelText: translate('common.cancel'),
      variant: 'destructive'
    });

    if (!confirmed) return;

    try {
      await apiDeletePermission(permission.id);
      addToast(translate('roles.toasts.permission_deleted'), 'success');
      await this.loadData();
    } catch (err) {
      addToast(getErrorMessage(err), 'error');
    }
  };

  updateSelectedRolePermissions = async (data: { permissions: string[] }): Promise<void> => {
    if (!this.selectedRole) return;

    this.isSubmittingRole = true;
    this.roleError = null;

    try {
      await apiUpdateRolePermissions(this.selectedRole.name, data);
      addToast(
        translate('roles.toasts.role_permissions_updated', {
          role: this.selectedRole.display_name
        }),
        'success'
      );
      this.closeRoleSheet();
      await this.loadData();
    } catch (err) {
      this.roleError = getErrorMessage(err);
    } finally {
      this.isSubmittingRole = false;
    }
  };
}
