export interface FieldDevicePermissionChecks {
  canPerform: (action: string, resource: string) => boolean;
  canPerformAny: (actions: string[], resource: string) => boolean;
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
  canPerformAny
}: FieldDevicePermissionPolicyOptions) {
  const canPerformProjectFieldDevice = (...actions: string[]) =>
    canPerformAny(actions, 'project.fielddevice');
  const canPerformProjectFieldDeviceSpecification = (...actions: string[]) =>
    canPerformAny(actions, 'project.fielddevice_specification');
  const canPerformProjectFieldDeviceBacnetObjects = (...actions: string[]) =>
    canPerformAny(actions, 'project.fielddevice.bacnetobjects');

  function canCreateFieldDevice(): boolean {
    return isProjectContext()
      ? canPerformProjectFieldDevice('create', 'edit')
      : canPerform('create', 'fielddevice');
  }

  function canUpdateFieldDevice(): boolean {
    return isProjectContext()
      ? canPerformProjectFieldDevice('update', 'edit')
      : canPerform('update', 'fielddevice');
  }

  function canDeleteFieldDevice(): boolean {
    return isProjectContext()
      ? canPerformProjectFieldDevice('delete', 'edit')
      : canPerform('delete', 'fielddevice');
  }

  function canUpdateFieldDeviceSpecification(): boolean {
    if (!isProjectContext()) {
      return canUpdateFieldDevice();
    }

    return (
      canPerformProjectFieldDeviceSpecification('update', 'edit') ||
      canPerformProjectFieldDevice('edit')
    );
  }

  function canUpdateFieldDeviceBacnetObjects(): boolean {
    if (!isProjectContext()) {
      return canUpdateFieldDevice();
    }

    return (
      canPerformProjectFieldDeviceBacnetObjects('update', 'edit') ||
      canPerformProjectFieldDevice('edit')
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
