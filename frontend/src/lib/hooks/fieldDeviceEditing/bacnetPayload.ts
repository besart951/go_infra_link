import type {
  BacnetObjectInput,
  BacnetObjectPatchInput,
  FieldDevice
} from '$lib/domain/facility/index.js';

export function buildBacnetObjectsPayload(
  device: FieldDevice,
  deviceEdits: Map<string, Partial<BacnetObjectInput>>
): BacnetObjectPatchInput[] {
  if (!device.bacnet_objects) return [];

  const patches: BacnetObjectPatchInput[] = [];
  for (const [objectId, edits] of deviceEdits.entries()) {
    const patch: BacnetObjectPatchInput = { id: objectId };
    let hasChanges = false;

    if ('text_fix' in edits) {
      patch.text_fix = edits.text_fix as string;
      hasChanges = true;
    }
    if ('description' in edits) {
      patch.description = edits.description as string | undefined;
      hasChanges = true;
    }
    if ('gms_visible' in edits) {
      patch.gms_visible = edits.gms_visible as boolean;
      hasChanges = true;
    }
    if ('optional' in edits) {
      patch.optional = edits.optional as boolean;
      hasChanges = true;
    }
    if ('text_individual' in edits) {
      patch.text_individual = edits.text_individual as string | undefined;
      hasChanges = true;
    }
    if ('software_type' in edits) {
      patch.software_type = edits.software_type as string;
      hasChanges = true;
    }
    if ('software_number' in edits) {
      patch.software_number = edits.software_number as number;
      hasChanges = true;
    }
    if ('hardware_type' in edits) {
      patch.hardware_type = edits.hardware_type as string;
      hasChanges = true;
    }
    if ('hardware_quantity' in edits) {
      patch.hardware_quantity = edits.hardware_quantity as number;
      hasChanges = true;
    }
    if ('software_reference_id' in edits) {
      patch.software_reference_id = edits.software_reference_id as string | undefined;
      hasChanges = true;
    }
    if ('state_text_id' in edits) {
      patch.state_text_id = edits.state_text_id as string | undefined;
      hasChanges = true;
    }
    if ('notification_class_id' in edits) {
      patch.notification_class_id = edits.notification_class_id as string | undefined;
      hasChanges = true;
    }
    if ('alarm_type_id' in edits) {
      patch.alarm_type_id = edits.alarm_type_id as string | undefined;
      hasChanges = true;
    }

    if (hasChanges) {
      patches.push(patch);
    }
  }

  return patches;
}
