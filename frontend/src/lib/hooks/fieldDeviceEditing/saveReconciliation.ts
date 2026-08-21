import type {
  BacnetObjectInput,
  BulkUpdateFieldDeviceItem,
  BulkUpdateFieldDeviceResponse,
  FieldDevice,
  SpecificationInput
} from '$lib/domain/facility/index.js';
import { findFieldPathSegment, getRelevantFieldPathSegments } from '$lib/api/fieldPath.js';
import { normalizeSpecificationForDisplay, toDisplayOptionalValue } from './specificationEdits.js';

export interface EditErrorInfo {
  message?: string;
  fields?: Record<string, string>;
  suggestions?: Record<string, number>;
  suggestionOptions?: Record<string, number[]>;
}

interface ReconcileFieldDeviceSaveResultInput {
  storeItems: FieldDevice[];
  updates: BulkUpdateFieldDeviceItem[];
  result: BulkUpdateFieldDeviceResponse;
  pendingEdits: Map<string, Partial<BulkUpdateFieldDeviceItem>>;
  pendingBacnetEdits: Map<string, Map<string, Partial<BacnetObjectInput>>>;
  pendingEditsSnapshot: Map<string, Partial<BulkUpdateFieldDeviceItem>>;
  pendingBacnetEditsSnapshot: Map<string, Map<string, Partial<BacnetObjectInput>>>;
  existingErrors: Map<string, EditErrorInfo>;
  localizeEditErrorInfo: (info?: EditErrorInfo) => EditErrorInfo | undefined;
  localizeFieldErrorMap: (fields: Record<string, string>) => Record<string, string>;
  nowIso?: string;
}

export interface ReconciledFieldDeviceSaveResult {
  editErrors: Map<string, EditErrorInfo>;
  bacnetFieldErrors: Map<string, Map<string, Record<string, string>>>;
  remainingEdits: Map<string, Partial<BulkUpdateFieldDeviceItem>>;
  remainingBacnetEdits: Map<string, Map<string, Partial<BacnetObjectInput>>>;
  successIds: Set<string>;
  optimisticUpdates: FieldDevice[];
}

export function reconcileFieldDeviceSaveResult({
  storeItems,
  updates,
  result,
  pendingEdits,
  pendingBacnetEdits,
  pendingEditsSnapshot,
  pendingBacnetEditsSnapshot,
  existingErrors,
  localizeEditErrorInfo,
  localizeFieldErrorMap,
  nowIso = new Date().toISOString()
}: ReconcileFieldDeviceSaveResultInput): ReconciledFieldDeviceSaveResult {
  const editErrors = new Map(existingErrors);
  const bacnetFieldErrors = new Map<string, Map<string, Record<string, string>>>();
  const remainingEdits = new Map(pendingEdits);
  const remainingBacnetEdits = new Map(pendingBacnetEdits);
  const successIds = new Set<string>();
  const optimisticUpdates: FieldDevice[] = [];

  for (const item of result.results) {
    if (item.success) {
      successIds.add(item.id);
      editErrors.delete(item.id);

      if (pendingEdits.get(item.id) === pendingEditsSnapshot.get(item.id)) {
        remainingEdits.delete(item.id);
      }
      if (pendingBacnetEdits.get(item.id) === pendingBacnetEditsSnapshot.get(item.id)) {
        remainingBacnetEdits.delete(item.id);
      }

      const device = storeItems.find((candidate) => candidate.id === item.id);
      if (item.field_device) {
        optimisticUpdates.push(item.field_device);
      } else if (device) {
        optimisticUpdates.push(
          applyAllEditsToDevice(device, pendingEdits, pendingBacnetEdits, nowIso)
        );
      }
      continue;
    }

    // A failed result means the backend rolled the complete FieldDevice
    // aggregate back. Keep every draft field for that device so the user can
    // fix the error and retry without silently losing otherwise valid edits.
    const localizedFields = item.fields ? localizeFieldErrorMap(item.fields) : undefined;
    editErrors.set(
      item.id,
      localizeEditErrorInfo({
        message: item.error,
        fields: localizedFields,
        suggestions: item.suggestions,
        suggestionOptions: item.suggestion_options
      }) ?? {}
    );

    if (!localizedFields) continue;
    const update = updates.find((candidate) => candidate.id === item.id);
    const objectErrors = getBacnetObjectErrors(localizedFields, update);
    if (objectErrors.size > 0) {
      bacnetFieldErrors.set(item.id, objectErrors);
    }
  }

  return {
    editErrors,
    bacnetFieldErrors,
    remainingEdits,
    remainingBacnetEdits,
    successIds,
    optimisticUpdates
  };
}

function getBacnetObjectErrors(
  fields: Record<string, string>,
  update?: BulkUpdateFieldDeviceItem
): Map<string, Record<string, string>> {
  const objectErrors = new Map<string, Record<string, string>>();
  for (const [fieldPath, message] of Object.entries(fields)) {
    const parsed = parseBacnetFieldPath(fieldPath, update);
    if (!parsed) continue;
    const existing = objectErrors.get(parsed.objectId) ?? {};
    existing[parsed.field] = message;
    objectErrors.set(parsed.objectId, existing);
  }
  return objectErrors;
}

function parseBacnetFieldPath(
  fieldPath: string,
  update?: BulkUpdateFieldDeviceItem
): { objectId: string; field: string } | undefined {
  const segments = getRelevantFieldPathSegments(fieldPath);
  const bacnetIndex = findFieldPathSegment(segments, ['bacnetobjects', 'bacnetobject']);
  if (bacnetIndex === undefined) return undefined;

  const objectRef = segments[bacnetIndex + 1];
  const field = segments.slice(bacnetIndex + 2).join('.');
  if (!objectRef || !field) return undefined;

  const objectIndex = Number(objectRef);
  const objectId = Number.isInteger(objectIndex)
    ? update?.bacnet_objects?.[objectIndex]?.id
    : objectRef;
  return objectId ? { objectId, field } : undefined;
}

function applyAllEditsToDevice(
  device: FieldDevice,
  pendingEdits: Map<string, Partial<BulkUpdateFieldDeviceItem>>,
  pendingBacnetEdits: Map<string, Map<string, Partial<BacnetObjectInput>>>,
  nowIso: string
): FieldDevice {
  const changes = pendingEdits.get(device.id);
  let updated = changes ? applyBaseEdits(device, changes) : device;
  if (changes?.specification) {
    updated = applySpecificationEdits(updated, changes.specification, nowIso);
  }

  const bacnetEdits = pendingBacnetEdits.get(device.id);
  if (bacnetEdits && updated.bacnet_objects) {
    updated = {
      ...updated,
      bacnet_objects: updated.bacnet_objects.map((object) => {
        const edits = bacnetEdits.get(object.id);
        return edits ? { ...object, ...edits } : object;
      })
    };
  }
  return updated;
}

function applyBaseEdits(
  device: FieldDevice,
  changes: Partial<BulkUpdateFieldDeviceItem>
): FieldDevice {
  let updated = device;
  if ('bmk' in changes) updated = { ...updated, bmk: toDisplayOptionalValue(changes.bmk) };
  if ('description' in changes) {
    updated = { ...updated, description: toDisplayOptionalValue(changes.description) };
  }
  if ('text_fix' in changes) {
    updated = { ...updated, text_fix: toDisplayOptionalValue(changes.text_fix) };
  }
  if ('apparat_nr' in changes && changes.apparat_nr !== undefined) {
    updated = { ...updated, apparat_nr: String(changes.apparat_nr) };
  }
  if ('apparat_id' in changes) {
    updated = { ...updated, apparat_id: changes.apparat_id as string };
  }
  if ('system_part_id' in changes) {
    updated = { ...updated, system_part_id: changes.system_part_id as string };
  }
  return updated;
}

function applySpecificationEdits(
  device: FieldDevice,
  changes: SpecificationInput,
  nowIso: string
): FieldDevice {
  const patch = normalizeSpecificationForDisplay(changes) ?? {};
  if (device.specification) {
    return { ...device, specification: { ...device.specification, ...patch } };
  }
  return {
    ...device,
    specification: {
      id: '',
      created_at: nowIso,
      updated_at: nowIso,
      field_device_id: device.id,
      ...patch
    }
  };
}
