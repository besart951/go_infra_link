/**
 * useFieldDeviceEditing - Composable for field device inline editing state
 *
 * Extracts all editing state + mutation functions from FieldDeviceListView.
 * Pattern follows useFormState.svelte.ts (getter-based for Svelte 5 reactivity).
 */

import {
  getFieldError as resolveFieldError,
  getFieldErrors,
  localizeErrorText,
  localizeFieldErrorMap
} from '$lib/api/client.js';
import {
  fieldErrorPathMatches,
  findFieldPathSegment,
  findIndexedFieldPathSegment,
  normalizeFieldPathSegment,
  splitFieldPath
} from '$lib/api/fieldPath.js';
import { fieldDeviceRepository } from '$lib/infrastructure/api/fieldDeviceRepository.js';
import { addToast } from '$lib/components/toast.svelte';
import { sessionStorage } from '$lib/services/sessionStorageService.js';
import { t as translate } from '$lib/i18n/index.js';
import { buildBacnetObjectsPayload } from './fieldDeviceEditing/bacnetPayload.js';
import { validateBacnetObjectEdits } from './fieldDeviceEditing/bacnetValidation.js';
import { reconcileFieldDeviceSaveResult } from './fieldDeviceEditing/saveReconciliation.js';
import {
  getFieldDeviceEditingStorageKey,
  loadPersistedFieldDeviceEditingState,
  removePersistedFieldDeviceEditingState,
  savePersistedFieldDeviceEditingState
} from './fieldDeviceEditing/persistence.js';
import {
  getChangedFieldsByDevice as collectChangedFieldsByDevice,
  getPendingDeviceIds as collectPendingDeviceIds,
  getPendingSpecValue as resolvePendingSpecValue,
  getPendingValue as resolvePendingValue,
  hasPendingBacnetEdits as detectPendingBacnetEdits,
  hasPendingBaseEdits as detectPendingBaseEdits,
  hasPendingSpecificationEdits as detectPendingSpecificationEdits,
  isFieldDirty as detectFieldDirty,
  isSpecFieldDirty as detectSpecFieldDirty
} from './fieldDeviceEditing/pendingInventory.js';
import {
  buildSpecificationPatch,
  normalizeSpecificationForDisplay,
  toDisplayOptionalValue
} from './fieldDeviceEditing/specificationEdits.js';
import { buildFieldDeviceUpdatePayload } from './fieldDeviceEditing/updatePayload.js';
import type {
  FieldDevice,
  UpdateFieldDeviceRequest,
  BulkUpdateFieldDeviceItem,
  SpecificationInput,
  BacnetObjectInput
} from '$lib/domain/facility/index.js';
import type { SharedFieldDeviceDraftDevice } from './fieldDeviceEditing/pendingInventory.js';
import type { EditErrorInfo } from './fieldDeviceEditing/saveReconciliation.js';

export type { EditErrorInfo } from './fieldDeviceEditing/saveReconciliation.js';

type ProjectIdInput = string | undefined | (() => string | undefined);

interface SharedFieldDeviceDraftState {
  devices: SharedFieldDeviceDraftDevice[];
}

interface UseFieldDeviceEditingOptions {
  projectId?: ProjectIdInput;
  onSharedStateChange?: (state: SharedFieldDeviceDraftState) => void;
  onSaveSuccess?: (deviceIds: string[]) => void;
}

function createBacnetObjectEditMap(
  entries?: Iterable<[string, Partial<BacnetObjectInput>]>
): Map<string, Partial<BacnetObjectInput>> {
  return new Map(entries);
}

const FIELD_DEVICE_ERROR_PREFIXES = [
  'fielddevice',
  'field_device',
  'field_devices',
  'specification',
  'specifications',
  'fielddevice.specification',
  'fielddevice.specifications',
  'field_device.specification',
  'field_device.specifications',
  'field_devices.specification',
  'field_devices.specifications',
  'updates.fielddevice',
  'updates.field_device',
  'updates.field_devices',
  'updates.specification',
  'updates.specifications'
];

export function useFieldDeviceEditing(options: UseFieldDeviceEditingOptions = {}) {
  const resolvedProjectId =
    typeof options.projectId === 'function' ? options.projectId() : options.projectId;
  const storageKey = getFieldDeviceEditingStorageKey(resolvedProjectId);

  // Load persisted state on initialization
  const persistedState = loadPersistedFieldDeviceEditingState(sessionStorage, storageKey);

  // Pending edits state for inline editing
  let pendingEdits = $state<Map<string, Partial<BulkUpdateFieldDeviceItem>>>(
    persistedState ? new Map(persistedState.edits) : new Map()
  );

  // BACnet pending edits: deviceId -> (objectId -> partial edits)
  let pendingBacnetEdits = $state<Map<string, Map<string, Partial<BacnetObjectInput>>>>(
    persistedState
      ? new Map(
          persistedState.bacnetEdits.map(([deviceId, entries]) => [deviceId, new Map(entries)])
        )
      : new Map()
  );

  // Error tracking state per device ID
  let editErrors = $state<Map<string, EditErrorInfo>>(new Map());

  // BACnet field errors from server: deviceId -> (objectId -> { field: error })
  let bacnetFieldErrors = $state<Map<string, Map<string, Record<string, string>>>>(new Map());
  // BACnet client-side validation errors: deviceId -> (objectId -> { field: error })
  let bacnetClientErrors = $state<Map<string, Map<string, Record<string, string>>>>(new Map());

  /**
   * Persistence: Save current state to sessionStorage
   */
  function savePersistedState() {
    savePersistedFieldDeviceEditingState(
      sessionStorage,
      storageKey,
      pendingEdits,
      pendingBacnetEdits
    );
  }

  function getPendingDeviceIds(): string[] {
    return collectPendingDeviceIds(pendingEdits, pendingBacnetEdits);
  }

  function getChangedFieldsByDevice(): SharedFieldDeviceDraftDevice[] {
    return collectChangedFieldsByDevice(pendingEdits, pendingBacnetEdits);
  }

  function emitSharedState() {
    options.onSharedStateChange?.({
      devices: getChangedFieldsByDevice()
    });
  }

  /**
   * Auto-save to sessionStorage whenever edits change
   */
  $effect(() => {
    // Track pendingEdits and pendingBacnetEdits for changes
    const _editsSize = pendingEdits.size;
    const _bacnetSize = pendingBacnetEdits.size;
    const _editIds = [...pendingEdits.keys()].join('|');
    const _bacnetIds = [...pendingBacnetEdits.keys()].join('|');

    // Save to sessionStorage
    savePersistedState();
    emitSharedState();
  });

  function setEditError(deviceId: string, info?: EditErrorInfo) {
    const next = new Map(editErrors);
    if (info) {
      next.set(deviceId, info);
    } else {
      next.delete(deviceId);
    }
    editErrors = next;
  }

  function queueEdit(deviceId: string, field: keyof BulkUpdateFieldDeviceItem, value: unknown) {
    const existing = pendingEdits.get(deviceId) || {};
    pendingEdits = new Map(pendingEdits).set(deviceId, { ...existing, [field]: value });
    // Clear any existing error for this device when editing
    setEditError(deviceId);
  }

  function queueSpecEdit(deviceId: string, field: keyof SpecificationInput, value: unknown) {
    const existing = pendingEdits.get(deviceId) || {};
    const existingSpec = existing.specification || {};
    const newSpec = { ...existingSpec, [field]: value };
    pendingEdits = new Map(pendingEdits).set(deviceId, {
      ...existing,
      specification: newSpec
    });
    // Clear any existing error
    setEditError(deviceId);
  }

  function hasPendingFieldDeviceEditsForDevice(deviceId: string): boolean {
    const changes = pendingEdits.get(deviceId);
    if (!changes) return false;

    for (const [key, value] of Object.entries(changes)) {
      if (key === 'specification') {
        if (value && Object.keys(value).length > 0) return true;
        continue;
      }
      return true;
    }

    return false;
  }

  function hasPendingBacnetEditsForDevice(deviceId: string): boolean {
    return (pendingBacnetEdits.get(deviceId)?.size ?? 0) > 0;
  }

  function hasPendingEditsForDevice(deviceId: string): boolean {
    return (
      hasPendingFieldDeviceEditsForDevice(deviceId) || hasPendingBacnetEditsForDevice(deviceId)
    );
  }

  function discardFieldEdit(deviceId: string, field: keyof BulkUpdateFieldDeviceItem): void {
    const changes = pendingEdits.get(deviceId);
    if (!changes || !(field in changes)) return;

    const { [field]: _removed, ...remaining } = changes;
    setPendingEditsForDevice(deviceId, remaining);
    clearEditErrorFields(deviceId, [
      String(field),
      `fielddevice.${String(field)}`,
      `data.fielddevice.${String(field)}`,
      `error.fielddevice.${String(field)}`
    ]);
  }

  function discardSpecEdit(deviceId: string, field: keyof SpecificationInput): void {
    const changes = pendingEdits.get(deviceId);
    const specChanges = changes?.specification;
    if (!changes || !specChanges || !(field in specChanges)) return;

    const { [field]: _removed, ...remainingSpec } = specChanges;
    const remaining: Partial<BulkUpdateFieldDeviceItem> = { ...changes };
    if (Object.keys(remainingSpec).length > 0) {
      remaining.specification = remainingSpec;
    } else {
      delete remaining.specification;
    }

    setPendingEditsForDevice(deviceId, remaining);
    clearEditErrorFields(deviceId, [
      String(field),
      `specification.${String(field)}`,
      `data.specification.${String(field)}`,
      `error.specification.${String(field)}`
    ]);
  }

  function discardDeviceFieldEdits(deviceId: string): void {
    if (!pendingEdits.has(deviceId)) return;
    const remaining = new Map(pendingEdits);
    remaining.delete(deviceId);
    pendingEdits = remaining;
    setEditError(deviceId);
  }

  function discardDeviceEdits(deviceId: string): void {
    discardDeviceFieldEdits(deviceId);
    discardDeviceBacnetEdits(deviceId);
  }

  function setPendingEditsForDevice(
    deviceId: string,
    changes: Partial<BulkUpdateFieldDeviceItem>
  ): void {
    const next = new Map(pendingEdits);
    if (hasFieldDeviceChangePayload(changes)) {
      next.set(deviceId, changes);
    } else {
      next.delete(deviceId);
    }
    pendingEdits = next;
  }

  function hasFieldDeviceChangePayload(changes: Partial<BulkUpdateFieldDeviceItem>): boolean {
    for (const [key, value] of Object.entries(changes)) {
      if (key === 'specification') {
        if (value && Object.keys(value).length > 0) return true;
        continue;
      }
      return true;
    }
    return false;
  }

  function clearEditErrorFields(deviceId: string, fieldKeys: string[]): void {
    const info = editErrors.get(deviceId);
    if (!info?.fields && !info?.suggestions && !info?.suggestionOptions) return;

    const fields = { ...(info.fields ?? {}) };
    const suggestions = { ...(info.suggestions ?? {}) };
    const suggestionOptions = { ...(info.suggestionOptions ?? {}) };
    for (const key of fieldKeys) {
      delete fields[key];
      delete suggestions[key];
      delete suggestionOptions[key];
    }
    for (const key of Object.keys(fields)) {
      if (fieldKeys.some((candidate) => fieldErrorPathMatches(key, candidate))) {
        delete fields[key];
      }
    }
    for (const key of Object.keys(suggestions)) {
      if (fieldKeys.some((candidate) => fieldErrorPathMatches(key, candidate))) {
        delete suggestions[key];
      }
    }
    for (const key of Object.keys(suggestionOptions)) {
      if (fieldKeys.some((candidate) => fieldErrorPathMatches(key, candidate))) {
        delete suggestionOptions[key];
      }
    }

    const next = new Map(editErrors);
    if (
      Object.keys(fields).length > 0 ||
      Object.keys(suggestions).length > 0 ||
      Object.keys(suggestionOptions).length > 0
    ) {
      next.set(deviceId, {
        ...info,
        fields: Object.keys(fields).length > 0 ? fields : undefined,
        suggestions: Object.keys(suggestions).length > 0 ? suggestions : undefined,
        suggestionOptions: Object.keys(suggestionOptions).length > 0 ? suggestionOptions : undefined
      });
    } else {
      next.delete(deviceId);
    }
    editErrors = next;
  }

  function buildUpdateForDevice(
    deviceId: string,
    storeItems: FieldDevice[],
    options: { includeBacnet: boolean }
  ): BulkUpdateFieldDeviceItem | null {
    return buildFieldDeviceUpdatePayload({
      deviceId,
      storeItems,
      pendingEdits,
      pendingBacnetEdits,
      includeBacnet: options.includeBacnet
    });
  }

  function applyEditsToDevice(
    device: FieldDevice,
    options: { includeBacnet: boolean }
  ): FieldDevice {
    const changes = pendingEdits.get(device.id);
    const bacnetEdits = pendingBacnetEdits.get(device.id);
    let updated: FieldDevice = { ...device };

    if (changes) {
      if ('bmk' in changes) {
        updated = { ...updated, bmk: toDisplayOptionalValue(changes.bmk) };
      }
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

      const specPatch = buildSpecificationPatch(changes.specification);
      if (specPatch) {
        const displaySpecPatch = normalizeSpecificationForDisplay(specPatch);
        if (!displaySpecPatch) {
          return updated;
        }
        if (updated.specification) {
          // Update existing specification
          updated = {
            ...updated,
            specification: { ...updated.specification, ...displaySpecPatch }
          };
        } else {
          // Create new specification optimistically
          updated = {
            ...updated,
            specification: {
              id: '', // Temporary, will be filled on next refresh
              created_at: new Date().toISOString(),
              updated_at: new Date().toISOString(),
              field_device_id: updated.id,
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
              ...displaySpecPatch
            }
          };
        }
      }
    }

    if (options.includeBacnet && bacnetEdits && bacnetEdits.size > 0 && device.bacnet_objects) {
      updated = {
        ...updated,
        bacnet_objects: device.bacnet_objects.map((obj) => {
          const edits = bacnetEdits.get(obj.id);
          return edits ? { ...obj, ...edits } : obj;
        })
      };
    }

    return updated;
  }

  function validatePendingEdits(deviceId: string): EditErrorInfo | null {
    const changes = pendingEdits.get(deviceId);
    if (!changes) return null;

    const fields: Record<string, string> = {};
    const bmk = changes.bmk;
    if (bmk !== undefined && bmk !== null && String(bmk).length > 10) {
      fields['fielddevice.bmk'] = translate('field_device.validation.bmk_max');
    }
    const description = changes.description;
    if (description !== undefined && description !== null && String(description).length > 250) {
      fields['fielddevice.description'] = translate('field_device.validation.description_max');
    }
    const apparatNr = changes.apparat_nr;
    if (apparatNr !== undefined && apparatNr !== null) {
      const nr = Number(apparatNr);
      if (Number.isNaN(nr) || nr < 1 || nr > 99) {
        fields['fielddevice.apparat_nr'] = translate('field_device.validation.apparat_nr_range');
      }
    }
    if ('apparat_id' in changes && !String(changes.apparat_id ?? '').trim()) {
      fields['fielddevice.apparat_id'] = translate('validation.required', {
        field: translate('field_device.table.apparat')
      });
    }
    if ('system_part_id' in changes && !String(changes.system_part_id ?? '').trim()) {
      fields['fielddevice.system_part_id'] = translate('validation.required', {
        field: translate('field_device.table.system_part')
      });
    }

    const spec = changes.specification;
    if (spec) {
      const checkMax = (key: keyof SpecificationInput, label: string) => {
        const value = spec[key];
        if (value !== undefined && value !== null && String(value).length > 250) {
          fields[`specification.${key}`] = translate('field_device.validation.spec_max', {
            label
          });
        }
      };
      checkMax('specification_supplier', translate('field_device.validation.supplier'));
      checkMax('specification_brand', translate('field_device.validation.brand'));
      checkMax('specification_type', translate('field_device.validation.type'));
      checkMax('additional_info_motor_valve', translate('field_device.validation.motor_valve'));
      checkMax(
        'additional_information_installation_location',
        translate('field_device.validation.install_location')
      );

      const acdc = spec.electrical_connection_acdc;
      if (acdc !== undefined && acdc !== null && String(acdc).length !== 2) {
        fields['specification.electrical_connection_acdc'] = translate(
          'field_device.validation.acdc_length'
        );
      }
    }

    if (Object.keys(fields).length > 0) {
      return { message: 'validation_error', fields };
    }
    return null;
  }

  function getFirstFieldValidationToast(
    fields: Record<string, string> | undefined,
    fallback = translate('field_device.editing.toasts.fix_validation')
  ): string {
    if (!fields) return fallback;
    const first = Object.entries(fields)[0];
    if (!first) return fallback;
    return localizeErrorText(first[1], first[0]);
  }

  function getFirstEditValidationToast(
    errors: Map<string, EditErrorInfo>,
    fallback = translate('field_device.editing.toasts.fix_validation')
  ): string {
    for (const info of errors.values()) {
      if (!info?.fields) continue;
      const first = Object.entries(info.fields)[0];
      if (first) {
        return localizeErrorText(first[1], first[0]);
      }
    }
    return fallback;
  }

  function getFirstBacnetClientValidationToast(
    deviceId?: string,
    fallback = translate('field_device.editing.toasts.fix_validation')
  ): string {
    if (deviceId) {
      const deviceErrors = bacnetClientErrors.get(deviceId);
      if (deviceErrors) {
        for (const [objectId, fieldErrors] of deviceErrors.entries()) {
          const first = Object.entries(fieldErrors)[0];
          if (first) {
            return first[1];
          }
        }
      }
      return fallback;
    }

    for (const deviceErrors of bacnetClientErrors.values()) {
      for (const fieldErrors of deviceErrors.values()) {
        const first = Object.entries(fieldErrors)[0];
        if (first) {
          return first[1];
        }
      }
    }

    return fallback;
  }

  function isFieldDirty(deviceId: string, field: keyof UpdateFieldDeviceRequest): boolean {
    return detectFieldDirty(pendingEdits, deviceId, field);
  }

  function isSpecFieldDirty(deviceId: string, field: keyof SpecificationInput): boolean {
    return detectSpecFieldDirty(pendingEdits, deviceId, field);
  }

  function getPendingValue(
    deviceId: string,
    field: keyof BulkUpdateFieldDeviceItem
  ): string | undefined {
    return resolvePendingValue(pendingEdits, deviceId, field);
  }

  function getPendingSpecValue(
    deviceId: string,
    field: keyof SpecificationInput
  ): string | undefined {
    return resolvePendingSpecValue(pendingEdits, deviceId, field);
  }

  function getFieldError(deviceId: string, field: string): string | undefined {
    const errorInfo = editErrors.get(deviceId);
    if (!errorInfo) return undefined;

    if (errorInfo.fields && Object.keys(errorInfo.fields).length > 0) {
      return resolveFieldError(errorInfo.fields, field, FIELD_DEVICE_ERROR_PREFIXES);
    }

    return undefined;
  }

  function getFieldPathVariants(field: string): string[] {
    return [
      field,
      `fielddevice.${field}`,
      `field_device.${field}`,
      `field_devices.${field}`,
      `specification.${field}`,
      `specifications.${field}`,
      `fielddevice.specification.${field}`,
      `fielddevice.specifications.${field}`,
      `field_device.specification.${field}`,
      `field_device.specifications.${field}`,
      `field_devices.specification.${field}`,
      `field_devices.specifications.${field}`,
      `data.fielddevice.${field}`,
      `error.fielddevice.${field}`,
      `data.specification.${field}`,
      `error.specification.${field}`
    ];
  }

  function resolveFieldMapValue<T>(
    values: Record<string, T> | undefined,
    field: string
  ): T | undefined {
    if (!values || Object.keys(values).length === 0) return undefined;
    for (const key of getFieldPathVariants(field)) {
      if (values[key] !== undefined) return values[key];
    }
    for (const [key, value] of Object.entries(values)) {
      for (const candidate of getFieldPathVariants(field)) {
        if (fieldErrorPathMatches(key, candidate)) return value;
      }
    }
    return undefined;
  }

  function getFieldSuggestion(
    deviceId: string,
    field: string,
    storeItems: FieldDevice[] = []
  ): number | undefined {
    const errorInfo = editErrors.get(deviceId);
    const staticSuggestion = resolveFieldMapValue(errorInfo?.suggestions, field);
    const options = resolveFieldMapValue(errorInfo?.suggestionOptions, field);

    if (field !== 'apparat_nr' || !options || options.length === 0) {
      return staticSuggestion;
    }

    return getDynamicApparatNrSuggestion(deviceId, options, storeItems) ?? staticSuggestion;
  }

  function getDynamicApparatNrSuggestion(
    deviceId: string,
    options: number[],
    storeItems: FieldDevice[]
  ): number | undefined {
    const targetDevice = storeItems.find((item) => item.id === deviceId);
    if (!targetDevice) return undefined;

    const targetScope = getEffectiveApparatNrScope(targetDevice);
    const occupied = new Set<number>();

    for (const device of storeItems) {
      if (device.id === deviceId) continue;
      if (!isSameApparatNrScope(targetScope, getEffectiveApparatNrScope(device))) continue;

      const changes = pendingEdits.get(device.id);
      if (!changes || !hasApparatNrConstraintDraft(changes)) continue;

      const nr = Number('apparat_nr' in changes ? changes.apparat_nr : device.apparat_nr);
      if (Number.isInteger(nr) && nr >= 1 && nr <= 99) {
        occupied.add(nr);
      }
    }

    return [...options].sort((a, b) => a - b).find((candidate) => !occupied.has(candidate));
  }

  function hasApparatNrConstraintDraft(changes: Partial<BulkUpdateFieldDeviceItem>): boolean {
    return 'apparat_nr' in changes || 'apparat_id' in changes || 'system_part_id' in changes;
  }

  function getEffectiveApparatNrScope(device: FieldDevice) {
    const changes = pendingEdits.get(device.id);
    return {
      spsControllerSystemTypeId: device.sps_controller_system_type_id,
      apparatId:
        typeof changes?.apparat_id === 'string' && changes.apparat_id
          ? changes.apparat_id
          : device.apparat_id,
      systemPartId:
        typeof changes?.system_part_id === 'string' && changes.system_part_id
          ? changes.system_part_id
          : (device.system_part_id ?? '')
    };
  }

  function isSameApparatNrScope(
    a: ReturnType<typeof getEffectiveApparatNrScope>,
    b: ReturnType<typeof getEffectiveApparatNrScope>
  ): boolean {
    return (
      a.spsControllerSystemTypeId === b.spsControllerSystemTypeId &&
      a.apparatId === b.apparatId &&
      a.systemPartId === b.systemPartId
    );
  }

  function localizeEditErrorInfo(info?: EditErrorInfo): EditErrorInfo | undefined {
    if (!info) return undefined;
    const localized = {
      message: info.message ? localizeErrorText(info.message) : info.message,
      fields: info.fields ? localizeFieldErrorMap(info.fields) : info.fields,
      suggestions: info.suggestions,
      suggestionOptions: info.suggestionOptions
    };
    if (
      !localized.message &&
      (!localized.fields || Object.keys(localized.fields).length === 0) &&
      (!localized.suggestions || Object.keys(localized.suggestions).length === 0) &&
      (!localized.suggestionOptions || Object.keys(localized.suggestionOptions).length === 0)
    ) {
      return undefined;
    }
    return localized;
  }

  // BACnet edit queuing
  function queueBacnetEdit(deviceId: string, objectId: string, field: string, value: unknown) {
    let deviceEdits = pendingBacnetEdits.get(deviceId);
    if (!deviceEdits) {
      deviceEdits = createBacnetObjectEditMap();
    }
    const objectEdits = deviceEdits.get(objectId) || {};
    deviceEdits.set(objectId, { ...objectEdits, [field]: value } as Partial<BacnetObjectInput>);
    const nextPendingBacnetEdits = new Map(pendingBacnetEdits);
    nextPendingBacnetEdits.set(deviceId, new Map(deviceEdits));
    pendingBacnetEdits = nextPendingBacnetEdits;
    clearBacnetFieldError(deviceId, objectId, field);
  }

  function clearBacnetFieldError(deviceId: string, objectId: string, field: string) {
    bacnetFieldErrors = clearNestedBacnetFieldError(bacnetFieldErrors, deviceId, objectId, field);
    bacnetClientErrors = clearNestedBacnetFieldError(bacnetClientErrors, deviceId, objectId, field);
  }

  function clearNestedBacnetFieldError(
    source: Map<string, Map<string, Record<string, string>>>,
    deviceId: string,
    objectId: string,
    field: string
  ): Map<string, Map<string, Record<string, string>>> {
    const deviceErrors = source.get(deviceId);
    if (!deviceErrors) return source;

    const objectErrors = deviceErrors.get(objectId);
    if (!objectErrors || !(field in objectErrors)) return source;

    const { [field]: _removed, ...remainingObjectErrors } = objectErrors;
    const nextDeviceErrors = new Map(deviceErrors);
    if (Object.keys(remainingObjectErrors).length > 0) {
      nextDeviceErrors.set(objectId, remainingObjectErrors);
    } else {
      nextDeviceErrors.delete(objectId);
    }

    const next = new Map(source);
    if (nextDeviceErrors.size > 0) {
      next.set(deviceId, nextDeviceErrors);
    } else {
      next.delete(deviceId);
    }
    return next;
  }

  function discardBacnetObjectFieldEdit(deviceId: string, objectId: string, field: string): void {
    const deviceEdits = pendingBacnetEdits.get(deviceId);
    const objectEdits = deviceEdits?.get(objectId);
    if (!deviceEdits || !objectEdits || !(field in objectEdits)) return;

    const remainingObjectEdits = { ...objectEdits };
    delete (remainingObjectEdits as Record<string, unknown>)[field];
    const nextDeviceEdits = new Map(deviceEdits);
    if (Object.keys(remainingObjectEdits).length > 0) {
      nextDeviceEdits.set(objectId, remainingObjectEdits);
    } else {
      nextDeviceEdits.delete(objectId);
    }

    const nextPendingBacnetEdits = new Map(pendingBacnetEdits);
    if (nextDeviceEdits.size > 0) {
      nextPendingBacnetEdits.set(deviceId, nextDeviceEdits);
    } else {
      nextPendingBacnetEdits.delete(deviceId);
    }
    pendingBacnetEdits = nextPendingBacnetEdits;
    clearBacnetFieldError(deviceId, objectId, field);
  }

  function discardBacnetObjectEdits(deviceId: string, objectId: string): void {
    const deviceEdits = pendingBacnetEdits.get(deviceId);
    if (!deviceEdits?.has(objectId)) return;

    const nextDeviceEdits = new Map(deviceEdits);
    nextDeviceEdits.delete(objectId);

    const nextPendingBacnetEdits = new Map(pendingBacnetEdits);
    if (nextDeviceEdits.size > 0) {
      nextPendingBacnetEdits.set(deviceId, nextDeviceEdits);
    } else {
      nextPendingBacnetEdits.delete(deviceId);
    }
    pendingBacnetEdits = nextPendingBacnetEdits;
    clearBacnetObjectErrors(deviceId, objectId);
  }

  function discardDeviceBacnetEdits(deviceId: string): void {
    if (!pendingBacnetEdits.has(deviceId)) return;

    const remainingBacnet = new Map(pendingBacnetEdits);
    remainingBacnet.delete(deviceId);
    pendingBacnetEdits = remainingBacnet;

    const nextFieldErrors = new Map(bacnetFieldErrors);
    nextFieldErrors.delete(deviceId);
    bacnetFieldErrors = nextFieldErrors;

    const nextClientErrors = new Map(bacnetClientErrors);
    nextClientErrors.delete(deviceId);
    bacnetClientErrors = nextClientErrors;
  }

  function clearBacnetObjectErrors(deviceId: string, objectId: string): void {
    const clearObject = (source: Map<string, Map<string, Record<string, string>>>) => {
      const deviceErrors = source.get(deviceId);
      if (!deviceErrors?.has(objectId)) return source;
      const nextDeviceErrors = new Map(deviceErrors);
      nextDeviceErrors.delete(objectId);
      const next = new Map(source);
      if (nextDeviceErrors.size > 0) {
        next.set(deviceId, nextDeviceErrors);
      } else {
        next.delete(deviceId);
      }
      return next;
    };

    bacnetFieldErrors = clearObject(bacnetFieldErrors);
    bacnetClientErrors = clearObject(bacnetClientErrors);
  }

  function validateBacnetEdits(items: FieldDevice[], deviceId: string): boolean {
    const device = items.find((d) => d.id === deviceId);
    const deviceEdits = pendingBacnetEdits.get(deviceId);
    const errors = validateBacnetObjectEdits({ device, deviceEdits, translate });

    if (errors.size > 0) {
      bacnetClientErrors = new Map(bacnetClientErrors).set(deviceId, errors);
      return false;
    }

    const newClientErrors = new Map(bacnetClientErrors);
    newClientErrors.delete(deviceId);
    bacnetClientErrors = newClientErrors;
    return true;
  }

  function applyBulkApiFieldErrors(
    updates: BulkUpdateFieldDeviceItem[],
    fieldErrors: Record<string, string>
  ): boolean {
    if (Object.keys(fieldErrors).length === 0) return false;

    const nextEditErrors = new Map(editErrors);
    const nextBacnetErrors = new Map(bacnetFieldErrors);
    let applied = false;

    for (const [path, message] of Object.entries(fieldErrors)) {
      const target = resolveBulkUpdateFieldError(updates, path);
      if (!target) continue;

      const existing = nextEditErrors.get(target.deviceId) ?? {
        message: 'validation_error'
      };
      nextEditErrors.set(target.deviceId, {
        ...existing,
        fields: {
          ...(existing.fields ?? {}),
          [target.fieldPath]: message
        }
      });

      if (target.bacnetObjectId && target.bacnetField) {
        const objectErrors = new Map(nextBacnetErrors.get(target.deviceId) ?? new Map());
        objectErrors.set(target.bacnetObjectId, {
          ...(objectErrors.get(target.bacnetObjectId) ?? {}),
          [target.bacnetField]: message
        });
        nextBacnetErrors.set(target.deviceId, objectErrors);
      }

      applied = true;
    }

    if (!applied) return false;

    editErrors = nextEditErrors;
    bacnetFieldErrors = nextBacnetErrors;
    return true;
  }

  function resolveBulkUpdateFieldError(
    updates: BulkUpdateFieldDeviceItem[],
    fieldPath: string
  ):
    | {
        deviceId: string;
        fieldPath: string;
        bacnetObjectId?: string;
        bacnetField?: string;
      }
    | undefined {
    const segments = splitFieldPath(fieldPath).filter((segment) => {
      const normalized = normalizeFieldPathSegment(segment);
      return normalized !== 'data' && normalized !== 'error' && normalized !== 'errors';
    });
    if (segments.length === 0) return undefined;

    const updateIndex = findIndexedFieldPathSegment(segments, ['updates', 'fielddevices']);
    const update =
      updateIndex !== undefined
        ? updates[updateIndex.index]
        : updates.length === 1
          ? updates[0]
          : undefined;
    if (!update) return undefined;

    const relevantSegments =
      updateIndex !== undefined ? segments.slice(updateIndex.segmentIndex + 2) : segments;
    const normalizedFirst = normalizeFieldPathSegment(relevantSegments[0] ?? '');

    const bacnetIndex = findFieldPathSegment(relevantSegments, ['bacnetobjects', 'bacnetobject']);
    if (bacnetIndex !== undefined) {
      const objectRef = relevantSegments[bacnetIndex + 1];
      const rawField = relevantSegments.slice(bacnetIndex + 2).join('.');
      const objectId = resolveBacnetObjectId(update, objectRef);
      if (!objectId || !rawField) return undefined;

      return {
        deviceId: update.id,
        fieldPath: `bacnet_objects.${objectId}.${rawField}`,
        bacnetObjectId: objectId,
        bacnetField: rawField
      };
    }

    if (normalizedFirst === 'fielddevice' || normalizedFirst === 'fielddevices') {
      const field = relevantSegments.slice(1).join('.');
      return field ? { deviceId: update.id, fieldPath: `fielddevice.${field}` } : undefined;
    }

    if (normalizedFirst === 'specification' || normalizedFirst === 'specifications') {
      const field = relevantSegments.slice(1).join('.');
      return field ? { deviceId: update.id, fieldPath: `specification.${field}` } : undefined;
    }

    const field = relevantSegments.join('.');
    return field ? { deviceId: update.id, fieldPath: field } : undefined;
  }

  function resolveBacnetObjectId(
    update: BulkUpdateFieldDeviceItem,
    objectRef: string | undefined
  ): string | undefined {
    if (!objectRef) return undefined;
    const objectIndex = Number(objectRef);
    if (Number.isInteger(objectIndex)) {
      return update.bacnet_objects?.[objectIndex]?.id;
    }
    return objectRef;
  }

  async function saveAllPendingEdits(
    storeItems: FieldDevice[],
    onSuccess?: (updated: FieldDevice[]) => void
  ): Promise<void> {
    if (pendingEdits.size === 0 && pendingBacnetEdits.size === 0) return;

    // Run client-side validation for all BACnet edits first
    let hasClientErrors = false;
    for (const deviceId of pendingBacnetEdits.keys()) {
      if (!validateBacnetEdits(storeItems, deviceId)) {
        hasClientErrors = true;
      }
    }
    if (hasClientErrors) {
      addToast(getFirstBacnetClientValidationToast(), 'error');
      return;
    }

    // Collect all device IDs that need updates
    const allDeviceIds = new Set([...pendingEdits.keys(), ...pendingBacnetEdits.keys()]);
    const updates: BulkUpdateFieldDeviceItem[] = [];
    const nextErrors = new Map(editErrors);

    for (const id of allDeviceIds) {
      const clientError = validatePendingEdits(id);
      if (clientError) {
        nextErrors.set(id, clientError);
        continue;
      }
      nextErrors.delete(id);

      const update = buildUpdateForDevice(id, storeItems, { includeBacnet: true });
      if (update) {
        updates.push(update);
      }
    }

    if (updates.length === 0) {
      editErrors = nextErrors;
      if (nextErrors.size === 0) {
        pendingEdits = new Map();
        return;
      }
      addToast(getFirstEditValidationToast(nextErrors), 'error');
      return;
    }

    const pendingSnapshot = new Map(pendingEdits);
    const pendingBacnetSnapshot = new Map(pendingBacnetEdits);

    try {
      const result = await fieldDeviceRepository.bulkUpdate({ updates });

      const reconciled = reconcileFieldDeviceSaveResult({
        storeItems,
        updates,
        result,
        pendingEdits,
        pendingBacnetEdits,
        pendingEditsSnapshot: pendingSnapshot,
        pendingBacnetEditsSnapshot: pendingBacnetSnapshot,
        existingErrors: nextErrors,
        localizeEditErrorInfo,
        localizeFieldErrorMap
      });

      pendingEdits = reconciled.remainingEdits;
      pendingBacnetEdits = reconciled.remainingBacnetEdits;
      editErrors = reconciled.editErrors;
      bacnetFieldErrors = reconciled.bacnetFieldErrors;

      const totalSuccessful = reconciled.successIds.size + reconciled.partialSuccessIds.size;
      if (totalSuccessful > 0) {
        options.onSaveSuccess?.([
          ...new Set([...reconciled.successIds, ...reconciled.partialSuccessIds])
        ]);
        if (reconciled.partialSuccessIds.size > 0) {
          addToast(
            translate('field_device.editing.toasts.partial_success', {
              complete: reconciled.successIds.size,
              partial: reconciled.partialSuccessIds.size
            }),
            'warning'
          );
        } else {
          addToast(
            translate('field_device.editing.toasts.success', {
              count: result.success_count
            }),
            'success'
          );
        }
        onSuccess?.(reconciled.optimisticUpdates);
      }
      if (result.failure_count > 0 && reconciled.partialSuccessIds.size === 0) {
        addToast(
          translate('field_device.editing.toasts.partial_failure', {
            count: result.failure_count
          }),
          'error'
        );
      }
    } catch (error: unknown) {
      const err = error as Error;
      const fieldErrors = getFieldErrors(error);
      if (applyBulkApiFieldErrors(updates, fieldErrors)) {
        addToast(getFirstFieldValidationToast(fieldErrors), 'error');
        return;
      }

      addToast(
        translate('field_device.editing.toasts.bulk_update_failed', {
          message: localizeErrorText(err.message)
        }),
        'error'
      );
    }
  }

  async function saveDeviceEdits(
    device: FieldDevice,
    onSuccess?: (updated: FieldDevice) => void
  ): Promise<void> {
    const update = buildUpdateForDevice(device.id, [device], { includeBacnet: false });
    if (!update) {
      const remaining = new Map(pendingEdits);
      remaining.delete(device.id);
      pendingEdits = remaining;
      setEditError(device.id);
      return;
    }

    const clientError = validatePendingEdits(device.id);
    if (clientError) {
      setEditError(device.id, clientError);
      addToast(getFirstFieldValidationToast(clientError.fields), 'error');
      return;
    }
    setEditError(device.id);

    const pendingSnapshot = pendingEdits.get(device.id);

    const optimistic = applyEditsToDevice(device, { includeBacnet: false });

    try {
      const result = await fieldDeviceRepository.bulkUpdate({ updates: [update] });
      const item = result.results.find((r) => r.id === device.id);
      if (item?.success) {
        if (pendingEdits.get(device.id) === pendingSnapshot) {
          const remaining = new Map(pendingEdits);
          remaining.delete(device.id);
          pendingEdits = remaining;
        }
        setEditError(device.id);
        options.onSaveSuccess?.([device.id]);
        onSuccess?.(optimistic);
        return;
      }

      setEditError(
        device.id,
        localizeEditErrorInfo({
          message: item?.error,
          fields: item?.fields,
          suggestions: item?.suggestions,
          suggestionOptions: item?.suggestion_options
        })
      );
      addToast(
        getFirstFieldValidationToast(
          item?.fields ? localizeFieldErrorMap(item.fields) : undefined,
          localizeErrorText(
            item?.error || translate('field_device.editing.toasts.update_failed_check_fields')
          )
        ),
        'error'
      );
    } catch (error: unknown) {
      const err = error as Error;
      const fieldErrors = getFieldErrors(error);
      if (applyBulkApiFieldErrors([update], fieldErrors)) {
        addToast(getFirstFieldValidationToast(fieldErrors), 'error');
        return;
      }

      addToast(
        translate('field_device.editing.toasts.update_failed', {
          message: localizeErrorText(err.message)
        }),
        'error'
      );
    }
  }

  async function saveDeviceBacnetEdits(
    device: FieldDevice,
    onSuccess?: (updated: FieldDevice) => void
  ): Promise<void> {
    const update = buildUpdateForDevice(device.id, [device], { includeBacnet: true });
    if (!update) {
      const remaining = new Map(pendingEdits);
      remaining.delete(device.id);
      pendingEdits = remaining;
      return;
    }

    if (!validateBacnetEdits([device], device.id)) {
      addToast(getFirstBacnetClientValidationToast(device.id), 'error');
      return;
    }

    const clientError = validatePendingEdits(device.id);
    if (clientError) {
      setEditError(device.id, clientError);
      addToast(getFirstFieldValidationToast(clientError.fields), 'error');
      return;
    }
    setEditError(device.id);

    const pendingEditsSnapshot = pendingEdits.get(device.id);
    const pendingBacnetSnapshot = pendingBacnetEdits.get(device.id);
    const optimistic = applyEditsToDevice(device, { includeBacnet: true });

    try {
      const result = await fieldDeviceRepository.bulkUpdate({ updates: [update] });
      const item = result.results.find((r) => r.id === device.id);
      if (item?.success) {
        if (pendingEdits.get(device.id) === pendingEditsSnapshot) {
          const remaining = new Map(pendingEdits);
          remaining.delete(device.id);
          pendingEdits = remaining;
        }
        if (pendingBacnetEdits.get(device.id) === pendingBacnetSnapshot) {
          const remainingBacnet = new Map(pendingBacnetEdits);
          remainingBacnet.delete(device.id);
          pendingBacnetEdits = remainingBacnet;
        }

        const nextBacnetErrors = new Map(bacnetFieldErrors);
        nextBacnetErrors.delete(device.id);
        bacnetFieldErrors = nextBacnetErrors;
        options.onSaveSuccess?.([device.id]);
        onSuccess?.(optimistic);
        return;
      }

      const errorFields = item?.fields ? localizeFieldErrorMap(item.fields) : {};
      const nextErrors = new Map(editErrors);
      nextErrors.set(device.id, {
        message: item?.error ? localizeErrorText(item.error) : item?.error,
        fields: errorFields,
        suggestions: item?.suggestions,
        suggestionOptions: item?.suggestion_options
      });
      editErrors = nextErrors;

      const objErrors = new Map<string, Record<string, string>>();
      for (const [fieldPath, msg] of Object.entries(errorFields)) {
        const target = resolveBulkUpdateFieldError([update], fieldPath);
        if (!target?.bacnetObjectId || !target.bacnetField) continue;
        const existing = objErrors.get(target.bacnetObjectId) || {};
        existing[target.bacnetField] = msg;
        objErrors.set(target.bacnetObjectId, existing);
      }
      if (objErrors.size > 0) {
        const nextBacnetErrors = new Map(bacnetFieldErrors);
        nextBacnetErrors.set(device.id, objErrors);
        bacnetFieldErrors = nextBacnetErrors;
      }

      addToast(
        getFirstFieldValidationToast(
          errorFields,
          localizeErrorText(
            item?.error || translate('field_device.editing.toasts.update_failed_check_fields')
          )
        ),
        'error'
      );
    } catch (error: unknown) {
      const err = error as Error;
      const fieldErrors = getFieldErrors(error);
      if (applyBulkApiFieldErrors([update], fieldErrors)) {
        addToast(getFirstFieldValidationToast(fieldErrors), 'error');
        return;
      }

      addToast(
        translate('field_device.editing.toasts.update_failed', {
          message: localizeErrorText(err.message)
        }),
        'error'
      );
    }
  }

  function discardAllEdits() {
    pendingEdits = new Map();
    pendingBacnetEdits = new Map();
    editErrors = new Map();
    bacnetFieldErrors = new Map();
    bacnetClientErrors = new Map();
    // Clear persisted state from sessionStorage
    removePersistedFieldDeviceEditingState(sessionStorage, storageKey);
  }

  function getBacnetPendingEdits(deviceId: string): Map<string, Partial<BacnetObjectInput>> {
    const edits = pendingBacnetEdits.get(deviceId);
    if (edits) return edits;
    return new Map();
  }

  function getBacnetFieldErrors(deviceId: string): Map<string, Record<string, string>> {
    const errors = bacnetFieldErrors.get(deviceId);
    if (errors) return errors;
    return new Map();
  }

  function getBacnetClientErrors(deviceId: string): Map<string, Record<string, string>> {
    const errors = bacnetClientErrors.get(deviceId);
    if (errors) return errors;
    return new Map();
  }

  function hasPendingBaseEdits(): boolean {
    return detectPendingBaseEdits(pendingEdits);
  }

  function hasPendingSpecificationEdits(): boolean {
    return detectPendingSpecificationEdits(pendingEdits);
  }

  function hasPendingBacnetEdits(): boolean {
    return detectPendingBacnetEdits(pendingBacnetEdits);
  }

  return {
    get hasUnsavedChanges() {
      return pendingEdits.size > 0 || pendingBacnetEdits.size > 0;
    },
    get pendingCount() {
      return pendingEdits.size + pendingBacnetEdits.size;
    },
    get pendingDeviceIds() {
      return getPendingDeviceIds();
    },
    get hasPendingBaseEdits() {
      return hasPendingBaseEdits();
    },
    get hasPendingSpecificationEdits() {
      return hasPendingSpecificationEdits();
    },
    get hasPendingBacnetEdits() {
      return hasPendingBacnetEdits();
    },
    queueEdit,
    queueSpecEdit,
    isFieldDirty,
    isSpecFieldDirty,
    getPendingValue,
    getPendingSpecValue,
    getFieldError,
    getFieldSuggestion,
    queueBacnetEdit,
    clearBacnetFieldError,
    validateBacnetEdits,
    buildBacnetObjectsPayload,
    saveAllPendingEdits,
    discardAllEdits,
    getBacnetPendingEdits,
    getBacnetFieldErrors,
    getBacnetClientErrors,
    saveDeviceEdits,
    saveDeviceBacnetEdits,
    hasPendingFieldDeviceEditsForDevice,
    hasPendingBacnetEditsForDevice,
    hasPendingEditsForDevice,
    discardFieldEdit,
    discardSpecEdit,
    discardDeviceFieldEdits,
    discardDeviceEdits,
    discardBacnetObjectFieldEdit,
    discardBacnetObjectEdits,
    discardDeviceBacnetEdits
  };
}
