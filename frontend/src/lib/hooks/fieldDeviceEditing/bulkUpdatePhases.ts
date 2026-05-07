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
    const phase = getFailedPhase(fieldPath);
    if (phase) {
      phases.add(phase);
    }
  }
  return phases;
}

function getFailedPhase(fieldPath: string): BulkUpdatePhase | undefined {
  const segments = fieldPath
    .replace(/\[([^\]]+)\]/g, '.$1')
    .split('.')
    .filter(Boolean)
    .filter((segment) => {
      const normalized = normalizeFieldPathSegment(segment);
      return normalized !== 'data' && normalized !== 'error' && normalized !== 'errors';
    });

  const indexed = findIndexedSegment(segments, ['updates', 'fielddevices']);
  const relevant = indexed ? segments.slice(indexed.segmentIndex + 2) : segments;
  if (relevant.length === 0) return undefined;

  const bacnetIndex = findSegment(relevant, ['bacnetobjects', 'bacnetobject']);
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

function normalizeFieldPathSegment(segment: string): string {
  return segment
    .trim()
    .toLowerCase()
    .replace(/[\s_-]/g, '');
}

function findSegment(segments: string[], names: string[]): number | undefined {
  const nameSet = new Set(names);
  const index = segments.findIndex((segment) => nameSet.has(normalizeFieldPathSegment(segment)));
  return index >= 0 ? index : undefined;
}

function findIndexedSegment(
  segments: string[],
  names: string[]
): { segmentIndex: number; index: number } | undefined {
  const segmentIndex = findSegment(segments, names);
  if (segmentIndex === undefined) return undefined;
  const rawIndex = Number(segments[segmentIndex + 1]);
  if (!Number.isInteger(rawIndex) || rawIndex < 0) return undefined;
  return { segmentIndex, index: rawIndex };
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
