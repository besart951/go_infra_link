import type {
  BacnetObjectInput,
  BulkUpdateFieldDeviceItem,
  FieldDevice
} from '$lib/domain/facility/index.js';
import { buildBacnetObjectsPayload } from './bacnetPayload.js';
import { buildSpecificationPatch } from './specificationEdits.js';

interface BuildFieldDeviceUpdatePayloadInput {
  deviceId: string;
  storeItems: FieldDevice[];
  pendingEdits: Map<string, Partial<BulkUpdateFieldDeviceItem>>;
  pendingBacnetEdits: Map<string, Map<string, Partial<BacnetObjectInput>>>;
  includeBacnet: boolean;
}

export function buildFieldDeviceUpdatePayload({
  deviceId,
  storeItems,
  pendingEdits,
  pendingBacnetEdits,
  includeBacnet
}: BuildFieldDeviceUpdatePayloadInput): BulkUpdateFieldDeviceItem | null {
  const changes = pendingEdits.get(deviceId);
  const bacnetEdits = pendingBacnetEdits.get(deviceId);
  const shouldIncludeBacnet = includeBacnet && bacnetEdits && bacnetEdits.size > 0;
  const update: BulkUpdateFieldDeviceItem = { id: deviceId };
  const device = storeItems.find((item) => item.id === deviceId);
  let hasChanges = false;

  if (changes) {
    if ('bmk' in changes) {
      update.bmk = changes.bmk;
      hasChanges = true;
    }
    if ('description' in changes) {
      update.description = changes.description;
      hasChanges = true;
    }
    if ('text_fix' in changes) {
      update.text_fix = changes.text_fix;
      hasChanges = true;
    }
    if ('apparat_nr' in changes) {
      update.apparat_nr = changes.apparat_nr;
      hasChanges = true;
    }
    if ('apparat_id' in changes) {
      update.apparat_id = changes.apparat_id;
      hasChanges = true;
    }
    if ('system_part_id' in changes) {
      update.system_part_id = changes.system_part_id;
      hasChanges = true;
    }

    const spec = buildSpecificationPatch(changes.specification);
    if (spec) {
      const hasPersistedSpecification = Boolean(device?.specification);
      const hasNonNullSpecValue = Object.values(spec).some((value) => value !== null);
      if (hasPersistedSpecification || hasNonNullSpecValue) {
        update.specification = spec;
        hasChanges = true;
      }
    }
  }

  if (shouldIncludeBacnet && device) {
    update.bacnet_objects = buildBacnetObjectsPayload(device, bacnetEdits);
    hasChanges = true;
  }

  return hasChanges ? update : null;
}
