import { describe, expect, it } from 'vitest';
import { createFieldDevicePermissionPolicy } from './fieldDevicePermissionPolicy.js';
import type { FieldDevicePendingEditState } from './fieldDevicePermissionPolicy.js';

function buildPolicy(projectContext: boolean, permissions: string[]) {
  const granted = new Set(permissions);
  const canPerform = (action: string, resource: string) => granted.has(`${resource}.${action}`);
  return createFieldDevicePermissionPolicy({
    isProjectContext: () => projectContext,
    canPerform,
    canPerformAny: (actions, resource) => actions.some((action) => canPerform(action, resource))
  });
}

const pending = (overrides: Partial<FieldDevicePendingEditState>): FieldDevicePendingEditState => ({
  hasUnsavedChanges: true,
  hasPendingBaseEdits: false,
  hasPendingSpecificationEdits: false,
  hasPendingBacnetEdits: false,
  ...overrides
});

describe('field-device permission policy', () => {
  it('uses global field-device update permission outside project context', () => {
    expect(
      buildPolicy(false, ['fielddevice.update']).canSavePendingEdits(
        pending({
          hasPendingBaseEdits: true,
          hasPendingSpecificationEdits: true,
          hasPendingBacnetEdits: true
        })
      )
    ).toBe(true);
    expect(buildPolicy(false, []).canUpdateFieldDevice()).toBe(false);
  });

  it('uses project field-device permissions for base edits in project context', () => {
    expect(
      buildPolicy(true, ['project.fielddevice.update']).canSavePendingEdits(
        pending({ hasPendingBaseEdits: true })
      )
    ).toBe(true);
    expect(
      buildPolicy(true, ['project.fielddevice_specification.update']).canSavePendingEdits(
        pending({ hasPendingBaseEdits: true })
      )
    ).toBe(false);
  });

  it('allows specification edits through specification or broad project edit permission', () => {
    expect(
      buildPolicy(true, ['project.fielddevice_specification.update']).canSavePendingEdits(
        pending({ hasPendingSpecificationEdits: true })
      )
    ).toBe(true);
    expect(
      buildPolicy(true, ['project.fielddevice.edit']).canSavePendingEdits(
        pending({ hasPendingSpecificationEdits: true })
      )
    ).toBe(true);
  });

  it('allows BACnet edits through BACnet or broad project edit permission', () => {
    expect(
      buildPolicy(true, ['project.fielddevice.bacnetobjects.update']).canSavePendingEdits(
        pending({ hasPendingBacnetEdits: true })
      )
    ).toBe(true);
    expect(
      buildPolicy(true, ['project.fielddevice.edit']).canSavePendingEdits(
        pending({ hasPendingBacnetEdits: true })
      )
    ).toBe(true);
    expect(
      buildPolicy(true, ['project.fielddevice_specification.update']).canSavePendingEdits(
        pending({ hasPendingBacnetEdits: true })
      )
    ).toBe(false);
  });

  it('requires every pending edit category to be allowed', () => {
    expect(
      buildPolicy(true, [
        'project.fielddevice.update',
        'project.fielddevice_specification.update'
      ]).canSavePendingEdits(
        pending({
          hasPendingBaseEdits: true,
          hasPendingSpecificationEdits: true,
          hasPendingBacnetEdits: true
        })
      )
    ).toBe(false);
  });
});
