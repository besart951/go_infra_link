import {
  BACNET_HARDWARE_TYPES,
  BACNET_SOFTWARE_TYPES
} from '$lib/domain/facility/bacnet-object.js';
import type { BacnetObjectInput, FieldDevice } from '$lib/domain/facility/index.js';

type Translate = (key: string) => string;

interface ValidateBacnetObjectEditsInput {
  device?: FieldDevice;
  deviceEdits?: Map<string, Partial<BacnetObjectInput>>;
  translate: Translate;
}

export function validateBacnetObjectEdits({
  device,
  deviceEdits,
  translate
}: ValidateBacnetObjectEditsInput): Map<string, Record<string, string>> {
  if (!device || !device.bacnet_objects || !deviceEdits || deviceEdits.size === 0) {
    return new Map();
  }

  const errors = new Map<string, Record<string, string>>();
  const validSoftwareTypes = new Set<string>(BACNET_SOFTWARE_TYPES.map((t) => t.value));
  const validHardwareTypes = new Set<string>(BACNET_HARDWARE_TYPES.map((t) => t.value));

  const textFixMap = new Map<string, string>();
  for (const obj of device.bacnet_objects) {
    const edits = deviceEdits.get(obj.id);
    const effectiveTextFix =
      edits && 'text_fix' in edits ? (edits.text_fix as string) : obj.text_fix;
    if (effectiveTextFix) {
      const existing = textFixMap.get(effectiveTextFix);
      if (existing && existing !== obj.id) {
        const objErrors = errors.get(obj.id) || {};
        objErrors['text_fix'] = translate('field_device.bacnet.validation.text_fix_unique');
        errors.set(obj.id, objErrors);
      }
      textFixMap.set(effectiveTextFix, obj.id);
    }
  }

  const softwareCombinationMap = new Map<string, string>();
  for (const obj of device.bacnet_objects) {
    const edits = deviceEdits.get(obj.id);
    const effectiveSoftwareTypeRaw =
      edits && 'software_type' in edits
        ? (edits.software_type as string | undefined)
        : obj.software_type;
    const effectiveSoftwareType = effectiveSoftwareTypeRaw?.trim().toLowerCase() ?? '';

    const effectiveSoftwareNumberRaw =
      edits && 'software_number' in edits
        ? (edits.software_number as number | undefined)
        : obj.software_number;
    const effectiveSoftwareNumber = Number(effectiveSoftwareNumberRaw);

    if (!validSoftwareTypes.has(effectiveSoftwareType)) continue;
    if (
      !Number.isFinite(effectiveSoftwareNumber) ||
      effectiveSoftwareNumber < 0 ||
      effectiveSoftwareNumber > 65535
    ) {
      continue;
    }

    const softwareKey = `${effectiveSoftwareType}:${effectiveSoftwareNumber}`;
    const existingObjectId = softwareCombinationMap.get(softwareKey);
    if (existingObjectId && existingObjectId !== obj.id) {
      const existingErrors = errors.get(existingObjectId) || {};
      existingErrors['software_number'] = translate(
        'field_device.bacnet.validation.software_unique'
      );
      errors.set(existingObjectId, existingErrors);

      const objErrors = errors.get(obj.id) || {};
      objErrors['software_number'] = translate('field_device.bacnet.validation.software_unique');
      errors.set(obj.id, objErrors);
      continue;
    }

    softwareCombinationMap.set(softwareKey, obj.id);
  }

  for (const [objectId, edits] of deviceEdits) {
    const objErrors = errors.get(objectId) || {};
    const currentObject = device.bacnet_objects.find((obj) => obj.id === objectId);

    if ('text_fix' in edits && !edits.text_fix) {
      objErrors['text_fix'] = translate('field_device.bacnet.validation.text_fix_required');
    }
    if ('software_number' in edits) {
      const num = edits.software_number as number;
      if (num < 0 || num > 65535) {
        objErrors['software_number'] = translate(
          'field_device.bacnet.validation.software_number_range'
        );
      }
    }
    if ('hardware_quantity' in edits) {
      const num = edits.hardware_quantity as number;
      if (num < 1 || num > 255) {
        objErrors['hardware_quantity'] = translate(
          'field_device.bacnet.validation.hardware_quantity_range'
        );
      }
    }
    if ('software_type' in edits && !validSoftwareTypes.has(edits.software_type as string)) {
      objErrors['software_type'] = translate('field_device.bacnet.validation.software_type');
    }
    if ('hardware_type' in edits && !validHardwareTypes.has(edits.hardware_type as string)) {
      objErrors['hardware_type'] = translate('field_device.bacnet.validation.hardware_type');
    }
    if ('text_individual' in edits) {
      const val = edits.text_individual as string | undefined;
      const hasExistingText = !!currentObject?.text_individual?.trim();
      if (hasExistingText && (!val || val.trim().length === 0)) {
        objErrors['text_individual'] = translate(
          'field_device.bacnet.validation.text_individual_required'
        );
      }
      if (val && val.length > 250) {
        objErrors['text_individual'] = translate(
          'field_device.bacnet.validation.text_individual_max'
        );
      }
    }

    if (Object.keys(objErrors).length > 0) {
      errors.set(objectId, objErrors);
    }
  }

  return errors;
}
