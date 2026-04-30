import type {
  BacnetObjectInput,
  BulkUpdateFieldDeviceItem,
  SpecificationInput,
  UpdateFieldDeviceRequest
} from '$lib/domain/facility/index.js';

export interface SharedFieldDeviceDraftDevice {
  device_id: string;
  changed_fields: string[];
  field_values?: Record<string, unknown>;
}

export function getPendingDeviceIds(
  pendingEdits: Map<string, Partial<BulkUpdateFieldDeviceItem>>,
  pendingBacnetEdits: Map<string, Map<string, Partial<BacnetObjectInput>>>
): string[] {
  const ids = new Set<string>();
  for (const id of pendingEdits.keys()) {
    ids.add(id);
  }
  for (const id of pendingBacnetEdits.keys()) {
    ids.add(id);
  }
  return [...ids].sort();
}

export function getChangedFieldsByDevice(
  pendingEdits: Map<string, Partial<BulkUpdateFieldDeviceItem>>,
  pendingBacnetEdits: Map<string, Map<string, Partial<BacnetObjectInput>>>
): SharedFieldDeviceDraftDevice[] {
  const result: SharedFieldDeviceDraftDevice[] = [];
  const deviceIds = getPendingDeviceIds(pendingEdits, pendingBacnetEdits);

  for (const deviceId of deviceIds) {
    const fields = new Set<string>();
    const values: Record<string, unknown> = {};
    const changes = pendingEdits.get(deviceId);

    if (changes) {
      collectBaseField(changes, fields, values, 'bmk');
      collectBaseField(changes, fields, values, 'description');
      collectBaseField(changes, fields, values, 'text_fix');
      collectBaseField(changes, fields, values, 'apparat_nr');
      collectBaseField(changes, fields, values, 'apparat_id');
      collectBaseField(changes, fields, values, 'system_part_id');

      if ('specification' in changes) {
        for (const [specKey, specValue] of Object.entries(changes.specification ?? {})) {
          const fieldKey = `specification.${specKey}`;
          fields.add(fieldKey);
          values[fieldKey] = specValue;
        }
      }
    }

    const bacnetEdits = pendingBacnetEdits.get(deviceId);
    if (bacnetEdits) {
      for (const [objectId, objectChanges] of bacnetEdits.entries()) {
        for (const [fieldName, fieldValue] of Object.entries(objectChanges)) {
          const fieldKey = `bacnet_objects.${objectId}.${fieldName}`;
          fields.add(fieldKey);
          values[fieldKey] = fieldValue;
        }
      }
    }

    if (fields.size > 0) {
      result.push({
        device_id: deviceId,
        changed_fields: [...fields].sort(),
        field_values: Object.keys(values).length > 0 ? values : undefined
      });
    }
  }

  return result;
}

function collectBaseField(
  changes: Partial<BulkUpdateFieldDeviceItem>,
  fields: Set<string>,
  values: Record<string, unknown>,
  field: keyof BulkUpdateFieldDeviceItem
): void {
  if (field in changes) {
    fields.add(field);
    values[field] = changes[field];
  }
}

export function isFieldDirty(
  pendingEdits: Map<string, Partial<BulkUpdateFieldDeviceItem>>,
  deviceId: string,
  field: keyof UpdateFieldDeviceRequest
): boolean {
  const edit = pendingEdits.get(deviceId);
  return edit ? field in edit : false;
}

export function isSpecFieldDirty(
  pendingEdits: Map<string, Partial<BulkUpdateFieldDeviceItem>>,
  deviceId: string,
  field: keyof SpecificationInput
): boolean {
  const edit = pendingEdits.get(deviceId);
  if (!edit) return false;
  const spec = edit.specification;
  return spec ? field in spec : false;
}

export function getPendingValue(
  pendingEdits: Map<string, Partial<BulkUpdateFieldDeviceItem>>,
  deviceId: string,
  field: keyof BulkUpdateFieldDeviceItem
): string | undefined {
  const edit = pendingEdits.get(deviceId);
  if (!edit || !(field in edit)) return undefined;
  const val = edit[field];
  if (val === null) return '';
  return val !== undefined ? String(val) : undefined;
}

export function getPendingSpecValue(
  pendingEdits: Map<string, Partial<BulkUpdateFieldDeviceItem>>,
  deviceId: string,
  field: keyof SpecificationInput
): string | undefined {
  const edit = pendingEdits.get(deviceId);
  if (!edit) return undefined;
  const spec = edit.specification;
  if (!spec || !(field in spec)) return undefined;
  const val = spec[field];
  if (val === null) return '';
  return val !== undefined ? String(val) : undefined;
}

export function hasPendingBaseEdits(
  pendingEdits: Map<string, Partial<BulkUpdateFieldDeviceItem>>
): boolean {
  for (const changes of pendingEdits.values()) {
    if (Object.keys(changes).some((key) => key !== 'specification')) {
      return true;
    }
  }
  return false;
}

export function hasPendingSpecificationEdits(
  pendingEdits: Map<string, Partial<BulkUpdateFieldDeviceItem>>
): boolean {
  for (const changes of pendingEdits.values()) {
    if (changes.specification && Object.keys(changes.specification).length > 0) {
      return true;
    }
  }
  return false;
}

export function hasPendingBacnetEdits(
  pendingBacnetEdits: Map<string, Map<string, Partial<BacnetObjectInput>>>
): boolean {
  for (const changes of pendingBacnetEdits.values()) {
    if (changes.size > 0) {
      return true;
    }
  }
  return false;
}
