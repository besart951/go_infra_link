const WRAPPER_PATH_SEGMENTS = new Set(['body', 'data', 'error', 'errors', 'payload', 'request']);

const FIELD_PATH_SEGMENT_ALIASES: Record<string, string> = {
  alarmtypes: 'alarmtype',
  bacnetobjects: 'bacnetobject',
  controlcabinets: 'controlcabinet',
  fielddevices: 'fielddevice',
  objectdatas: 'objectdata',
  spscontrollers: 'spscontroller',
  spscontrollersystemtypes: 'spscontrollersystemtype',
  specifications: 'specification',
  updates: 'update'
};

const INDEXED_FIELD_PATH_SEGMENTS = new Set([
  'alarmtype',
  'alarmtypes',
  'bacnetobject',
  'bacnetobjects',
  'fielddevices',
  'spscontrollersystemtypes',
  'updates'
]);

export interface IndexedFieldPathSegment {
  segmentIndex: number;
  index: number;
}

export function splitFieldPath(path: string): string[] {
  return path
    .replace(/\[([^\]]+)\]/g, '.$1')
    .split('.')
    .filter(Boolean);
}

export function normalizeFieldPathSegment(segment: string): string {
  return segment
    .trim()
    .toLowerCase()
    .replace(/[\s_-]/g, '')
    .replace(/^\[(.*)\]$/, '$1');
}

export function canonicalFieldPathSegment(segment: string): string {
  const normalized = normalizeFieldPathSegment(segment);
  return FIELD_PATH_SEGMENT_ALIASES[normalized] ?? normalized;
}

export function isWrapperFieldPathSegment(segment: string): boolean {
  return WRAPPER_PATH_SEGMENTS.has(normalizeFieldPathSegment(segment));
}

export function unwrapFieldPathSegments(path: string): string[] {
  return splitFieldPath(path).filter((segment) => !isWrapperFieldPathSegment(segment));
}

export function findFieldPathSegment(segments: string[], names: string[]): number | undefined {
  const nameSet = new Set(names.map(normalizeFieldPathSegment));
  const index = segments.findIndex((segment) => nameSet.has(normalizeFieldPathSegment(segment)));
  return index >= 0 ? index : undefined;
}

export function findIndexedFieldPathSegment(
  segments: string[],
  names: string[]
): IndexedFieldPathSegment | undefined {
  const segmentIndex = findFieldPathSegment(segments, names);
  if (segmentIndex === undefined) return undefined;

  const rawIndex = Number(segments[segmentIndex + 1]);
  if (!Number.isInteger(rawIndex) || rawIndex < 0) return undefined;
  return { segmentIndex, index: rawIndex };
}

export function getRelevantFieldPathSegments(
  path: string,
  indexedCollectionNames: string[] = ['updates', 'fielddevices']
): string[] {
  const segments = unwrapFieldPathSegments(path);
  const indexed = findIndexedFieldPathSegment(segments, indexedCollectionNames);
  return indexed ? segments.slice(indexed.segmentIndex + 2) : segments;
}

function isIndexOrIdentifierSegment(segment: string): boolean {
  return (
    /^\d+$/.test(segment) ||
    /^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$/i.test(segment)
  );
}

function canonicalFieldPathSegments(path: string): string[] {
  const rawSegments = splitFieldPath(path);
  const segments: string[] = [];

  for (let index = 0; index < rawSegments.length; index += 1) {
    const rawSegment = rawSegments[index];
    const segment = canonicalFieldPathSegment(rawSegment);
    if (!segment || WRAPPER_PATH_SEGMENTS.has(segment)) continue;

    segments.push(segment);

    const rawNormalized = normalizeFieldPathSegment(rawSegment);
    const nextSegment = rawSegments[index + 1];
    const hasFieldAfterNext = index + 2 < rawSegments.length;
    const hasIndexedCollection =
      rawNormalized in FIELD_PATH_SEGMENT_ALIASES || INDEXED_FIELD_PATH_SEGMENTS.has(rawNormalized);

    if (hasIndexedCollection && nextSegment && hasFieldAfterNext) {
      const next = canonicalFieldPathSegment(nextSegment);
      if (next && !WRAPPER_PATH_SEGMENTS.has(next)) {
        index += 1;
        continue;
      }
    }

    const next = nextSegment ? canonicalFieldPathSegment(nextSegment) : '';
    if (next && isIndexOrIdentifierSegment(next) && hasFieldAfterNext) {
      index += 1;
    }
  }

  return segments;
}

function canonicalFieldPath(path: string): string {
  return canonicalFieldPathSegments(path).join('.');
}

export function fieldErrorPathMatches(errorPath: string, candidatePath: string): boolean {
  const error = canonicalFieldPath(errorPath);
  const candidate = canonicalFieldPath(candidatePath);
  if (!error || !candidate) return false;
  if (error === candidate) return true;

  const candidateSegments = candidate.split('.');
  if (candidateSegments.length <= 1) return false;

  return error.endsWith(`.${candidate}`);
}

export function resolveFieldError(
  errors: Record<string, string>,
  field: string,
  prefixes: string[] = []
): string | undefined {
  const candidates = [field, ...prefixes.map((prefix) => `${prefix}.${field}`)];

  for (const candidate of candidates) {
    if (errors[candidate]) return errors[candidate];
  }

  const entries = Object.entries(errors);
  for (const candidate of candidates) {
    for (const [path, value] of entries) {
      if (fieldErrorPathMatches(path, candidate)) return value;
    }
  }

  return undefined;
}
