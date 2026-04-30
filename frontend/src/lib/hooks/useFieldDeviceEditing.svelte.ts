/**
 * useFieldDeviceEditing - Composable for field device inline editing state
 *
 * Extracts all editing state + mutation functions from FieldDeviceListView.
 * Pattern follows useFormState.svelte.ts (getter-based for Svelte 5 reactivity).
 */

import { localizeErrorText, localizeFieldErrorMap } from '$lib/api/client.js';
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
      if (errorInfo.fields[field]) return errorInfo.fields[field];
      if (errorInfo.fields[`fielddevice.${field}`]) return errorInfo.fields[`fielddevice.${field}`];
      if (errorInfo.fields[`specification.${field}`]) {
        return errorInfo.fields[`specification.${field}`];
      }
      if (errorInfo.fields[`data.fielddevice.${field}`]) {
        return errorInfo.fields[`data.fielddevice.${field}`];
      }
      if (errorInfo.fields[`error.fielddevice.${field}`]) {
        return errorInfo.fields[`error.fielddevice.${field}`];
      }
      if (errorInfo.fields[`data.specification.${field}`]) {
        return errorInfo.fields[`data.specification.${field}`];
      }
      if (errorInfo.fields[`error.specification.${field}`]) {
        return errorInfo.fields[`error.specification.${field}`];
      }
      return undefined;
    }

    return undefined;
  }

  function localizeEditErrorInfo(info?: EditErrorInfo): EditErrorInfo | undefined {
    if (!info) return undefined;
    const localized = {
      message: info.message ? localizeErrorText(info.message) : info.message,
      fields: info.fields ? localizeFieldErrorMap(info.fields) : info.fields
    };
    if (!localized.message && (!localized.fields || Object.keys(localized.fields).length === 0)) {
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
    const deviceServerErrs = bacnetFieldErrors.get(deviceId);
    if (deviceServerErrs) {
      const objErrs = deviceServerErrs.get(objectId);
      if (objErrs && field in objErrs) {
        const { [field]: _, ...rest } = objErrs;
        const newDeviceErrs = new Map(deviceServerErrs);
        if (Object.keys(rest).length > 0) {
          newDeviceErrs.set(objectId, rest);
        } else {
          newDeviceErrs.delete(objectId);
        }
        bacnetFieldErrors = new Map(bacnetFieldErrors).set(deviceId, newDeviceErrs);
      }
    }
    const deviceClientErrs = bacnetClientErrors.get(deviceId);
    if (deviceClientErrs) {
      const objErrs = deviceClientErrs.get(objectId);
      if (objErrs && field in objErrs) {
        const { [field]: _, ...rest } = objErrs;
        const newDeviceErrs = new Map(deviceClientErrs);
        if (Object.keys(rest).length > 0) {
          newDeviceErrs.set(objectId, rest);
        } else {
          newDeviceErrs.delete(objectId);
        }
        bacnetClientErrors = new Map(bacnetClientErrors).set(deviceId, newDeviceErrs);
      }
    }
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
        localizeEditErrorInfo({ message: item?.error, fields: item?.fields })
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
        fields: errorFields
      });
      editErrors = nextErrors;

      const objErrors = new Map<string, Record<string, string>>();
      for (const [fieldPath, msg] of Object.entries(errorFields)) {
        const match = fieldPath.match(/(?:^|\.)bacnet_objects\.([0-9a-f-]+)\.(.+)$/i);
        if (!match) continue;
        const objId = match[1];
        const field = match[2];
        const existing = objErrors.get(objId) || {};
        existing[field] = msg;
        objErrors.set(objId, existing);
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
    saveDeviceBacnetEdits
  };
}
