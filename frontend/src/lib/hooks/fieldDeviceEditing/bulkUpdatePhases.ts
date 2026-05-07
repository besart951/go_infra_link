import type { BulkUpdateFieldDeviceItem } from '$lib/domain/facility/index.js';
import {
  findFieldPathSegment,
  getRelevantFieldPathSegments,
  normalizeFieldPathSegment
} from '$lib/api/fieldPath.js';

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
    const phase = getFailedPhase(fieldPath);
    if (phase) {
      phases.add(phase);
    }
  }
  return phases;
}

function getFailedPhase(fieldPath: string): BulkUpdatePhase | undefined {
  const relevant = getRelevantFieldPathSegments(fieldPath);
  if (relevant.length === 0) return undefined;

  const bacnetIndex = findFieldPathSegment(relevant, ['bacnetobjects', 'bacnetobject']);
  if (bacnetIndex !== undefined) {
    return 'bacnet_objects';
  }

  const first = normalizeFieldPathSegment(relevant[0]);
  if (first === 'fielddevice' || first === 'fielddevices') {
    const second = normalizeFieldPathSegment(relevant[1] ?? '');
    return second === 'specification' || second === 'specifications'
      ? 'specification'
      : 'fielddevice';
  }

  if (first === 'specification' || first === 'specifications') {
    return 'specification';
  }

  if (isBaseField(first)) {
    return 'fielddevice';
  }

  return undefined;
}

function isBaseField(field: string): boolean {
  return [
    'bmk',
    'description',
    'textfix',
    'apparatnr',
    'apparatid',
    'systempartid',
    'spscontrollersystemtypeid'
  ].includes(field);
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
