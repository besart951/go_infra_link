import type { BulkUpdateFieldDeviceItem } from '$lib/domain/facility/index.js';

export type BulkUpdatePhase = 'fielddevice' | 'specification' | 'bacnet_objects';

export function getUpdatePhases(update: BulkUpdateFieldDeviceItem): Set<BulkUpdatePhase> {
  const phases = new Set<BulkUpdatePhase>();
  if (
    'bmk' in update ||
    'description' in update ||
    'text_fix' in update ||
    'apparat_nr' in update ||
    'apparat_id' in update ||
    'system_part_id' in update
  ) {
    phases.add('fielddevice');
  }
  if (update.specification) {
    phases.add('specification');
  }
  if (update.bacnet_objects) {
    phases.add('bacnet_objects');
  }
  return phases;
}

export function getFailedPhases(fields: Record<string, string>): Set<BulkUpdatePhase> {
  const phases = new Set<BulkUpdatePhase>();
  for (const fieldPath of Object.keys(fields)) {
    if (fieldPath === 'fielddevice' || fieldPath.startsWith('fielddevice.')) {
      phases.add('fielddevice');
    } else if (fieldPath === 'specification' || fieldPath.startsWith('specification.')) {
      phases.add('specification');
    } else if (fieldPath === 'bacnet_objects' || fieldPath.startsWith('bacnet_objects.')) {
      phases.add('bacnet_objects');
    }
  }
  return phases;
}

export function hasPartialPhaseSuccess(
  update: BulkUpdateFieldDeviceItem,
  fields: Record<string, string>
): boolean {
  const phases = getUpdatePhases(update);
  const failedPhases = getFailedPhases(fields);
  for (const phase of phases) {
    if (!failedPhases.has(phase)) {
      return true;
    }
  }
  return false;
}
