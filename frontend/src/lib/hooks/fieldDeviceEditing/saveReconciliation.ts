import type {
  BacnetObjectInput,
  BulkUpdateFieldDeviceItem,
  BulkUpdateFieldDeviceResponse,
  FieldDevice,
  SpecificationInput
} from '$lib/domain/facility/index.js';
import { hasPartialPhaseSuccess } from './bulkUpdatePhases.js';
import { normalizeSpecificationForDisplay, toDisplayOptionalValue } from './specificationEdits.js';

export interface EditErrorInfo {
  message?: string;
  fields?: Record<string, string>;
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
  partialSuccessIds: Set<string>;
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
  const successIds = new Set<string>();
  const partialSuccessIds = new Set<string>();

  for (const item of result.results) {
    if (item.success) {
      successIds.add(item.id);
      continue;
    }

    if (item.error) {
      editErrors.set(
        item.id,
        localizeEditErrorInfo({ message: item.error, fields: item.fields }) ?? {}
      );
    }

    if (!item.fields) continue;

    const localizedFields = localizeFieldErrorMap(item.fields);
    const objectErrors = getBacnetObjectErrors(localizedFields);
    if (objectErrors.size > 0) {
      bacnetFieldErrors.set(item.id, objectErrors);
    }

    const update = updates.find((candidate) => candidate.id === item.id);
    if (update && hasPartialPhaseSuccess(update, item.fields)) {
      partialSuccessIds.add(item.id);
    }
  }

  const optimisticUpdates: FieldDevice[] = [];
  for (const id of successIds) {
    const device = storeItems.find((item) => item.id === id);
    if (device) {
      optimisticUpdates.push(
        applyAllEditsToDevice(device, pendingEdits, pendingBacnetEdits, nowIso)
      );
    }
  }

  for (const id of partialSuccessIds) {
    const device = storeItems.find((item) => item.id === id);
    if (!device) continue;

    const resultItem = result.results.find((item) => item.id === id);
    const failed = parseFailedFieldPaths(resultItem?.fields ?? {});
    const updated = applyPartialEditsToDevice(
      device,
      pendingEdits,
      pendingBacnetEdits,
      failed,
      nowIso
    );
    if (updated !== device) {
      upsertOptimisticUpdate(optimisticUpdates, updated);
    }
  }

  const remainingEdits = new Map(pendingEdits);
  const remainingBacnetEdits = new Map(pendingBacnetEdits);

  for (const id of successIds) {
    if (pendingEdits.get(id) === pendingEditsSnapshot.get(id)) {
      remainingEdits.delete(id);
    }
    if (pendingBacnetEdits.get(id) === pendingBacnetEditsSnapshot.get(id)) {
      remainingBacnetEdits.delete(id);
    }
    editErrors.delete(id);
  }

  for (const id of partialSuccessIds) {
    const resultItem = result.results.find((item) => item.id === id);
    if (!resultItem?.fields) continue;

    const failed = parseFailedFieldPaths(resultItem.fields);
    retainFailedBaseAndSpecificationEdits(remainingEdits, pendingEdits, id, failed);
    retainFailedBacnetEdits(remainingBacnetEdits, pendingBacnetEdits, id, failed);
  }

  return {
    editErrors,
    bacnetFieldErrors,
    remainingEdits,
    remainingBacnetEdits,
    successIds,
    partialSuccessIds,
    optimisticUpdates
  };
}

function getBacnetObjectErrors(
  fields: Record<string, string>
): Map<string, Record<string, string>> {
  const objectErrors = new Map<string, Record<string, string>>();
  for (const [fieldPath, message] of Object.entries(fields)) {
    const match = fieldPath.match(/(?:^|\.)bacnet_objects\.([0-9a-f-]+)\.(.+)$/i);
    if (!match) continue;
    const objectId = match[1];
    const field = match[2];
    const existing = objectErrors.get(objectId) || {};
    existing[field] = message;
    objectErrors.set(objectId, existing);
  }
  return objectErrors;
}

interface FailedFieldPaths {
  base: Set<string>;
  specification: Set<string>;
  bacnetObjects: Set<string>;
  entireBase: boolean;
  entireSpecification: boolean;
}

function parseFailedFieldPaths(fields: Record<string, string>): FailedFieldPaths {
  const failed: FailedFieldPaths = {
    base: new Set(),
    specification: new Set(),
    bacnetObjects: new Set(),
    entireBase: false,
    entireSpecification: false
  };

  for (const fieldPath of Object.keys(fields)) {
    if (fieldPath === 'fielddevice') {
      failed.entireBase = true;
    } else if (fieldPath.startsWith('fielddevice.')) {
      failed.entireBase = true;
      failed.base.add(fieldPath.replace('fielddevice.', ''));
    } else if (fieldPath === 'specification') {
      failed.entireSpecification = true;
    } else if (fieldPath.startsWith('specification.')) {
      failed.entireSpecification = true;
      failed.specification.add(fieldPath.replace('specification.', ''));
    } else if (fieldPath.startsWith('bacnet_objects.')) {
      const match = fieldPath.match(/^bacnet_objects\.([0-9a-f-]+)/);
      if (match) failed.bacnetObjects.add(match[1]);
    }
  }

  return failed;
}

function applyAllEditsToDevice(
  device: FieldDevice,
  pendingEdits: Map<string, Partial<BulkUpdateFieldDeviceItem>>,
  pendingBacnetEdits: Map<string, Map<string, Partial<BacnetObjectInput>>>,
  nowIso: string
): FieldDevice {
  return applyPartialEditsToDevice(
    device,
    pendingEdits,
    pendingBacnetEdits,
    {
      base: new Set(),
      specification: new Set(),
      bacnetObjects: new Set(),
      entireBase: false,
      entireSpecification: false
    },
    nowIso
  );
}

function applyPartialEditsToDevice(
  device: FieldDevice,
  pendingEdits: Map<string, Partial<BulkUpdateFieldDeviceItem>>,
  pendingBacnetEdits: Map<string, Map<string, Partial<BacnetObjectInput>>>,
  failed: FailedFieldPaths,
  nowIso: string
): FieldDevice {
  const changes = pendingEdits.get(device.id);
  let updated: FieldDevice = device;

  if (changes) {
    updated = applySuccessfulBaseEdits(updated, changes, failed);
    updated = applySuccessfulSpecificationEdits(updated, changes, failed, nowIso);
  }

  const bacnetEdits = pendingBacnetEdits.get(device.id);
  if (bacnetEdits && device.bacnet_objects) {
    updated = {
      ...updated,
      bacnet_objects: device.bacnet_objects.map((object) => {
        if (failed.bacnetObjects.has(object.id)) return object;
        const edits = bacnetEdits.get(object.id);
        return edits ? { ...object, ...edits } : object;
      })
    };
  }

  return updated;
}

function applySuccessfulBaseEdits(
  device: FieldDevice,
  changes: Partial<BulkUpdateFieldDeviceItem>,
  failed: FailedFieldPaths
): FieldDevice {
  if (failed.entireBase) return device;

  let updated = device;
  if ('bmk' in changes && !failed.base.has('bmk')) {
    updated = { ...updated, bmk: toDisplayOptionalValue(changes.bmk) };
  }
  if ('description' in changes && !failed.base.has('description')) {
    updated = { ...updated, description: toDisplayOptionalValue(changes.description) };
  }
  if ('text_fix' in changes && !failed.base.has('text_fix')) {
    updated = { ...updated, text_fix: toDisplayOptionalValue(changes.text_fix) };
  }
  if (
    'apparat_nr' in changes &&
    !failed.base.has('apparat_nr') &&
    changes.apparat_nr !== undefined
  ) {
    updated = { ...updated, apparat_nr: String(changes.apparat_nr) };
  }
  if ('apparat_id' in changes && !failed.base.has('apparat_id')) {
    updated = { ...updated, apparat_id: changes.apparat_id as string };
  }
  if ('system_part_id' in changes && !failed.base.has('system_part_id')) {
    updated = { ...updated, system_part_id: changes.system_part_id as string };
  }
  return updated;
}

function applySuccessfulSpecificationEdits(
  device: FieldDevice,
  changes: Partial<BulkUpdateFieldDeviceItem>,
  failed: FailedFieldPaths,
  nowIso: string
): FieldDevice {
  const specChanges = changes.specification;
  if (
    !specChanges ||
    failed.entireSpecification ||
    failed.specification.size >= Object.keys(specChanges).length
  ) {
    return device;
  }

  const successfulSpecPatch: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(specChanges)) {
    if (!failed.specification.has(key) && value !== undefined) {
      successfulSpecPatch[key] = value;
    }
  }
  if (Object.keys(successfulSpecPatch).length === 0) return device;

  const displayPatch = normalizeSpecificationForDisplay(successfulSpecPatch as SpecificationInput);
  if (device.specification) {
    return {
      ...device,
      specification: {
        ...device.specification,
        ...(displayPatch ?? {})
      }
    };
  }

  return {
    ...device,
    specification: {
      id: '',
      created_at: nowIso,
      updated_at: nowIso,
      field_device_id: device.id,
      specification_supplier: undefined,
      specification_brand: undefined,
      specification_type: undefined,
      additional_info_motor_valve: undefined,
      additional_info_size: undefined,
      additional_information_installation_location: undefined,
      electrical_connection_ph: undefined,
      electrical_connection_acdc: undefined,
      electrical_connection_amperage: undefined,
      electrical_connection_power: undefined,
      electrical_connection_rotation: undefined,
      ...(displayPatch ?? {})
    }
  };
}

function upsertOptimisticUpdate(updates: FieldDevice[], update: FieldDevice): void {
  const existingIndex = updates.findIndex((item) => item.id === update.id);
  if (existingIndex >= 0) {
    updates[existingIndex] = update;
  } else {
    updates.push(update);
  }
}

function retainFailedBaseAndSpecificationEdits(
  remainingEdits: Map<string, Partial<BulkUpdateFieldDeviceItem>>,
  pendingEdits: Map<string, Partial<BulkUpdateFieldDeviceItem>>,
  deviceId: string,
  failed: FailedFieldPaths
): void {
  const changes = pendingEdits.get(deviceId);
  if (!changes) return;

  const onlyFailedFields: Partial<BulkUpdateFieldDeviceItem> = {};
  retainFailedBaseFields(onlyFailedFields, changes, failed);
  retainFailedSpecificationFields(onlyFailedFields, changes, failed);

  if (Object.keys(onlyFailedFields).length > 0) {
    remainingEdits.set(deviceId, onlyFailedFields);
  } else {
    remainingEdits.delete(deviceId);
  }
}

function retainFailedBaseFields(
  target: Partial<BulkUpdateFieldDeviceItem>,
  changes: Partial<BulkUpdateFieldDeviceItem>,
  failed: FailedFieldPaths
): void {
  const shouldKeep = (field: string) => failed.entireBase || failed.base.has(field);
  if ('bmk' in changes && shouldKeep('bmk')) target.bmk = changes.bmk;
  if ('description' in changes && shouldKeep('description'))
    target.description = changes.description;
  if ('text_fix' in changes && shouldKeep('text_fix')) target.text_fix = changes.text_fix;
  if ('apparat_nr' in changes && shouldKeep('apparat_nr')) target.apparat_nr = changes.apparat_nr;
  if ('apparat_id' in changes && shouldKeep('apparat_id')) target.apparat_id = changes.apparat_id;
  if ('system_part_id' in changes && shouldKeep('system_part_id')) {
    target.system_part_id = changes.system_part_id;
  }
}

function retainFailedSpecificationFields(
  target: Partial<BulkUpdateFieldDeviceItem>,
  changes: Partial<BulkUpdateFieldDeviceItem>,
  failed: FailedFieldPaths
): void {
  const specChanges = changes.specification;
  if (!specChanges) return;

  if (failed.entireSpecification) {
    target.specification = specChanges;
    return;
  }

  if (failed.specification.size === 0) return;

  const onlyFailedSpecFields: SpecificationInput = {};
  for (const [key, value] of Object.entries(specChanges)) {
    if (failed.specification.has(key)) {
      (onlyFailedSpecFields as Record<string, unknown>)[key] = value;
    }
  }
  if (Object.keys(onlyFailedSpecFields).length > 0) {
    target.specification = onlyFailedSpecFields;
  }
}

function retainFailedBacnetEdits(
  remainingBacnetEdits: Map<string, Map<string, Partial<BacnetObjectInput>>>,
  pendingBacnetEdits: Map<string, Map<string, Partial<BacnetObjectInput>>>,
  deviceId: string,
  failed: FailedFieldPaths
): void {
  const bacnetEdits = pendingBacnetEdits.get(deviceId);
  if (!bacnetEdits) return;

  const onlyFailedBacnetEdits = new Map<string, Partial<BacnetObjectInput>>();
  for (const [objectId, edits] of bacnetEdits.entries()) {
    if (failed.bacnetObjects.has(objectId)) {
      onlyFailedBacnetEdits.set(objectId, edits);
    }
  }

  if (onlyFailedBacnetEdits.size > 0) {
    remainingBacnetEdits.set(deviceId, onlyFailedBacnetEdits);
  } else {
    remainingBacnetEdits.delete(deviceId);
  }
}
