import type { ProjectPermissionName } from '$lib/api/generated/permissions.js';

export interface FieldDevicePermissionChecks {
  canPerform: (action: string, resource: string) => boolean;
  canProject: (permission: ProjectPermissionName) => boolean;
}

export interface FieldDevicePendingEditState {
  hasUnsavedChanges: boolean;
  hasPendingBaseEdits: boolean;
  hasPendingSpecificationEdits: boolean;
  hasPendingBacnetEdits: boolean;
}

interface FieldDevicePermissionPolicyOptions extends FieldDevicePermissionChecks {
  isProjectContext: () => boolean;
}

export function createFieldDevicePermissionPolicy({
  isProjectContext,
  canPerform,
  canProject
}: FieldDevicePermissionPolicyOptions) {
  const canPerformProjectFieldDevice = (action: 'create' | 'update' | 'delete') =>
    canProject(`project.fielddevice.${action}`);
  const canPerformProjectFieldDeviceSpecification = (action: 'update') =>
    canProject(`project.fielddevice_specification.${action}`);
  const canPerformProjectFieldDeviceBacnetObjects = (action: 'update') =>
    canProject(`project.fielddevice.bacnetobjects.${action}`);

  function canCreateFieldDevice(): boolean {
    return isProjectContext()
      ? canPerformProjectFieldDevice('create')
      : canPerform('create', 'fielddevice');
  }

  function canUpdateFieldDevice(): boolean {
    return isProjectContext()
      ? canPerformProjectFieldDevice('update') && canPerform('update', 'fielddevice')
      : canPerform('update', 'fielddevice');
  }

  function canDeleteFieldDevice(): boolean {
    return isProjectContext()
      ? canPerformProjectFieldDevice('delete')
      : canPerform('delete', 'fielddevice');
  }

  function canUpdateFieldDeviceSpecification(): boolean {
    if (!isProjectContext()) {
      return canUpdateFieldDevice();
    }

    return (
      canPerformProjectFieldDeviceSpecification('update') ||
      canPerformProjectFieldDevice('update')
    );
  }

  function canUpdateFieldDeviceBacnetObjects(): boolean {
    if (!isProjectContext()) {
      return canUpdateFieldDevice();
    }

    return (
      canPerformProjectFieldDeviceBacnetObjects('update') ||
      canPerformProjectFieldDevice('update')
    );
  }

  function canOpenBulkEditPanel(): boolean {
    if (!isProjectContext()) {
      return canUpdateFieldDevice();
    }

    return canUpdateFieldDevice() || canUpdateFieldDeviceSpecification();
  }

  function canSavePendingEdits(pending: FieldDevicePendingEditState): boolean {
    if (!pending.hasUnsavedChanges) {
      return false;
    }

    if (!isProjectContext()) {
      return canUpdateFieldDevice();
    }

    if (pending.hasPendingBaseEdits && !canUpdateFieldDevice()) {
      return false;
    }

    if (pending.hasPendingSpecificationEdits && !canUpdateFieldDeviceSpecification()) {
      return false;
    }

    if (pending.hasPendingBacnetEdits && !canUpdateFieldDeviceBacnetObjects()) {
      return false;
    }

    return true;
  }

  return {
    canCreateFieldDevice,
    canUpdateFieldDevice,
    canDeleteFieldDevice,
    canUpdateFieldDeviceSpecification,
    canUpdateFieldDeviceBacnetObjects,
    canOpenBulkEditPanel,
    canSavePendingEdits
  };
}
